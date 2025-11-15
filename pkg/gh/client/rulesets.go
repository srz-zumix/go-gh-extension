package client

import (
	"context"

	"github.com/google/go-github/v79/github"
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

// ListOrgRulesets retrieves all rulesets for a specific organization
func (g *GitHubClient) ListOrgRulesets(ctx context.Context, org string) ([]*github.RepositoryRuleset, error) {
	allRulesets := []*github.RepositoryRuleset{}
	opt := &github.ListOptions{
		PerPage: defaultPerPage,
	}

	for {
		rulesets, resp, err := g.client.Organizations.GetAllRepositoryRulesets(ctx, org, opt)
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

// GetOrgRuleset retrieves a single ruleset for a specific organization by ruleset ID
func (g *GitHubClient) GetOrgRuleset(ctx context.Context, org string, rulesetID int64) (*github.RepositoryRuleset, error) {
	ruleset, _, err := g.client.Organizations.GetRepositoryRuleset(ctx, org, rulesetID)
	if err != nil {
		return nil, err
	}

	return ruleset, nil
}

// CreateOrgRuleset creates a new ruleset for a specific organization
func (g *GitHubClient) CreateOrgRuleset(ctx context.Context, org string, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	createdRuleset, _, err := g.client.Organizations.CreateRepositoryRuleset(ctx, org, *ruleset)
	if err != nil {
		return nil, err
	}

	return createdRuleset, nil
}

// UpdateOrgRuleset updates an existing ruleset for a specific organization
func (g *GitHubClient) UpdateOrgRuleset(ctx context.Context, org string, rulesetID int64, ruleset *github.RepositoryRuleset) (*github.RepositoryRuleset, error) {
	updatedRuleset, _, err := g.client.Organizations.UpdateRepositoryRuleset(ctx, org, rulesetID, *ruleset)
	if err != nil {
		return nil, err
	}

	return updatedRuleset, nil
}

// DeleteOrgRuleset deletes a single ruleset for a specific organization by ruleset ID
func (g *GitHubClient) DeleteOrgRuleset(ctx context.Context, org string, rulesetID int64) error {
	_, err := g.client.Organizations.DeleteRepositoryRuleset(ctx, org, rulesetID)
	if err != nil {
		return err
	}

	return nil
}
