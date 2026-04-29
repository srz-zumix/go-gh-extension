package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// CodeSecurityFeatureStates is the list of valid feature enablement states.
var CodeSecurityFeatureStates = []string{
	"enabled",
	"disabled",
	"not_set",
}

// CodeSecurityAdvancedSecurityStates is the list of valid advanced_security states.
var CodeSecurityAdvancedSecurityStates = []string{
	"enabled",
	"disabled",
	"code_security",
	"secret_protection",
}

// CodeSecurityEnforcementStates is the list of valid enforcement states.
var CodeSecurityEnforcementStates = []string{
	"enforced",
	"unenforced",
}

// CodeSecurityListTargetTypes is the list of valid target_type filters for listing configurations.
var CodeSecurityListTargetTypes = []string{
	"global",
	"all",
}

// CodeSecurityAttachScopes is the list of valid attach scopes for an organization.
var CodeSecurityAttachScopes = []string{
	"all",
	"all_without_configurations",
	"public",
	"private_or_internal",
	"selected",
}

// CodeSecurityDefaultForNewRepos is the list of valid default_for_new_repos values.
var CodeSecurityDefaultForNewRepos = []string{
	"all",
	"none",
	"private_and_internal",
	"public",
}

// CodeSecurityRepositoryStatuses is the list of valid repository attachment statuses for filtering.
var CodeSecurityRepositoryStatuses = []string{
	"all",
	"attached",
	"attaching",
	"detached",
	"removed",
	"enforced",
	"failed",
	"updating",
	"removed_by_enterprise",
}

// ListCodeSecurityConfigurationsOptions holds filter options for listing configurations.
type ListCodeSecurityConfigurationsOptions struct {
	TargetType string
}

// ListCodeSecurityConfigurationRepositoriesOptions holds filter options for listing configuration repositories.
type ListCodeSecurityConfigurationRepositoriesOptions struct {
	Status string
}

// CodeSecurityConfigurationOptions holds the editable fields for create/update of a code security configuration.
// Empty string fields are omitted from the request.
type CodeSecurityConfigurationOptions struct {
	Name                                  string
	Description                           string
	AdvancedSecurity                      string
	CodeSecurity                          string
	DependencyGraph                       string
	DependencyGraphAutosubmitAction       string
	DependabotAlerts                      string
	DependabotSecurityUpdates             string
	CodeScanningDefaultSetup              string
	CodeScanningDelegatedAlertDismissal   string
	SecretProtection                      string
	SecretScanning                        string
	SecretScanningPushProtection          string
	SecretScanningDelegatedBypass         string
	SecretScanningValidityChecks          string
	SecretScanningNonProviderPatterns     string
	SecretScanningGenericSecrets          string
	SecretScanningDelegatedAlertDismissal string
	PrivateVulnerabilityReporting         string
	Enforcement                           string
}

