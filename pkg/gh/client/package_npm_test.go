package client

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roundTripFunc is an http.RoundTripper backed by a function, for use in tests.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// captureTransport records the last request passed to the inner transport and
// returns a fixed response. Used to assert that specific headers were (or were not) set.
type captureTransport struct {
	lastReq  *http.Request
	response *http.Response
}

func (c *captureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.lastReq = r
	return c.response, nil
}

// tgzEntry represents a single file entry in a .tgz archive for testing.
type tgzEntry struct {
	name    string
	content []byte
}

// buildSampleTgz creates an in-memory .tgz archive containing the given entries.
func buildSampleTgz(t *testing.T, entries []tgzEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Size: int64(len(e.content)), Mode: 0600}
		require.NoError(t, tw.WriteHeader(hdr))
		_, err := tw.Write(e.content)
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gzw.Close())
	return buf.Bytes()
}

// buildNPMMetadataJSON returns a minimal npm metadata JSON body for the given
// version and tarball URL, matching the npmMetadata struct.
func buildNPMMetadataJSON(t *testing.T, version, tarballURL string) string {
	t.Helper()
	meta := npmMetadata{
		Versions: map[string]npmVersionMeta{
			version: {Dist: npmVersionDist{Tarball: tarballURL}},
		},
	}
	b, err := json.Marshal(meta)
	require.NoError(t, err)
	return string(b)
}

// --- NPMRegistryBase ---

func TestNPMRegistryBase(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{"github.com", "github.com", "https://npm.pkg.github.com/"},
		{"empty host treated as github.com", "", "https://npm.pkg.github.com/"},
		{"GHES host", "ghe.example.com", "https://npm.ghe.example.com/"},
		{"GHES short host", "ghe.internal", "https://npm.ghe.internal/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NPMRegistryBase(tt.host))
		})
	}
}

// --- NPMPackageURL ---

func TestNPMPackageURL(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		owner       string
		packageName string
		expected    string
	}{
		{
			name:        "unscoped name with owner is scoped as @owner/pkg",
			host:        "github.com",
			owner:       "myorg",
			packageName: "mypkg",
			expected:    "https://npm.pkg.github.com/@myorg/mypkg",
		},
		{
			name:        "already scoped name is not double-scoped",
			host:        "github.com",
			owner:       "myorg",
			packageName: "@org/mypkg",
			expected:    "https://npm.pkg.github.com/@org/mypkg",
		},
		{
			name:        "empty owner leaves name unscoped",
			host:        "github.com",
			owner:       "",
			packageName: "mypkg",
			expected:    "https://npm.pkg.github.com/mypkg",
		},
		{
			name:        "empty host treated as github.com",
			host:        "",
			owner:       "myorg",
			packageName: "mypkg",
			expected:    "https://npm.pkg.github.com/@myorg/mypkg",
		},
		{
			name:        "GHES host",
			host:        "ghe.example.com",
			owner:       "myorg",
			packageName: "mypkg",
			expected:    "https://npm.ghe.example.com/@myorg/mypkg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NPMPackageURL(tt.host, tt.owner, tt.packageName))
		})
	}
}

// --- npmBearerAuthTransport ---

func TestNPMBearerAuthTransport_AddsHeaderForRegistryHost(t *testing.T) {
	inner := &captureTransport{
		response: &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))},
	}
	tr := &npmBearerAuthTransport{
		token:            "secret-token",
		registryHostname: "npm.pkg.github.com",
		registryPort:     "",
		transport:        inner,
	}
	req, err := http.NewRequest(http.MethodGet, "https://npm.pkg.github.com/@owner/mypkg", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, "Bearer secret-token", inner.lastReq.Header.Get("Authorization"))
}

func TestNPMBearerAuthTransport_NoHeaderForOtherHost(t *testing.T) {
	// Requests to CDN/S3 tarball URLs must NOT receive the Authorization header,
	// because pre-signed URLs embed auth in query params and reject extra headers.
	inner := &captureTransport{
		response: &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))},
	}
	tr := &npmBearerAuthTransport{
		token:            "secret-token",
		registryHostname: "npm.pkg.github.com",
		registryPort:     "",
		transport:        inner,
	}
	req, err := http.NewRequest(http.MethodGet, "https://cdn.example.com/tarballs/pkg.tgz", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Empty(t, inner.lastReq.Header.Get("Authorization"))
}

