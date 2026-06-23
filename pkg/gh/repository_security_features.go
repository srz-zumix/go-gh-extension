package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v88/github"
)

// RepositorySecurityFeatureStatus holds the enabled/paused status of a repository security feature.
type RepositorySecurityFeatureStatus struct {
	Feature string `json:"feature"`
	Enabled bool   `json:"enabled"`
	Paused  *bool  `json:"paused,omitempty"`
}

// GetVulnerabilityAlerts checks if vulnerability alerts are enabled for a repository.
func GetVulnerabilityAlerts(ctx context.Context, g *GitHubClient, repo repository.Repository) (*RepositorySecurityFeatureStatus, error) {
	enabled, err := g.GetVulnerabilityAlerts(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get vulnerability alerts status for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return &RepositorySecurityFeatureStatus{Feature: "vulnerability-alerts", Enabled: enabled}, nil
}

// EnableVulnerabilityAlerts enables vulnerability alerts for a repository.
func EnableVulnerabilityAlerts(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	if err := g.EnableVulnerabilityAlerts(ctx, repo.Owner, repo.Name); err != nil {
		return fmt.Errorf("failed to enable vulnerability alerts for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return nil
}

// DisableVulnerabilityAlerts disables vulnerability alerts for a repository.
func DisableVulnerabilityAlerts(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	if err := g.DisableVulnerabilityAlerts(ctx, repo.Owner, repo.Name); err != nil {
		return fmt.Errorf("failed to disable vulnerability alerts for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return nil
}

// GetAutomatedSecurityFixes checks if automated security fixes are enabled for a repository.
func GetAutomatedSecurityFixes(ctx context.Context, g *GitHubClient, repo repository.Repository) (*RepositorySecurityFeatureStatus, error) {
	result, err := g.GetAutomatedSecurityFixes(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get automated security fixes status for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	status := &RepositorySecurityFeatureStatus{Feature: "automated-security-fixes"}
	if result != nil {
		status.Enabled = result.GetEnabled()
		status.Paused = result.Paused
	}
	return status, nil
}

// EnableAutomatedSecurityFixes enables automated security fixes for a repository.
func EnableAutomatedSecurityFixes(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	if err := g.EnableAutomatedSecurityFixes(ctx, repo.Owner, repo.Name); err != nil {
		return fmt.Errorf("failed to enable automated security fixes for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return nil
}

// DisableAutomatedSecurityFixes disables automated security fixes for a repository.
func DisableAutomatedSecurityFixes(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	if err := g.DisableAutomatedSecurityFixes(ctx, repo.Owner, repo.Name); err != nil {
		return fmt.Errorf("failed to disable automated security fixes for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return nil
}

// GetPrivateVulnerabilityReporting checks if private vulnerability reporting is enabled for a repository.
func GetPrivateVulnerabilityReporting(ctx context.Context, g *GitHubClient, repo repository.Repository) (*RepositorySecurityFeatureStatus, error) {
	enabled, err := g.IsPrivateReportingEnabled(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get private vulnerability reporting status for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return &RepositorySecurityFeatureStatus{Feature: "private-vulnerability-reporting", Enabled: enabled}, nil
}

// EnablePrivateVulnerabilityReporting enables private vulnerability reporting for a repository.
func EnablePrivateVulnerabilityReporting(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	if err := g.EnablePrivateReporting(ctx, repo.Owner, repo.Name); err != nil {
		return fmt.Errorf("failed to enable private vulnerability reporting for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return nil
}

// DisablePrivateVulnerabilityReporting disables private vulnerability reporting for a repository.
func DisablePrivateVulnerabilityReporting(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	if err := g.DisablePrivateReporting(ctx, repo.Owner, repo.Name); err != nil {
		return fmt.Errorf("failed to disable private vulnerability reporting for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return nil
}

// AutomatedSecurityFixes is an alias for github.AutomatedSecurityFixes for external use.
type AutomatedSecurityFixes = github.AutomatedSecurityFixes
