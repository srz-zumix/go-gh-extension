package client

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubTokenTransport is a test transport that satisfies the tokenGetter interface.
type stubTokenTransport struct {
	token string
}

func (s stubTokenTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s stubTokenTransport) Token() string {
	return s.token
}

// stubWrappingTransport wraps an inner transport and exposes Unwrap(), mimicking
// the getOnlyRoundTripper pattern used in the factory package.
type stubWrappingTransport struct {
	inner http.RoundTripper
}

func (s *stubWrappingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return s.inner.RoundTrip(r)
}

func (s *stubWrappingTransport) Unwrap() http.RoundTripper {
	return s.inner
}

// newTestClient builds a *GitHubClient with its REST BaseURL set to baseURL.
func newTestClient(t *testing.T, baseURL string, transport http.RoundTripper) *GitHubClient {
	t.Helper()
	gc := github.NewClient(&http.Client{Transport: transport})
	parsed, err := url.Parse(baseURL)
	require.NoError(t, err)
	if parsed.Path == "" || parsed.Path[len(parsed.Path)-1] != '/' {
		parsed.Path += "/"
	}
	gc.BaseURL = parsed
	g, err := NewClient(gc)
	require.NoError(t, err)
	return g
}

func TestBearerToken(t *testing.T) {
	t.Run("direct token transport returns token", func(t *testing.T) {
		g := newTestClient(t, "https://api.github.com/", stubTokenTransport{token: "ghs_direct"})
		assert.Equal(t, "ghs_direct", g.bearerToken())
	})

	t.Run("wrapped transport unwraps to token getter", func(t *testing.T) {
		inner := stubTokenTransport{token: "ghs_wrapped"}
		wrapped := &stubWrappingTransport{inner: inner}
		g := newTestClient(t, "https://api.github.com/", wrapped)
		assert.Equal(t, "ghs_wrapped", g.bearerToken())
	})

	t.Run("transport with no token getter returns empty string", func(t *testing.T) {
		g := newTestClient(t, "https://api.github.com/", http.DefaultTransport)
		assert.Equal(t, "", g.bearerToken())
	})

	t.Run("wrapping transport whose inner has no token getter returns empty string", func(t *testing.T) {
		wrapped := &stubWrappingTransport{inner: http.DefaultTransport}
		g := newTestClient(t, "https://api.github.com/", wrapped)
		assert.Equal(t, "", g.bearerToken())
	})
}

func TestGitAuthEnvs(t *testing.T) {
	const token = "ghs_testtoken"
	expectedCreds := base64.StdEncoding.EncodeToString([]byte("x-access-token:" + token))

	tests := []struct {
		name     string
		rawURL   string
		token    string
		expected []string
	}{
		{
			name:   "github.com URL produces correctly scoped config key",
			rawURL: "https://github.com/owner/repo.git",
			token:  token,
			expected: []string{
				"GIT_CONFIG_COUNT=1",
				"GIT_CONFIG_KEY_0=http.https://github.com/.extraHeader",
				"GIT_CONFIG_VALUE_0=Authorization: Basic " + expectedCreds,
			},
		},
		{
			name:   "GHES URL scopes config key to GHES host",
			rawURL: "https://ghe.example.com/owner/repo.git",
			token:  token,
			expected: []string{
				"GIT_CONFIG_COUNT=1",
				"GIT_CONFIG_KEY_0=http.https://ghe.example.com/.extraHeader",
				"GIT_CONFIG_VALUE_0=Authorization: Basic " + expectedCreds,
			},
		},
		{
			name:     "empty token returns nil",
			rawURL:   "https://github.com/owner/repo.git",
			token:    "",
			expected: nil,
		},
		{
			name:     "empty URL returns nil",
			rawURL:   "",
			token:    token,
			expected: nil,
		},
		{
			name:     "unparseable URL returns nil",
			rawURL:   "://bad url\x00",
			token:    token,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestClient(t, "https://api.github.com/", stubTokenTransport{token: tt.token})
			got := g.GitAuthEnvs(tt.rawURL)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestHost(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{
			name:     "api.github.com is normalized to github.com",
			baseURL:  "https://api.github.com/",
			expected: "github.com",
		},
		{
			name:     "GHES hostname is returned as-is",
			baseURL:  "https://ghe.example.com/api/v3/",
			expected: "ghe.example.com",
		},
		{
			name:     "GHES short hostname is returned as-is",
			baseURL:  "https://ghe.internal/api/v3/",
			expected: "ghe.internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestClient(t, tt.baseURL, http.DefaultTransport)
			assert.Equal(t, tt.expected, g.Host())
		})
	}
}

func TestBasicAuthHTTPClient_PreservesClientSettings(t *testing.T) {
	const wantTimeout = 42 * time.Second

	// Build a GitHubClient whose underlying http.Client has a custom Timeout.
	gc := github.NewClient(&http.Client{
		Transport: http.DefaultTransport,
		Timeout:   wantTimeout,
	})
	parsed, err := url.Parse("https://api.github.com/")
	require.NoError(t, err)
	gc.BaseURL = parsed
	g, err := NewClient(gc)
	require.NoError(t, err)

	got := g.basicAuthHTTPClient()

	// Timeout must be carried over from the base client.
	assert.Equal(t, wantTimeout, got.Timeout)
	// Transport must be basicAuthTransport using the raw transport as base.
	bat, ok := got.Transport.(*basicAuthTransport)
	require.True(t, ok, "expected Transport to be *basicAuthTransport")
	assert.Equal(t, http.DefaultTransport, bat.base)
	// No token configured on the underlying transport, so token must be empty.
	assert.Empty(t, bat.token)
}