func TestNPMBearerAuthTransport_NoHeaderWhenTokenEmpty(t *testing.T) {
	inner := &captureTransport{
		response: &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))},
	}
	tr := &npmBearerAuthTransport{
		token:            "",
		registryHostname: "npm.pkg.github.com",
		registryPort:     "",
		transport:        inner,
	}
	req, err := http.NewRequest(http.MethodGet, "https://npm.pkg.github.com/@owner/mypkg", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Empty(t, inner.lastReq.Header.Get("Authorization"))
}

func TestNPMBearerAuthTransport_ExplicitPortMismatch(t *testing.T) {
	// registryPort="" (no explicit port in registry URL) must NOT match a request
	// that includes an explicit port in its URL, since Port() strings differ ("" vs "443").
	inner := &captureTransport{
		response: &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))},
	}
	tr := &npmBearerAuthTransport{
		token:            "secret-token",
		registryHostname: "npm.pkg.github.com",
		registryPort:     "",
		transport:        inner,
	}
	req, err := http.NewRequest(http.MethodGet, "https://npm.pkg.github.com:443/@owner/mypkg", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Empty(t, inner.lastReq.Header.Get("Authorization"))
}

func TestNPMBearerAuthTransport_ExplicitPortMatch(t *testing.T) {
	// When the registry URL was parsed with an explicit port, a matching explicit
	// port in the request URL should trigger the Authorization header.
	inner := &captureTransport{
		response: &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))},
	}
	tr := &npmBearerAuthTransport{
		token:            "secret-token",
		registryHostname: "ghe.example.com",
		registryPort:     "8080",
		transport:        inner,
	}
	req, err := http.NewRequest(http.MethodGet, "http://ghe.example.com:8080/@owner/mypkg", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, "Bearer secret-token", inner.lastReq.Header.Get("Authorization"))
}

// --- rewritePackageJSON ---

func TestRewritePackageJSON_UpdatesRepository(t *testing.T) {
	input := []byte(`{"name":"mypkg","version":"1.0.0","description":"A package"}`)
	result, err := rewritePackageJSON(input, "https://github.com/myorg/myrepo")
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(result, &got))

	// Other fields must be preserved
	assert.Equal(t, "mypkg", got["name"])
	assert.Equal(t, "1.0.0", got["version"])
	assert.Equal(t, "A package", got["description"])

	// repository field must be set correctly
	repo, ok := got["repository"].(map[string]any)
	require.True(t, ok, "repository should be an object")
	assert.Equal(t, "git", repo["type"])
	assert.Equal(t, "https://github.com/myorg/myrepo", repo["url"])
}

func TestRewritePackageJSON_EmptyRepoURLSkipsUpdate(t *testing.T) {
	input := []byte(`{"name":"mypkg","version":"1.0.0","repository":{"type":"git","url":"https://github.com/old/repo"}}`)
	result, err := rewritePackageJSON(input, "")
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(result, &got))

	// repository must be unchanged when repoURL is empty
	repo, ok := got["repository"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "https://github.com/old/repo", repo["url"])
}

func TestRewritePackageJSON_InvalidJSONReturnsError(t *testing.T) {
	_, err := rewritePackageJSON([]byte(`{invalid json`), "https://github.com/myorg/myrepo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse package.json")
}

// --- RewriteNPMPackageJSON ---

