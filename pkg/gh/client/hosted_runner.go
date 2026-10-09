package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// ListOrgHostedRunners lists the configured GitHub-hosted runner pools.
func (g *GitHubClient) ListOrgHostedRunners(ctx context.Context, owner string) ([]*github.HostedRunner, error) {
	return g.listHostedRunners(ctx, owner, g.client.Actions.ListHostedRunners)
}

func (g *GitHubClient) ListEnterpriseHostedRunners(ctx context.Context, enterprise string) ([]*github.HostedRunner, error) {
	return g.listHostedRunners(ctx, enterprise, g.client.Enterprise.ListHostedRunners)
}

// ListOrgRunnerGroupHostedRunners lists the GitHub-hosted runner pools of an organization runner group.
func (g *GitHubClient) ListOrgRunnerGroupHostedRunners(ctx context.Context, owner string, groupID int64) ([]*github.HostedRunner, error) {
	return g.listHostedRunners(ctx, owner, func(ctx context.Context, owner string, opts *github.ListOptions) (*github.HostedRunners, *github.Response, error) {
		return g.client.Actions.ListRunnerGroupHostedRunners(ctx, owner, groupID, opts)
	})
}

func (g *GitHubClient) listHostedRunners(ctx context.Context, scope string, list func(context.Context, string, *github.ListOptions) (*github.HostedRunners, *github.Response, error)) ([]*github.HostedRunner, error) {
	var all []*github.HostedRunner
	opts := &github.ListOptions{PerPage: defaultPerPage}
	for {
		runners, resp, err := list(ctx, scope, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, runners.Runners...)
		if resp.NextPage == 0 {
			return all, nil
		}
		opts.Page = resp.NextPage
	}
}
