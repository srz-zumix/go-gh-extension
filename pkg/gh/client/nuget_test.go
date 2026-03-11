package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			expected: "https://ghe.example.com/_registry/nuget/myorg",
		},
		{
			name:     "GHES short host",
			host:     "ghe.internal",
			owner:    "myorg",
			expected: "https://ghe.internal/_registry/nuget/myorg",
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
			expected:    "https://ghe.example.com/_registry/nuget/myorg/download/mypackage/1.0.0/mypackage.1.0.0.nupkg",
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
			expected: "https://ghe.example.com/_registry/nuget/myorg",
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
