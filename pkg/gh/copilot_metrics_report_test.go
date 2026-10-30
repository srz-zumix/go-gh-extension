package gh

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v93/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedRequest struct {
	Method  string
	URL     string
	Version string
	Body    string
}

type fixtureResponse struct {
	Status int
	Body   string
	Link   string
}

// fixtureTransport serves canned responses keyed by "METHOD URL" and records every request.
type fixtureTransport struct {
	mu        sync.Mutex
	responses map[string]fixtureResponse
	requests  []recordedRequest
}

func (f *fixtureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
	}
	key := req.Method + " " + req.URL.String()
	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{Method: req.Method, URL: req.URL.String(), Version: req.Header.Get("X-GitHub-Api-Version"), Body: string(body)})
	res, ok := f.responses[key]
	f.mu.Unlock()
	if !ok {
		res = fixtureResponse{Status: http.StatusNotFound, Body: `{"message":"Not Found"}`}
	}
	header := http.Header{"Content-Type": []string{"application/json"}}
	if res.Link != "" {
		header.Set("Link", res.Link)
	}
	return &http.Response{StatusCode: res.Status, Header: header, Body: io.NopCloser(bytes.NewBufferString(res.Body)), Request: req}, nil
}

func newFixtureClient(t *testing.T, responses map[string]fixtureResponse) (*GitHubClient, *fixtureTransport) {
	t.Helper()
	tr := &fixtureTransport{responses: responses}
	base := "https://api.example.test/"
	gc, err := github.NewClient(github.WithHTTPClient(&http.Client{Transport: tr}), github.WithURLs(&base, nil))
	require.NoError(t, err)
	g, err := client.NewClient(gc)
	require.NoError(t, err)
	return g, tr
}

func ok(body string) fixtureResponse {
	return fixtureResponse{Status: http.StatusOK, Body: body}
}

const fixtureAPI = "https://api.example.test/"

func TestGetCopilotLatestMetricsMergesLinks(t *testing.T) {
	g, _ := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "orgs/octo/copilot/metrics/reports/organization-28-day/latest": ok(`{"download_links":["https://dl.example.test/a","https://dl.example.test/b"],"report_start_day":"2030-01-01","report_end_day":"2030-01-28"}`),
		"GET https://dl.example.test/a": ok(`{"report_start_day":"2030-01-01","report_end_day":"2030-01-28","day_totals":[{"day":"2030-01-01","daily_active_users":3}]}`),
		"GET https://dl.example.test/b": ok(`{"report_start_day":"2030-01-01","report_end_day":"2030-01-28","day_totals":[{"day":"2030-01-02","daily_active_users":4}]}`),
	})
	m, err := GetCopilotLatestMetrics(t.Context(), g, CopilotMetricsSource{Owner: repository.Repository{Owner: "octo"}})
	require.NoError(t, err)
	assert.Equal(t, CopilotMetricsWindow{StartDay: "2030-01-01", EndDay: "2030-01-28"}, m.Window)
	require.Len(t, m.DayTotals, 2)
	assert.Equal(t, "2030-01-02", m.DayTotals[1].Day)
}

func TestGetCopilotDailyMetricsEnterpriseAndMissing(t *testing.T) {
	g, _ := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "enterprises/acme/copilot/metrics/reports/enterprise-1-day?day=2030-01-01": ok(`{"download_links":["https://dl.example.test/e1"],"report_day":"2030-01-01"}`),
		"GET https://dl.example.test/e1": ok(`{"day":"2030-01-01","daily_active_users":9}`),
		"GET " + fixtureAPI + "enterprises/acme/copilot/metrics/reports/enterprise-1-day?day=2030-01-03": ok(`{"download_links":[],"report_day":"2030-01-03"}`),
	})
	src := CopilotMetricsSource{Owner: repository.Repository{Owner: "acme"}, Enterprise: true}
	assert.Equal(t, "acme (enterprise)", src.String())

	records, found, err := GetCopilotDailyMetrics(t.Context(), g, src, "2030-01-01")
	require.NoError(t, err)
	assert.True(t, found)
	require.Len(t, records, 1)
	assert.Equal(t, 9, *records[0].DailyActiveUsers)

	_, found, err = GetCopilotDailyMetrics(t.Context(), g, src, "2030-01-02")
	require.NoError(t, err, "404 is reported as a missing day")
	assert.False(t, found)

	_, found, err = GetCopilotDailyMetrics(t.Context(), g, src, "2030-01-03")
	require.NoError(t, err)
	assert.False(t, found, "empty download links are reported as a missing day")
}

