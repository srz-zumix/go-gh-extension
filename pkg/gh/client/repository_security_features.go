package client

import (
	"context"

	"github.com/google/go-github/v88/github"
)

// GetVulnerabilityAlerts checks if vulnerability alerts are enabled for a repository.
func (g *GitHubClient) GetVulnerabilityAlerts(ctx context.Context, owner, repo string) (bool, error) {
	enabled, _, err := g.client.Repositories.GetVulnerabilityAlerts(ctx, owner, repo)
	if err != nil {
		return false, err
	}
	return enabled, nil
}

// EnableVulnerabilityAlerts enables vulnerability alerts and the dependency graph for a repository.
func (g *GitHubClient) EnableVulnerabilityAlerts(ctx context.Context, owner, repo string) error {
	_, err := g.client.Repositories.EnableVulnerabilityAlerts(ctx, owner, repo)
	return err
}

// DisableVulnerabilityAlerts disables vulnerability alerts and the dependency graph for a repository.
func (g *GitHubClient) DisableVulnerabilityAlerts(ctx context.Context, owner, repo string) error {
	_, err := g.client.Repositories.DisableVulnerabilityAlerts(ctx, owner, repo)
	return err
}

// GetAutomatedSecurityFixes checks if automated security fixes are enabled for a repository.
func (g *GitHubClient) GetAutomatedSecurityFixes(ctx context.Context, owner, repo string) (*github.AutomatedSecurityFixes, error) {
	result, _, err := g.client.Repositories.GetAutomatedSecurityFixes(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// EnableAutomatedSecurityFixes enables automated security fixes for a repository.
func (g *GitHubClient) EnableAutomatedSecurityFixes(ctx context.Context, owner, repo string) error {
	_, err := g.client.Repositories.EnableAutomatedSecurityFixes(ctx, owner, repo)
	return err
}

// DisableAutomatedSecurityFixes disables automated security fixes for a repository.
func (g *GitHubClient) DisableAutomatedSecurityFixes(ctx context.Context, owner, repo string) error {
	_, err := g.client.Repositories.DisableAutomatedSecurityFixes(ctx, owner, repo)
	return err
}

// IsPrivateReportingEnabled checks if private vulnerability reporting is enabled for a repository.
func (g *GitHubClient) IsPrivateReportingEnabled(ctx context.Context, owner, repo string) (bool, error) {
	enabled, _, err := g.client.Repositories.IsPrivateReportingEnabled(ctx, owner, repo)
	if err != nil {
		return false, err
	}
	return enabled, nil
}

// EnablePrivateReporting enables private vulnerability reporting for a repository.
func (g *GitHubClient) EnablePrivateReporting(ctx context.Context, owner, repo string) error {
	_, err := g.client.Repositories.EnablePrivateReporting(ctx, owner, repo)
	return err
}

// DisablePrivateReporting disables private vulnerability reporting for a repository.
func (g *GitHubClient) DisablePrivateReporting(ctx context.Context, owner, repo string) error {
	_, err := g.client.Repositories.DisablePrivateReporting(ctx, owner, repo)
	return err
}
