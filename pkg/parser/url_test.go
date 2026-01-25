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

func ptrValue(v any) any {
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

func TestParseIssueURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *IssueURL
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
			name:  "valid HTTPS issue URL",
			input: "https://github.com/owner/repo/issues/123",
			want: &IssueURL{
				Url:    mustParseURL("https://github.com/owner/repo/issues/123"),
				Number: intPtr(123),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "valid HTTP issue URL",
			input: "http://github.com/owner/repo/issues/456",
			want: &IssueURL{
				Url:    mustParseURL("http://github.com/owner/repo/issues/456"),
				Number: intPtr(456),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "valid HTTPS pull request URL",
			input: "https://github.com/owner/repo/pull/789",
			want: &IssueURL{
				Url:    mustParseURL("https://github.com/owner/repo/pull/789"),
				Number: intPtr(789),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "valid HTTP pull request URL",
			input: "http://github.com/owner/repo/pull/101",
			want: &IssueURL{
				Url:    mustParseURL("http://github.com/owner/repo/pull/101"),
				Number: intPtr(101),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "issue URL with trailing slash",
			input: "https://github.com/owner/repo/issues/200/",
			want: &IssueURL{
				Url:    mustParseURL("https://github.com/owner/repo/issues/200/"),
				Number: intPtr(200),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "issue URL with query parameters",
			input: "https://github.com/owner/repo/issues/300?tab=timeline",
			want: &IssueURL{
				Url:    mustParseURL("https://github.com/owner/repo/issues/300?tab=timeline"),
				Number: intPtr(300),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "issue URL with fragment",
			input: "https://github.com/owner/repo/issues/400#issuecomment-123",
			want: &IssueURL{
				Url:    mustParseURL("https://github.com/owner/repo/issues/400#issuecomment-123"),
				Number: intPtr(400),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "issue URL with whitespace",
			input: "  https://github.com/owner/repo/issues/500  ",
			want: &IssueURL{
				Url:    mustParseURL("https://github.com/owner/repo/issues/500"),
				Number: intPtr(500),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "GitHub Enterprise issue URL",
			input: "https://github.example.com/owner/repo/issues/600",
			want: &IssueURL{
				Url:    mustParseURL("https://github.example.com/owner/repo/issues/600"),
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
			input: "https://github.com/my-org/repo/issues/700",
			want: &IssueURL{
				Url:    mustParseURL("https://github.com/my-org/repo/issues/700"),
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
			input: "https://github.com/owner/repo.name/issues/800",
			want: &IssueURL{
				Url:    mustParseURL("https://github.com/owner/repo.name/issues/800"),
				Number: intPtr(800),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo.name",
				},
			},
		},
		{
			name:    "URL with invalid issue number (zero)",
			input:   "https://github.com/owner/repo/issues/0",
			wantErr: true,
		},
		{
			name:    "URL with invalid issue number (negative)",
			input:   "https://github.com/owner/repo/issues/-1",
			wantErr: true,
		},
		{
			name:    "URL with non-numeric issue number",
			input:   "https://github.com/owner/repo/issues/abc",
			wantErr: true,
		},
		{
			name:    "URL without issue number",
			input:   "https://github.com/owner/repo/issues/",
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
			name:    "not an issue or pull request URL",
			input:   "https://github.com/owner/repo/actions/runs/123",
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
			got, err := ParseIssueURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIssueURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if (got == nil) != (tt.want == nil) {
				t.Errorf("ParseIssueURL() = %v, want %v", got, tt.want)
				return
			}

			if got == nil {
				return
			}

			// Compare URL
			if got.Url.String() != tt.want.Url.String() {
				t.Errorf("ParseIssueURL() Url = %v, want %v", got.Url, tt.want.Url)
			}

			// Compare Number
			if !compareIntPtr(got.Number, tt.want.Number) {
				t.Errorf("ParseIssueURL() Number = %v, want %v", ptrValue(got.Number), ptrValue(tt.want.Number))
			}

			// Compare Repo
			if !compareRepo(got.Repo, tt.want.Repo) {
				t.Errorf("ParseIssueURL() Repo = %v, want %v", got.Repo, tt.want.Repo)
			}
		})
	}
}

func TestParseDiscussionURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *DiscussionURL
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
			name:  "valid HTTPS discussion URL",
			input: "https://github.com/owner/repo/discussions/123",
			want: &DiscussionURL{
				Url:    mustParseURL("https://github.com/owner/repo/discussions/123"),
				Number: intPtr(123),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "valid HTTP discussion URL",
			input: "http://github.com/owner/repo/discussions/456",
			want: &DiscussionURL{
				Url:    mustParseURL("http://github.com/owner/repo/discussions/456"),
				Number: intPtr(456),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "discussion URL with trailing slash",
			input: "https://github.com/owner/repo/discussions/789/",
			want: &DiscussionURL{
				Url:    mustParseURL("https://github.com/owner/repo/discussions/789/"),
				Number: intPtr(789),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "discussion URL with query parameters",
			input: "https://github.com/owner/repo/discussions/300?sort=top",
			want: &DiscussionURL{
				Url:    mustParseURL("https://github.com/owner/repo/discussions/300?sort=top"),
				Number: intPtr(300),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "discussion URL with fragment",
			input: "https://github.com/owner/repo/discussions/400#discussioncomment-123",
			want: &DiscussionURL{
				Url:    mustParseURL("https://github.com/owner/repo/discussions/400#discussioncomment-123"),
				Number: intPtr(400),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "discussion URL with whitespace",
			input: "  https://github.com/owner/repo/discussions/500  ",
			want: &DiscussionURL{
				Url:    mustParseURL("https://github.com/owner/repo/discussions/500"),
				Number: intPtr(500),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "GitHub Enterprise discussion URL",
			input: "https://github.example.com/owner/repo/discussions/600",
			want: &DiscussionURL{
				Url:    mustParseURL("https://github.example.com/owner/repo/discussions/600"),
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
			input: "https://github.com/my-org/repo/discussions/700",
			want: &DiscussionURL{
				Url:    mustParseURL("https://github.com/my-org/repo/discussions/700"),
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
			input: "https://github.com/owner/repo.name/discussions/800",
			want: &DiscussionURL{
				Url:    mustParseURL("https://github.com/owner/repo.name/discussions/800"),
				Number: intPtr(800),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo.name",
				},
			},
		},
		{
			name:    "URL with invalid discussion number (zero)",
			input:   "https://github.com/owner/repo/discussions/0",
			wantErr: true,
		},
		{
			name:    "URL with invalid discussion number (negative)",
			input:   "https://github.com/owner/repo/discussions/-1",
			wantErr: true,
		},
		{
			name:    "URL with non-numeric discussion number",
			input:   "https://github.com/owner/repo/discussions/abc",
			wantErr: true,
		},
		{
			name:    "URL without discussion number",
			input:   "https://github.com/owner/repo/discussions/",
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
			name:    "not a discussion URL (issues)",
			input:   "https://github.com/owner/repo/issues/123",
			wantErr: true,
		},
		{
			name:    "not a discussion URL (pull)",
			input:   "https://github.com/owner/repo/pull/123",
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
			got, err := ParseDiscussionURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDiscussionURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if (got == nil) != (tt.want == nil) {
				t.Errorf("ParseDiscussionURL() = %v, want %v", got, tt.want)
				return
			}

			if got == nil {
				return
			}

			// Compare URL
			if got.Url.String() != tt.want.Url.String() {
				t.Errorf("ParseDiscussionURL() Url = %v, want %v", got.Url, tt.want.Url)
			}

			// Compare Number
			if !compareIntPtr(got.Number, tt.want.Number) {
				t.Errorf("ParseDiscussionURL() Number = %v, want %v", ptrValue(got.Number), ptrValue(tt.want.Number))
			}

			// Compare Repo
			if !compareRepo(got.Repo, tt.want.Repo) {
				t.Errorf("ParseDiscussionURL() Repo = %v, want %v", got.Repo, tt.want.Repo)
			}
		})
	}
}

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *GitHubURL
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
			name:  "owner/repo format (not a URL)",
			input: "owner/repo",
			want:  nil,
		},
		{
			name:  "valid HTTPS repository URL",
			input: "https://github.com/owner/repo",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo"},
			},
		},
		{
			name:  "valid HTTP repository URL",
			input: "http://github.com/owner/repo",
			want: &GitHubURL{
				Url: mustParseURL("http://github.com/owner/repo"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo"},
			},
		},
		{
			name:  "repository URL with trailing slash",
			input: "https://github.com/owner/repo/",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo/"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo"},
			},
		},
		{
			name:  "pull request URL",
			input: "https://github.com/owner/repo/pull/123",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo/pull/123"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo", "pull", "123"},
			},
		},
		{
			name:  "issue URL",
			input: "https://github.com/owner/repo/issues/456",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo/issues/456"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo", "issues", "456"},
			},
		},
		{
			name:  "discussion URL",
			input: "https://github.com/owner/repo/discussions/789",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo/discussions/789"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo", "discussions", "789"},
			},
		},
		{
			name:  "blob URL with file path",
			input: "https://github.com/owner/repo/blob/main/src/file.go",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo/blob/main/src/file.go"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo", "blob", "main", "src", "file.go"},
			},
		},
		{
			name:  "actions run URL",
			input: "https://github.com/owner/repo/actions/runs/123456",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo/actions/runs/123456"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo", "actions", "runs", "123456"},
			},
		},
		{
			name:  "URL with query parameters",
			input: "https://github.com/owner/repo/pull/100?tab=files",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo/pull/100?tab=files"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo", "pull", "100"},
			},
		},
		{
			name:  "URL with fragment",
			input: "https://github.com/owner/repo/issues/200#issuecomment-123",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo/issues/200#issuecomment-123"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo", "issues", "200"},
			},
		},
		{
			name:  "URL with whitespace",
			input: "  https://github.com/owner/repo  ",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo"},
			},
		},
		{
			name:  "GitHub Enterprise URL",
			input: "https://github.example.com/owner/repo",
			want: &GitHubURL{
				Url: mustParseURL("https://github.example.com/owner/repo"),
				Repo: &repository.Repository{
					Host:  "github.example.com",
					Owner: "owner",
					Name:  "repo",
				},
				PathParts: []string{"owner", "repo"},
			},
		},
		{
			name:  "owner name with hyphen",
			input: "https://github.com/my-org/repo",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/my-org/repo"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "my-org",
					Name:  "repo",
				},
				PathParts: []string{"my-org", "repo"},
			},
		},
		{
			name:  "repo name with dots",
			input: "https://github.com/owner/repo.name",
			want: &GitHubURL{
				Url: mustParseURL("https://github.com/owner/repo.name"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo.name",
				},
				PathParts: []string{"owner", "repo.name"},
			},
		},
		{
			name:    "URL with only owner",
			input:   "https://github.com/owner",
			wantErr: true,
		},
		{
			name:    "URL with empty path",
			input:   "https://github.com/",
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
			got, err := ParseGitHubURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitHubURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if (got == nil) != (tt.want == nil) {
				t.Errorf("ParseGitHubURL() = %v, want %v", got, tt.want)
				return
			}

			if got == nil {
				return
			}

			// Compare URL
			if got.Url.String() != tt.want.Url.String() {
				t.Errorf("ParseGitHubURL() Url = %v, want %v", got.Url, tt.want.Url)
			}

			// Compare Repo
			if !compareRepo(got.Repo, tt.want.Repo) {
				t.Errorf("ParseGitHubURL() Repo = %v, want %v", got.Repo, tt.want.Repo)
			}

			// Compare PathParts
			if !compareStringSlice(got.PathParts, tt.want.PathParts) {
				t.Errorf("ParseGitHubURL() PathParts = %v, want %v", got.PathParts, tt.want.PathParts)
			}
		})
	}
}

func compareStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestParseNumberFromPath(t *testing.T) {
	tests := []struct {
		name      string
		pathParts []string
		index     int
		wantNum   int
		wantOk    bool
	}{
		{
			name:      "valid positive number",
			pathParts: []string{"owner", "repo", "pull", "123"},
			index:     3,
			wantNum:   123,
			wantOk:    true,
		},
		{
			name:      "valid single digit number",
			pathParts: []string{"owner", "repo", "issues", "1"},
			index:     3,
			wantNum:   1,
			wantOk:    true,
		},
		{
			name:      "valid large number",
			pathParts: []string{"owner", "repo", "discussions", "999999"},
			index:     3,
			wantNum:   999999,
			wantOk:    true,
		},
		{
			name:      "index out of bounds",
			pathParts: []string{"owner", "repo"},
			index:     3,
			wantNum:   0,
			wantOk:    false,
		},
		{
			name:      "index at boundary (equal to length)",
			pathParts: []string{"owner", "repo", "pull"},
			index:     3,
			wantNum:   0,
			wantOk:    false,
		},
		{
			name:      "zero number",
			pathParts: []string{"owner", "repo", "pull", "0"},
			index:     3,
			wantNum:   0,
			wantOk:    false,
		},
		{
			name:      "negative number",
			pathParts: []string{"owner", "repo", "pull", "-1"},
			index:     3,
			wantNum:   0,
			wantOk:    false,
		},
		{
			name:      "non-numeric string",
			pathParts: []string{"owner", "repo", "pull", "abc"},
			index:     3,
			wantNum:   0,
			wantOk:    false,
		},
		{
			name:      "empty string at index",
			pathParts: []string{"owner", "repo", "pull", ""},
			index:     3,
			wantNum:   0,
			wantOk:    false,
		},
		{
			name:      "mixed alphanumeric string",
			pathParts: []string{"owner", "repo", "pull", "123abc"},
			index:     3,
			wantNum:   0,
			wantOk:    false,
		},
		{
			name:      "number with leading zeros",
			pathParts: []string{"owner", "repo", "pull", "000123"},
			index:     3,
			wantNum:   123,
			wantOk:    true,
		},
		{
			name:      "number with plus sign",
			pathParts: []string{"owner", "repo", "pull", "+123"},
			index:     3,
			wantNum:   123,
			wantOk:    true,
		},
		{
			name:      "empty pathParts slice",
			pathParts: []string{},
			index:     0,
			wantNum:   0,
			wantOk:    false,
		},
		{
			name:      "negative index",
			pathParts: []string{"owner", "repo", "pull", "456"},
			index:     -1,
			wantNum:   0,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNum, gotOk := parseNumberFromPath(tt.pathParts, tt.index)
			if gotNum != tt.wantNum {
				t.Errorf("parseNumberFromPath() gotNum = %v, want %v", gotNum, tt.wantNum)
			}
			if gotOk != tt.wantOk {
				t.Errorf("parseNumberFromPath() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestParseMilestoneURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *MilestoneURL
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
			name:  "valid HTTPS milestone URL",
			input: "https://github.com/owner/repo/milestone/123",
			want: &MilestoneURL{
				Url:    mustParseURL("https://github.com/owner/repo/milestone/123"),
				Number: intPtr(123),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "valid HTTP milestone URL",
			input: "http://github.com/owner/repo/milestone/456",
			want: &MilestoneURL{
				Url:    mustParseURL("http://github.com/owner/repo/milestone/456"),
				Number: intPtr(456),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "milestone URL with trailing slash",
			input: "https://github.com/owner/repo/milestone/789/",
			want: &MilestoneURL{
				Url:    mustParseURL("https://github.com/owner/repo/milestone/789/"),
				Number: intPtr(789),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "milestone URL with query parameters",
			input: "https://github.com/owner/repo/milestone/300?state=open",
			want: &MilestoneURL{
				Url:    mustParseURL("https://github.com/owner/repo/milestone/300?state=open"),
				Number: intPtr(300),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "milestone URL with fragment",
			input: "https://github.com/owner/repo/milestone/400#milestone-details",
			want: &MilestoneURL{
				Url:    mustParseURL("https://github.com/owner/repo/milestone/400#milestone-details"),
				Number: intPtr(400),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "milestone URL with whitespace",
			input: "  https://github.com/owner/repo/milestone/500  ",
			want: &MilestoneURL{
				Url:    mustParseURL("https://github.com/owner/repo/milestone/500"),
				Number: intPtr(500),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "GitHub Enterprise milestone URL",
			input: "https://github.example.com/owner/repo/milestone/600",
			want: &MilestoneURL{
				Url:    mustParseURL("https://github.example.com/owner/repo/milestone/600"),
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
			input: "https://github.com/my-org/repo/milestone/700",
			want: &MilestoneURL{
				Url:    mustParseURL("https://github.com/my-org/repo/milestone/700"),
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
			input: "https://github.com/owner/repo.name/milestone/800",
			want: &MilestoneURL{
				Url:    mustParseURL("https://github.com/owner/repo.name/milestone/800"),
				Number: intPtr(800),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo.name",
				},
			},
		},
		{
			name:    "URL with invalid milestone number (zero)",
			input:   "https://github.com/owner/repo/milestone/0",
			wantErr: true,
		},
		{
			name:    "URL with invalid milestone number (negative)",
			input:   "https://github.com/owner/repo/milestone/-1",
			wantErr: true,
		},
		{
			name:    "URL with non-numeric milestone number",
			input:   "https://github.com/owner/repo/milestone/abc",
			wantErr: true,
		},
		{
			name:    "URL without milestone number",
			input:   "https://github.com/owner/repo/milestone/",
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
			name:    "not a milestone URL (issues)",
			input:   "https://github.com/owner/repo/issues/123",
			wantErr: true,
		},
		{
			name:    "not a milestone URL (pull)",
			input:   "https://github.com/owner/repo/pull/123",
			wantErr: true,
		},
		{
			name:    "not a milestone URL (discussions)",
			input:   "https://github.com/owner/repo/discussions/123",
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
			got, err := ParseMilestoneURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMilestoneURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if (got == nil) != (tt.want == nil) {
				t.Errorf("ParseMilestoneURL() = %v, want %v", got, tt.want)
				return
			}

			if got == nil {
				return
			}

			// Compare URL
			if got.Url.String() != tt.want.Url.String() {
				t.Errorf("ParseMilestoneURL() Url = %v, want %v", got.Url, tt.want.Url)
			}

			// Compare Number
			if !compareIntPtr(got.Number, tt.want.Number) {
				t.Errorf("ParseMilestoneURL() Number = %v, want %v", ptrValue(got.Number), ptrValue(tt.want.Number))
			}

			// Compare Repo
			if !compareRepo(got.Repo, tt.want.Repo) {
				t.Errorf("ParseMilestoneURL() Repo = %v, want %v", got.Repo, tt.want.Repo)
			}
		})
	}
}
