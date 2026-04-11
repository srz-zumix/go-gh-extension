package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListTagProtection retrieves all tag protection settings for a repository.
func (g *GitHubClient) ListTagProtection(ctx context.Context, owner, repo string) ([]*github.TagProtection, error) {
	tagProtections, _, err := g.client.Repositories.ListTagProtection(ctx, owner, repo) //nolint:staticcheck // SA1019: deprecated but no replacement available in go-github yet
	if err != nil {
		return nil, err
	}
	return tagProtections, nil
}

// DeleteTagProtection removes a tag protection setting by its ID.
func (g *GitHubClient) DeleteTagProtection(ctx context.Context, owner, repo string, tagProtectionID int64) error {
	_, err := g.client.Repositories.DeleteTagProtection(ctx, owner, repo, tagProtectionID) //nolint:staticcheck // SA1019: deprecated but no replacement available in go-github yet
	return err
}
