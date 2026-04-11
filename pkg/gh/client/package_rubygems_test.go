package client

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildGemMetadataGz creates a gzip-compressed metadata YAML for use in gem archives.
func buildGemMetadataGz(t *testing.T, yaml string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := io.WriteString(gw, yaml)
	require.NoError(t, err)
	require.NoError(t, gw.Close())
	return buf.Bytes()
}

// gemBuildEntry represents a single file in a .gem (tar) archive.
type gemBuildEntry struct {
	name    string
	content []byte
}

// buildGemArchive creates an in-memory .gem tar archive from the given entries.
func buildGemArchive(t *testing.T, entries []gemBuildEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Size: int64(len(e.content)), Mode: 0600}
		require.NoError(t, tw.WriteHeader(hdr))
		_, err := tw.Write(e.content)
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	return buf.Bytes()
}

// buildGemWithChecksums creates a .gem archive containing metadata.gz and checksums.yaml.gz.
// The checksums are computed from the provided entries (excluding checksums.yaml.gz itself).
func buildGemWithChecksums(t *testing.T, metaYAML string, extraEntries []gemBuildEntry) []byte {
	t.Helper()
	metaGz := buildGemMetadataGz(t, metaYAML)
	entries := append([]gemBuildEntry{{name: "metadata.gz", content: metaGz}}, extraEntries...)

	// Build checksums.yaml.gz
	var csYAML bytes.Buffer
	csYAML.WriteString("---\nSHA256:\n")
	for _, e := range entries {
		s := sha256.Sum256(e.content)
		fmt.Fprintf(&csYAML, "  %s: %s\n", e.name, hex.EncodeToString(s[:]))
	}
	csYAML.WriteString("SHA512:\n")
	for _, e := range entries {
		s := sha512.Sum512(e.content)
		fmt.Fprintf(&csYAML, "  %s: '%s'\n", e.name, hex.EncodeToString(s[:]))
	}
	var csBuf bytes.Buffer
	gw := gzip.NewWriter(&csBuf)
	_, err := gw.Write(csYAML.Bytes())
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	entries = append(entries, gemBuildEntry{name: "checksums.yaml.gz", content: csBuf.Bytes()})
	return buildGemArchive(t, entries)
}

// readGemMetadataYAML extracts and decompresses metadata.gz from a .gem archive.
func readGemMetadataYAML(t *testing.T, gemData []byte) string {
	t.Helper()
	r := tar.NewReader(bytes.NewReader(gemData))
	for {
		hdr, err := r.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if hdr.Name != "metadata.gz" {
			continue
		}
		data, err := io.ReadAll(r)
		require.NoError(t, err)
		gr, err := gzip.NewReader(bytes.NewReader(data))
		require.NoError(t, err)
		yaml, err := io.ReadAll(gr)
		require.NoError(t, err)
		require.NoError(t, gr.Close())
		return string(yaml)
	}
	t.Fatal("metadata.gz not found in gem archive")
	return ""
}

// readGemChecksumsYAML extracts and decompresses checksums.yaml.gz from a .gem archive.
func readGemChecksumsYAML(t *testing.T, gemData []byte) string {
	t.Helper()
	r := tar.NewReader(bytes.NewReader(gemData))
	for {
		hdr, err := r.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if hdr.Name != "checksums.yaml.gz" {
			continue
		}
		data, err := io.ReadAll(r)
		require.NoError(t, err)
		gr, err := gzip.NewReader(bytes.NewReader(data))
		require.NoError(t, err)
		yaml, err := io.ReadAll(gr)
		require.NoError(t, err)
		require.NoError(t, gr.Close())
		return string(yaml)
	}
	t.Fatal("checksums.yaml.gz not found in gem archive")
	return ""
}

// --- RubyGemsRegistryBase ---

func TestRubyGemsRegistryBase(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		owner    string
		expected string
	}{
		{
			name:     "github.com",
			host:     "github.com",
			owner:    "myorg",
			expected: "https://rubygems.pkg.github.com/myorg",
		},
		{
			name:     "empty host treated as github.com",
			host:     "",
			owner:    "myorg",
			expected: "https://rubygems.pkg.github.com/myorg",
		},
		{
			name:     "GHES host",
			host:     "ghe.example.com",
			owner:    "myorg",
			expected: "https://rubygems.ghe.example.com/myorg",
		},
		{
			name:     "GHES short host",
			host:     "ghe.internal",
			owner:    "myorg",
			expected: "https://rubygems.ghe.internal/myorg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RubyGemsRegistryBase(tt.host, tt.owner))
		})
	}
}

// --- RubyGemsDownloadURL ---

