package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// authInjectingTransport models the authenticated GitHub transport: it injects an
// Authorization header on every request and forwards to a raw transport. It exposes
// RawTransport() so the cross-host stripping logic can reach the header-free transport.
type authInjectingTransport struct {
	raw   http.RoundTripper
	token string
}

func (t *authInjectingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	if t.token != "" {
		r2.Header.Set("Authorization", "token "+t.token)
	}
	return t.raw.RoundTrip(r2)
}

func (t *authInjectingTransport) RawTransport() http.RoundTripper {
	return t.raw
}

// newArchiveTestClient builds a *GitHubClient whose REST base URL is baseURL and whose
// transport injects an Authorization header, so that cross-host header stripping can be
// exercised. baseURL should end with "/".
func newArchiveTestClient(t *testing.T, baseURL string) *GitHubClient {
	t.Helper()
	return newTestClient(t, baseURL, &authInjectingTransport{raw: http.DefaultTransport, token: "secret"})
}

func TestDownloadRepositoryArchive_PreservesAuthOnSameHostWithPort(t *testing.T) {
	var downloadAuth string
	var downloadHit bool

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/tarball/"):
			// Archive-link endpoint: redirect to a download URL on the same host:port.
			http.Redirect(w, r, srv.URL+"/download/archive.tar.gz", http.StatusFound)
		case r.URL.Path == "/download/archive.tar.gz":
			downloadHit = true
			downloadAuth = r.Header.Get("Authorization")
			_, _ = io.WriteString(w, "archive-bytes")
		default:
			http.Error(w, "unexpected path", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	g := newArchiveTestClient(t, srv.URL+"/")
	// The httptest server listens on 127.0.0.1:<port>; Host() therefore includes a
	// non-default port. The download URL is on the same host, so authentication must
	// be preserved despite the port.
	require.Contains(t, g.Host(), ":")

	body, err := g.DownloadRepositoryArchive(t.Context(), "owner", "repo", github.Tarball, "main")
	require.NoError(t, err)
	defer func() { _ = body.Close() }()

	got, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.Equal(t, "archive-bytes", string(got))
	assert.True(t, downloadHit, "download endpoint was not hit")
	assert.Equal(t, "token secret", downloadAuth, "auth header must be preserved on same-host (with port) download")
}

func TestDownloadRepositoryArchive_StripsAuthOnCrossHost(t *testing.T) {
	var downloadAuth string
	var downloadHit bool

	download := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downloadHit = true
		downloadAuth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, "archive-bytes")
	}))
	defer download.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/tarball/") {
			http.Redirect(w, r, crossHostURL(t, download.URL)+"/archive.tar.gz", http.StatusFound)
			return
		}
		http.Error(w, "unexpected path", http.StatusNotFound)
	}))
	defer api.Close()

	g := newArchiveTestClient(t, api.URL+"/")

	body, err := g.DownloadRepositoryArchive(t.Context(), "owner", "repo", github.Tarball, "main")
	require.NoError(t, err)
	defer func() { _ = body.Close() }()

	_, err = io.ReadAll(body)
	require.NoError(t, err)
	assert.True(t, downloadHit, "download endpoint was not hit")
	assert.Empty(t, downloadAuth, "auth header must be stripped on cross-host download")
}

// crossHostURL rewrites a 127.0.0.1 httptest URL to use the "localhost" hostname
// (which resolves to the same address) so it is classified as a different host than
// the API server while remaining reachable.
func crossHostURL(t *testing.T, rawURL string) string {
	t.Helper()
	u, err := url.Parse(rawURL)
	require.NoError(t, err)
	u.Host = "localhost:" + u.Port()
	return u.String()
}

// instrumentedBody is an io.ReadCloser that records how many bytes were read and
// whether Close was called, so tests can assert client-side cleanup behavior that
// an httptest.Server cannot observe.
type instrumentedBody struct {
	r      *strings.Reader
	read   int
	closed bool
}

func (b *instrumentedBody) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	b.read += n
	return n, err
}

func (b *instrumentedBody) Close() error {
	b.closed = true
	return nil
}

// archiveMockTransport returns a 302 to downloadURL for the archive-link API call and
// serves the instrumented non-2xx response for the subsequent download request.
type archiveMockTransport struct {
	downloadURL string
	body        *instrumentedBody
}

func (t *archiveMockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if strings.Contains(r.URL.Path, "/tarball/") {
		h := make(http.Header)
		h.Set("Location", t.downloadURL)
		return &http.Response{StatusCode: http.StatusFound, Header: h, Body: http.NoBody, Request: r}, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Status:     "404 Not Found",
		Header:     make(http.Header),
		Body:       t.body,
		Request:    r,
	}, nil
}

func TestDownloadRepositoryArchive_NonSuccessDrainsAndClosesBody(t *testing.T) {
	body := &instrumentedBody{r: strings.NewReader("error details")}
	tr := &archiveMockTransport{
		downloadURL: "https://codeload.github.com/owner/repo/tar.gz/main",
		body:        body,
	}
	g := newTestClient(t, "https://api.github.com/", tr)

	rc, err := g.DownloadRepositoryArchive(t.Context(), "owner", "repo", github.Tarball, "main")
	require.Error(t, err)
	assert.Nil(t, rc)
	assert.Contains(t, err.Error(), "unexpected http status")
	assert.True(t, body.closed, "error response body must be closed")
	assert.Equal(t, len("error details"), body.read, "error response body must be drained to EOF")
}
