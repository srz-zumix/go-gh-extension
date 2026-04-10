package client

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMavenRegistryBase(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		owner    string
		repo     string
		expected string
	}{
		{
			name:     "github.com",
			host:     "github.com",
			owner:    "myorg",
			repo:     "myrepo",
			expected: "https://maven.pkg.github.com/myorg/myrepo",
		},
		{
			name:     "empty host treated as github.com",
			host:     "",
			owner:    "myorg",
			repo:     "myrepo",
			expected: "https://maven.pkg.github.com/myorg/myrepo",
		},
		{
			name:     "GHES host",
			host:     "ghe.example.com",
			owner:    "myorg",
			repo:     "myrepo",
			expected: "https://maven.ghe.example.com/myorg/myrepo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, MavenRegistryBase(tt.host, tt.owner, tt.repo))
		})
	}
}

func TestParseMavenPackageName(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantGroupID    string
		wantArtifactID string
		wantErr        bool
	}{
		{
			name:           "colon-separated",
			input:          "com.example:my-artifact",
			wantGroupID:    "com.example",
			wantArtifactID: "my-artifact",
		},
		{
			name:           "dot-separated (GitHub Packages format)",
			input:          "com.example.my-artifact",
			wantGroupID:    "com.example",
			wantArtifactID: "my-artifact",
		},
		{
			name:    "no separator",
			input:   "artifact",
			wantErr: true,
		},
		{
			name:    "empty groupId (colon at start)",
			input:   ":artifact",
			wantErr: true,
		},
		{
			name:    "empty artifactId (colon at end)",
			input:   "com.example:",
			wantErr: true,
		},
		{
			name:    "multiple colons (GAV coordinate rejected)",
			input:   "com.example:my-artifact:1.0",
			wantErr: true,
		},
		{
			name:    "multiple colons (two separators)",
			input:   "a:b:c",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gid, aid, err := ParseMavenPackageName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantGroupID, gid)
			assert.Equal(t, tt.wantArtifactID, aid)
		})
	}
}

func TestMavenArtifactURL(t *testing.T) {
	got := MavenArtifactURL("github.com", "myorg", "myrepo", "com.example", "my-artifact", "1.0.0", "", "jar")
	assert.Equal(t, "https://maven.pkg.github.com/myorg/myrepo/com/example/my-artifact/1.0.0/my-artifact-1.0.0.jar", got)

	gotClassifier := MavenArtifactURL("github.com", "myorg", "myrepo", "com.example", "my-artifact", "1.0.0", "sources", "jar")
	assert.Equal(t, "https://maven.pkg.github.com/myorg/myrepo/com/example/my-artifact/1.0.0/my-artifact-1.0.0-sources.jar", gotClassifier)
}

func TestMavenDownloadError(t *testing.T) {
	err := &MavenDownloadError{StatusCode: http.StatusNotFound, Message: "not found"}
	assert.Equal(t, "download failed with status 404: not found", err.Error())
	assert.True(t, err.IsNotFound())

	err500 := &MavenDownloadError{StatusCode: http.StatusInternalServerError, Message: "server error"}
	assert.False(t, err500.IsNotFound())
}

// --- basicAuthTransport ---

func TestBasicAuthTransport_InjectsBasicAuth(t *testing.T) {
	var capturedAuth string
	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		capturedAuth = r.Header.Get("Authorization")
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	tr := &basicAuthTransport{base: inner, token: "fake-token"}
	req, err := http.NewRequest(http.MethodGet, "https://maven.pkg.github.com/owner/repo/pom.xml", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req)
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(capturedAuth, "Basic "), "expected Basic auth, got: %s", capturedAuth)
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(capturedAuth, "Basic "))
	require.NoError(t, err)
	assert.Equal(t, "x-token:fake-token", string(decoded))
}

func TestBasicAuthTransport_NoTokenNoAuthHeader(t *testing.T) {
	var capturedAuth string
	inner := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		capturedAuth = r.Header.Get("Authorization")
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	tr := &basicAuthTransport{base: inner, token: ""}
	req, err := http.NewRequest(http.MethodGet, "https://maven.pkg.github.com/owner/repo/pom.xml", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Empty(t, capturedAuth)
}

// --- fetchMavenArtifactBody: redirect handling ---

func TestFetchMavenArtifactBody_DirectOK(t *testing.T) {
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("artifact-data")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)
	body, err := g.fetchMavenArtifactBody(t.Context(), "https://maven.pkg.github.com/owner/repo/artifact.jar")
	require.NoError(t, err)
	defer func() { require.NoError(t, body.Close()) }()
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.Equal(t, "artifact-data", string(data))
}

