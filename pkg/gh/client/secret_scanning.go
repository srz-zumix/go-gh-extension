package client

import (
	"context"

	"github.com/google/go-github/v88/github"
)

// ListOrgSecretScanningPatternConfigs lists secret scanning pattern configurations for an organization.
func (g *GitHubClient) ListOrgSecretScanningPatternConfigs(ctx context.Context, org string) (*github.SecretScanningPatternConfigs, error) {
	configs, _, err := g.client.SecretScanning.ListPatternConfigsForOrg(ctx, org)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// UpdateOrgSecretScanningPatternConfigs updates secret scanning pattern configurations for an organization.
func (g *GitHubClient) UpdateOrgSecretScanningPatternConfigs(ctx context.Context, org string, opts *github.SecretScanningPatternConfigsUpdateOptions) (*github.SecretScanningPatternConfigsUpdate, error) {
	update, _, err := g.client.SecretScanning.UpdatePatternConfigsForOrg(ctx, org, opts)
	if err != nil {
		return nil, err
	}
	return update, nil
}

// ListOrgSecretScanningAlerts lists all secret scanning alerts for an organization.
func (g *GitHubClient) ListOrgSecretScanningAlerts(ctx context.Context, org string, opts *github.SecretScanningAlertListOptions) ([]*github.SecretScanningAlert, error) {
	var allAlerts []*github.SecretScanningAlert
	if opts == nil {
		opts = &github.SecretScanningAlertListOptions{}
	}
	opts.ListOptions = github.ListOptions{PerPage: defaultPerPage}

	for {
		alerts, resp, err := g.client.SecretScanning.ListAlertsForOrg(ctx, org, opts)
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

// ListRepoSecretScanningAlerts lists all secret scanning alerts for a repository.
func (g *GitHubClient) ListRepoSecretScanningAlerts(ctx context.Context, owner, repo string, opts *github.SecretScanningAlertListOptions) ([]*github.SecretScanningAlert, error) {
	var allAlerts []*github.SecretScanningAlert
	if opts == nil {
		opts = &github.SecretScanningAlertListOptions{}
	}
	opts.ListOptions = github.ListOptions{PerPage: defaultPerPage}

	for {
		alerts, resp, err := g.client.SecretScanning.ListAlertsForRepo(ctx, owner, repo, opts)
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

// GetSecretScanningAlert gets a single secret scanning alert for a repository.
func (g *GitHubClient) GetSecretScanningAlert(ctx context.Context, owner, repo string, number int64) (*github.SecretScanningAlert, error) {
	alert, _, err := g.client.SecretScanning.GetAlert(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return alert, nil
}

// UpdateSecretScanningAlert updates a secret scanning alert for a repository.
func (g *GitHubClient) UpdateSecretScanningAlert(ctx context.Context, owner, repo string, number int64, opts *github.SecretScanningAlertUpdateOptions) (*github.SecretScanningAlert, error) {
	alert, _, err := g.client.SecretScanning.UpdateAlert(ctx, owner, repo, number, opts)
	if err != nil {
		return nil, err
	}
	return alert, nil
}

// ListSecretScanningAlertLocations lists all locations for a secret scanning alert.
func (g *GitHubClient) ListSecretScanningAlertLocations(ctx context.Context, owner, repo string, number int64) ([]*github.SecretScanningAlertLocation, error) {
	var allLocations []*github.SecretScanningAlertLocation
	listOpts := &github.ListOptions{PerPage: defaultPerPage}

	for {
		locations, resp, err := g.client.SecretScanning.ListLocationsForAlert(ctx, owner, repo, number, listOpts)
		if err != nil {
			return nil, err
		}
		allLocations = append(allLocations, locations...)
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}
	return allLocations, nil
}

// GetSecretScanningScanHistory gets the secret scanning scan history for a repository.
func (g *GitHubClient) GetSecretScanningScanHistory(ctx context.Context, owner, repo string) (*github.SecretScanningScanHistory, error) {
	history, _, err := g.client.SecretScanning.GetScanHistory(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return history, nil
}
