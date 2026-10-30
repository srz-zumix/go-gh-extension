package client

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/google/go-github/v93/github"
)

// copilotMetricsRequestAttempts is the maximum number of attempts for a single metrics report request.
const copilotMetricsRequestAttempts = 3

// copilotMetricsRetryBackoff returns the wait before the given retry (1-based). It is a variable
// so tests can shorten it.
var copilotMetricsRetryBackoff = func(retry int) time.Duration {
	return time.Duration(retry) * 2 * time.Second
}

// isRetryableCopilotMetricsError reports whether a metrics report request failure is transient:
// a client timeout (the server occasionally stalls before sending response headers), an
// interrupted body, or a throttling or server error response.
func isRetryableCopilotMetricsError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		return true
	}
	var errResp *github.ErrorResponse
	if errors.As(err, &errResp) && errResp.Response != nil {
		code := errResp.Response.StatusCode
		return code == http.StatusTooManyRequests || code >= http.StatusInternalServerError
	}
	return false
}

// retryCopilotMetrics runs a metrics report request and retries transient failures.
// The report endpoints and the storage backend occasionally stall, so idempotent GETs are retried.
func retryCopilotMetrics[T any](ctx context.Context, request func() (T, *github.Response, error)) (T, error) {
	var zero T
	for attempt := 1; ; attempt++ {
		m, _, err := request()
		if err == nil {
			return m, nil
		}
		if attempt >= copilotMetricsRequestAttempts || ctx.Err() != nil || !isRetryableCopilotMetricsError(err) {
			return zero, err
		}
		timer := time.NewTimer(copilotMetricsRetryBackoff(attempt))
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, err
		case <-timer.C:
		}
	}
}

// downloadCopilotMetrics runs a payload download with the unauthenticated download client and
// retries transient failures.
func downloadCopilotMetrics[T any](ctx context.Context, g *GitHubClient, download func(*github.Client) (T, *github.Response, error)) (T, error) {
	c, err := g.copilotMetricsDownloadClient()
	if err != nil {
		var zero T
		return zero, err
	}
	return retryCopilotMetrics(ctx, func() (T, *github.Response, error) { return download(c) })
}

// copilotMetricsDownloadClient returns a go-github client for downloading report payloads.
// Download links are pre-signed storage URLs, and the storage backend may reject a request
// that also carries the GitHub Authorization header, so payloads are fetched without the
// GitHub credentials.
func (g *GitHubClient) copilotMetricsDownloadClient() (*github.Client, error) {
	clone := *g.client.Client()
	clone.Transport = g.rawHTTPTransport()
	return github.NewClient(github.WithHTTPClient(&clone))
}

// GetOrganizationCopilotMetricsReport gets the latest 28-day organization-level usage metrics report links.
func (g *GitHubClient) GetOrganizationCopilotMetricsReport(ctx context.Context, org string) (*github.CopilotMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetOrganizationMetricsReport(ctx, org)
	})
}

// GetEnterpriseCopilotMetricsReport gets the latest 28-day enterprise-level usage metrics report links.
func (g *GitHubClient) GetEnterpriseCopilotMetricsReport(ctx context.Context, enterprise string) (*github.CopilotMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetEnterpriseMetricsReport(ctx, enterprise)
	})
}

// GetOrganizationCopilotDailyMetricsReport gets the 1-day organization-level usage metrics report links.
func (g *GitHubClient) GetOrganizationCopilotDailyMetricsReport(ctx context.Context, org, day string) (*github.CopilotDailyMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotDailyMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetOrganizationDailyMetricsReport(ctx, org, &github.CopilotMetricsReportOptions{Day: day})
	})
}

// GetEnterpriseCopilotDailyMetricsReport gets the 1-day enterprise-level usage metrics report links.
func (g *GitHubClient) GetEnterpriseCopilotDailyMetricsReport(ctx context.Context, enterprise, day string) (*github.CopilotDailyMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotDailyMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetEnterpriseDailyMetricsReport(ctx, enterprise, &github.CopilotMetricsReportOptions{Day: day})
	})
}

// GetOrganizationCopilotUsersMetricsReport gets the latest 28-day users usage metrics report links of an organization.
func (g *GitHubClient) GetOrganizationCopilotUsersMetricsReport(ctx context.Context, org string) (*github.CopilotMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetOrganizationUsersMetricsReport(ctx, org)
	})
}

// GetEnterpriseCopilotUsersMetricsReport gets the latest 28-day users usage metrics report links of an enterprise.
func (g *GitHubClient) GetEnterpriseCopilotUsersMetricsReport(ctx context.Context, enterprise string) (*github.CopilotMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetEnterpriseUsersMetricsReport(ctx, enterprise)
	})
}

// GetOrganizationCopilotUsersDailyMetricsReport gets the 1-day users usage metrics report links of an organization.
func (g *GitHubClient) GetOrganizationCopilotUsersDailyMetricsReport(ctx context.Context, org, day string) (*github.CopilotDailyMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotDailyMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetOrganizationUsersDailyMetricsReport(ctx, org, &github.CopilotMetricsReportOptions{Day: day})
	})
}

