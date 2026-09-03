package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

// ListCodespacesRepoSecrets lists all Codespaces secrets in a repository (wrapper).
func ListCodespacesRepoSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	return g.ListCodespacesRepoSecrets(ctx, repo.Owner, repo.Name)
}

// ListCodespacesOrgSecrets lists all Codespaces secrets in an organization (wrapper).
func ListCodespacesOrgSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	return g.ListCodespacesOrgSecrets(ctx, repo.Owner)
}

// ListCodespacesUserSecrets lists all Codespaces secrets of the authenticated user (wrapper).
func ListCodespacesUserSecrets(ctx context.Context, g *GitHubClient) ([]*github.Secret, error) {
	return g.ListCodespacesUserSecrets(ctx)
}
