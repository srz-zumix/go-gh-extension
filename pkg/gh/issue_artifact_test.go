package gh

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListIssueArtifactsPagination(t *testing.T) {
	page1 := fixtureAPI + "repos/example-org/example-repo/issues/5/artifacts?page=2&per_page=100&type=generic"
	g, tr := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "repos/example-org/example-repo/issues/5/artifacts?per_page=100&type=generic": {Status: http.StatusOK, Body: `[{"number":1,"type":"generic","name":"a"}]`, Link: `<` + page1 + `>; rel="next"`},
		"GET " + page1: ok(`[{"number":2,"type":"generic","name":"b"}]`),
	})
	artifacts, err := ListIssueArtifacts(t.Context(), g, repository.Repository{Owner: "example-org", Name: "example-repo"}, 5, IssueArtifactTypeGeneric)
	require.NoError(t, err)
	require.Len(t, artifacts, 2)
	assert.Equal(t, 2, artifacts[1].Number)
	for _, r := range tr.requests {
		assert.Equal(t, "2026-03-10", r.Version)
	}
}

func TestUpsertIssueArtifact(t *testing.T) {
	repo := repository.Repository{Owner: "example-org", Name: "example-repo"}
	list := "GET " + fixtureAPI + "repos/example-org/example-repo/issues/5/artifacts?per_page=100&type=embedded_html"

	t.Run("update", func(t *testing.T) {
		g, tr := newFixtureClient(t, map[string]fixtureResponse{
			list: ok(`[{"number":3,"type":"embedded_html","name":"other"},{"number":4,"type":"embedded_html","name":"report"}]`),
			"PATCH " + fixtureAPI + "repos/example-org/example-repo/issues/5/artifacts/4": ok(`{"number":4,"type":"embedded_html","name":"report","html_url":"https://example.test/a/4"}`),
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
			"POST " + fixtureAPI + "repos/example-org/example-repo/issues/5/artifacts": {Status: http.StatusCreated, Body: `{"number":9,"type":"embedded_html","name":"report"}`},
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
			"POST " + fixtureAPI + "repos/example-org/example-repo/issues/5/artifacts": {Status: http.StatusUnprocessableEntity, Body: `{"message":"Validation Failed"}`},
		})
		_, created, err := UpsertIssueArtifact(t.Context(), g, repo, 5, &IssueArtifactRequest{Type: IssueArtifactTypeEmbeddedHTML, Name: "report", Body: "x"})
		require.Error(t, err)
		assert.False(t, created)
	})
}