// GetEnterpriseCopilotUsersDailyMetricsReport gets the 1-day users usage metrics report links of an enterprise.
func (g *GitHubClient) GetEnterpriseCopilotUsersDailyMetricsReport(ctx context.Context, enterprise, day string) (*github.CopilotDailyMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotDailyMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetEnterpriseUsersDailyMetricsReport(ctx, enterprise, &github.CopilotMetricsReportOptions{Day: day})
	})
}

// GetOrganizationCopilotRepositoriesDailyMetricsReport gets the 1-day repository usage metrics report links of an organization.
func (g *GitHubClient) GetOrganizationCopilotRepositoriesDailyMetricsReport(ctx context.Context, org, day string) (*github.CopilotDailyMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotDailyMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetOrganizationRepositoriesDailyMetricsReport(ctx, org, &github.CopilotMetricsReportOptions{Day: day})
	})
}

// GetEnterpriseCopilotRepositoriesDailyMetricsReport gets the 1-day repository usage metrics report links of an enterprise.
func (g *GitHubClient) GetEnterpriseCopilotRepositoriesDailyMetricsReport(ctx context.Context, enterprise, day string) (*github.CopilotDailyMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotDailyMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetEnterpriseRepositoriesDailyMetricsReport(ctx, enterprise, &github.CopilotMetricsReportOptions{Day: day})
	})
}

// GetOrganizationCopilotUserTeamsDailyMetricsReport gets the 1-day user-team membership report links of an organization.
func (g *GitHubClient) GetOrganizationCopilotUserTeamsDailyMetricsReport(ctx context.Context, org, day string) (*github.CopilotDailyMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotDailyMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetOrganizationUserTeamsDailyMetricsReport(ctx, org, &github.CopilotMetricsReportOptions{Day: day})
	})
}

// GetEnterpriseCopilotUserTeamsDailyMetricsReport gets the 1-day user-team membership report links of an enterprise.
func (g *GitHubClient) GetEnterpriseCopilotUserTeamsDailyMetricsReport(ctx context.Context, enterprise, day string) (*github.CopilotDailyMetricsReport, error) {
	return retryCopilotMetrics(ctx, func() (*github.CopilotDailyMetricsReport, *github.Response, error) {
		return g.client.Copilot.GetEnterpriseUserTeamsDailyMetricsReport(ctx, enterprise, &github.CopilotMetricsReportOptions{Day: day})
	})
}

// DownloadCopilotPeriodicMetrics downloads a multi-day org/enterprise-level metrics payload.
func (g *GitHubClient) DownloadCopilotPeriodicMetrics(ctx context.Context, url string) (*github.CopilotPeriodicMetrics, error) {
	return downloadCopilotMetrics(ctx, g, func(c *github.Client) (*github.CopilotPeriodicMetrics, *github.Response, error) {
		return c.Copilot.DownloadPeriodicMetrics(ctx, url)
	})
}

// DownloadCopilotDailyMetrics downloads a 1-day org/enterprise-level metrics payload.
func (g *GitHubClient) DownloadCopilotDailyMetrics(ctx context.Context, url string) (*github.CopilotDailyMetrics, error) {
	return downloadCopilotMetrics(ctx, g, func(c *github.Client) (*github.CopilotDailyMetrics, *github.Response, error) {
		return c.Copilot.DownloadDailyMetrics(ctx, url)
	})
}

// DownloadCopilotUserPeriodicMetrics downloads a multi-day users metrics payload.
func (g *GitHubClient) DownloadCopilotUserPeriodicMetrics(ctx context.Context, url string) ([]*github.CopilotUserPeriodicMetrics, error) {
	return downloadCopilotMetrics(ctx, g, func(c *github.Client) ([]*github.CopilotUserPeriodicMetrics, *github.Response, error) {
		return c.Copilot.DownloadUserPeriodicMetrics(ctx, url)
	})
}

// DownloadCopilotUserDailyMetrics downloads a 1-day users metrics payload.
func (g *GitHubClient) DownloadCopilotUserDailyMetrics(ctx context.Context, url string) ([]*github.CopilotUserDailyMetrics, error) {
	return downloadCopilotMetrics(ctx, g, func(c *github.Client) ([]*github.CopilotUserDailyMetrics, *github.Response, error) {
		return c.Copilot.DownloadUserDailyMetrics(ctx, url)
	})
}

// DownloadCopilotRepositoryDailyMetrics downloads a 1-day repository metrics payload.
func (g *GitHubClient) DownloadCopilotRepositoryDailyMetrics(ctx context.Context, url string) ([]*github.CopilotRepositoryDailyMetrics, error) {
	return downloadCopilotMetrics(ctx, g, func(c *github.Client) ([]*github.CopilotRepositoryDailyMetrics, *github.Response, error) {
		return c.Copilot.DownloadRepositoryDailyMetrics(ctx, url)
	})
}

// DownloadCopilotUserTeamsDailyMetrics downloads a 1-day user-team membership payload.
func (g *GitHubClient) DownloadCopilotUserTeamsDailyMetrics(ctx context.Context, url string) ([]*github.CopilotUserTeamsDailyMetrics, error) {
	return downloadCopilotMetrics(ctx, g, func(c *github.Client) ([]*github.CopilotUserTeamsDailyMetrics, *github.Response, error) {
		return c.Copilot.DownloadUserTeamsDailyMetrics(ctx, url)
	})
}
