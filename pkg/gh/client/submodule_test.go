package client

import (
	"testing"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveGitURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		gitURL   string
		expected string
		wantErr  bool
	}{
		{
			name:     "absolute https URL is returned as-is",
			baseURL:  "https://github.com/owner/repo",
			gitURL:   "https://github.com/other/repo",
			expected: "https://github.com/other/repo",
		},
		{
			name:     "absolute ssh URL is returned as-is",
			baseURL:  "https://github.com/owner/repo",
			gitURL:   "git@github.com:other/repo.git",
			expected: "git@github.com:other/repo.git",
		},
		{
			name:     "../sibling resolves to sibling repo",
			baseURL:  "https://github.com/owner/repo",
			gitURL:   "../sibling",
			expected: "https://github.com/owner/sibling",
		},
		{
			name:     "../../other-owner/other-repo resolves to different owner",
			baseURL:  "https://github.com/owner/repo",
			gitURL:   "../../other-owner/other-repo",
			expected: "https://github.com/other-owner/other-repo",
		},
		{
			name:     "./child resolves relative to base",
			baseURL:  "https://github.com/owner/repo",
			gitURL:   "./child",
			expected: "https://github.com/owner/repo/child",
		},
		{
			name:     "../sibling on GHES host",
			baseURL:  "https://ghe.example.com/owner/repo",
			gitURL:   "../sibling",
			expected: "https://ghe.example.com/owner/sibling",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveGitURL(tt.baseURL, tt.gitURL)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestConvertSubmodule(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		node       repositorySubmoduleObject
		wantGitUrl string
		wantOwner  string
		wantRepo   string
		wantBranch string
		wantPath   string
		wantName   string
		wantErr    bool
	}{
		{
			name:    "absolute URL submodule",
			baseURL: "https://github.com/owner/repo",
			node: repositorySubmoduleObject{
				Name:   githubv4.String("mylib"),
				GitUrl: githubv4.String("https://github.com/owner/mylib"),
				Branch: githubv4.String("main"),
				Path:   githubv4.String("vendor/mylib"),
			},
			wantGitUrl: "https://github.com/owner/mylib",
			wantOwner:  "owner",
			wantRepo:   "mylib",
			wantBranch: "main",
			wantPath:   "vendor/mylib",
			wantName:   "mylib",
		},
		{
			name:    "relative URL ../ resolves to sibling",
			baseURL: "https://github.com/owner/repo",
			node: repositorySubmoduleObject{
				Name:   githubv4.String("sibling"),
				GitUrl: githubv4.String("../sibling"),
				Branch: githubv4.String("develop"),
				Path:   githubv4.String("libs/sibling"),
			},
			wantGitUrl: "../sibling",
			wantOwner:  "owner",
			wantRepo:   "sibling",
			wantBranch: "develop",
			wantPath:   "libs/sibling",
			wantName:   "sibling",
		},
		{
			name:    "relative URL ../../ resolves to different owner",
			baseURL: "https://github.com/owner/repo",
			node: repositorySubmoduleObject{
				Name:   githubv4.String("other-repo"),
				GitUrl: githubv4.String("../../other-owner/other-repo"),
				Branch: githubv4.String(""),
				Path:   githubv4.String("ext/other-repo"),
			},
			wantGitUrl: "../../other-owner/other-repo",
			wantOwner:  "other-owner",
			wantRepo:   "other-repo",
			wantBranch: "",
			wantPath:   "ext/other-repo",
			wantName:   "other-repo",
		},
		{
			name:    "unparseable URL returns error",
			baseURL: "https://github.com/owner/repo",
			node: repositorySubmoduleObject{
				Name:   githubv4.String("bad"),
				GitUrl: githubv4.String("not-a-valid-github-url"),
				Branch: githubv4.String(""),
				Path:   githubv4.String("bad"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertSubmodule(tt.node, tt.baseURL)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantGitUrl, got.GitUrl)
			assert.Equal(t, tt.wantOwner, got.Repository.Owner)
			assert.Equal(t, tt.wantRepo, got.Repository.Name)
			assert.Equal(t, tt.wantBranch, got.Branch)
			assert.Equal(t, tt.wantPath, got.Path)
			assert.Equal(t, tt.wantName, got.Name)
		})
	}
}
