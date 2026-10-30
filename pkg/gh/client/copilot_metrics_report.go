package client

import (
	"context"

	"github.com/google/go-github/v93/github"
)

// GetOrganizationCopilotMetricsReport gets the latest 28-day organization-level usage metrics report links.
func (g *GitHubClient) GetOrganizationCopilotMetricsReport(ctx context.Context, org string) (*github.CopilotMetricsReport, error) {
	report, _, err := g.client.Copilot.GetOrganizationMetricsReport(ctx, org)
	return report, err
}

// GetEnterpriseCopilotMetricsReport gets the latest 28-day enterprise-level usage metrics report links.
func (g *GitHubClient) GetEnterpriseCopilotMetricsReport(ctx context.Context, enterprise string) (*github.CopilotMetricsReport, error) {
	report, _, err := g.client.Copilot.GetEnterpriseMetricsReport(ctx, enterprise)
	return report, err
}

// GetOrganizationCopilotDailyMetricsReport gets the 1-day organization-level usage metrics report links.
func (g *GitHubClient) GetOrganizationCopilotDailyMetricsReport(ctx context.Context, org, day string) (*github.CopilotDailyMetricsReport, error) {
	report, _, err := g.client.Copilot.GetOrganizationDailyMetricsReport(ctx, org, &github.CopilotMetricsReportOptions{Day: day})
	return report, err
}

// GetEnterpriseCopilotDailyMetricsReport gets the 1-day enterprise-level usage metrics report links.
func (g *GitHubClient) GetEnterpriseCopilotDailyMetricsReport(ctx context.Context, enterprise, day string) (*github.CopilotDailyMetricsReport, error) {
	report, _, err := g.client.Copilot.GetEnterpriseDailyMetricsReport(ctx, enterprise, &github.CopilotMetricsReportOptions{Day: day})
	return report, err
}

// GetOrganizationCopilotUsersMetricsReport gets the latest 28-day users usage metrics report links of an organization.
func (g *GitHubClient) GetOrganizationCopilotUsersMetricsReport(ctx context.Context, org string) (*github.CopilotMetricsReport, error) {
	report, _, err := g.client.Copilot.GetOrganizationUsersMetricsReport(ctx, org)
	return report, err
}

// GetEnterpriseCopilotUsersMetricsReport gets the latest 28-day users usage metrics report links of an enterprise.
func (g *GitHubClient) GetEnterpriseCopilotUsersMetricsReport(ctx context.Context, enterprise string) (*github.CopilotMetricsReport, error) {
	report, _, err := g.client.Copilot.GetEnterpriseUsersMetricsReport(ctx, enterprise)
	return report, err
}

// GetOrganizationCopilotUsersDailyMetricsReport gets the 1-day users usage metrics report links of an organization.
func (g *GitHubClient) GetOrganizationCopilotUsersDailyMetricsReport(ctx context.Context, org, day string) (*github.CopilotDailyMetricsReport, error) {
	report, _, err := g.client.Copilot.GetOrganizationUsersDailyMetricsReport(ctx, org, &github.CopilotMetricsReportOptions{Day: day})
	return report, err
}

// GetEnterpriseCopilotUsersDailyMetricsReport gets the 1-day users usage metrics report links of an enterprise.
func (g *GitHubClient) GetEnterpriseCopilotUsersDailyMetricsReport(ctx context.Context, enterprise, day string) (*github.CopilotDailyMetricsReport, error) {
	report, _, err := g.client.Copilot.GetEnterpriseUsersDailyMetricsReport(ctx, enterprise, &github.CopilotMetricsReportOptions{Day: day})
	return report, err
}

// GetOrganizationCopilotRepositoriesDailyMetricsReport gets the 1-day repository usage metrics report links of an organization.
func (g *GitHubClient) GetOrganizationCopilotRepositoriesDailyMetricsReport(ctx context.Context, org, day string) (*github.CopilotDailyMetricsReport, error) {
	report, _, err := g.client.Copilot.GetOrganizationRepositoriesDailyMetricsReport(ctx, org, &github.CopilotMetricsReportOptions{Day: day})
	return report, err
}

