package client

import (
	"context"

	"github.com/google/go-github/v79/github"
)

// GetBranchProtection retrieves the branch protection settings for the given branch.
func (g *GitHubClient) GetBranchProtection(ctx context.Context, owner, repo, branch string) (*github.Protection, error) {
	protection, _, err := g.client.Repositories.GetBranchProtection(ctx, owner, repo, branch)
	if err != nil {
		return nil, err
	}
	return protection, nil
}

// RemoveBranchProtection removes the protection of a given branch.
func (g *GitHubClient) RemoveBranchProtection(ctx context.Context, owner, repo, branch string) error {
	_, err := g.client.Repositories.RemoveBranchProtection(ctx, owner, repo, branch)
	return err
}

// ListProtectedBranches retrieves all protected branches for a repository.
func (g *GitHubClient) ListProtectedBranches(ctx context.Context, owner, repo string) ([]*github.Branch, error) {
	return g.ListBranches(ctx, owner, repo, github.Ptr(true))
}
