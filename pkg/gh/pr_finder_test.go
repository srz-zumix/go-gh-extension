package gh

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
)

func TestParsePRIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *PRIdentifier
		wantErr bool
	}{
		{
			name:  "empty string",
			input: "",
			want:  &PRIdentifier{},
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  &PRIdentifier{},
		},
		{
			name:  "PR number",
			input: "123",
			want: &PRIdentifier{
				Number: intPtr(123),
			},
		},
		{
			name:  "PR number with # prefix",
			input: "#456",
			want: &PRIdentifier{
				Number: intPtr(456),
			},
		},
		{
			name:  "PR number with # prefix and whitespace",
			input: " #789 ",
			want: &PRIdentifier{
				Number: intPtr(789),
			},
		},
		{
			name:  "zero is not valid PR number",
			input: "0",
			want: &PRIdentifier{
				Head: stringPtr("0"),
			},
		},
		{
			name:  "negative number treated as branch",
			input: "-1",
			want: &PRIdentifier{
				Head: stringPtr("-1"),
			},
		},
		{
			name:  "HTTPS URL",
			input: "https://github.com/owner/repo/pull/123",
			want: &PRIdentifier{
				Number: intPtr(123),
				URL:    stringPtr("https://github.com/owner/repo/pull/123"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "HTTP URL",
			input: "http://github.com/owner/repo/pull/456",
			want: &PRIdentifier{
				Number: intPtr(456),
				URL:    stringPtr("http://github.com/owner/repo/pull/456"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "URL with trailing slash",
			input: "https://github.com/owner/repo/pull/789/",
			want: &PRIdentifier{
				Number: intPtr(789),
				URL:    stringPtr("https://github.com/owner/repo/pull/789/"),
				Repo: &repository.Repository{
					Host:  "github.com",
					Owner: "owner",
					Name:  "repo",
				},
			},
		},
		{
			name:  "URL with additional path",
			input: "https://github.com/owner/repo/pull/100/files",
			want: &PRIdentifier{
				Number: intPtr(100),
				URL:    stringPtr("https://github.com/owner/repo/pull/100/files"),
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
			name:    "URL with invalid PR number",
			input:   "https://github.com/owner/repo/pull/abc",
			wantErr: true,
		},
		{
			name:  "branch name",
			input: "feature/add-feature",
			want: &PRIdentifier{
				Head: stringPtr("feature/add-feature"),
			},
		},
		{
			name:  "branch name with special characters",
			input: "fix/bug-#123",
			want: &PRIdentifier{
				Head: stringPtr("fix/bug-#123"),
			},
		},
		{
			name:  "branch name starting with number",
			input: "123-feature",
			want: &PRIdentifier{
				Head: stringPtr("123-feature"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePRIdentifier(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePRIdentifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Compare Number
			if !compareIntPtr(got.Number, tt.want.Number) {
				t.Errorf("ParsePRIdentifier() Number = %v, want %v", ptrToString(got.Number), ptrToString(tt.want.Number))
			}

			// Compare Head
			if !compareStringPtr(got.Head, tt.want.Head) {
				t.Errorf("ParsePRIdentifier() Head = %v, want %v", ptrToString(got.Head), ptrToString(tt.want.Head))
			}

			// Compare URL
			if !compareStringPtr(got.URL, tt.want.URL) {
				t.Errorf("ParsePRIdentifier() URL = %v, want %v", ptrToString(got.URL), ptrToString(tt.want.URL))
			}

			// Compare Repo
			if !compareRepo(got.Repo, tt.want.Repo) {
				t.Errorf("ParsePRIdentifier() Repo = %v, want %v", got.Repo, tt.want.Repo)
			}
		})
	}
}

func TestPRIdentifierString(t *testing.T) {
	tests := []struct {
		name string
		pri  *PRIdentifier
		want string
	}{
		{
			name: "empty identifier",
			pri:  &PRIdentifier{},
			want: "<empty>",
		},
		{
			name: "number only",
			pri: &PRIdentifier{
				Number: intPtr(123),
			},
			want: "#123",
		},
		{
			name: "number with repo",
			pri: &PRIdentifier{
				Number: intPtr(123),
				Repo: &repository.Repository{
					Owner: "owner",
					Name:  "repo",
				},
			},
			want: "owner/repo #123",
		},
		{
			name: "URL only",
			pri: &PRIdentifier{
				URL: stringPtr("https://github.com/owner/repo/pull/123"),
			},
			want: "https://github.com/owner/repo/pull/123",
		},
		{
			name: "head only",
			pri: &PRIdentifier{
				Head: stringPtr("feature-branch"),
			},
			want: "feature-branch",
		},
		{
			name: "head with repo",
			pri: &PRIdentifier{
				Head: stringPtr("feature-branch"),
				Repo: &repository.Repository{
					Owner: "owner",
					Name:  "repo",
				},
			},
			want: "owner/repo feature-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pri.String()
			if got != tt.want {
				t.Errorf("PRIdentifier.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
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

func compareStringPtr(a, b *string) bool {
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

func ptrToString(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	switch val := v.(type) {
	case *int:
		if val == nil {
			return "<nil>"
		}
		return string(rune(*val))
	case *string:
		if val == nil {
			return "<nil>"
		}
		return *val
	default:
		return "<unknown>"
	}
}
