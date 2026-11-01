package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v93/github"
)

// CopilotMetricsSource identifies the organization or enterprise whose Copilot usage metrics are read.
// Owner.Host selects the GitHub host and Owner.Owner is the organization or enterprise slug.
type CopilotMetricsSource struct {
	Owner      repository.Repository
	Enterprise bool
}

// String returns a human-readable label such as "acme (organization)".
func (s CopilotMetricsSource) String() string {
	kind := "organization"
	if s.Enterprise {
		kind = "enterprise"
	}
	return fmt.Sprintf("%s (%s)", s.Owner.Owner, kind)
}

// CopilotMetricsWindow is the inclusive day range (YYYY-MM-DD) covered by a multi-day report.
type CopilotMetricsWindow struct {
	StartDay string
	EndDay   string
}

// CopilotLatestMetrics is the latest 28-day org/enterprise-level report with all download links merged.
type CopilotLatestMetrics struct {
	Window    CopilotMetricsWindow
	DayTotals []*github.CopilotDailyMetrics
}

// CopilotLatestUserMetrics is the latest 28-day users report with all download links merged.
// Each record is a single user-day.
type CopilotLatestUserMetrics struct {
	Window  CopilotMetricsWindow
	Records []*github.CopilotUserPeriodicMetrics
}

func reportDailyLinks(report *github.CopilotDailyMetricsReport) []string {
	if report == nil {
		return nil
	}
	return report.DownloadLinks
}

