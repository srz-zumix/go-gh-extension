package gh

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

func TestGetRepositoryFromGitHubRepository(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expected  repository.Repository
		wantError bool
	}{
		{
			name:  "pointer repository",
			input: &github.Repository{Owner: &github.User{Login: github.Ptr("octocat")}, Name: github.Ptr("hello-world")},
			expected: repository.Repository{
				Owner: "octocat",
				Name:  "hello-world",
			},
		},
		{
			name:  "value repository",
			input: github.Repository{Owner: &github.User{Login: github.Ptr("octocat")}, Name: github.Ptr("hello-world")},
			expected: repository.Repository{
				Owner: "octocat",
				Name:  "hello-world",
			},
		},
		{
			name:      "nil repository",
			input:     (*github.Repository)(nil),
			wantError: true,
		},
		{
			name:      "nil owner",
			input:     &github.Repository{Name: github.Ptr("hello-world")},
			wantError: true,
		},
		{
			name:      "empty owner",
			input:     &github.Repository{Owner: &github.User{Login: github.Ptr("")}, Name: github.Ptr("hello-world")},
			wantError: true,
		},
		{
			name:      "empty name",
			input:     &github.Repository{Owner: &github.User{Login: github.Ptr("octocat")}, Name: github.Ptr("")},
			wantError: true,
		},
		{
			name:      "unsupported type",
			input:     100,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRepositoryFromGitHubRepository(tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Owner != tt.expected.Owner || result.Name != tt.expected.Name {
				t.Fatalf("unexpected repository: got %+v, want %+v", result, tt.expected)
			}
		})
	}
}
