package client

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNuGetRegistryBase(t *testing.T) {
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
			expected: "https://nuget.pkg.github.com/myorg",
		},
		{
			name:     "empty host treated as github.com",
			host:     "",
			owner:    "myorg",
			expected: "https://nuget.pkg.github.com/myorg",
		},
		{
			name:     "GHES host",
			host:     "ghe.example.com",
			owner:    "myorg",
			expected: "https://nuget.ghe.example.com/myorg",
		},
		{
			name:     "GHES short host",
			host:     "ghe.internal",
			owner:    "myorg",
			expected: "https://nuget.ghe.internal/myorg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NuGetRegistryBase(tt.host, tt.owner))
		})
	}
}

func TestNuGetDownloadURL(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		owner       string
		packageName string
		version     string
		expected    string
	}{
		{
			name:        "github.com lowercase package",
			host:        "github.com",
			owner:       "myorg",
			packageName: "MyPackage",
			version:     "1.0.0",
			expected:    "https://nuget.pkg.github.com/myorg/download/mypackage/1.0.0/mypackage.1.0.0.nupkg",
		},
		{
			name:        "github.com already lowercase",
			host:        "github.com",
			owner:       "myorg",
			packageName: "mypackage",
			version:     "2.3.4",
			expected:    "https://nuget.pkg.github.com/myorg/download/mypackage/2.3.4/mypackage.2.3.4.nupkg",
		},
		{
			name:        "empty host treated as github.com",
			host:        "",
			owner:       "myorg",
			packageName: "Pkg",
			version:     "1.0.0",
			expected:    "https://nuget.pkg.github.com/myorg/download/pkg/1.0.0/pkg.1.0.0.nupkg",
		},
		{
			name:        "GHES host",
			host:        "ghe.example.com",
			owner:       "myorg",
			packageName: "MyPackage",
			version:     "1.0.0",
			expected:    "https://nuget.ghe.example.com/myorg/download/mypackage/1.0.0/mypackage.1.0.0.nupkg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NuGetDownloadURL(tt.host, tt.owner, tt.packageName, tt.version))
		})
	}
}

func TestNuGetPushURL(t *testing.T) {
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
			expected: "https://nuget.pkg.github.com/myorg",
		},
		{
			name:     "empty host treated as github.com",
			host:     "",
			owner:    "myorg",
			expected: "https://nuget.pkg.github.com/myorg",
		},
		{
			name:     "GHES host",
			host:     "ghe.example.com",
			owner:    "myorg",
			expected: "https://nuget.ghe.example.com/myorg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NuGetPushURL(tt.host, tt.owner))
		})
	}
}

func TestRewriteNuspecRepository_ReplaceExisting(t *testing.T) {
	nuspec := []byte(`<metadata><repository type="git" url="https://github.com/old/repo" /></metadata>`)
	result := rewriteNuspecRepository(nuspec, "https://github.com/new/repo")
	assert.Contains(t, string(result), `url="https://github.com/new/repo"`)
	assert.NotContains(t, string(result), "old/repo")
}

func TestRewriteNuspecRepository_InsertMissing(t *testing.T) {
	nuspec := []byte("<metadata><id>MyPkg</id></metadata>")
	result := rewriteNuspecRepository(nuspec, "https://github.com/myorg/myrepo")
	assert.Contains(t, string(result), `<repository type="git" url="https://github.com/myorg/myrepo" />`)
	assert.Contains(t, string(result), "</metadata>")
}

func TestRewriteNuspecRepository_ReplaceMultilineElement(t *testing.T) {
	nuspec := []byte("<metadata>\n\t<repository type=\"git\" url=\"https://old.example.com/repo\">\n\t</repository>\n</metadata>")
	result := rewriteNuspecRepository(nuspec, "https://github.com/new/repo")
	assert.Contains(t, string(result), `url="https://github.com/new/repo"`)
	assert.NotContains(t, string(result), "old.example.com")
}

