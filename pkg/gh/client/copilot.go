package client

import (
	"context"
	"time"

	"github.com/google/go-github/v73/github"
)

// GetCopilotTeamMetrics retrieves Copilot metrics for a team via REST API (not supported by go-github)
func (g *GitHubClient) GetCopilotTeamMetrics(ctx context.Context, org, teamSlug string, since, until *time.Time) ([]*github.CopilotMetrics, error) {
	allMetrics := []*github.CopilotMetrics{}
	opt := &github.CopilotMetricsListOptions{
		Since: since,
		Until: until,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	for {
		result, resp, err := g.client.Copilot.GetOrganizationTeamMetrics(ctx, org, teamSlug, opt)
		if err != nil {
			return nil, err
		}
		allMetrics = append(allMetrics, result...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allMetrics, nil
}

func (g *GitHubClient) GetEnterpriseTeamMetrics(ctx context.Context, enterprise, teamSlug string, since, until *time.Time) ([]*github.CopilotMetrics, error) {
	allMetrics := []*github.CopilotMetrics{}
	opt := &github.CopilotMetricsListOptions{
		Since: since,
		Until: until,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	for {
		result, resp, err := g.client.Copilot.GetEnterpriseTeamMetrics(ctx, enterprise, teamSlug, opt)
		if err != nil {
			return nil, err
		}
		allMetrics = append(allMetrics, result...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allMetrics, nil
}
