package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// EnableAdvancedSecurity enables GitHub Advanced Security for all eligible repositories in an organization.
func EnableAdvancedSecurity(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	return EnableDisableOrganizationSecurityFeature(ctx, g, repo, "advanced_security", "enable_all", "")
}

// DisableAdvancedSecurity disables GitHub Advanced Security for all eligible repositories in an organization.
func DisableAdvancedSecurity(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	return EnableDisableOrganizationSecurityFeature(ctx, g, repo, "advanced_security", "disable_all", "")
}
