package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListRepoCodeScanningAlerts lists all code scanning alerts for a repository.
func (g *GitHubClient) ListRepoCodeScanningAlerts(ctx context.Context, owner, repo string, opts *github.AlertListOptions) ([]*github.Alert, error) {
	var allAlerts []*github.Alert
	if opts == nil {
		opts = &github.AlertListOptions{}
	}
	opts.ListOptions = github.ListOptions{PerPage: defaultPerPage}

	for {
		alerts, resp, err := g.client.CodeScanning.ListAlertsForRepo(ctx, owner, repo, opts)
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

// GetRepoCodeScanningAlert gets a single code scanning alert for a repository.
func (g *GitHubClient) GetRepoCodeScanningAlert(ctx context.Context, owner, repo string, number int64) (*github.Alert, error) {
	alert, _, err := g.client.CodeScanning.GetAlert(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return alert, nil
}

// UpdateRepoCodeScanningAlert updates a code scanning alert for a repository.
func (g *GitHubClient) UpdateRepoCodeScanningAlert(ctx context.Context, owner, repo string, number int64, state *github.CodeScanningAlertState) (*github.Alert, error) {
	alert, _, err := g.client.CodeScanning.UpdateAlert(ctx, owner, repo, number, state)
	if err != nil {
		return nil, err
	}
	return alert, nil
}
