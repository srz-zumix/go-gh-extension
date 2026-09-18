package gh

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// archiveLinkFailTransport fails the archive-link API call so the wrapper's
// error-context formatting can be exercised without any real download.
type archiveLinkFailTransport struct{}

func (archiveLinkFailTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Status:     "404 Not Found",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"message":"Not Found"}`)),
		Request:    r,
	}, nil
}

func newArchiveWrapperTestClient(t *testing.T) *GitHubClient {
	t.Helper()
	base := "https://api.github.com/"
	gc, err := github.NewClient(
		github.WithHTTPClient(&http.Client{Transport: archiveLinkFailTransport{}}),
		github.WithURLs(&base, nil),
	)
	require.NoError(t, err)
	g, err := client.NewClient(gc)
	require.NoError(t, err)
	return g
}

// TestDownloadRepositoryArchive_WrapsErrorWithRepoContext verifies the pkg/gh wrapper
// adds repository and ref context to errors returned by the client layer.
func TestDownloadRepositoryArchive_WrapsErrorWithRepoContext(t *testing.T) {
	g := newArchiveWrapperTestClient(t)

	body, err := DownloadRepositoryArchive(t.Context(), g, repository.Repository{Owner: "owner", Name: "repo"}, "main", github.Tarball)
	require.Error(t, err)
	assert.Nil(t, body)
	assert.Contains(t, err.Error(), "failed to download")
	assert.Contains(t, err.Error(), "owner/repo")
	assert.Contains(t, err.Error(), "ref 'main'")
}
