package gh

// Deploy key wrapper functions for the GitHub API.

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// ListDeployKeys retrieves all deploy keys for a repository.
func ListDeployKeys(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Key, error) {
	return g.ListDeployKeys(ctx, repo.Owner, repo.Name)
}

// GetDeployKey retrieves a single deploy key by ID.
func GetDeployKey(ctx context.Context, g *GitHubClient, repo repository.Repository, id int64) (*github.Key, error) {
	return g.GetDeployKey(ctx, repo.Owner, repo.Name, id)
}

// CreateDeployKey adds a deploy key to a repository.
func CreateDeployKey(ctx context.Context, g *GitHubClient, repo repository.Repository, title, key string, readOnly bool) (*github.Key, error) {
	input := &github.Key{
		Title:    &title,
		Key:      &key,
		ReadOnly: &readOnly,
	}
	return g.CreateDeployKey(ctx, repo.Owner, repo.Name, input)
}

// DeleteDeployKey removes a deploy key from a repository by ID.
func DeleteDeployKey(ctx context.Context, g *GitHubClient, repo repository.Repository, id int64) error {
	return g.DeleteDeployKey(ctx, repo.Owner, repo.Name, id)
}