func TestFetchMavenArtifactBody_RedirectDropsAuthHeader(t *testing.T) {
	// Use URL-based routing in the transport so both the registry (302) and CDN (200)
	// requests go through the same mock — no real network or httptest.Server needed.
	const cdnURL = "http://cdn.example.internal/artifact.jar"
	var cdnAuthHeader string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "maven.pkg.github.com" {
			// Registry: return 302 redirect to a fake CDN URL.
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{cdnURL}},
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
		// CDN: record the Authorization header and return the artifact.
		cdnAuthHeader = r.Header.Get("Authorization")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("cdn-artifact-data")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	body, err := g.fetchMavenArtifactBody(t.Context(), "https://maven.pkg.github.com/owner/repo/artifact.jar")
	require.NoError(t, err)
	defer func() { require.NoError(t, body.Close()) }()
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.Equal(t, "cdn-artifact-data", string(data))
	// Authorization header must NOT be forwarded to the CDN (third-party storage).
	assert.Empty(t, cdnAuthHeader, "Authorization header must not be forwarded to CDN after redirect")
}

func TestFetchMavenArtifactBody_Non200ReturnsDownloadError(t *testing.T) {
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("forbidden")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)
	_, err := g.fetchMavenArtifactBody(t.Context(), "https://maven.pkg.github.com/owner/repo/artifact.jar")
	require.Error(t, err)
	var dlErr *MavenDownloadError
	require.True(t, errors.As(err, &dlErr))
	assert.Equal(t, http.StatusForbidden, dlErr.StatusCode)
	assert.Equal(t, "forbidden", dlErr.Message)
	assert.False(t, dlErr.IsNotFound())
}

func TestFetchMavenArtifactBody_404ReturnsNotFoundError(t *testing.T) {
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("not found")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)
	_, err := g.fetchMavenArtifactBody(t.Context(), "https://maven.pkg.github.com/owner/repo/artifact.jar")
	require.Error(t, err)
	var dlErr *MavenDownloadError
	require.True(t, errors.As(err, &dlErr))
	assert.True(t, dlErr.IsNotFound())
}

// --- DownloadMavenArtifacts ---

func TestDownloadMavenArtifacts_JarNotFound_Skipped(t *testing.T) {
	// 404 for .jar, 200 for .pom — the .jar miss should be silently skipped.
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".jar"):
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("not found")),
			}, nil
		case strings.HasSuffix(r.URL.Path, ".pom"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("<project/>")),
			}, nil
		default:
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	artifacts, err := g.DownloadMavenArtifacts(t.Context(), "myorg", "myrepo", "com.example:my-artifact", "1.0.0")
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	assert.Equal(t, "pom", artifacts[0].Ext)
	assert.Equal(t, []byte("<project/>"), artifacts[0].Data)
}

func TestDownloadMavenArtifacts_JarServerError_Surfaced(t *testing.T) {
	// 500 for .jar — the error should NOT be silently swallowed.
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, ".jar") {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("internal error")),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("<project/>")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	_, dlErr := g.DownloadMavenArtifacts(t.Context(), "myorg", "myrepo", "com.example:my-artifact", "1.0.0")
	require.Error(t, dlErr)
	var mavenErr *MavenDownloadError
	assert.True(t, errors.As(dlErr, &mavenErr), "expected MavenDownloadError in error chain")
	assert.Equal(t, http.StatusInternalServerError, mavenErr.StatusCode)
}

func TestDownloadMavenArtifacts_JarUnauthorized_Surfaced(t *testing.T) {
	// 401 for .jar — auth errors must also be surfaced, not ignored.
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, ".jar") {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader("unauthorized")),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("<project/>")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	_, dlErr := g.DownloadMavenArtifacts(t.Context(), "myorg", "myrepo", "com.example:my-artifact", "1.0.0")
	require.Error(t, dlErr)
	var mavenErr *MavenDownloadError
	assert.True(t, errors.As(dlErr, &mavenErr))
	assert.Equal(t, http.StatusUnauthorized, mavenErr.StatusCode)
}

func TestDownloadMavenArtifacts_BothArtifacts(t *testing.T) {
	// 200 for both .jar and .pom — both should be returned.
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".jar"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("fake-jar-bytes")),
			}, nil
		case strings.HasSuffix(r.URL.Path, ".pom"):
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("<project/>")),
			}, nil
		default:
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	artifacts, err := g.DownloadMavenArtifacts(t.Context(), "myorg", "myrepo", "com.example:my-artifact", "1.0.0")
	require.NoError(t, err)
	require.Len(t, artifacts, 2)
	exts := map[string]bool{}
	for _, a := range artifacts {
		exts[a.Ext] = true
	}
	assert.True(t, exts["jar"])
	assert.True(t, exts["pom"])
}

func TestDownloadMavenArtifacts_InvalidPackageName(t *testing.T) {
	g := newTestClient(t, "https://api.github.com/", roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil
	}))
	_, err := g.DownloadMavenArtifacts(t.Context(), "myorg", "myrepo", "invalid-no-separator", "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Maven package name")
}
