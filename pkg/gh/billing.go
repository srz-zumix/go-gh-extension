package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v93/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// BillingUsageOptions selects the period of a billing usage report.
type BillingUsageOptions = client.BillingUsageOptions

// BillingUsageSummaryOptions selects the period and filters of a billing usage summary.
type BillingUsageSummaryOptions = client.BillingUsageSummaryOptions

// BillingUsageSummaryItem is a usage line of a billing usage summary.
type BillingUsageSummaryItem = client.BillingUsageSummaryItem

// BillingUsageSummary is the billing usage summary of an organization, user or enterprise.
type BillingUsageSummary = client.BillingUsageSummary

// GetOrgUsageReport gets the billing usage report of the organization that owns repo (wrapper).
func GetOrgUsageReport(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *BillingUsageOptions) (*github.UsageReport, error) {
	return g.GetOrgUsageReport(ctx, repo.Owner, opts)
}

// GetUserUsageReport gets the billing usage report of the user that owns repo (wrapper).
func GetUserUsageReport(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *BillingUsageOptions) (*github.UsageReport, error) {
	return g.GetUserUsageReport(ctx, repo.Owner, opts)
}

// GetEnterpriseUsageReport gets the billing usage report of an enterprise whose slug is enterprise.Owner (wrapper).
func GetEnterpriseUsageReport(ctx context.Context, g *GitHubClient, enterprise repository.Repository, opts *BillingUsageOptions) (*github.EnterpriseUsageReport, error) {
	return g.GetEnterpriseUsageReport(ctx, enterprise.Owner, opts)
}

// GetOrgUsageSummary gets the billing usage summary of the organization that owns repo (wrapper).
func GetOrgUsageSummary(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
	return g.GetOrgUsageSummary(ctx, repo.Owner, opts)
}

// GetUserUsageSummary gets the billing usage summary of the user that owns repo (wrapper).
func GetUserUsageSummary(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
	return g.GetUserUsageSummary(ctx, repo.Owner, opts)
}

// GetEnterpriseUsageSummary gets the billing usage summary of an enterprise whose slug is enterprise.Owner (wrapper).
func GetEnterpriseUsageSummary(ctx context.Context, g *GitHubClient, enterprise repository.Repository, opts *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
	return g.GetEnterpriseUsageSummary(ctx, enterprise.Owner, opts)
}
