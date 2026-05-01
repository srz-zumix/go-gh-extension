package client

import (
	"context"

	"github.com/google/go-github/v84/github"
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
