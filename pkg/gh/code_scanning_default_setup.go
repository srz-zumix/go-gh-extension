package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v88/github"
)

// EnableCodeScanningDefaultSetup enables code scanning default setup for all eligible repositories in an organization.
// querySuite may be "default" or "extended" (or empty to use the GitHub default).
func EnableCodeScanningDefaultSetup(ctx context.Context, g *GitHubClient, repo repository.Repository, querySuite string) error {
	return EnableDisableOrganizationSecurityFeature(ctx, g, repo, "code_scanning_default_setup", "enable_all", querySuite)
}

// DisableCodeScanningDefaultSetup disables code scanning default setup for all eligible repositories in an organization.
func DisableCodeScanningDefaultSetup(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	return EnableDisableOrganizationSecurityFeature(ctx, g, repo, "code_scanning_default_setup", "disable_all", "")
}

// GetCodeScanningDefaultSetupConfiguration gets the code scanning default setup configuration for a repository.
func GetCodeScanningDefaultSetupConfiguration(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.DefaultSetupConfiguration, error) {
	config, err := g.GetDefaultSetupConfiguration(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get code scanning default setup configuration for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return config, nil
}

// UpdateCodeScanningDefaultSetupConfigurationOptions holds options for updating a
// repository's code scanning default setup configuration.
type UpdateCodeScanningDefaultSetupConfigurationOptions struct {
	State      string
	QuerySuite string
	Languages  []string
}

// toGitHubUpdateDefaultSetupConfigurationOptions converts UpdateCodeScanningDefaultSetupConfigurationOptions
// to github.UpdateDefaultSetupConfigurationOptions.
func toGitHubUpdateDefaultSetupConfigurationOptions(opts *UpdateCodeScanningDefaultSetupConfigurationOptions) *github.UpdateDefaultSetupConfigurationOptions {
	if opts == nil {
		return nil
	}
	o := &github.UpdateDefaultSetupConfigurationOptions{
		State: opts.State,
	}
	if opts.QuerySuite != "" {
		o.QuerySuite = &opts.QuerySuite
	}
	if len(opts.Languages) > 0 {
		o.Languages = opts.Languages
	}
	return o
}

// UpdateCodeScanningDefaultSetupConfiguration updates the code scanning default setup configuration for a repository.
func UpdateCodeScanningDefaultSetupConfiguration(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *UpdateCodeScanningDefaultSetupConfigurationOptions) (*github.UpdateDefaultSetupConfigurationResponse, error) {
	result, err := g.UpdateDefaultSetupConfiguration(ctx, repo.Owner, repo.Name, toGitHubUpdateDefaultSetupConfigurationOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to update code scanning default setup configuration for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return result, nil
}
