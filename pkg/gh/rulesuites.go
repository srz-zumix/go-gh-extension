package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// ListRepositoryRuleSuites retrieves all rule suites for a specific repository
func ListRepositoryRuleSuites(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *client.ListRuleSuitesOptions) ([]*client.RuleSuite, error) {
	return g.ListRepositoryRuleSuites(ctx, repo.Owner, repo.Name, opts)
}

// GetRepositoryRuleSuite retrieves a single rule suite for a specific repository by rule suite ID
func GetRepositoryRuleSuite(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleSuiteID int64) (*client.RuleSuite, error) {
	return g.GetRepositoryRuleSuite(ctx, repo.Owner, repo.Name, ruleSuiteID)
}

// ListOrgRuleSuites retrieves all rule suites for a specific organization
func ListOrgRuleSuites(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *client.ListRuleSuitesOptions) ([]*client.RuleSuite, error) {
	return g.ListOrgRuleSuites(ctx, repo.Owner, opts)
}

// GetOrgRuleSuite retrieves a single rule suite for a specific organization by rule suite ID
func GetOrgRuleSuite(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleSuiteID int64) (*client.RuleSuite, error) {
	return g.GetOrgRuleSuite(ctx, repo.Owner, ruleSuiteID)
}
