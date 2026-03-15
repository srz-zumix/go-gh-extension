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

// ListProtectedBranches retrieves all protected branches for a repository.
func (g *GitHubClient) ListProtectedBranches(ctx context.Context, owner, repo string) ([]*github.Branch, error) {
	protected := true
	opt := &github.BranchListOptions{
		Protected: &protected,
		ListOptions: github.ListOptions{
			PerPage: defaultPerPage,
		},
	}
	var allBranches []*github.Branch
	for {
		branches, resp, err := g.client.Repositories.ListBranches(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allBranches, nil
}
