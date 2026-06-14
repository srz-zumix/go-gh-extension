package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
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
