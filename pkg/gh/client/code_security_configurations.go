package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListOrgCodeSecurityConfigurations lists code security configurations for an organization.
func (g *GitHubClient) ListOrgCodeSecurityConfigurations(ctx context.Context, org string, opts *github.ListOrgCodeSecurityConfigurationOptions) ([]*github.CodeSecurityConfiguration, error) {
	var all []*github.CodeSecurityConfiguration
	if opts == nil {
		opts = &github.ListOrgCodeSecurityConfigurationOptions{}
	}
	opts.PerPage = defaultPerPage
	for {
		configs, resp, err := g.client.Organizations.ListCodeSecurityConfigurations(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, configs...)
		if resp.After == "" {
			break
		}
		opts.After = resp.After
	}
	return all, nil
}

// CreateOrgCodeSecurityConfiguration creates a code security configuration for an organization.
func (g *GitHubClient) CreateOrgCodeSecurityConfiguration(ctx context.Context, org string, config github.CodeSecurityConfiguration) (*github.CodeSecurityConfiguration, error) {
	c, _, err := g.client.Organizations.CreateCodeSecurityConfiguration(ctx, org, config)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ListOrgDefaultCodeSecurityConfigurations lists default code security configurations for an organization.
func (g *GitHubClient) ListOrgDefaultCodeSecurityConfigurations(ctx context.Context, org string) ([]*github.CodeSecurityConfigurationWithDefaultForNewRepos, error) {
	defaults, _, err := g.client.Organizations.ListDefaultCodeSecurityConfigurations(ctx, org)
	if err != nil {
		return nil, err
	}
	return defaults, nil
}

// DetachOrgCodeSecurityConfigurations detaches code security configurations from a set of repositories.
func (g *GitHubClient) DetachOrgCodeSecurityConfigurations(ctx context.Context, org string, repoIDs []int64) error {
	_, err := g.client.Organizations.DetachCodeSecurityConfigurationsFromRepositories(ctx, org, repoIDs)
	return err
}

// GetOrgCodeSecurityConfiguration gets a code security configuration for an organization.
func (g *GitHubClient) GetOrgCodeSecurityConfiguration(ctx context.Context, org string, configurationID int64) (*github.CodeSecurityConfiguration, error) {
	c, _, err := g.client.Organizations.GetCodeSecurityConfiguration(ctx, org, configurationID)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// UpdateOrgCodeSecurityConfiguration updates a code security configuration for an organization.
func (g *GitHubClient) UpdateOrgCodeSecurityConfiguration(ctx context.Context, org string, configurationID int64, config github.CodeSecurityConfiguration) (*github.CodeSecurityConfiguration, error) {
	c, _, err := g.client.Organizations.UpdateCodeSecurityConfiguration(ctx, org, configurationID, config)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// DeleteOrgCodeSecurityConfiguration deletes a code security configuration for an organization.
func (g *GitHubClient) DeleteOrgCodeSecurityConfiguration(ctx context.Context, org string, configurationID int64) error {
	_, err := g.client.Organizations.DeleteCodeSecurityConfiguration(ctx, org, configurationID)
	return err
}

// AttachOrgCodeSecurityConfiguration attaches a code security configuration to repositories in an organization.
func (g *GitHubClient) AttachOrgCodeSecurityConfiguration(ctx context.Context, org string, configurationID int64, scope string, repoIDs []int64) error {
	_, err := g.client.Organizations.AttachCodeSecurityConfigurationToRepositories(ctx, org, configurationID, scope, repoIDs)
	return err
}

// SetOrgDefaultCodeSecurityConfiguration sets a code security configuration as a default for an organization.
func (g *GitHubClient) SetOrgDefaultCodeSecurityConfiguration(ctx context.Context, org string, configurationID int64, defaultForNewRepos string) (*github.CodeSecurityConfigurationWithDefaultForNewRepos, error) {
	d, _, err := g.client.Organizations.SetDefaultCodeSecurityConfiguration(ctx, org, configurationID, defaultForNewRepos)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// ListOrgCodeSecurityConfigurationRepositories lists repositories associated with a code security configuration.
func (g *GitHubClient) ListOrgCodeSecurityConfigurationRepositories(ctx context.Context, org string, configurationID int64, opts *github.ListCodeSecurityConfigurationRepositoriesOptions) ([]*github.RepositoryAttachment, error) {
	var all []*github.RepositoryAttachment
	if opts == nil {
		opts = &github.ListCodeSecurityConfigurationRepositoriesOptions{}
	}
	opts.PerPage = defaultPerPage
	for {
		attachments, resp, err := g.client.Organizations.ListCodeSecurityConfigurationRepositories(ctx, org, configurationID, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, attachments...)
		if resp.After == "" {
			break
		}
		opts.After = resp.After
	}
	return all, nil
}

// GetRepoCodeSecurityConfiguration gets the code security configuration associated with a repository.
func (g *GitHubClient) GetRepoCodeSecurityConfiguration(ctx context.Context, owner, repo string) (*github.RepositoryCodeSecurityConfiguration, error) {
	c, _, err := g.client.Organizations.GetCodeSecurityConfigurationForRepository(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return c, nil
}
