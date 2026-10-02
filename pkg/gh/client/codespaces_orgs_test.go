package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCodespacesOrgAccess(t *testing.T) {
	var gotMethod, gotPath string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"visibility":"selected_members"}`)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	access, err := g.GetCodespacesOrgAccess(t.Context(), "my-org")
	require.NoError(t, err)
	require.NotNil(t, access)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/orgs/my-org/codespaces/access", gotPath)
	assert.Equal(t, "selected_members", access.Visibility)
}

func TestGetCodespacesOrgAccess_APIError(t *testing.T) {
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"message":"Not Found"}`)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	access, err := g.GetCodespacesOrgAccess(t.Context(), "my-org")
	require.Error(t, err)
	assert.Nil(t, access)
	var errResp *github.ErrorResponse
	require.ErrorAs(t, err, &errResp)
	assert.Equal(t, http.StatusNotFound, errResp.Response.StatusCode)
}
