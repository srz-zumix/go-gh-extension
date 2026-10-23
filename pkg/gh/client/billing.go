package client

// GitHub billing usage API functions
// See: https://docs.github.com/rest/billing/usage

import (
	"context"
	"fmt"

	"github.com/google/go-github/v93/github"
)

// BillingUsageOptions selects the period of a billing usage report. Zero values are omitted from the request.
type BillingUsageOptions struct {
	Year  int
	Month int
	Day   int
	// Hour (0-23) is supported by organization and user reports only; nil omits it.
	Hour *int
	// CostCenterID is supported by enterprise reports only.
	CostCenterID string
}

// BillingUsageSummaryOptions selects the period and filters of a billing usage summary. Zero values are omitted from the request.
type BillingUsageSummaryOptions struct {
	Year  int `url:"year,omitempty"`
	Month int `url:"month,omitempty"`
	Day   int `url:"day,omitempty"`
	// CostCenterID is supported by enterprise summaries only.
	CostCenterID string `url:"cost_center_id,omitempty"`
	// Organization is supported by enterprise summaries only.
	Organization string `url:"organization,omitempty"`
	// Repository is in the OWNER/REPO format.
	Repository string `url:"repository,omitempty"`
	Product    string `url:"product,omitempty"`
	SKU        string `url:"sku,omitempty"`
}

// BillingUsageSummaryItem is a usage line of a billing usage summary.
type BillingUsageSummaryItem struct {
	Product          string  `json:"product"`
	SKU              string  `json:"sku"`
	UnitType         string  `json:"unitType"`
	PricePerUnit     float64 `json:"pricePerUnit"`
	GrossQuantity    float64 `json:"grossQuantity"`
	GrossAmount      float64 `json:"grossAmount"`
	DiscountQuantity float64 `json:"discountQuantity"`
	DiscountAmount   float64 `json:"discountAmount"`
	NetQuantity      float64 `json:"netQuantity"`
	NetAmount        float64 `json:"netAmount"`
}

// BillingUsageSummaryTimePeriod is the period covered by a billing usage summary.
type BillingUsageSummaryTimePeriod struct {
	Year  int  `json:"year"`
	Month *int `json:"month,omitempty"`
	Day   *int `json:"day,omitempty"`
}

// BillingUsageSummary is the billing usage summary of an organization, user or enterprise.
type BillingUsageSummary struct {
	TimePeriod   BillingUsageSummaryTimePeriod `json:"timePeriod"`
	Enterprise   *string                       `json:"enterprise,omitempty"`
	Organization *string                       `json:"organization,omitempty"`
	User         *string                       `json:"user,omitempty"`
	Repository   *string                       `json:"repository,omitempty"`
	Product      *string                       `json:"product,omitempty"`
	SKU          *string                       `json:"sku,omitempty"`
	CostCenter   *github.BillingCostCenter     `json:"costCenter,omitempty"`
	UsageItems   []*BillingUsageSummaryItem    `json:"usageItems"`
}

func (o *BillingUsageOptions) toUsageReportOptions() *github.UsageReportOptions {
	opts := &github.UsageReportOptions{}
	if o == nil {
		return opts
	}
	if o.Year != 0 {
		opts.Year = github.Ptr(o.Year)
	}
	if o.Month != 0 {
		opts.Month = github.Ptr(o.Month)
	}
	if o.Day != 0 {
		opts.Day = github.Ptr(o.Day)
	}
	opts.Hour = o.Hour
	return opts
}

// GetOrgUsageReport gets the billing usage report of an organization.
func (g *GitHubClient) GetOrgUsageReport(ctx context.Context, org string, opts *BillingUsageOptions) (*github.UsageReport, error) {
	report, _, err := g.client.Billing.GetOrganizationUsageReport(ctx, org, opts.toUsageReportOptions())
	return report, err
}

// GetUserUsageReport gets the billing usage report of a user.
func (g *GitHubClient) GetUserUsageReport(ctx context.Context, user string, opts *BillingUsageOptions) (*github.UsageReport, error) {
	report, _, err := g.client.Billing.GetUsageReport(ctx, user, opts.toUsageReportOptions())
	return report, err
}