func TestRubyGemsDownloadURL(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		owner       string
		packageName string
		version     string
		expected    string
	}{
		{
			name:        "github.com",
			host:        "github.com",
			owner:       "myorg",
			packageName: "mygem",
			version:     "1.0.0",
			expected:    "https://rubygems.pkg.github.com/myorg/gems/mygem-1.0.0.gem",
		},
		{
			name:        "empty host treated as github.com",
			host:        "",
			owner:       "myorg",
			packageName: "mygem",
			version:     "2.3.4",
			expected:    "https://rubygems.pkg.github.com/myorg/gems/mygem-2.3.4.gem",
		},
		{
			name:        "GHES host",
			host:        "ghe.example.com",
			owner:       "myorg",
			packageName: "mygem",
			version:     "1.0.0",
			expected:    "https://rubygems.ghe.example.com/myorg/gems/mygem-1.0.0.gem",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RubyGemsDownloadURL(tt.host, tt.owner, tt.packageName, tt.version))
		})
	}
}

// --- RubyGemsPushURL ---

func TestRubyGemsPushURL(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		owner    string
		expected string
	}{
		{
			name:     "github.com",
			host:     "github.com",
			owner:    "myorg",
			expected: "https://rubygems.pkg.github.com/myorg/api/v1/gems",
		},
		{
			name:     "empty host treated as github.com",
			host:     "",
			owner:    "myorg",
			expected: "https://rubygems.pkg.github.com/myorg/api/v1/gems",
		},
		{
			name:     "GHES host",
			host:     "ghe.example.com",
			owner:    "myorg",
			expected: "https://rubygems.ghe.example.com/myorg/api/v1/gems",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RubyGemsPushURL(tt.host, tt.owner))
		})
	}
}

// --- DownloadRubyGemsPackage ---

func TestDownloadRubyGemsPackage_DirectOK(t *testing.T) {
	gemData := buildGemArchive(t, []gemBuildEntry{
		{name: "metadata.gz", content: buildGemMetadataGz(t, "name: mygem\nversion: 1.0.0\n")},
	})
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.String(), "mygem-1.0.0.gem")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(gemData)),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	got, err := g.DownloadRubyGemsPackage(context.Background(), "myorg", "mygem", "1.0.0")
	require.NoError(t, err)
	assert.Equal(t, gemData, got)
}

func TestDownloadRubyGemsPackage_RedirectDropsAuthHeader(t *testing.T) {
	gemData := []byte("gem-binary-data")
	var cdnAuthHeader string

	// Use a local test server to act as the CDN (avoids real DNS lookups).
	cdnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdnAuthHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(gemData)
	}))
	defer cdnServer.Close()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		// Registry returns redirect to the local test CDN server.
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": []string{cdnServer.URL + "/mygem-1.0.0.gem"}},
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	got, err := g.DownloadRubyGemsPackage(context.Background(), "myorg", "mygem", "1.0.0")
	require.NoError(t, err)
	assert.Equal(t, gemData, got)
	assert.Empty(t, cdnAuthHeader, "auth header must not be forwarded to CDN")
}

func TestDownloadRubyGemsPackage_RedirectMissingLocation(t *testing.T) {
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	_, err := g.DownloadRubyGemsPackage(context.Background(), "myorg", "mygem", "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Location")
}

func TestDownloadRubyGemsPackage_ErrorStatus(t *testing.T) {
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("not found")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	_, err := g.DownloadRubyGemsPackage(context.Background(), "myorg", "mygem", "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestDownloadRubyGemsPackage_RedirectCDNErrorStatus(t *testing.T) {
	// Use a local test server that returns 403 to simulate a CDN error.
	cdnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, "forbidden")
	}))
	defer cdnServer.Close()

	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": []string{cdnServer.URL + "/mygem-1.0.0.gem"}},
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	_, err := g.DownloadRubyGemsPackage(context.Background(), "myorg", "mygem", "1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

// --- PushRubyGemsPackage ---

func TestPushRubyGemsPackage_Success201(t *testing.T) {
	gemData := []byte("gem-binary-data")
	var capturedReq *http.Request
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		capturedReq = r
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	err := g.PushRubyGemsPackage(context.Background(), "myorg", bytes.NewReader(gemData))
	require.NoError(t, err)
	assert.Equal(t, "POST", capturedReq.Method)
	assert.Equal(t, "application/octet-stream", capturedReq.Header.Get("Content-Type"))
	assert.Contains(t, capturedReq.URL.String(), "/api/v1/gems")
}

func TestPushRubyGemsPackage_Success200(t *testing.T) {
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	err := g.PushRubyGemsPackage(context.Background(), "myorg", bytes.NewReader([]byte("gem-data")))
	require.NoError(t, err)
}

func TestPushRubyGemsPackage_ErrorStatus(t *testing.T) {
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       io.NopCloser(strings.NewReader("invalid gem")),
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", transport)

	err := g.PushRubyGemsPackage(context.Background(), "myorg", bytes.NewReader([]byte("bad data")))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "422")
}

