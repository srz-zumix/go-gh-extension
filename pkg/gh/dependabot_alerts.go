package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// DependabotAlertEcosystems is the list of valid ecosystem values for Dependabot alerts.
var DependabotAlertEcosystems = []string{
	"composer",
	"go",
	"maven",
	"npm",
	"nuget",
	"pip",
	"pub",
	"rubygems",
	"rust",
}

// DependabotAlertStates is the list of valid state values for Dependabot alerts.
var DependabotAlertStates = []string{
	"auto_dismissed",
	"dismissed",
	"fixed",
	"open",
}

// DependabotAlertSeverities is the list of valid severity values for Dependabot alerts.
var DependabotAlertSeverities = []string{
	"low",
	"medium",
	"high",
	"critical",
}

// DependabotAlertScopes is the list of valid scope values for Dependabot alerts.
var DependabotAlertScopes = []string{
	"development",
	"runtime",
}

// DependabotAlertSortOptions is the list of valid sort values for Dependabot alerts.
var DependabotAlertSortOptions = []string{
	"created",
	"updated",
	"epss_percentage",
}

// ListDependabotAlerts lists Dependabot alerts for a repository.
func ListDependabotAlerts(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *github.ListAlertsOptions) ([]*github.DependabotAlert, error) {
	alerts, err := g.ListRepoDependabotAlerts(ctx, repo.Owner, repo.Name, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list Dependabot alerts for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return alerts, nil
}

// GetDependabotAlert gets a single Dependabot alert for a repository.
func GetDependabotAlert(ctx context.Context, g *GitHubClient, repo repository.Repository, number int) (*github.DependabotAlert, error) {
	alert, err := g.GetRepoDependabotAlert(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get Dependabot alert #%d for %s/%s: %w", number, repo.Owner, repo.Name, err)
	}
	return alert, nil
}

// DependabotAlertUpdateStates is the list of valid state values for updating a Dependabot alert.
var DependabotAlertUpdateStates = []string{
	"dismissed",
	"open",
}

// DependabotAlertDismissedReasons is the list of valid dismissed_reason values.
var DependabotAlertDismissedReasons = []string{
	"fix_started",
	"inaccurate",
	"no_bandwidth",
	"not_used",
	"tolerable_risk",
}

// UpdateDependabotAlert updates a Dependabot alert for a repository.
func UpdateDependabotAlert(ctx context.Context, g *GitHubClient, repo repository.Repository, number int, stateInfo *github.DependabotAlertState) (*github.DependabotAlert, error) {
	alert, err := g.UpdateRepoDependabotAlert(ctx, repo.Owner, repo.Name, number, stateInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to update Dependabot alert #%d for %s/%s: %w", number, repo.Owner, repo.Name, err)
	}
	return alert, nil
}