// GetEnterpriseUsageReport gets the billing usage report of an enterprise.
func (g *GitHubClient) GetEnterpriseUsageReport(ctx context.Context, enterprise string, opts *BillingUsageOptions) (*github.EnterpriseUsageReport, error) {
	apiOpts := &github.EnterpriseUsageReportOptions{}
	if opts != nil {
		apiOpts.Year = opts.Year
		apiOpts.Month = opts.Month
		apiOpts.Day = opts.Day
		apiOpts.CostCenterID = opts.CostCenterID
	}
	report, _, err := g.client.Enterprise.GetUsageReport(ctx, enterprise, apiOpts)
	return report, err
}

// toUsageSummaryOptions converts the options for the organization and user summary endpoints.
// The fields that only the enterprise summary endpoint accepts are dropped.
func (o *BillingUsageSummaryOptions) toUsageSummaryOptions() *github.UsageSummaryOptions {
	if o == nil {
		return nil
	}
	return &github.UsageSummaryOptions{
		Year:       o.Year,
		Month:      o.Month,
		Day:        o.Day,
		Repository: o.Repository,
		Product:    o.Product,
		SKU:        o.SKU,
	}
}

// usageSummaryFromReport converts an organization or user usage summary into the
// BillingUsageSummary type shared with the enterprise summary.
func usageSummaryFromReport(report *github.UsageSummaryReport) *BillingUsageSummary {
	if report == nil {
		return &BillingUsageSummary{}
	}
	summary := &BillingUsageSummary{
		TimePeriod: BillingUsageSummaryTimePeriod{
			Year:  report.TimePeriod.Year,
			Month: report.TimePeriod.Month,
			Day:   report.TimePeriod.Day,
		},
		Organization: report.Organization,
		User:         report.User,
		Repository:   report.Repository,
		Product:      report.Product,
		SKU:          report.SKU,
	}
	if report.UsageItems != nil {
		summary.UsageItems = make([]*BillingUsageSummaryItem, 0, len(report.UsageItems))
	}
	for _, item := range report.UsageItems {
		if item == nil {
			summary.UsageItems = append(summary.UsageItems, nil)
			continue
		}
		summary.UsageItems = append(summary.UsageItems, &BillingUsageSummaryItem{
			Product:          item.Product,
			SKU:              item.SKU,
			UnitType:         item.UnitType,
			PricePerUnit:     item.PricePerUnit,
			GrossQuantity:    item.GrossQuantity,
			GrossAmount:      item.GrossAmount,
			DiscountQuantity: item.DiscountQuantity,
			DiscountAmount:   item.DiscountAmount,
			NetQuantity:      item.NetQuantity,
			NetAmount:        item.NetAmount,
		})
	}
	return summary
}

// GetOrgUsageSummary gets the billing usage summary of an organization.
func (g *GitHubClient) GetOrgUsageSummary(ctx context.Context, org string, opts *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
	report, _, err := g.client.Billing.GetOrganizationUsageSummary(ctx, org, opts.toUsageSummaryOptions())
	if err != nil {
		return nil, err
	}
	return usageSummaryFromReport(report), nil
}

// GetUserUsageSummary gets the billing usage summary of a user.
func (g *GitHubClient) GetUserUsageSummary(ctx context.Context, user string, opts *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
	report, _, err := g.client.Billing.GetUserUsageSummary(ctx, user, opts.toUsageSummaryOptions())
	if err != nil {
		return nil, err
	}
	return usageSummaryFromReport(report), nil
}

// GetEnterpriseUsageSummary gets the billing usage summary of an enterprise.
func (g *GitHubClient) GetEnterpriseUsageSummary(ctx context.Context, enterprise string, opts *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
	return g.getUsageSummary(ctx, fmt.Sprintf("enterprises/%v/settings/billing/usage/summary", enterprise), opts)
}

func (g *GitHubClient) getUsageSummary(ctx context.Context, path string, opts *BillingUsageSummaryOptions) (*BillingUsageSummary, error) {
	u, err := addOptions(path, opts)
	if err != nil {
		return nil, err
	}
	req, err := g.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	summary := new(BillingUsageSummary)
	if _, err := g.client.Do(req, summary); err != nil {
		return nil, err
	}
	return summary, nil
}
