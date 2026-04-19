package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListRepoDependabotAlerts lists all Dependabot alerts of a repository.
func (g *GitHubClient) ListRepoDependabotAlerts(ctx context.Context, owner, repo string, opts *github.ListAlertsOptions) ([]*github.DependabotAlert, error) {
	var allAlerts []*github.DependabotAlert
	if opts == nil {
		opts = &github.ListAlertsOptions{}
	}
	opts.ListOptions = github.ListOptions{PerPage: defaultPerPage}

	for {
		alerts, resp, err := g.client.Dependabot.ListRepoAlerts(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allAlerts = append(allAlerts, alerts...)
		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return allAlerts, nil
}

// GetRepoDependabotAlert gets a single Dependabot alert for a repository.
func (g *GitHubClient) GetRepoDependabotAlert(ctx context.Context, owner, repo string, number int) (*github.DependabotAlert, error) {
	alert, _, err := g.client.Dependabot.GetRepoAlert(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return alert, nil
}

// ListOrgDependabotAlerts lists all Dependabot alerts of an organization.
func (g *GitHubClient) ListOrgDependabotAlerts(ctx context.Context, org string, opts *github.ListAlertsOptions) ([]*github.DependabotAlert, error) {
	var allAlerts []*github.DependabotAlert
	if opts == nil {
		opts = &github.ListAlertsOptions{}
	}
	opts.ListOptions = github.ListOptions{PerPage: defaultPerPage}

	for {
		alerts, resp, err := g.client.Dependabot.ListOrgAlerts(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		allAlerts = append(allAlerts, alerts...)
		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return allAlerts, nil
}

// UpdateRepoDependabotAlert updates a Dependabot alert for a repository.
func (g *GitHubClient) UpdateRepoDependabotAlert(ctx context.Context, owner, repo string, number int, stateInfo *github.DependabotAlertState) (*github.DependabotAlert, error) {
	alert, _, err := g.client.Dependabot.UpdateAlert(ctx, owner, repo, number, stateInfo)
	if err != nil {
		return nil, err
	}
	return alert, nil
}
