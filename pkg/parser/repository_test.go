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

func TestRepositoryInOwner(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		owner     repository.Repository
		expected  repository.Repository
		expectErr bool
	}{
		{
			name:     "repository name",
			input:    "example-repo",
			owner:    repository.Repository{Host: "github.com", Owner: "example-org"},
			expected: repository.Repository{Host: "github.com", Owner: "example-org", Name: "example-repo"},
		},
		{
			name:     "owner and repository",
			input:    "example-org/example-repo",
			owner:    repository.Repository{Host: "github.com", Owner: "example-org"},
			expected: repository.Repository{Host: "github.com", Owner: "example-org", Name: "example-repo"},
		},
		{
			name:      "different owner",
			input:     "other-org/example-repo",
			owner:     repository.Repository{Host: "github.com", Owner: "example-org"},
			expectErr: true,
		},
		{
			name:      "different host",
			input:     "ghe.example.com/example-org/example-repo",
			owner:     repository.Repository{Host: "github.com", Owner: "example-org"},
			expectErr: true,
		},
		{
			name:     "owner without host",
			input:    "example-org/example-repo",
			owner:    repository.Repository{Owner: "example-org"},
			expected: repository.Repository{Host: "github.com", Owner: "example-org", Name: "example-repo"},
		},
		{
			name:     "matching enterprise host",
			input:    "ghe.example.com/example-org/example-repo",
			owner:    repository.Repository{Host: "ghe.example.com", Owner: "example-org"},
			expected: repository.Repository{Host: "ghe.example.com", Owner: "example-org", Name: "example-repo"},
		},
		{
			name:     "owner and host differ only in case",
			input:    "GitHub.com/Example-Org/example-repo",
			owner:    repository.Repository{Host: "github.com", Owner: "example-org"},
			expected: repository.Repository{Host: "GitHub.com", Owner: "Example-Org", Name: "example-repo"},
		},
		{
			name:      "empty repository",
			input:     "",
			owner:     repository.Repository{Host: "github.com", Owner: "example-org"},
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RepositoryInOwner(tt.input, tt.owner)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
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

func TestRepositoryOwnerOrRepo(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		initial   repository.Repository
		expectErr bool
		expected  repository.Repository
	}{
		{
			name:      "Owner only",
			input:     "owner",
			initial:   repository.Repository{},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "owner", Name: ""},
		},
		{
			name:      "Owner and repo",
			input:     "owner/repo",
			initial:   repository.Repository{},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "owner", Name: "repo"},
		},
		{
			name:      "Host, owner and repo",
			input:     "github.com/owner/repo",
			initial:   repository.Repository{},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "owner", Name: "repo"},
		},
		{
			name:      "GHE host, owner and repo",
			input:     "ghe.example.com/owner/repo",
			initial:   repository.Repository{},
			expectErr: false,
			expected:  repository.Repository{Host: "ghe.example.com", Owner: "owner", Name: "repo"},
		},
		{
			name:      "GHE host and owner only",
			input:     "ghe.example.com/owner",
			initial:   repository.Repository{},
			expectErr: false,
			expected:  repository.Repository{Host: "ghe.example.com", Owner: "owner", Name: ""},
		},
		{
			name:      "Empty input",
			input:     "",
			initial:   repository.Repository{},
			expectErr: false,
			expected:  repository.Repository{},
		},
		{
			name:      "Conflicting owner",
			input:     "other-owner",
			initial:   repository.Repository{Host: "github.com", Owner: "owner"},
			expectErr: true,
		},
		{
			name:      "Conflicting host",
			input:     "gitlab.com/owner/repo",
			initial:   repository.Repository{Host: "github.com"},
			expectErr: true,
		},
		{
			name:      "Conflicting repo name",
			input:     "owner/other-repo",
			initial:   repository.Repository{Host: "github.com", Owner: "owner", Name: "repo"},
			expectErr: true,
		},
		{
			name:      "Same owner no conflict",
			input:     "owner",
			initial:   repository.Repository{Host: "github.com", Owner: "owner"},
			expectErr: false,
			expected:  repository.Repository{Host: "github.com", Owner: "owner", Name: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := RepositoryOwnerOrRepo(tt.input)
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

func TestRepositoryOptionsIgnoreCaseWhenCheckingConflicts(t *testing.T) {
	got, err := Repository(RepositoryInput("github.com/Example-Org/Example-Repo"), RepositoryOwner("example-org"))
	assert.NoError(t, err)
	assert.Equal(t, "example-org", got.Owner)

	_, err = Repository(RepositoryInput("github.com/example-org/example-repo"), RepositoryOwnerOrRepo("GitHub.com/Example-Org/Example-Repo"))
	assert.NoError(t, err)

	_, err = Repository(RepositoryInput("github.com/example-org/example-repo"), RepositoryOwner("other-org"))
	assert.Error(t, err)
}
