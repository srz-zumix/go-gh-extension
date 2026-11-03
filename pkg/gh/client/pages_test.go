package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPagesSiteRequests(t *testing.T) {
	var requests []struct {
		method string
		path   string
		body   map[string]any
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := struct {
			method string
			path   string
			body   map[string]any
		}{method: r.Method, path: r.URL.Path}
		if r.Method != http.MethodGet {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request.body))
		}
		requests = append(requests, request)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		_, err := w.Write([]byte(`{"public":false,"build_type":"workflow"}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	g := newTestClient(t, server.URL, server.Client().Transport)
	site, err := g.GetPagesSite(t.Context(), "example-org", "example-repo")
	require.NoError(t, err)
	require.NotNil(t, site.Public)
	assert.False(t, *site.Public)
	_, err = g.CreatePagesSite(t.Context(), "example-org", "example-repo", &github.Pages{BuildType: github.Ptr("workflow")})
	require.NoError(t, err)
	err = g.UpdatePagesSite(t.Context(), "example-org", "example-repo", &github.PagesUpdateWithoutCNAME{Public: github.Ptr(false)})
	require.NoError(t, err)
	err = g.UpdatePagesSite(t.Context(), "example-org", "example-repo", &github.PagesUpdateWithoutCNAME{
		BuildType: github.Ptr("legacy"),
		Source:    &github.PagesSource{Branch: github.Ptr("metrics-pages"), Path: github.Ptr("/")},
	})
	require.NoError(t, err)

	require.Len(t, requests, 4)
	for _, request := range requests {
		assert.Equal(t, "/repos/example-org/example-repo/pages", request.path)
	}
	assert.Equal(t, http.MethodGet, requests[0].method)
	assert.Equal(t, map[string]any{"build_type": "workflow"}, requests[1].body)
	assert.Equal(t, map[string]any{"public": false}, requests[2].body)
	assert.Equal(t, map[string]any{
		"build_type": "legacy",
		"source":     map[string]any{"branch": "metrics-pages", "path": "/"},
	}, requests[3].body)
}

func TestPagesSiteErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, err := w.Write([]byte(`{"message":"permission denied"}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	g := newTestClient(t, server.URL, server.Client().Transport)
	_, err := g.GetPagesSite(t.Context(), "example-org", "example-repo")
	var responseErr *github.ErrorResponse
	require.ErrorAs(t, err, &responseErr)
	assert.Equal(t, http.StatusForbidden, responseErr.Response.StatusCode)
}

func TestCreateRootTreeAndCommit(t *testing.T) {
	var requests []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		requests = append(requests, body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte(`{"sha":"1111111111111111111111111111111111111111"}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	g := newTestClient(t, server.URL, server.Client().Transport)
	tree, err := g.CreateGitTree(t.Context(), "example-org", "example-repo", "", []*github.TreeEntry{{
		Path: github.Ptr(".nojekyll"), Mode: github.Ptr("100644"), Type: github.Ptr("blob"), Content: github.Ptr(""),
	}})
	require.NoError(t, err)
	_, err = g.CreateGitCommit(t.Context(), "example-org", "example-repo", github.Commit{
		Message: github.Ptr("Initialize metrics site"), Tree: &github.Tree{SHA: tree.SHA},
	})
	require.NoError(t, err)

	require.Len(t, requests, 2)
	assert.NotContains(t, requests[0], "base_tree")
	assert.Equal(t, []any{map[string]any{"path": ".nojekyll", "mode": "100644", "type": "blob", "content": ""}}, requests[0]["tree"])
	assert.NotContains(t, requests[1], "parents")
	assert.Equal(t, "1111111111111111111111111111111111111111", requests[1]["tree"])
}