// GetEnterpriseCopilotRepositoriesDailyMetricsReport gets the 1-day repository usage metrics report links of an enterprise.
func (g *GitHubClient) GetEnterpriseCopilotRepositoriesDailyMetricsReport(ctx context.Context, enterprise, day string) (*github.CopilotDailyMetricsReport, error) {
	report, _, err := g.client.Copilot.GetEnterpriseRepositoriesDailyMetricsReport(ctx, enterprise, &github.CopilotMetricsReportOptions{Day: day})
	return report, err
}

// GetOrganizationCopilotUserTeamsDailyMetricsReport gets the 1-day user-team membership report links of an organization.
func (g *GitHubClient) GetOrganizationCopilotUserTeamsDailyMetricsReport(ctx context.Context, org, day string) (*github.CopilotDailyMetricsReport, error) {
	report, _, err := g.client.Copilot.GetOrganizationUserTeamsDailyMetricsReport(ctx, org, &github.CopilotMetricsReportOptions{Day: day})
	return report, err
}

// GetEnterpriseCopilotUserTeamsDailyMetricsReport gets the 1-day user-team membership report links of an enterprise.
func (g *GitHubClient) GetEnterpriseCopilotUserTeamsDailyMetricsReport(ctx context.Context, enterprise, day string) (*github.CopilotDailyMetricsReport, error) {
	report, _, err := g.client.Copilot.GetEnterpriseUserTeamsDailyMetricsReport(ctx, enterprise, &github.CopilotMetricsReportOptions{Day: day})
	return report, err
}

// DownloadCopilotPeriodicMetrics downloads a multi-day org/enterprise-level metrics payload.
func (g *GitHubClient) DownloadCopilotPeriodicMetrics(ctx context.Context, url string) (*github.CopilotPeriodicMetrics, error) {
	m, _, err := g.client.Copilot.DownloadPeriodicMetrics(ctx, url)
	return m, err
}

// DownloadCopilotDailyMetrics downloads a 1-day org/enterprise-level metrics payload.
func (g *GitHubClient) DownloadCopilotDailyMetrics(ctx context.Context, url string) (*github.CopilotDailyMetrics, error) {
	m, _, err := g.client.Copilot.DownloadDailyMetrics(ctx, url)
	return m, err
}

// DownloadCopilotUserPeriodicMetrics downloads a multi-day users metrics payload.
func (g *GitHubClient) DownloadCopilotUserPeriodicMetrics(ctx context.Context, url string) ([]*github.CopilotUserPeriodicMetrics, error) {
	m, _, err := g.client.Copilot.DownloadUserPeriodicMetrics(ctx, url)
	return m, err
}

// DownloadCopilotUserDailyMetrics downloads a 1-day users metrics payload.
func (g *GitHubClient) DownloadCopilotUserDailyMetrics(ctx context.Context, url string) ([]*github.CopilotUserDailyMetrics, error) {
	m, _, err := g.client.Copilot.DownloadUserDailyMetrics(ctx, url)
	return m, err
}

// DownloadCopilotRepositoryDailyMetrics downloads a 1-day repository metrics payload.
func (g *GitHubClient) DownloadCopilotRepositoryDailyMetrics(ctx context.Context, url string) ([]*github.CopilotRepositoryDailyMetrics, error) {
	m, _, err := g.client.Copilot.DownloadRepositoryDailyMetrics(ctx, url)
	return m, err
}

// DownloadCopilotUserTeamsDailyMetrics downloads a 1-day user-team membership payload.
func (g *GitHubClient) DownloadCopilotUserTeamsDailyMetrics(ctx context.Context, url string) ([]*github.CopilotUserTeamsDailyMetrics, error) {
	m, _, err := g.client.Copilot.DownloadUserTeamsDailyMetrics(ctx, url)
	return m, err
}
