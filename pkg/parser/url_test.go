package parser

import (
	"net/url"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
)

func TestParsePullRequestURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *PullRequestURL
		wantErr bool
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  nil,
		},
		{
			name:  "not a URL",
			input: "feature-branch",
			want:  nil,
		},
		{
			name:  "valid HTTPS PR URL",
			input: "https://github.com/owner/repo/pull/123",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo/pull/123"),
				Number: intPtr(123),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "valid HTTP PR URL",
			input: "http://github.com/owner/repo/pull/456",
			want: &PullRequestURL{
				Url:    mustParseURL("http://github.com/owner/repo/pull/456"),
				Number: intPtr(456),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "PR URL with trailing slash",
			input: "https://github.com/owner/repo/pull/789/",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo/pull/789/"),
				Number: intPtr(789),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "PR URL with additional path (files tab)",
			input: "https://github.com/owner/repo/pull/100/files",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo/pull/100/files"),
				Number: intPtr(100),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "PR URL with additional path (commits tab)",
			input: "https://github.com/owner/repo/pull/200/commits",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo/pull/200/commits"),
				Number: intPtr(200),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "PR URL with query parameters",
			input: "https://github.com/owner/repo/pull/300?tab=files",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo/pull/300?tab=files"),
				Number: intPtr(300),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "PR URL with fragment",
			input: "https://github.com/owner/repo/pull/400#issuecomment-123",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo/pull/400#issuecomment-123"),
				Number: intPtr(400),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "PR URL with whitespace",
			input: "  https://github.com/owner/repo/pull/500  ",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo/pull/500"),
				Number: intPtr(500),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "GitHub Enterprise URL",
			input: "https://github.example.com/owner/repo/pull/600",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.example.com/owner/repo/pull/600"),
				Number: intPtr(600),
				Repo: &repository.Repository{
					Host:  "github.example.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "owner name with hyphen",
			input: "https://github.com/my-org/repo/pull/700",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/my-org/repo/pull/700"),
				Number: intPtr(700),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "my-org",
					Name:  "repo",
				},
			},
		},
		{
			name:  "repo name with dots",
			input: "https://github.com/owner/repo.name/pull/800",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo.name/pull/800"),
				Number: intPtr(800),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo.name",
				},
			},
		},
		{
			name:  "PR URL from actions run with pr query parameter",
			input: "https://github.com/srz-zumix/go-gh-extension/actions/runs/20018947899/job/57401849098?pr=69",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/srz-zumix/go-gh-extension/actions/runs/20018947899/job/57401849098?pr=69"),
				Number: intPtr(69),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "srz-zumix",
					Name:  "go-gh-extension",
				},
			},
		},
		{
			name:  "PR URL from workflow run with pr query parameter",
			input: "https://github.com/owner/repo/actions/runs/123456?pr=999",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo/actions/runs/123456?pr=999"),
				Number: intPtr(999),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "PR URL with pr query parameter and other parameters",
			input: "https://github.com/owner/repo/actions/runs/123?tab=artifacts&pr=42",
			want: &PullRequestURL{
				Url:    mustParseURL("https://github.com/owner/repo/actions/runs/123?tab=artifacts&pr=42"),
				Number: intPtr(42),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:    "invalid URL format",
			input:   "https://github.com/owner/repo/issues/123",
			wantErr: true,
		},
		{
			name:    "URL without PR number",
			input:   "https://github.com/owner/repo/pull/",
			wantErr: true,
		},
		{
			name:    "URL with invalid PR number (zero)",
			input:   "https://github.com/owner/repo/pull/0",
			wantErr: true,
		},
		{
			name:    "URL with invalid PR number (negative)",
			input:   "https://github.com/owner/repo/pull/-1",
			wantErr: true,
		},
		{
			name:    "URL with non-numeric PR number",
			input:   "https://github.com/owner/repo/pull/abc",
			wantErr: true,
		},
		{
			name:    "URL with invalid pr query parameter (zero)",
			input:   "https://github.com/owner/repo/actions/runs/123?pr=0",
			wantErr: true,
		},
		{
			name:    "URL with invalid pr query parameter (negative)",
			input:   "https://github.com/owner/repo/actions/runs/123?pr=-1",
			wantErr: true,
		},
		{
			name:    "URL with non-numeric pr query parameter",
			input:   "https://github.com/owner/repo/actions/runs/123?pr=abc",
			wantErr: true,
		},
		{
			name:    "URL with too short path",
			input:   "https://github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "URL with only owner",
			input:   "https://github.com/owner",
			wantErr: true,
		},
		{
			name:    "malformed URL",
			input:   "https://not a valid url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePullRequestURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePullRequestURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if (got == nil) != (tt.want == nil) {
				t.Errorf("ParsePullRequestURL() = %v, want %v", got, tt.want)
				return
			}

			if got == nil {
				return
			}

			// Compare URL
			if got.Url.String() != tt.want.Url.String() {
				t.Errorf("ParsePullRequestURL() Url = %v, want %v", got.Url, tt.want.Url)
			}

			// Compare Number
			if !compareIntPtr(got.Number, tt.want.Number) {
				t.Errorf("ParsePullRequestURL() Number = %v, want %v", ptrValue(got.Number), ptrValue(tt.want.Number))
			}

			// Compare Repo
			if !compareRepo(got.Repo, tt.want.Repo) {
				t.Errorf("ParsePullRequestURL() Repo = %v, want %v", got.Repo, tt.want.Repo)
			}
		})
	}
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

func compareIntPtr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func compareRepo(a, b *repository.Repository) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Host == b.Host && a.Owner == b.Owner && a.Name == b.Name
}

func ptrValue(v interface{}) interface{} {
	if v == nil {
		return "<nil>"
	}
	switch val := v.(type) {
	case *int:
		if val == nil {
			return "<nil>"
		}
		return *val
	case *string:
		if val == nil {
			return "<nil>"
		}
		return *val
	default:
		return v
	}
}