// stringPtrIfSet returns a pointer to s if non-empty, otherwise nil.
func stringPtrIfSet(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// toGitHubCodeSecurityConfiguration converts CodeSecurityConfigurationOptions to github.CodeSecurityConfiguration.
func toGitHubCodeSecurityConfiguration(opts *CodeSecurityConfigurationOptions) github.CodeSecurityConfiguration {
	if opts == nil {
		return github.CodeSecurityConfiguration{}
	}
	return github.CodeSecurityConfiguration{
		Name:                                  opts.Name,
		Description:                           opts.Description,
		AdvancedSecurity:                      stringPtrIfSet(opts.AdvancedSecurity),
		CodeSecurity:                          stringPtrIfSet(opts.CodeSecurity),
		DependencyGraph:                       stringPtrIfSet(opts.DependencyGraph),
		DependencyGraphAutosubmitAction:       stringPtrIfSet(opts.DependencyGraphAutosubmitAction),
		DependabotAlerts:                      stringPtrIfSet(opts.DependabotAlerts),
		DependabotSecurityUpdates:             stringPtrIfSet(opts.DependabotSecurityUpdates),
		CodeScanningDefaultSetup:              stringPtrIfSet(opts.CodeScanningDefaultSetup),
		CodeScanningDelegatedAlertDismissal:   stringPtrIfSet(opts.CodeScanningDelegatedAlertDismissal),
		SecretProtection:                      stringPtrIfSet(opts.SecretProtection),
		SecretScanning:                        stringPtrIfSet(opts.SecretScanning),
		SecretScanningPushProtection:          stringPtrIfSet(opts.SecretScanningPushProtection),
		SecretScanningDelegatedBypass:         stringPtrIfSet(opts.SecretScanningDelegatedBypass),
		SecretScanningValidityChecks:          stringPtrIfSet(opts.SecretScanningValidityChecks),
		SecretScanningNonProviderPatterns:     stringPtrIfSet(opts.SecretScanningNonProviderPatterns),
		SecretScanningGenericSecrets:          stringPtrIfSet(opts.SecretScanningGenericSecrets),
		SecretScanningDelegatedAlertDismissal: stringPtrIfSet(opts.SecretScanningDelegatedAlertDismissal),
		PrivateVulnerabilityReporting:         stringPtrIfSet(opts.PrivateVulnerabilityReporting),
		Enforcement:                           stringPtrIfSet(opts.Enforcement),
	}
}

// ListCodeSecurityConfigurations lists code security configurations for an organization.
func ListCodeSecurityConfigurations(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListCodeSecurityConfigurationsOptions) ([]*github.CodeSecurityConfiguration, error) {
	var ghOpts *github.ListOrgCodeSecurityConfigurationOptions
	if opts != nil && opts.TargetType != "" {
		ghOpts = &github.ListOrgCodeSecurityConfigurationOptions{TargetType: opts.TargetType}
	}
	configs, err := g.ListOrgCodeSecurityConfigurations(ctx, repo.Owner, ghOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list code security configurations for %s: %w", repo.Owner, err)
	}
	return configs, nil
}

// CreateCodeSecurityConfiguration creates a code security configuration for an organization.
func CreateCodeSecurityConfiguration(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *CodeSecurityConfigurationOptions) (*github.CodeSecurityConfiguration, error) {
	c, err := g.CreateOrgCodeSecurityConfiguration(ctx, repo.Owner, toGitHubCodeSecurityConfiguration(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to create code security configuration for %s: %w", repo.Owner, err)
	}
	return c, nil
}

// ListDefaultCodeSecurityConfigurations lists default code security configurations for an organization.
func ListDefaultCodeSecurityConfigurations(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.CodeSecurityConfigurationWithDefaultForNewRepos, error) {
	defaults, err := g.ListOrgDefaultCodeSecurityConfigurations(ctx, repo.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to list default code security configurations for %s: %w", repo.Owner, err)
	}
	return defaults, nil
}

// DetachCodeSecurityConfigurations detaches configurations from a set of repositories.
func DetachCodeSecurityConfigurations(ctx context.Context, g *GitHubClient, repo repository.Repository, repoIDs []int64) error {
	if err := g.DetachOrgCodeSecurityConfigurations(ctx, repo.Owner, repoIDs); err != nil {
		return fmt.Errorf("failed to detach code security configurations for %s: %w", repo.Owner, err)
	}
	return nil
}

// GetCodeSecurityConfiguration gets a single code security configuration for an organization.
func GetCodeSecurityConfiguration(ctx context.Context, g *GitHubClient, repo repository.Repository, configurationID int64) (*github.CodeSecurityConfiguration, error) {
	c, err := g.GetOrgCodeSecurityConfiguration(ctx, repo.Owner, configurationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get code security configuration #%d for %s: %w", configurationID, repo.Owner, err)
	}
	return c, nil
}

// UpdateCodeSecurityConfiguration updates a code security configuration for an organization.
func UpdateCodeSecurityConfiguration(ctx context.Context, g *GitHubClient, repo repository.Repository, configurationID int64, opts *CodeSecurityConfigurationOptions) (*github.CodeSecurityConfiguration, error) {
	c, err := g.UpdateOrgCodeSecurityConfiguration(ctx, repo.Owner, configurationID, toGitHubCodeSecurityConfiguration(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to update code security configuration #%d for %s: %w", configurationID, repo.Owner, err)
	}
	return c, nil
}

// DeleteCodeSecurityConfiguration deletes a code security configuration for an organization.
func DeleteCodeSecurityConfiguration(ctx context.Context, g *GitHubClient, repo repository.Repository, configurationID int64) error {
	if err := g.DeleteOrgCodeSecurityConfiguration(ctx, repo.Owner, configurationID); err != nil {
		return fmt.Errorf("failed to delete code security configuration #%d for %s: %w", configurationID, repo.Owner, err)
	}
	return nil
}

// AttachCodeSecurityConfiguration attaches a configuration to repositories in an organization.
func AttachCodeSecurityConfiguration(ctx context.Context, g *GitHubClient, repo repository.Repository, configurationID int64, scope string, repoIDs []int64) error {
	if err := g.AttachOrgCodeSecurityConfiguration(ctx, repo.Owner, configurationID, scope, repoIDs); err != nil {
		return fmt.Errorf("failed to attach code security configuration #%d for %s: %w", configurationID, repo.Owner, err)
	}
	return nil
}

// SetDefaultCodeSecurityConfiguration sets a code security configuration as default for an organization.
func SetDefaultCodeSecurityConfiguration(ctx context.Context, g *GitHubClient, repo repository.Repository, configurationID int64, defaultForNewRepos string) (*github.CodeSecurityConfigurationWithDefaultForNewRepos, error) {
	d, err := g.SetOrgDefaultCodeSecurityConfiguration(ctx, repo.Owner, configurationID, defaultForNewRepos)
	if err != nil {
		return nil, fmt.Errorf("failed to set default code security configuration #%d for %s: %w", configurationID, repo.Owner, err)
	}
	return d, nil
}

// ListCodeSecurityConfigurationRepositories lists repositories associated with a configuration.
func ListCodeSecurityConfigurationRepositories(ctx context.Context, g *GitHubClient, repo repository.Repository, configurationID int64, opts *ListCodeSecurityConfigurationRepositoriesOptions) ([]*github.RepositoryAttachment, error) {
	var ghOpts *github.ListCodeSecurityConfigurationRepositoriesOptions
	if opts != nil && opts.Status != "" {
		ghOpts = &github.ListCodeSecurityConfigurationRepositoriesOptions{Status: opts.Status}
	}
	attachments, err := g.ListOrgCodeSecurityConfigurationRepositories(ctx, repo.Owner, configurationID, ghOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for code security configuration #%d in %s: %w", configurationID, repo.Owner, err)
	}
	return attachments, nil
}

// GetRepoCodeSecurityConfiguration gets the code security configuration associated with a repository.
func GetRepoCodeSecurityConfiguration(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.RepositoryCodeSecurityConfiguration, error) {
	c, err := g.GetRepoCodeSecurityConfiguration(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get code security configuration for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return c, nil
}
