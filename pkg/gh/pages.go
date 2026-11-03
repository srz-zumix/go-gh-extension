package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v93/github"
)

// GetPagesSite retrieves a repository's GitHub Pages configuration.
func GetPagesSite(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.Pages, error) {
	return g.GetPagesSite(ctx, repo.Owner, repo.Name)
}

// CreatePagesSite enables GitHub Pages for a repository.
func CreatePagesSite(ctx context.Context, g *GitHubClient, repo repository.Repository, pages *github.Pages) (*github.Pages, error) {
	return g.CreatePagesSite(ctx, repo.Owner, repo.Name, pages)
}

// UpdatePagesSite updates Pages settings without modifying a custom domain.
func UpdatePagesSite(ctx context.Context, g *GitHubClient, repo repository.Repository, update *github.PagesUpdateWithoutCNAME) error {
	return g.UpdatePagesSite(ctx, repo.Owner, repo.Name, update)
}
