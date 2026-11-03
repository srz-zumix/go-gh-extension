package client

import (
	"context"

	"github.com/google/go-github/v93/github"
)

// GetPagesSite retrieves a repository's GitHub Pages configuration.
func (g *GitHubClient) GetPagesSite(ctx context.Context, owner, repo string) (*github.Pages, error) {
	pages, _, err := g.client.Repositories.GetPagesInfo(ctx, owner, repo)
	return pages, err
}

// CreatePagesSite enables GitHub Pages for a repository.
func (g *GitHubClient) CreatePagesSite(ctx context.Context, owner, repo string, pages *github.Pages) (*github.Pages, error) {
	result, _, err := g.client.Repositories.EnablePages(ctx, owner, repo, pages)
	return result, err
}

// UpdatePagesSite updates Pages settings without modifying a custom domain.
func (g *GitHubClient) UpdatePagesSite(ctx context.Context, owner, repo string, update *github.PagesUpdateWithoutCNAME) error {
	// This method uses the same REST endpoint as UpdatePages but omits cname.
	_, err := g.client.Repositories.UpdatePagesGHES(ctx, owner, repo, update)
	return err
}
