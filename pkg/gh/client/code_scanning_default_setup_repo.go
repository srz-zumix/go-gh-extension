package client

import (
	"context"

	"github.com/google/go-github/v88/github"
)

// GetDefaultSetupConfiguration gets a code scanning default setup configuration for a repository.
func (g *GitHubClient) GetDefaultSetupConfiguration(ctx context.Context, owner, repo string) (*github.DefaultSetupConfiguration, error) {
	config, _, err := g.client.CodeScanning.GetDefaultSetupConfiguration(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// UpdateDefaultSetupConfiguration updates a code scanning default setup configuration for a repository.
// This method may return an AcceptedError (HTTP 202) while GitHub schedules the update;
// this is treated as a non-fatal success and the response body is still returned.
func (g *GitHubClient) UpdateDefaultSetupConfiguration(ctx context.Context, owner, repo string, opts *github.UpdateDefaultSetupConfigurationOptions) (*github.UpdateDefaultSetupConfigurationResponse, error) {
	result := new(github.UpdateDefaultSetupConfigurationResponse)
	updated, _, err := g.client.CodeScanning.UpdateDefaultSetupConfiguration(ctx, owner, repo, opts)
	if err := handleAcceptedError(err, result); err != nil {
		return nil, err
	}
	if updated != nil {
		result = updated
	}
	return result, nil
}