// --- RewriteGemGitHubRepo ---

func TestRewriteGemGitHubRepo_NoGitHubRepoField(t *testing.T) {
	// When metadata has no github_repo field, the gem should be returned unchanged.
	metaYAML := "name: mygem\nversion: 1.0.0\nauthor: someone\n"
	gemData := buildGemArchive(t, []gemBuildEntry{
		{name: "metadata.gz", content: buildGemMetadataGz(t, metaYAML)},
	})

	result, err := RewriteGemGitHubRepo(gemData, "github.com", "neworg", "newrepo")
	require.NoError(t, err)
	// Should be identical to input when no github_repo field is present.
	assert.Equal(t, gemData, result)
}

func TestRewriteGemGitHubRepo_RewritesGitHubRepoField(t *testing.T) {
	metaYAML := "name: mygem\nversion: 1.0.0\ngithub_repo: ssh://github.com/oldorg/oldrepo\n"
	gemData := buildGemArchive(t, []gemBuildEntry{
		{name: "metadata.gz", content: buildGemMetadataGz(t, metaYAML)},
	})

	result, err := RewriteGemGitHubRepo(gemData, "github.com", "neworg", "newrepo")
	require.NoError(t, err)

	gotYAML := readGemMetadataYAML(t, result)
	assert.Contains(t, gotYAML, "ssh://github.com/neworg/newrepo")
	assert.NotContains(t, gotYAML, "oldorg/oldrepo")
}

func TestRewriteGemGitHubRepo_GHESHost(t *testing.T) {
	metaYAML := "name: mygem\nversion: 1.0.0\ngithub_repo: ssh://github.com/oldorg/oldrepo\n"
	gemData := buildGemArchive(t, []gemBuildEntry{
		{name: "metadata.gz", content: buildGemMetadataGz(t, metaYAML)},
	})

	result, err := RewriteGemGitHubRepo(gemData, "ghe.internal", "neworg", "newrepo")
	require.NoError(t, err)

	gotYAML := readGemMetadataYAML(t, result)
	assert.Contains(t, gotYAML, "ssh://ghe.internal/neworg/newrepo")
	assert.NotContains(t, gotYAML, "oldorg")
}

func TestRewriteGemGitHubRepo_UpdatesChecksums(t *testing.T) {
	metaYAML := "name: mygem\nversion: 1.0.0\ngithub_repo: ssh://github.com/oldorg/oldrepo\n"
	extraEntries := []gemBuildEntry{
		{name: "data.tar.gz", content: []byte("fake-data-tar-gz")},
	}
	gemData := buildGemWithChecksums(t, metaYAML, extraEntries)

	result, err := RewriteGemGitHubRepo(gemData, "github.com", "neworg", "newrepo")
	require.NoError(t, err)

	// Verify metadata.gz was rewritten.
	gotYAML := readGemMetadataYAML(t, result)
	assert.Contains(t, gotYAML, "ssh://github.com/neworg/newrepo")

	// Verify checksums.yaml.gz was rebuilt with the new metadata.gz hash.
	csYAML := readGemChecksumsYAML(t, result)

	// Extract the new metadata.gz content and compute expected SHA256.
	r := tar.NewReader(bytes.NewReader(result))
	var newMetaGz []byte
	for {
		hdr, err := r.Next()
		if err != nil {
			break
		}
		if hdr.Name == "metadata.gz" {
			newMetaGz, err = io.ReadAll(r)
			require.NoError(t, err)
			break
		}
	}
	require.NotNil(t, newMetaGz)
	expectedHash := sha256.Sum256(newMetaGz)
	assert.Contains(t, csYAML, hex.EncodeToString(expectedHash[:]))
}

func TestRewriteGemGitHubRepo_InvalidGemArchive(t *testing.T) {
	_, err := RewriteGemGitHubRepo([]byte("not a tar archive"), "github.com", "org", "repo")
	require.Error(t, err)
}

func TestRewriteGemGitHubRepo_InvalidMetadataGz(t *testing.T) {
	gemData := buildGemArchive(t, []gemBuildEntry{
		{name: "metadata.gz", content: []byte("not gzip data")},
	})

	_, err := RewriteGemGitHubRepo(gemData, "github.com", "org", "repo")
	require.Error(t, err)
}
