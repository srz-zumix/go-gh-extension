package gh

import (
	"context"
	"time"

	"github.com/google/go-github/v73/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// GetCopilotTeamMetrics is a wrapper to get Copilot metrics for a team
func GetCopilotTeamMetrics(ctx context.Context, g *client.GitHubClient, org, teamSlug string, since, until *time.Time) ([]*github.CopilotMetrics, error) {
	return g.GetCopilotTeamMetrics(ctx, org, teamSlug, since, until)
}

// GetEnterpriseTeamMetrics is a wrapper to get Copilot metrics for an enterprise team
func GetEnterpriseTeamMetrics(ctx context.Context, g *client.GitHubClient, enterprise, teamSlug string, since, until *time.Time) ([]*github.CopilotMetrics, error) {
	return g.GetEnterpriseTeamMetrics(ctx, enterprise, teamSlug, since, until)
}
