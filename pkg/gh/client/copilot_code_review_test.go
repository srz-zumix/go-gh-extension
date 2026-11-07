package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Body   map[string]any
}

// copilotTestTransport records requests and answers with the response chosen by respond.
func copilotTestTransport(t *testing.T, reqs *[]recordedRequest, respond func(r *http.Request) (int, http.Header, string)) roundTripFunc {
	return func(r *http.Request) (*http.Response, error) {
		rec := recordedRequest{Method: r.Method, Path: r.URL.EscapedPath(), Query: r.URL.RawQuery}
		if r.Body != nil {
			data, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			if len(data) > 0 {
				require.NoError(t, json.Unmarshal(data, &rec.Body))
			}
		}
		*reqs = append(*reqs, rec)
		status, header, body := respond(r)
		if header == nil {
			header = make(http.Header)
		}
		return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	}
}

func TestListPullRequestReviewCommentsWithCopilotAttribution(t *testing.T) {
	var reqs []recordedRequest
	tr := copilotTestTransport(t, &reqs, func(r *http.Request) (int, http.Header, string) {
		if r.URL.Query().Get("page") == "1" {
			h := make(http.Header)
			h.Set("Link", fmt.Sprintf(`<%s?per_page=100&page=2>; rel="next"`, "https://api.github.com/repos/octo/demo/pulls/7/comments"))
			return http.StatusOK, h, `[{"id":11,"body":"first","is_copilot_code_review":true,"copilot_code_review_sources":[{"kind":"SKILL","name":"code-review","path":".github/skills/code-review/SKILL.md"},{"kind":"MCP","name":"github"}]}]`
		}
		return http.StatusOK, nil, `[{"id":12,"body":"second"}]`
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	comments, err := g.ListPullRequestReviewCommentsWithCopilotAttribution(t.Context(), "octo", "demo", 7)
	require.NoError(t, err)
	require.Len(t, reqs, 2)
	assert.Equal(t, "/repos/octo/demo/pulls/7/comments", reqs[0].Path)
	require.Len(t, comments, 2)
	assert.Equal(t, int64(11), comments[0].GetID())
	assert.Equal(t, "first", comments[0].GetBody())
	require.NotNil(t, comments[0].IsCopilotCodeReview)
	assert.True(t, *comments[0].IsCopilotCodeReview)
	assert.Equal(t, []CopilotCodeReviewSource{
		{Kind: CopilotCodeReviewSourceKindSkill, Name: "code-review", Path: ".github/skills/code-review/SKILL.md"},
		{Kind: CopilotCodeReviewSourceKindMCP, Name: "github"},
	}, comments[0].CopilotCodeReviewSources)
	assert.Nil(t, comments[1].IsCopilotCodeReview)
	assert.Empty(t, comments[1].CopilotCodeReviewSources)
}

func TestListPullRequestReviewsWithCopilotAttribution(t *testing.T) {
	var reqs []recordedRequest
	tr := copilotTestTransport(t, &reqs, func(r *http.Request) (int, http.Header, string) {
		return http.StatusOK, nil, `[{"id":21,"state":"COMMENTED","is_copilot_code_review":true,"copilot_code_review_effort":"balanced"},{"id":22,"state":"APPROVED","is_copilot_code_review":false,"copilot_code_review_effort":null}]`
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	reviews, err := g.ListPullRequestReviewsWithCopilotAttribution(t.Context(), "octo", "demo", 7)
	require.NoError(t, err)
	require.Len(t, reqs, 1)
	assert.Equal(t, "/repos/octo/demo/pulls/7/reviews", reqs[0].Path)
	require.Len(t, reviews, 2)
	assert.Equal(t, "COMMENTED", reviews[0].GetState())
	assert.Equal(t, "balanced", *reviews[0].CopilotCodeReviewEffort)
	assert.True(t, *reviews[0].IsCopilotCodeReview)
	assert.False(t, *reviews[1].IsCopilotCodeReview)
	assert.Nil(t, reviews[1].CopilotCodeReviewEffort)
}

func TestRequestCopilotCodeReview(t *testing.T) {
	var reqs []recordedRequest
	tr := copilotTestTransport(t, &reqs, func(r *http.Request) (int, http.Header, string) {
		return http.StatusCreated, nil, `{"node_id":"RRR_test","state":"PENDING"}`
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	result, err := g.RequestCopilotCodeReview(t.Context(), "octo", "demo", 7, &CopilotCodeReviewRequestBody{ReviewEffort: "max"})
	require.NoError(t, err)
	require.Len(t, reqs, 1)
	assert.Equal(t, http.MethodPost, reqs[0].Method)
	assert.Equal(t, "/repos/octo/demo/pulls/7/copilot-code-review", reqs[0].Path)
	assert.Equal(t, map[string]any{"review_effort": "max"}, reqs[0].Body)
	assert.Equal(t, "PENDING", result.State)

	reqs = nil
	_, err = g.RequestCopilotCodeReview(t.Context(), "octo", "demo", 7, &CopilotCodeReviewRequestBody{})
	require.NoError(t, err)
	assert.Empty(t, reqs[0].Body)
}

func TestRequestCopilotCodeReview_APIError(t *testing.T) {
	var reqs []recordedRequest
	tr := copilotTestTransport(t, &reqs, func(r *http.Request) (int, http.Header, string) {
		return http.StatusUnprocessableEntity, nil, `{"message":"review_effort must be one of: lite, balanced","failure_reason":"INVALID_REVIEW_EFFORT"}`
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	result, err := g.RequestCopilotCodeReview(t.Context(), "octo", "demo", 7, &CopilotCodeReviewRequestBody{ReviewEffort: "max"})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "lite, balanced")
}

func TestRepositoryCopilotCodeReviewSettings(t *testing.T) {
	var reqs []recordedRequest
	settingsJSON := `{"review_effort":{"configured":null,"effective":"balanced"},"available_review_efforts":["lite","balanced"],"severity_filter":"low","custom_instructions_enabled":true,"mcp_enabled":false}`
	tr := copilotTestTransport(t, &reqs, func(r *http.Request) (int, http.Header, string) {
		return http.StatusOK, nil, settingsJSON
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	settings, err := g.GetRepositoryCopilotCodeReviewSettings(t.Context(), "octo", "demo")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, reqs[0].Method)
	assert.Equal(t, "/repos/octo/demo/copilot/code-review/settings", reqs[0].Path)
	assert.Nil(t, settings.ReviewEffort.Configured)
	assert.Equal(t, "balanced", settings.ReviewEffort.Effective)
	assert.Equal(t, []string{"lite", "balanced"}, settings.AvailableReviewEfforts)
	assert.Equal(t, "low", settings.SeverityFilter)
	assert.True(t, settings.CustomInstructionsEnabled)
	assert.False(t, settings.MCPEnabled)

	reqs = nil
	effort := "lite"
	mcp := false
	_, err = g.UpdateRepositoryCopilotCodeReviewSettings(t.Context(), "octo", "demo", &RepositoryCopilotCodeReviewSettingsUpdate{ReviewEffort: &effort, MCPEnabled: &mcp})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, reqs[0].Method)
	assert.Equal(t, map[string]any{"review_effort": "lite", "mcp_enabled": false}, reqs[0].Body)
}

func TestOrganizationCopilotCodeReviewSettings(t *testing.T) {
	var reqs []recordedRequest
	tr := copilotTestTransport(t, &reqs, func(r *http.Request) (int, http.Header, string) {
		return http.StatusOK, nil, `{"default_review_effort":{"configured":"max","effective":"max","github_default":"balanced","editable":true,"managed_by":"organization"},"available_review_efforts":["lite","balanced","max"]}`
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	settings, err := g.GetOrganizationCopilotCodeReviewSettings(t.Context(), "octo-org")
	require.NoError(t, err)
	assert.Equal(t, "/orgs/octo-org/copilot/code-review/settings", reqs[0].Path)
	assert.Equal(t, "max", *settings.DefaultReviewEffort.Configured)
	assert.Equal(t, "organization", settings.DefaultReviewEffort.ManagedBy)

	reqs = nil
	effort := "inherit"
	_, err = g.UpdateOrganizationCopilotCodeReviewSettings(t.Context(), "octo-org", &OrganizationCopilotCodeReviewSettingsUpdate{DefaultReviewEffort: &effort})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, reqs[0].Method)
	assert.Equal(t, map[string]any{"default_review_effort": "inherit"}, reqs[0].Body)
}

func TestOrganizationCopilotCodeReviewRunner(t *testing.T) {
	var reqs []recordedRequest
	tr := copilotTestTransport(t, &reqs, func(r *http.Request) (int, http.Header, string) {
		return http.StatusOK, nil, `{"runner_type":"labeled","runner_labels":["self-hosted"],"runner_group_name":null,"allow_repository_customization":true,"configured_by":"coding_agent"}`
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	runner, err := g.GetOrganizationCopilotCodeReviewRunner(t.Context(), "octo-org")
	require.NoError(t, err)
	assert.Equal(t, "/orgs/octo-org/copilot/code-review/runner", reqs[0].Path)
	assert.Equal(t, "labeled", runner.RunnerType)
	assert.Nil(t, runner.RunnerGroupName)
	assert.Equal(t, "coding_agent", runner.ConfiguredBy)

	reqs = nil
	allow := false
	_, err = g.SetOrganizationCopilotCodeReviewRunner(t.Context(), "octo-org", &OrganizationCopilotCodeReviewRunnerUpdate{RunnerType: "labeled", RunnerLabels: []string{"gpu"}, AllowRepositoryCustomization: &allow})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, reqs[0].Method)
	assert.Equal(t, map[string]any{"runner_type": "labeled", "runner_labels": []any{"gpu"}, "allow_repository_customization": false}, reqs[0].Body)
}

func TestCopilotCodeReviewPolicies(t *testing.T) {
	var reqs []recordedRequest
	tr := copilotTestTransport(t, &reqs, func(r *http.Request) (int, http.Header, string) {
		if r.Method == http.MethodPut {
			return http.StatusNoContent, nil, ""
		}
		if strings.HasPrefix(r.URL.Path, "/enterprises/") {
			return http.StatusOK, nil, `{"policy_state":null}`
		}
		return http.StatusOK, nil, `{"configured_state":null,"effective_state":"enabled","editable":false,"managed_by":"enterprise"}`
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	orgPolicy, err := g.GetOrganizationCopilotCodeReviewPolicy(t.Context(), "octo-org")
	require.NoError(t, err)
	assert.Equal(t, "/orgs/octo-org/copilot/policies/code-review", reqs[0].Path)
	assert.Nil(t, orgPolicy.ConfiguredState)
	assert.Equal(t, "enabled", orgPolicy.EffectiveState)
	assert.Equal(t, "enterprise", orgPolicy.ManagedBy)

	require.NoError(t, g.SetOrganizationCopilotCodeReviewPolicy(t.Context(), "octo-org", "disabled"))
	assert.Equal(t, http.MethodPut, reqs[1].Method)
	assert.Equal(t, map[string]any{"state": "disabled"}, reqs[1].Body)

	entPolicy, err := g.GetEnterpriseCopilotCodeReviewPolicy(t.Context(), "octo-ent")
	require.NoError(t, err)
	assert.Equal(t, "/enterprises/octo-ent/copilot/policies/code-review", reqs[2].Path)
	assert.Nil(t, entPolicy.PolicyState)

	require.NoError(t, g.SetEnterpriseCopilotCodeReviewPolicy(t.Context(), "octo-ent", "configured_by_organization_admins"))
	assert.Equal(t, map[string]any{"policy_state": "configured_by_organization_admins"}, reqs[3].Body)
}