func TestRewriteNPMPackageJSON_UpdatesPackageJSON(t *testing.T) {
	pkgJSON := []byte(`{"name":"mypkg","version":"1.0.0","repository":{"type":"git","url":"https://github.com/old/repo"}}`)
	tgz := buildSampleTgz(t, []tgzEntry{
		{name: "package/package.json", content: pkgJSON},
	})

	result, err := RewriteNPMPackageJSON(bytes.NewReader(tgz), "https://github.com/new/repo")
	require.NoError(t, err)

	gzr, err := gzip.NewReader(bytes.NewReader(result))
	require.NoError(t, err)
	tr := tar.NewReader(gzr)

	hdr, err := tr.Next()
	require.NoError(t, err)
	assert.Equal(t, "package/package.json", hdr.Name)

	content, err := io.ReadAll(tr)
	require.NoError(t, err)
	assert.Contains(t, string(content), "https://github.com/new/repo")
	assert.NotContains(t, string(content), "https://github.com/old/repo")
	// Non-repository fields must be preserved
	assert.Contains(t, string(content), `"name"`)
	assert.Contains(t, string(content), "mypkg")
	assert.Contains(t, string(content), "1.0.0")
}

func TestRewriteNPMPackageJSON_PreservesOtherEntries(t *testing.T) {
	pkgJSON := []byte(`{"name":"mypkg","version":"1.0.0"}`)
	jsContent := []byte(`module.exports = {}`)
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF}
	tgz := buildSampleTgz(t, []tgzEntry{
		{name: "package/package.json", content: pkgJSON},
		{name: "package/index.js", content: jsContent},
		{name: "package/lib/native.node", content: binaryContent},
	})

	result, err := RewriteNPMPackageJSON(bytes.NewReader(tgz), "https://github.com/new/repo")
	require.NoError(t, err)

	gzr, err := gzip.NewReader(bytes.NewReader(result))
	require.NoError(t, err)
	tarR := tar.NewReader(gzr)

	entries := map[string][]byte{}
	for {
		hdr, err := tarR.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		data, err := io.ReadAll(tarR)
		require.NoError(t, err)
		entries[hdr.Name] = data
	}

	assert.Len(t, entries, 3)
	assert.Equal(t, jsContent, entries["package/index.js"])
	assert.Equal(t, binaryContent, entries["package/lib/native.node"])
}

// --- DownloadNPMPackage: Content-Type validation ---

func TestDownloadNPMPackage_AcceptsVendorJSONContentType(t *testing.T) {
	tarballData := buildSampleTgz(t, []tgzEntry{
		{name: "package/package.json", content: []byte(`{"name":"mypkg","version":"1.0.0"}`)},
	})
	const tarballURL = "https://npm.pkg.github.com/download/@owner/mypkg/1.0.0/mypkg.tgz"
	metaBody := buildNPMMetadataJSON(t, "1.0.0", tarballURL)

	reqCount := 0
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		reqCount++
		if reqCount == 1 {
			// Metadata request: return vendor +json content type as GitHub npm registry does
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/vnd.npm.install-v1+json"}},
				Body:       io.NopCloser(strings.NewReader(metaBody)),
			}, nil
		}
		// Tarball request
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/octet-stream"}},
			Body:       io.NopCloser(bytes.NewReader(tarballData)),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	data, err := g.DownloadNPMPackage(context.Background(), "owner", "mypkg", "1.0.0")
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Equal(t, 2, reqCount, "expected metadata + tarball requests")
}

func TestDownloadNPMPackage_AcceptsApplicationJSON(t *testing.T) {
	tarballData := buildSampleTgz(t, []tgzEntry{
		{name: "package/package.json", content: []byte(`{"name":"mypkg","version":"1.0.0"}`)},
	})
	const tarballURL = "https://npm.pkg.github.com/download/@owner/mypkg/1.0.0/mypkg.tgz"
	metaBody := buildNPMMetadataJSON(t, "1.0.0", tarballURL)

	reqCount := 0
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		reqCount++
		if reqCount == 1 {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(metaBody)),
			}, nil
		}
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/octet-stream"}},
			Body:       io.NopCloser(bytes.NewReader(tarballData)),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	data, err := g.DownloadNPMPackage(context.Background(), "owner", "mypkg", "1.0.0")
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestDownloadNPMPackage_RejectsHTMLContentType(t *testing.T) {
	// text/html indicates the registry redirected to a login page or proxy error page.
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
			Body:       io.NopCloser(strings.NewReader("<html>Login</html>")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	_, err := g.DownloadNPMPackage(context.Background(), "owner", "mypkg", "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected content type")
}