// dailyReportLinks fetches the download links of a 1-day report. A 404 response or an empty
// link list is reported as found=false so callers can treat the day as missing.
func dailyReportLinks(fetch func() (*github.CopilotDailyMetricsReport, error)) ([]string, bool, error) {
	report, err := fetch()
	if err != nil {
		if IsHTTPNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	links := reportDailyLinks(report)
	return links, len(links) > 0, nil
}

// GetCopilotLatestMetrics gets the latest 28-day org/enterprise-level usage metrics and downloads every payload.
func GetCopilotLatestMetrics(ctx context.Context, g *GitHubClient, src CopilotMetricsSource) (*CopilotLatestMetrics, error) {
	var report *github.CopilotMetricsReport
	var err error
	if src.Enterprise {
		report, err = g.GetEnterpriseCopilotMetricsReport(ctx, src.Owner.Owner)
	} else {
		report, err = g.GetOrganizationCopilotMetricsReport(ctx, src.Owner.Owner)
	}
	if err != nil {
		return nil, err
	}
	result := &CopilotLatestMetrics{}
	if report == nil {
		return result, nil
	}
	result.Window = CopilotMetricsWindow{StartDay: report.ReportStartDay, EndDay: report.ReportEndDay}
	for _, link := range report.DownloadLinks {
		m, err := g.DownloadCopilotPeriodicMetrics(ctx, link)
		if err != nil {
			return nil, fmt.Errorf("failed to download metrics report payload: %w", err)
		}
		if m == nil {
			continue
		}
		if result.Window.StartDay == "" {
			result.Window = CopilotMetricsWindow{StartDay: m.ReportStartDay, EndDay: m.ReportEndDay}
		}
		result.DayTotals = append(result.DayTotals, m.DayTotals...)
	}
	return result, nil
}

// GetCopilotDailyMetrics gets the 1-day org/enterprise-level usage metrics for day (YYYY-MM-DD).
// found is false when no report exists for the day.
func GetCopilotDailyMetrics(ctx context.Context, g *GitHubClient, src CopilotMetricsSource, day string) ([]*github.CopilotDailyMetrics, bool, error) {
	links, found, err := dailyReportLinks(func() (*github.CopilotDailyMetricsReport, error) {
		if src.Enterprise {
			return g.GetEnterpriseCopilotDailyMetricsReport(ctx, src.Owner.Owner, day)
		}
		return g.GetOrganizationCopilotDailyMetricsReport(ctx, src.Owner.Owner, day)
	})
	if err != nil || !found {
		return nil, found, err
	}
	var result []*github.CopilotDailyMetrics
	for _, link := range links {
		m, err := g.DownloadCopilotDailyMetrics(ctx, link)
		if err != nil {
			return nil, true, fmt.Errorf("failed to download daily metrics payload for %s: %w", day, err)
		}
		if m != nil {
			result = append(result, m)
		}
	}
	return result, true, nil
}

// GetCopilotLatestUserMetrics gets the latest 28-day users usage metrics and downloads every payload.
func GetCopilotLatestUserMetrics(ctx context.Context, g *GitHubClient, src CopilotMetricsSource) (*CopilotLatestUserMetrics, error) {
	var report *github.CopilotMetricsReport
	var err error
	if src.Enterprise {
		report, err = g.GetEnterpriseCopilotUsersMetricsReport(ctx, src.Owner.Owner)
	} else {
		report, err = g.GetOrganizationCopilotUsersMetricsReport(ctx, src.Owner.Owner)
	}
	if err != nil {
		return nil, err
	}
	result := &CopilotLatestUserMetrics{}
	if report == nil {
		return result, nil
	}
	result.Window = CopilotMetricsWindow{StartDay: report.ReportStartDay, EndDay: report.ReportEndDay}
	for _, link := range report.DownloadLinks {
		records, err := g.DownloadCopilotUserPeriodicMetrics(ctx, link)
		if err != nil {
			return nil, fmt.Errorf("failed to download users metrics report payload: %w", err)
		}
		result.Records = append(result.Records, records...)
	}
	return result, nil
}

// GetCopilotDailyUserMetrics gets the 1-day users usage metrics for day (YYYY-MM-DD).
// found is false when no report exists for the day.
func GetCopilotDailyUserMetrics(ctx context.Context, g *GitHubClient, src CopilotMetricsSource, day string) ([]*github.CopilotUserDailyMetrics, bool, error) {
	links, found, err := dailyReportLinks(func() (*github.CopilotDailyMetricsReport, error) {
		if src.Enterprise {
			return g.GetEnterpriseCopilotUsersDailyMetricsReport(ctx, src.Owner.Owner, day)
		}
		return g.GetOrganizationCopilotUsersDailyMetricsReport(ctx, src.Owner.Owner, day)
	})
	if err != nil || !found {
		return nil, found, err
	}
	var result []*github.CopilotUserDailyMetrics
	for _, link := range links {
		records, err := g.DownloadCopilotUserDailyMetrics(ctx, link)
		if err != nil {
			return nil, true, fmt.Errorf("failed to download users daily metrics payload for %s: %w", day, err)
		}
		result = append(result, records...)
	}
	return result, true, nil
}

// GetCopilotDailyRepositoryMetrics gets the 1-day repository usage metrics for day (YYYY-MM-DD).
// found is false when no report exists for the day.
func GetCopilotDailyRepositoryMetrics(ctx context.Context, g *GitHubClient, src CopilotMetricsSource, day string) ([]*github.CopilotRepositoryDailyMetrics, bool, error) {
	links, found, err := dailyReportLinks(func() (*github.CopilotDailyMetricsReport, error) {
		if src.Enterprise {
			return g.GetEnterpriseCopilotRepositoriesDailyMetricsReport(ctx, src.Owner.Owner, day)
		}
		return g.GetOrganizationCopilotRepositoriesDailyMetricsReport(ctx, src.Owner.Owner, day)
	})
	if err != nil || !found {
		return nil, found, err
	}
	var result []*github.CopilotRepositoryDailyMetrics
	for _, link := range links {
		records, err := g.DownloadCopilotRepositoryDailyMetrics(ctx, link)
		if err != nil {
			return nil, true, fmt.Errorf("failed to download repository daily metrics payload for %s: %w", day, err)
		}
		result = append(result, records...)
	}
	return result, true, nil
}

// GetCopilotDailyUserTeamsMetrics gets the 1-day user-team membership records for day (YYYY-MM-DD).
// found is false when no report exists for the day.
func GetCopilotDailyUserTeamsMetrics(ctx context.Context, g *GitHubClient, src CopilotMetricsSource, day string) ([]*github.CopilotUserTeamsDailyMetrics, bool, error) {
	links, found, err := dailyReportLinks(func() (*github.CopilotDailyMetricsReport, error) {
		if src.Enterprise {
			return g.GetEnterpriseCopilotUserTeamsDailyMetricsReport(ctx, src.Owner.Owner, day)
		}
		return g.GetOrganizationCopilotUserTeamsDailyMetricsReport(ctx, src.Owner.Owner, day)
	})
	if err != nil || !found {
		return nil, found, err
	}
	var result []*github.CopilotUserTeamsDailyMetrics
	for _, link := range links {
		records, err := g.DownloadCopilotUserTeamsDailyMetrics(ctx, link)
		if err != nil {
			return nil, true, fmt.Errorf("failed to download user-teams daily metrics payload for %s: %w", day, err)
		}
		result = append(result, records...)
	}
	return result, true, nil
}
