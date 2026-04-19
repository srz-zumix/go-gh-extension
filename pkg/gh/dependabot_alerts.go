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

// ListDependabotAlertsOptions holds filter/sort options for listing Dependabot alerts.
type ListDependabotAlertsOptions struct {
	State     string
	Severity  string
	Ecosystem string
	Scope     string
	Sort      string
	Direction string
}

// toGitHubListAlertsOptions converts ListDependabotAlertsOptions to github.ListAlertsOptions.
func toGitHubListAlertsOptions(opts *ListDependabotAlertsOptions) *github.ListAlertsOptions {
	if opts == nil {
		return nil
	}
	o := &github.ListAlertsOptions{}
	if opts.State != "" {
		o.State = &opts.State
	}
	if opts.Severity != "" {
		o.Severity = &opts.Severity
	}
	if opts.Ecosystem != "" {
		o.Ecosystem = &opts.Ecosystem
	}
	if opts.Scope != "" {
		o.Scope = &opts.Scope
	}
	if opts.Sort != "" {
		o.Sort = &opts.Sort
	}
	if opts.Direction != "" {
		o.Direction = &opts.Direction
	}
	return o
}

// UpdateDependabotAlertOptions holds options for updating a Dependabot alert.
type UpdateDependabotAlertOptions struct {
	State            string
	DismissedReason  string
	DismissedComment string
}

// toGitHubDependabotAlertState converts UpdateDependabotAlertOptions to github.DependabotAlertState.
func toGitHubDependabotAlertState(opts *UpdateDependabotAlertOptions) *github.DependabotAlertState {
	if opts == nil {
		return nil
	}
	s := &github.DependabotAlertState{
		State: opts.State,
	}
	if opts.DismissedReason != "" {
		s.DismissedReason = &opts.DismissedReason
	}
	if opts.DismissedComment != "" {
		s.DismissedComment = &opts.DismissedComment
	}
	return s
}

// ListDependabotAlerts lists Dependabot alerts for a repository.
func ListDependabotAlerts(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListDependabotAlertsOptions) ([]*github.DependabotAlert, error) {
	alerts, err := g.ListRepoDependabotAlerts(ctx, repo.Owner, repo.Name, toGitHubListAlertsOptions(opts))
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

// UpdateDependabotAlert updates a Dependabot alert for a repository.
func UpdateDependabotAlert(ctx context.Context, g *GitHubClient, repo repository.Repository, number int, opts *UpdateDependabotAlertOptions) (*github.DependabotAlert, error) {
	alert, err := g.UpdateRepoDependabotAlert(ctx, repo.Owner, repo.Name, number, toGitHubDependabotAlertState(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to update Dependabot alert #%d for %s/%s: %w", number, repo.Owner, repo.Name, err)
	}
	return alert, nil
}