func TestGetCopilotDailyUserMetricsNDJSON(t *testing.T) {
	g, _ := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "orgs/octo/copilot/metrics/reports/users-1-day?day=2030-01-01": ok(`{"download_links":["https://dl.example.test/u1","https://dl.example.test/u2"],"report_day":"2030-01-01"}`),
		"GET https://dl.example.test/u1": ok("{\"user_id\":1,\"user_login\":\"user-a\",\"day\":\"2030-01-01\"}\n{\"user_id\":2,\"user_login\":\"user-b\",\"day\":\"2030-01-01\"}\n"),
		"GET https://dl.example.test/u2": ok("{\"user_id\":3,\"user_login\":\"user-c\",\"day\":\"2030-01-01\"}\n"),
	})
	records, found, err := GetCopilotDailyUserMetrics(t.Context(), g, CopilotMetricsSource{Owner: repository.Repository{Owner: "octo"}}, "2030-01-01")
	require.NoError(t, err)
	assert.True(t, found)
	require.Len(t, records, 3)
	assert.Equal(t, "user-c", records[2].UserLogin)
}

func TestGetCopilotDailyMetricsServerError(t *testing.T) {
	g, _ := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "orgs/octo/copilot/metrics/reports/repos-1-day?day=2030-01-01": {Status: http.StatusForbidden, Body: `{"message":"Forbidden"}`},
	})
	_, _, err := GetCopilotDailyRepositoryMetrics(t.Context(), g, CopilotMetricsSource{Owner: repository.Repository{Owner: "octo"}}, "2030-01-01")
	require.Error(t, err)
}

func TestListIssueArtifactsPagination(t *testing.T) {
	page1 := fixtureAPI + "repos/octo/reports/issues/5/artifacts?page=2&per_page=100&type=generic"
	g, tr := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "repos/octo/reports/issues/5/artifacts?per_page=100&type=generic": {Status: http.StatusOK, Body: `[{"number":1,"type":"generic","name":"a"}]`, Link: `<` + page1 + `>; rel="next"`},
		"GET " + page1: ok(`[{"number":2,"type":"generic","name":"b"}]`),
	})
	artifacts, err := ListIssueArtifacts(t.Context(), g, repository.Repository{Owner: "octo", Name: "reports"}, 5, IssueArtifactTypeGeneric)
	require.NoError(t, err)
	require.Len(t, artifacts, 2)
	assert.Equal(t, 2, artifacts[1].Number)
	for _, r := range tr.requests {
		assert.Equal(t, "2026-03-10", r.Version)
	}
}

func TestUpsertIssueArtifact(t *testing.T) {
	repo := repository.Repository{Owner: "octo", Name: "reports"}
	list := "GET " + fixtureAPI + "repos/octo/reports/issues/5/artifacts?per_page=100&type=embedded_html"

	t.Run("update", func(t *testing.T) {
		g, tr := newFixtureClient(t, map[string]fixtureResponse{
			list: ok(`[{"number":3,"type":"embedded_html","name":"other"},{"number":4,"type":"embedded_html","name":"report"}]`),
			"PATCH " + fixtureAPI + "repos/octo/reports/issues/5/artifacts/4": ok(`{"number":4,"type":"embedded_html","name":"report","html_url":"https://example.test/a/4"}`),
		})
		a, created, err := UpsertIssueArtifact(t.Context(), g, repo, 5, &IssueArtifactRequest{Type: IssueArtifactTypeEmbeddedHTML, Name: "report", Body: "<p>x</p>"})
		require.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, 4, a.Number)
		require.Len(t, tr.requests, 2)
		patch := tr.requests[1]
		assert.Equal(t, "2026-03-10", patch.Version)
		var body map[string]any
		require.NoError(t, json.Unmarshal([]byte(patch.Body), &body))
		assert.Equal(t, map[string]any{"name": "report", "body": "<p>x</p>"}, body, "PATCH never sends type")
	})

	t.Run("create", func(t *testing.T) {
		g, tr := newFixtureClient(t, map[string]fixtureResponse{
			list: ok(`[]`),
			"POST " + fixtureAPI + "repos/octo/reports/issues/5/artifacts": {Status: http.StatusCreated, Body: `{"number":9,"type":"embedded_html","name":"report"}`},
		})
		a, created, err := UpsertIssueArtifact(t.Context(), g, repo, 5, &IssueArtifactRequest{Type: IssueArtifactTypeEmbeddedHTML, Name: "report", Body: "<p>x</p>"})
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, 9, a.Number)
		var body map[string]any
		require.NoError(t, json.Unmarshal([]byte(tr.requests[1].Body), &body))
		assert.Equal(t, "embedded_html", body["type"])
	})

	t.Run("error", func(t *testing.T) {
		g, _ := newFixtureClient(t, map[string]fixtureResponse{
			list: ok(`[]`),
			"POST " + fixtureAPI + "repos/octo/reports/issues/5/artifacts": {Status: http.StatusUnprocessableEntity, Body: `{"message":"Validation Failed"}`},
		})
		_, created, err := UpsertIssueArtifact(t.Context(), g, repo, 5, &IssueArtifactRequest{Type: IssueArtifactTypeEmbeddedHTML, Name: "report", Body: "x"})
		require.Error(t, err)
		assert.False(t, created)
	})
}
