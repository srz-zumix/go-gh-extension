package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListRequestedReviewers(t *testing.T) {
	tests := []struct {
		name     string
		response string
		users    []string
		teams    []string
		wantErr  bool
	}{
		{
			name: "users teams and Copilot bot",
			response: `{"data":{"repository":{"pullRequest":{"reviewRequests":{"nodes":[
				{"requestedReviewer":{"login":"alice"}},
				{"requestedReviewer":{"login":"copilot-pull-request-reviewer"}},
				{"requestedReviewer":{"slug":"reviewers"}},
				{"requestedReviewer":null}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}}}`,
			users: []string{"alice", "copilot-pull-request-reviewer"},
			teams: []string{"reviewers"},
		},
		{
			name: "no pending requests",
			response: `{"data":{"repository":{"pullRequest":{"reviewRequests":{
				"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}}}`,
		},
		{
			name:     "missing pull request",
			response: `{"data":{"repository":{"pullRequest":null}}}`,
			wantErr:  true,
		},
		{
			name:     "API error",
			response: `{"errors":[{"message":"permission denied"}]}`,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Query string `json:"query"`
				}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
				assert.Contains(t, request.Query, "... on Bot")
				assert.Contains(t, request.Query, "... on User")
				assert.Contains(t, request.Query, "... on Team")
				_, err := w.Write([]byte(tt.response))
				assert.NoError(t, err)
			}))
			defer server.Close()
			g := newTestClient(t, server.URL, server.Client().Transport)
			g.graphql = githubv4.NewEnterpriseClient(server.URL, server.Client())
			reviewers, err := g.ListRequestedReviewers(context.Background(), "owner", "repo", 37)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var users, teams []string
			for _, user := range reviewers.Users {
				users = append(users, user.GetLogin())
			}
			for _, team := range reviewers.Teams {
				teams = append(teams, team.GetSlug())
			}
			assert.Equal(t, tt.users, users)
			assert.Equal(t, tt.teams, teams)
		})
	}
}

func TestListRequestedReviewersPagination(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Variables struct {
				Cursor *string `json:"cursor"`
			} `json:"variables"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		calls++
		if calls == 1 {
			assert.Nil(t, request.Variables.Cursor)
			_, err := w.Write([]byte(`{"data":{"repository":{"pullRequest":{"reviewRequests":{
				"nodes":[{"requestedReviewer":{"login":"alice"}}],
				"pageInfo":{"hasNextPage":true,"endCursor":"next"}}}}}}`))
			assert.NoError(t, err)
		} else {
			require.NotNil(t, request.Variables.Cursor)
			assert.Equal(t, "next", *request.Variables.Cursor)
			_, err := w.Write([]byte(`{"data":{"repository":{"pullRequest":{"reviewRequests":{
				"nodes":[{"requestedReviewer":{"login":"copilot-pull-request-reviewer"}}],
				"pageInfo":{"hasNextPage":false,"endCursor":"last"}}}}}}`))
			assert.NoError(t, err)
		}
	}))
	defer server.Close()
	g := newTestClient(t, server.URL, server.Client().Transport)
	g.graphql = githubv4.NewEnterpriseClient(server.URL, server.Client())
	reviewers, err := g.ListRequestedReviewers(context.Background(), "owner", "repo", 37)
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
	require.Len(t, reviewers.Users, 2)
	assert.Equal(t, "alice", reviewers.Users[0].GetLogin())
	assert.Equal(t, "copilot-pull-request-reviewer", reviewers.Users[1].GetLogin())
}