// buildSampleNuPkg creates a minimal in-memory .nupkg (ZIP) archive for testing.
// The archive contains a single .nuspec file with the given nuspec content,
// plus an optional extra entry (name="lib/net6.0/sample.dll", content=dllContent)
// if dllContent is non-nil.
func buildSampleNuPkg(t *testing.T, nuspecContent string, dllContent []byte) ([]byte, int64) {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	fw, err := w.Create("sample.nuspec")
	require.NoError(t, err)
	_, err = fw.Write([]byte(nuspecContent))
	require.NoError(t, err)

	if dllContent != nil {
		fw2, err := w.Create("lib/net6.0/sample.dll")
		require.NoError(t, err)
		_, err = fw2.Write(dllContent)
		require.NoError(t, err)
	}

	require.NoError(t, w.Close())
	data := buf.Bytes()
	return data, int64(len(data))
}

func TestRewriteNuPkgRepository_SelfClosingTag(t *testing.T) {
	nuspec := `<?xml version="1.0"?>
<package>
  <metadata>
    <id>SamplePkg</id>
    <version>1.0.0</version>
    <repository type="git" url="https://github.com/old/repo" />
  </metadata>
</package>`
	data, size := buildSampleNuPkg(t, nuspec, nil)

	var out bytes.Buffer
	err := RewriteNuPkgRepository(bytes.NewReader(data), size, &out, "https://github.com/new/repo")
	require.NoError(t, err)

	outData := out.Bytes()
	r, err := zip.NewReader(bytes.NewReader(outData), int64(len(outData)))
	require.NoError(t, err)
	require.Len(t, r.File, 1)

	rc, err := r.File[0].Open()
	require.NoError(t, err)
	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	_ = rc.Close()

	assert.Contains(t, string(content), `url="https://github.com/new/repo"`)
	assert.NotContains(t, string(content), "old/repo")
}

func TestRewriteNuPkgRepository_NoRepositoryElement(t *testing.T) {
	nuspec := `<?xml version="1.0"?>
<package>
  <metadata>
    <id>SamplePkg</id>
    <version>1.0.0</version>
  </metadata>
</package>`
	data, size := buildSampleNuPkg(t, nuspec, nil)

	var out bytes.Buffer
	err := RewriteNuPkgRepository(bytes.NewReader(data), size, &out, "https://github.com/myorg/myrepo")
	require.NoError(t, err)

	outData := out.Bytes()
	r, err := zip.NewReader(bytes.NewReader(outData), int64(len(outData)))
	require.NoError(t, err)
	require.Len(t, r.File, 1)

	rc, err := r.File[0].Open()
	require.NoError(t, err)
	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	_ = rc.Close()

	assert.Contains(t, string(content), `url="https://github.com/myorg/myrepo"`)
}

func TestRewriteNuPkgRepository_WithBinaryEntry(t *testing.T) {
	nuspec := `<?xml version="1.0"?>
<package>
  <metadata>
    <id>SamplePkg</id>
    <version>1.0.0</version>
    <repository type="git" url="https://github.com/old/repo" />
  </metadata>
</package>`
	dllContent := []byte{0x4D, 0x5A, 0x90, 0x00} // PE header magic bytes
	data, size := buildSampleNuPkg(t, nuspec, dllContent)

	var out bytes.Buffer
	err := RewriteNuPkgRepository(bytes.NewReader(data), size, &out, "https://github.com/new/repo")
	require.NoError(t, err)

	outData := out.Bytes()
	r, err := zip.NewReader(bytes.NewReader(outData), int64(len(outData)))
	require.NoError(t, err)
	require.Len(t, r.File, 2)

	// Verify nuspec was rewritten
	var nuspecEntry *zip.File
	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".nuspec") {
			nuspecEntry = f
		}
	}
	require.NotNil(t, nuspecEntry)
	rc, err := nuspecEntry.Open()
	require.NoError(t, err)
	content, err := io.ReadAll(rc)
	require.NoError(t, err)
	_ = rc.Close()
	assert.Contains(t, string(content), `url="https://github.com/new/repo"`)

	// Verify binary entry was preserved intact
	var dllEntry *zip.File
	for _, f := range r.File {
		if f.Name == "lib/net6.0/sample.dll" {
			dllEntry = f
		}
	}
	require.NotNil(t, dllEntry)
	rc2, err := dllEntry.Open()
	require.NoError(t, err)
	got, err := io.ReadAll(rc2)
	require.NoError(t, err)
	_ = rc2.Close()
	assert.Equal(t, dllContent, got)
}
