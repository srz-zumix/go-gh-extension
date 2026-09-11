package client

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGitTreeRecursive_SingleCallWhenNotTruncated(t *testing.T) {
	var calls int
	var gotPath, gotQuery string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		gotPath = r.URL.EscapedPath()
		gotQuery = r.URL.RawQuery
		body, _ := json.Marshal(map[string]any{
			"sha": "root-sha",
			"tree": []map[string]any{
				{"path": "dir/file.txt", "type": "blob", "sha": "blob-sha", "size": 42},
			},
			"truncated": false,
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(string(body))),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	tree, err := g.GetGitTreeRecursive(t.Context(), "owner", "repo", "root-sha")
	require.NoError(t, err)
	require.NotNil(t, tree)

	// Exactly one API call, using the recursive query parameter.
	assert.Equal(t, 1, calls)
	assert.Equal(t, "/repos/owner/repo/git/trees/root-sha", gotPath)
	assert.Equal(t, "recursive=1", gotQuery)
	require.Len(t, tree.Entries, 1)
	assert.Equal(t, "dir/file.txt", tree.Entries[0].GetPath())
}

func TestGetGitTreeRecursive_FallsBackToWalkWhenTruncated(t *testing.T) {
	var calls int
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		path := r.URL.EscapedPath()
		var body []byte
		switch path {
		case "/repos/owner/repo/git/trees/root-sha":
			if r.URL.RawQuery == "recursive=1" {
				// Single-call attempt reports truncation.
				body, _ = json.Marshal(map[string]any{
					"sha":       "root-sha",
					"tree":      []map[string]any{},
					"truncated": true,
				})
			} else {
				// Non-recursive walk of the root tree.
				body, _ = json.Marshal(map[string]any{
					"sha": "root-sha",
					"tree": []map[string]any{
						{"path": "file.txt", "type": "blob", "sha": "blob-sha", "size": 1},
						{"path": "dir", "type": "tree", "sha": "dir-sha"},
					},
					"truncated": false,
				})
			}
		case "/repos/owner/repo/git/trees/dir-sha":
			body, _ = json.Marshal(map[string]any{
				"sha": "dir-sha",
				"tree": []map[string]any{
					{"path": "nested.txt", "type": "blob", "sha": "nested-sha", "size": 2},
				},
				"truncated": false,
			})
		default:
			t.Fatalf("unexpected request path %q", path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(string(body))),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	tree, err := g.GetGitTreeRecursive(t.Context(), "owner", "repo", "root-sha")
	require.NoError(t, err)
	require.NotNil(t, tree)

	// One call for the truncated recursive attempt, then two more walking
	// the root and its single subtree.
	assert.Equal(t, 3, calls)
	paths := make([]string, 0, len(tree.Entries))
	for _, e := range tree.Entries {
		paths = append(paths, e.GetPath())
	}
	assert.ElementsMatch(t, []string{"file.txt", "dir", "dir/nested.txt"}, paths)
}

func TestGetGitTreeRecursive_ErrorsWhenWalkedDirectoryTruncated(t *testing.T) {
	var gotWalkRecursive bool
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		path := r.URL.EscapedPath()
		var body []byte
		switch path {
		case "/repos/owner/repo/git/trees/root-sha":
			if r.URL.RawQuery == "recursive=1" {
				// Single-call attempt reports truncation, forcing the walk.
				body, _ = json.Marshal(map[string]any{
					"sha":       "root-sha",
					"tree":      []map[string]any{},
					"truncated": true,
				})
			} else {
				gotWalkRecursive = gotWalkRecursive || strings.Contains(r.URL.RawQuery, "recursive")
				body, _ = json.Marshal(map[string]any{
					"sha": "root-sha",
					"tree": []map[string]any{
						{"path": "dir", "type": "tree", "sha": "dir-sha"},
					},
					"truncated": false,
				})
			}
		case "/repos/owner/repo/git/trees/dir-sha":
			// An oversized directory whose own listing is truncated.
			body, _ = json.Marshal(map[string]any{
				"sha":       "dir-sha",
				"tree":      []map[string]any{},
				"truncated": true,
			})
		default:
			t.Fatalf("unexpected request path %q", path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(string(body))),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	tree, err := g.GetGitTreeRecursive(t.Context(), "owner", "repo", "root-sha")
	require.Error(t, err)
	assert.Nil(t, tree)
	assert.Contains(t, err.Error(), "dir-sha")
	assert.False(t, gotWalkRecursive, "walk requests must not use the recursive parameter")
}
