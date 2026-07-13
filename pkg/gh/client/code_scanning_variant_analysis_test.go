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

func TestCreateCodeQLVariantAnalysis(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		if r.Body != nil {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		}
		return &http.Response{
			StatusCode: http.StatusCreated,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"id":99,"status":"in_progress","query_language":"go"}`)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	opts := &CreateCodeQLVariantAnalysisOptions{
		Language:     "go",
		QueryPack:    "base64pack",
		Repositories: []string{"octo/repo1", "octo/repo2"},
	}
	result, err := g.CreateCodeQLVariantAnalysis(t.Context(), "owner", "repo", opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repos/owner/repo/code-scanning/codeql/variant-analyses", gotPath)
	assert.Equal(t, "go", gotBody["language"])
	assert.Equal(t, "base64pack", gotBody["query_pack"])
	assert.Equal(t, []any{"octo/repo1", "octo/repo2"}, gotBody["repositories"])
	assert.Equal(t, int64(99), result.ID)
	assert.Equal(t, "in_progress", result.Status)
}

func TestGetCodeQLVariantAnalysis(t *testing.T) {
	var gotMethod, gotPath string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"id":99,"status":"succeeded"}`)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	result, err := g.GetCodeQLVariantAnalysis(t.Context(), "owner", "repo", 99)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/repos/owner/repo/code-scanning/codeql/variant-analyses/99", gotPath)
	assert.Equal(t, int64(99), result.ID)
	assert.Equal(t, "succeeded", result.Status)
}

func TestGetCodeQLVariantAnalysisRepoStatus(t *testing.T) {
	var gotMethod, gotPath string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"analysis_status":"succeeded"}`)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	result, err := g.GetCodeQLVariantAnalysisRepoStatus(t.Context(), "owner", "repo", 99, "octo", "target")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/repos/owner/repo/code-scanning/codeql/variant-analyses/99/repos/octo/target", gotPath)
	assert.Equal(t, "succeeded", result.AnalysisStatus)
}
