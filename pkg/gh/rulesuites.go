package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

type RuleSuite = client.RuleSuite

// RuleSuiteTimePeriodList is the list of valid time period values for filtering rule suites.
var RuleSuiteTimePeriodList = []string{
	"hour",
	"day",
	"week",
	"month",
}

// RuleSuiteResultList is the list of valid result values for filtering rule suites.
var RuleSuiteResultList = []string{
	"pass",
	"fail",
	"bypass",
	"all",
}

type ListRuleSuitesOptions struct {
	Ref             string
	TimePeriod      string
	ActorName       string
	RuleSuiteResult string
}

func (opts *ListRuleSuitesOptions) ToGitHubOptions() *client.ListRuleSuitesOptions {
	if opts == nil {
		return &client.ListRuleSuitesOptions{}
	}
	return &client.ListRuleSuitesOptions{
		Ref:             opts.Ref,
		TimePeriod:      opts.TimePeriod,
		ActorName:       opts.ActorName,
		RuleSuiteResult: opts.RuleSuiteResult,
	}
}

// ListRepositoryRuleSuites retrieves all rule suites for a specific repository
func ListRepositoryRuleSuites(ctx context.Context, g *GitHubClient, repo repository.Repository, options *ListRuleSuitesOptions) ([]*client.RuleSuite, error) {
	opts := options.ToGitHubOptions()
	return g.ListRepositoryRuleSuites(ctx, repo.Owner, repo.Name, opts)
}

// GetRepositoryRuleSuite retrieves a single rule suite for a specific repository by rule suite ID
func GetRepositoryRuleSuite(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleSuiteID int64) (*RuleSuite, error) {
	return g.GetRepositoryRuleSuite(ctx, repo.Owner, repo.Name, ruleSuiteID)
}

// ListOrgRuleSuites retrieves all rule suites for a specific organization
func ListOrgRuleSuites(ctx context.Context, g *GitHubClient, repo repository.Repository, options *ListRuleSuitesOptions) ([]*RuleSuite, error) {
	opts := options.ToGitHubOptions()
	return g.ListOrgRuleSuites(ctx, repo.Owner, opts)
}

// GetOrgRuleSuite retrieves a single rule suite for a specific organization by rule suite ID
func GetOrgRuleSuite(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleSuiteID int64) (*RuleSuite, error) {
	return g.GetOrgRuleSuite(ctx, repo.Owner, ruleSuiteID)
}

// ListRuleSuites dispatches to the organization or repository rule-suite
// listing API based on whether repo.Name is set.
func ListRuleSuites(ctx context.Context, g *GitHubClient, repo repository.Repository, options *ListRuleSuitesOptions) ([]*RuleSuite, error) {
	if repo.Name == "" {
		return ListOrgRuleSuites(ctx, g, repo, options)
	}
	return ListRepositoryRuleSuites(ctx, g, repo, options)
}

// GetRuleSuite dispatches to the organization or repository rule-suite getter
// API based on whether repo.Name is set.
func GetRuleSuite(ctx context.Context, g *GitHubClient, repo repository.Repository, ruleSuiteID int64) (*RuleSuite, error) {
	if repo.Name == "" {
		return GetOrgRuleSuite(ctx, g, repo, ruleSuiteID)
	}
	return GetRepositoryRuleSuite(ctx, g, repo, ruleSuiteID)
}
