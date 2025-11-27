package parser

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		initial   repository.Repository
		expectErr bool
		expected  repository.Repository
	}{
		{
			name:      "Valid input",
			input:     "github.com/owner/repo",
			initial:   repository.Repository{},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "owner", Name: "repo"},
		},
		{
			name:      "Conflicting host",
			input:     "gitlab.com/owner/repo",
			initial:   repository.Repository{Host: "github.com"},
			expectErr: true,
		},
		{
			name:      "Empty input",
			input:     "",
			initial:   repository.Repository{},
			expectErr: false,
			expected:  repository.Repository{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := RepositoryInput(tt.input)
			r := tt.initial
			err := op(&r)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, r)
			}
		})
	}
}

func TestRepositoryOwner(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		initial   repository.Repository
		expectErr bool
		expected  repository.Repository
	}{
		{
			name:      "Valid owner",
			input:     "owner",
			initial:   repository.Repository{},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "owner"},
		},
		{
			name:      "Conflicting owner",
			input:     "new-owner",
			initial:   repository.Repository{Host: "github.com", Owner: "owner"},
			expectErr: true,
		},
		{
			name:      "Empty input",
			input:     "",
			initial:   repository.Repository{Host: "github.com", Owner: "owner"},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "owner"},
		},
		{
			name:      "Owner with slash",
			input:     "owner/repo",
			initial:   repository.Repository{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := RepositoryOwner(tt.input)
			r := tt.initial
			err := op(&r)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, r)
			}
		})
	}
}

func TestRepository(t *testing.T) {
	tests := []struct {
		name      string
		options   []RepositoryOption
		env       map[string]string
		expectErr bool
		expected  repository.Repository
	}{
		{
			name:      "Valid options",
			options:   []RepositoryOption{RepositoryInput("github.com/owner/repo"), RepositoryOwner("owner")},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "owner", Name: "repo"},
		},
		{
			name:      "Valid url options",
			options:   []RepositoryOption{RepositoryInput("https://github.com/owner/repo"), RepositoryOwner("owner")},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "owner", Name: "repo"},
		},
		{
			name:      "Conflicting options",
			options:   []RepositoryOption{RepositoryInput("github.com/owner/repo"), RepositoryOwner("new-owner")},
			expectErr: true,
		},
		{
			name:    "Actions Context fallback",
			options: []RepositoryOption{},
			env: map[string]string{
				"GITHUB_REPOSITORY": "owner/repo",
				"GITHUB_SERVER_URL": "https://hoge.com",
				"GIT_DIR":           "/path/to/git/dir",
			},
			expectErr: false,
			expected:  repository.Repository{Host: "hoge.com", Owner: "owner", Name: "repo"},
		},
		{
			name:      "Empty options",
			options:   []RepositoryOption{},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "srz-zumix", Name: "go-gh-extension"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			r, err := Repository(tt.options...)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, r)
			}
		})
	}
}
