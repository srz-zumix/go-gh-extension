package gh

import (
	"context"
	"fmt"
	"slices"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// OrgSecurityProducts is the list of valid security products for organization-wide enable/disable operations.
var OrgSecurityProducts = []string{
	"dependency_graph",
	"dependabot_alerts",
	"dependabot_security_updates",
	"advanced_security",
	"code_scanning_default_setup",
	"secret_scanning",
	"secret_scanning_push_protection",
}

// OrgSecurityEnablements is the list of valid organization-wide security enablement actions.
var OrgSecurityEnablements = []string{
	"enable_all",
	"disable_all",
}

// CodeScanningQuerySuites is the list of valid query suites for code_scanning_default_setup.
var CodeScanningQuerySuites = []string{
	"default",
	"extended",
}

// EnableDisableOrganizationSecurityFeature enables or disables a security feature for all eligible repositories in an organization.
func EnableDisableOrganizationSecurityFeature(ctx context.Context, g *GitHubClient, repo repository.Repository, securityProduct, enablement, querySuite string) error {
	if !slices.Contains(OrgSecurityProducts, securityProduct) {
		return fmt.Errorf("invalid security product %q", securityProduct)
	}
	if !slices.Contains(OrgSecurityEnablements, enablement) {
		return fmt.Errorf("invalid enablement %q", enablement)
	}
	if querySuite != "" {
		if securityProduct != "code_scanning_default_setup" {
			return fmt.Errorf("query suite is only supported for code_scanning_default_setup")
		}
		if !slices.Contains(CodeScanningQuerySuites, querySuite) {
			return fmt.Errorf("invalid query suite %q", querySuite)
		}
	}

	if err := g.EnableDisableOrgSecurityFeature(ctx, repo.Owner, securityProduct, enablement, querySuite); err != nil {
		return fmt.Errorf("failed to apply organization-wide security feature %s/%s for %s: %w", securityProduct, enablement, repo.Owner, err)
	}

	return nil
}
