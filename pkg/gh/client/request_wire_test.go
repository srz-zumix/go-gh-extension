package client

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureRequest records the method, path, and decoded JSON body of the last request it sees.
type captureRequest struct {
	method string
	path   string
	body   map[string]any
}

func newCaptureTransport(t *testing.T, cap *captureRequest) http.RoundTripper {
	t.Helper()
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		cap.method = r.Method
		cap.path = r.URL.EscapedPath()
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		cap.body = nil
		if len(raw) > 0 {
			require.NoError(t, json.Unmarshal(raw, &cap.body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{}`)),
			Request:    r,
		}, nil
	})
}

func TestUpdateTeam_RemoveParentTeamSendsNull(t *testing.T) {
	var cap captureRequest
	g := newTestClient(t, "https://api.github.com/", newCaptureTransport(t, &cap))

	_, err := g.UpdateTeam(t.Context(), "example-org", "example-team", &github.UpdateTeamRequest{
		Name:             github.Ptr("example-team"),
		RemoveParentTeam: true,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, cap.method)
	assert.Equal(t, "/orgs/example-org/teams/example-team", cap.path)
	assert.Equal(t, "example-team", cap.body["name"])
	assert.Contains(t, cap.body, "parent_team_id")
	assert.Nil(t, cap.body["parent_team_id"])
	assert.Contains(t, cap.body, "parent_team_slug")
	assert.Nil(t, cap.body["parent_team_slug"])
}

func TestUpdateTeam_WithoutRemoveParentTeamOmitsParent(t *testing.T) {
	var cap captureRequest
	g := newTestClient(t, "https://api.github.com/", newCaptureTransport(t, &cap))

	_, err := g.UpdateTeam(t.Context(), "example-org", "example-team", &github.UpdateTeamRequest{
		Name: github.Ptr("example-team"),
	})
	require.NoError(t, err)
	assert.NotContains(t, cap.body, "parent_team_id")
	assert.NotContains(t, cap.body, "parent_team_slug")
}

func TestAddTeamRepo_EmptyPermissionIsOmitted(t *testing.T) {
	var cap captureRequest
	g := newTestClient(t, "https://api.github.com/", newCaptureTransport(t, &cap))

	require.NoError(t, g.AddTeamRepo(t.Context(), "example-org", "example-team", "example-org", "example-repo", ""))
	assert.Equal(t, http.MethodPut, cap.method)
	assert.Equal(t, "/orgs/example-org/teams/example-team/repos/example-org/example-repo", cap.path)
	assert.NotContains(t, cap.body, "permission")
}

func TestAddOrUpdateTeamMembership_EmptyRoleIsOmitted(t *testing.T) {
	var cap captureRequest
	g := newTestClient(t, "https://api.github.com/", newCaptureTransport(t, &cap))

	_, err := g.AddOrUpdateTeamMembership(t.Context(), "example-org", "example-team", "example-user", "")
	require.NoError(t, err)
	assert.NotContains(t, cap.body, "role")
}

func TestCreatePullRequestComment_RequestPayload(t *testing.T) {
	for _, tt := range []struct {
		name    string
		comment github.CreatePullRequestCommentRequest
		want    map[string]any
	}{
		{
			name: "new comment",
			comment: github.CreatePullRequestCommentRequest{
				Body: "review body", CommitID: strings.Repeat("0", 40), Path: "example.go",
				Line: github.Ptr(5), Side: github.Ptr("RIGHT"),
			},
			want: map[string]any{
				"body": "review body", "commit_id": strings.Repeat("0", 40), "path": "example.go",
				"line": float64(5), "side": "RIGHT",
			},
		},
		{
			name:    "reply without commit or path",
			comment: github.CreatePullRequestCommentRequest{Body: "reply body", InReplyTo: github.Ptr(int64(42))},
			want:    map[string]any{"body": "reply body", "in_reply_to": float64(42)},
		},
		{
			name: "reply ignores creation fields",
			comment: github.CreatePullRequestCommentRequest{
				Body: "reply body", InReplyTo: github.Ptr(int64(42)),
				CommitID: strings.Repeat("0", 40), Path: "example.go",
				Line: github.Ptr(5), Side: github.Ptr("RIGHT"),
			},
			want: map[string]any{"body": "reply body", "in_reply_to": float64(42)},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var cap captureRequest
			g := newTestClient(t, "https://api.github.com/", newCaptureTransport(t, &cap))
			before, err := json.Marshal(tt.comment)
			require.NoError(t, err)

			_, err = g.CreatePullRequestComment(t.Context(), "example-org", "example-repo", 7, &tt.comment)
			require.NoError(t, err)
			assert.Equal(t, http.MethodPost, cap.method)
			assert.Equal(t, "/repos/example-org/example-repo/pulls/7/comments", cap.path)
			assert.Equal(t, tt.want, cap.body)
			after, err := json.Marshal(tt.comment)
			require.NoError(t, err)
			assert.JSONEq(t, string(before), string(after))
		})
	}
}

func TestCreatePullRequestComment_ReplyResponse(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		body   string
	}{
		{"created", http.StatusCreated, `{"id":99,"body":"reply body","html_url":"https://example.test/reply"}`},
		{"API error", http.StatusUnprocessableEntity, `{"message":"invalid reply"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestClient(t, "https://api.github.com/", roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tt.status,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(tt.body)),
					Request:    r,
				}, nil
			}))
			got, err := g.CreatePullRequestComment(t.Context(), "example-org", "example-repo", 7,
				&github.CreatePullRequestCommentRequest{Body: "reply body", InReplyTo: github.Ptr(int64(42))})
			if tt.status == http.StatusUnprocessableEntity {
				var apiErr *github.ErrorResponse
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, tt.status, apiErr.Response.StatusCode)
				assert.Equal(t, "invalid reply", apiErr.Message)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, int64(99), got.GetID())
			assert.Equal(t, "reply body", got.GetBody())
			assert.Equal(t, "https://example.test/reply", got.GetHTMLURL())
		})
	}
}

func TestEditPullRequestComment_SendsBodyOnly(t *testing.T) {
	var cap captureRequest
	g := newTestClient(t, "https://api.github.com/", newCaptureTransport(t, &cap))

	_, err := g.EditPullRequestComment(t.Context(), "example-org", "example-repo", 99, "edited body")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, cap.method)
	assert.Equal(t, "/repos/example-org/example-repo/pulls/comments/99", cap.path)
	assert.Equal(t, map[string]any{"body": "edited body"}, cap.body)
}

func TestEditMilestone_OmitsUnsetFields(t *testing.T) {
	var cap captureRequest
	g := newTestClient(t, "https://api.github.com/", newCaptureTransport(t, &cap))

	_, err := g.EditMilestone(t.Context(), "example-org", "example-repo", 3, &github.UpdateMilestoneRequest{
		Title: github.Ptr("v1.0"),
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, cap.method)
	assert.Equal(t, "/repos/example-org/example-repo/milestones/3", cap.path)
	assert.Equal(t, map[string]any{"title": "v1.0"}, cap.body)
}
