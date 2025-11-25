package parser

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

// Helper function to create repository pointers
func repoPtr(owner, name string) *repository.Repository {
	return &repository.Repository{
		Owner: owner,
		Name:  name,
	}
}

func TestParseContentPathFromUses(t *testing.T) {
	tests := []struct {
		name    string
		uses    string
		want    *ContentPath
		wantErr bool
	}{
		{
			name: "local path",
			uses: "./path/to/.github/labeler.yml",
			want: &ContentPath{
				Repo: nil,
				Path: strPtr("path/to/.github/labeler.yml"),
				Ref:  nil,
			},
			wantErr: false,
		},
		{
			name:    "local path without dot",
			uses:    ".github/labeler.yml",
			want:    nil,
			wantErr: true,
		},
		{
			name: "owner/repo@ref format",
			uses: "actions/checkout@v2",
			want: &ContentPath{
				Repo: repoPtr("actions", "checkout"),
				Path: nil,
				Ref:  strPtr("v2"),
			},
			wantErr: false,
		},
		{
			name: "owner/repo/path@ref format",
			uses: "actions/workflows/deploy@main",
			want: &ContentPath{
				Repo: repoPtr("actions", "workflows"),
				Path: strPtr("deploy"),
				Ref:  strPtr("main"),
			},
			wantErr: false,
		},
		{
			name: "owner/repo/nested/path@ref format",
			uses: "actions/workflows/nested/deploy@v1.0.0",
			want: &ContentPath{
				Repo: repoPtr("actions", "workflows"),
				Path: strPtr("nested/deploy"),
				Ref:  strPtr("v1.0.0"),
			},
			wantErr: false,
		},
		{
			name:    "invalid format - no @ref",
			uses:    "actions/checkout",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - missing repo",
			uses:    "actions",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - only slash",
			uses:    "/",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty string",
			uses:    "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseContentPathFromUses(tt.uses)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContentPathFromUses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if (got.Repo == nil) != (tt.want.Repo == nil) {
					t.Errorf("ParseContentPathFromUses() Repo = %v, want %v", got.Repo, tt.want.Repo)
				} else if got.Repo != nil {
					if got.Repo.Owner != tt.want.Repo.Owner {
						t.Errorf("ParseContentPathFromUses() Repo.Owner = %v, want %v", got.Repo.Owner, tt.want.Repo.Owner)
					}
					if got.Repo.Name != tt.want.Repo.Name {
						t.Errorf("ParseContentPathFromUses() Repo.Name = %v, want %v", got.Repo.Name, tt.want.Repo.Name)
					}
				}
				if (got.Path == nil) != (tt.want.Path == nil) {
					t.Errorf("ParseContentPathFromUses() Path = %v, want %v", got.Path, tt.want.Path)
				} else if got.Path != nil && *got.Path != *tt.want.Path {
					t.Errorf("ParseContentPathFromUses() Path = %v, want %v", *got.Path, *tt.want.Path)
				}
				if (got.Ref == nil) != (tt.want.Ref == nil) {
					t.Errorf("ParseContentPathFromUses() Ref = %v, want %v", got.Ref, tt.want.Ref)
				} else if got.Ref != nil && *got.Ref != *tt.want.Ref {
					t.Errorf("ParseContentPathFromUses() Ref = %v, want %v", *got.Ref, *tt.want.Ref)
				}
			}
		})
	}
}

