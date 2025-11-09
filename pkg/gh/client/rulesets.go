package client

import (
	"context"

	"github.com/google/go-github/v73/github"
)

// ListRepositoryRulesets retrieves all rulesets for a specific repository
func (g *GitHubClient) ListRepositoryRulesets(ctx context.Context, owner string, repo string, includesParents bool) ([]*github.RepositoryRuleset, error) {
	allRulesets := []*github.RepositoryRuleset{}
	opt := &github.RepositoryListRulesetsOptions{
		IncludesParents: github.Ptr(includesParents),
		ListOptions: github.ListOptions{
			PerPage: defaultPerPage,
		},
	}

	for {
		rulesets, resp, err := g.client.Repositories.GetAllRulesets(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allRulesets = append(allRulesets, rulesets...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRulesets, nil
}

// GetRepositoryRuleset retrieves a single ruleset for a specific repository by ruleset ID
func (g *GitHubClient) GetRepositoryRuleset(ctx context.Context, owner string, repo string, rulesetID int64, includesParents bool) (*github.RepositoryRuleset, error) {
	ruleset, _, err := g.client.Repositories.GetRuleset(ctx, owner, repo, rulesetID, includesParents)
	if err != nil {
		return nil, err
	}

	return ruleset, nil
}

// CreateRepositoryRuleset creates a new ruleset for a specific repository
func (g *GitHubClient) CreateRepositoryRuleset(ctx context.Context, owner string, repo string, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	createdRuleset, _, err := g.client.Repositories.CreateRuleset(ctx, owner, repo, *ruleset)
	if err != nil {
		return nil, err
	}

	return createdRuleset, nil
}

// UpdateRepositoryRuleset updates an existing ruleset for a specific repository
func (g *GitHubClient) UpdateRepositoryRuleset(ctx context.Context, owner string, repo string, rulesetID int64, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	updatedRuleset, _, err := g.client.Repositories.UpdateRuleset(ctx, owner, repo, rulesetID, *ruleset)
	if err != nil {
		return nil, err
	}

	return updatedRuleset, nil
}

// DeleteRepositoryRuleset deletes a single ruleset for a specific repository by ruleset ID
func (g *GitHubClient) DeleteRepositoryRuleset(ctx context.Context, owner string, repo string, rulesetID int64) error {
	_, err := g.client.Repositories.DeleteRuleset(ctx, owner, repo, rulesetID)
	if err != nil {
		return err
	}

	return nil
}
