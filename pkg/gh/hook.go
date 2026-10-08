package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

func ListOrgHooks(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Hook, error) {
	return g.ListOrgHooks(ctx, repo.Owner)
}

func ListRepoHooks(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Hook, error) {
	return g.ListRepoHooks(ctx, repo.Owner, repo.Name)
}