func TestParseContentPathFromURL(t *testing.T) {
	tests := []struct {
		name    string
		htmlUrl string
		want    *ContentPath
		wantErr bool
	}{
		{
			name:    "repository URL",
			htmlUrl: "https://github.com/owner/repo",
			want: &ContentPath{
				Repo: repoPtr("owner", "repo"),
				Path: nil,
				Ref:  nil,
			},
			wantErr: false,
		},
		{
			name:    "blob URL with file path",
			htmlUrl: "https://github.com/owner/repo/blob/main/README.md",
			want: &ContentPath{
				Repo: repoPtr("owner", "repo"),
				Path: strPtr("README.md"),
				Ref:  strPtr("main"),
			},
			wantErr: false,
		},
		{
			name:    "blob URL with nested file path",
			htmlUrl: "https://github.com/owner/repo/blob/main/path/to/file.go",
			want: &ContentPath{
				Repo: repoPtr("owner", "repo"),
				Path: strPtr("path/to/file.go"),
				Ref:  strPtr("main"),
			},
			wantErr: false,
		},
		{
			name:    "tree URL with branch",
			htmlUrl: "https://github.com/owner/repo/tree/main",
			want: &ContentPath{
				Repo: repoPtr("owner", "repo"),
				Path: nil,
				Ref:  strPtr("main"),
			},
			wantErr: false,
		},
		{
			name:    "tree URL with nested path",
			htmlUrl: "https://github.com/owner/repo/tree/main/pkg/parser",
			want: &ContentPath{
				Repo: repoPtr("owner", "repo"),
				Path: nil,
				Ref:  strPtr("main/pkg/parser"),
			},
			wantErr: false,
		},
		{
			name:    "issues URL - not blob or tree",
			htmlUrl: "https://github.com/owner/repo/issues/123",
			want: &ContentPath{
				Repo: repoPtr("owner", "repo"),
				Path: nil,
				Ref:  nil,
			},
			wantErr: false,
		},
		{
			name:    "pull request URL - not blob or tree",
			htmlUrl: "https://github.com/owner/repo/pull/456",
			want: &ContentPath{
				Repo: repoPtr("owner", "repo"),
				Path: nil,
				Ref:  nil,
			},
			wantErr: false,
		},
		{
			name:    "blob URL without file path",
			htmlUrl: "https://github.com/owner/repo/blob/main",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid URL format - missing repo",
			htmlUrl: "https://github.com/owner",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid URL",
			htmlUrl: "not-a-valid-url",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "URL with only owner",
			htmlUrl: "https://github.com/owner/",
			want: &ContentPath{
				Repo: repoPtr("owner", ""),
				Path: nil,
				Ref:  nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseContentPathFromURL(tt.htmlUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContentPathFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if (got.Repo == nil) != (tt.want.Repo == nil) {
					t.Errorf("ParseContentPathFromURL() Repo = %v, want %v", got.Repo, tt.want.Repo)
				} else if got.Repo != nil {
					if got.Repo.Owner != tt.want.Repo.Owner {
						t.Errorf("ParseContentPathFromURL() Repo.Owner = %v, want %v", got.Repo.Owner, tt.want.Repo.Owner)
					}
					if got.Repo.Name != tt.want.Repo.Name {
						t.Errorf("ParseContentPathFromURL() Repo.Name = %v, want %v", got.Repo.Name, tt.want.Repo.Name)
					}
				}
				if (got.Path == nil) != (tt.want.Path == nil) {
					t.Errorf("ParseContentPathFromURL() Path = %v, want %v", got.Path, tt.want.Path)
				} else if got.Path != nil && *got.Path != *tt.want.Path {
					t.Errorf("ParseContentPathFromURL() Path = %v, want %v", *got.Path, *tt.want.Path)
				}
				if (got.Ref == nil) != (tt.want.Ref == nil) {
					t.Errorf("ParseContentPathFromURL() Ref = %v, want %v", got.Ref, tt.want.Ref)
				} else if got.Ref != nil && *got.Ref != *tt.want.Ref {
					t.Errorf("ParseContentPathFromURL() Ref = %v, want %v", *got.Ref, *tt.want.Ref)
				}
			}
		})
	}
}

func TestParseContentPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *ContentPath
		wantErr bool
	}{
		{
			name:  "HTTP URL",
			input: "http://github.com/owner/repo",
			want: &ContentPath{
				Repo: repoPtr("owner", "repo"),
				Path: nil,
				Ref:  nil,
			},
			wantErr: false,
		},
		{
			name:  "HTTPS URL",
			input: "https://github.com/owner/repo/blob/main/README.md",
			want: &ContentPath{
				Repo: repoPtr("owner", "repo"),
				Path: strPtr("README.md"),
				Ref:  strPtr("main"),
			},
			wantErr: false,
		},
		{
			name:  "uses format",
			input: "actions/checkout@v2",
			want: &ContentPath{
				Repo: repoPtr("actions", "checkout"),
				Path: nil,
				Ref:  strPtr("v2"),
			},
			wantErr: false,
		},
		{
			name:  "local relative path",
			input: "./path/to/action",
			want: &ContentPath{
				Repo: nil,
				Path: strPtr("path/to/action"),
				Ref:  nil,
			},
			wantErr: false,
		},
		{
			name:  "local path",
			input: ".github/path/to/action",
			want: &ContentPath{
				Repo: nil,
				Path: strPtr(".github/path/to/action"),
				Ref:  nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseContentPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContentPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if (got.Repo == nil) != (tt.want.Repo == nil) {
					t.Errorf("ParseContentPath() Repo = %v, want %v", got.Repo, tt.want.Repo)
				} else if got.Repo != nil {
					if got.Repo.Owner != tt.want.Repo.Owner {
						t.Errorf("ParseContentPath() Repo.Owner = %v, want %v", got.Repo.Owner, tt.want.Repo.Owner)
					}
					if got.Repo.Name != tt.want.Repo.Name {
						t.Errorf("ParseContentPath() Repo.Name = %v, want %v", got.Repo.Name, tt.want.Repo.Name)
					}
				}
				if (got.Path == nil) != (tt.want.Path == nil) {
					t.Errorf("ParseContentPath() Path = %v, want %v", got.Path, tt.want.Path)
				} else if got.Path != nil && *got.Path != *tt.want.Path {
					t.Errorf("ParseContentPath() Path = %v, want %v", *got.Path, *tt.want.Path)
				}
				if (got.Ref == nil) != (tt.want.Ref == nil) {
					t.Errorf("ParseContentPath() Ref = %v, want %v", got.Ref, tt.want.Ref)
				} else if got.Ref != nil && *got.Ref != *tt.want.Ref {
					t.Errorf("ParseContentPath() Ref = %v, want %v", *got.Ref, *tt.want.Ref)
				}
			}
		})
	}
}
