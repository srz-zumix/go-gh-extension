package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// CodeScanningAlertStates is the list of valid state values for filtering code scanning alerts.
var CodeScanningAlertStates = []string{
	"open",
	"closed",
	"dismissed",
	"fixed",
}

// CodeScanningAlertSeverities is the list of valid severity values for code scanning alerts.
var CodeScanningAlertSeverities = []string{
	"critical",
	"high",
	"medium",
	"low",
	"warning",
	"note",
	"error",
}

// CodeScanningAlertUpdateStates is the list of valid state values for updating a code scanning alert.
var CodeScanningAlertUpdateStates = []string{
	"open",
	"dismissed",
}

// CodeScanningAlertDismissedReasons is the list of valid dismissed_reason values.
var CodeScanningAlertDismissedReasons = []string{
	"false positive",
	"won't fix",
	"used in tests",
}

// CodeScanningAlertSortOptions is the list of valid sort values for code scanning alerts.
var CodeScanningAlertSortOptions = []string{
	"created",
	"updated",
}

// ListCodeScanningAlertsOptions holds filter/sort options for listing code scanning alerts.
type ListCodeScanningAlertsOptions struct {
	State     string
	Severity  string
	ToolName  string
	ToolGUID  string
	Ref       string
	Sort      string
	Direction string
}

// toGitHubAlertListOptions converts ListCodeScanningAlertsOptions to github.AlertListOptions.
func toGitHubAlertListOptions(opts *ListCodeScanningAlertsOptions) *github.AlertListOptions {
	if opts == nil {
		return nil
	}
	o := &github.AlertListOptions{}
	if opts.State != "" {
		o.State = opts.State
	}
	if opts.Severity != "" {
		o.Severity = opts.Severity
	}
	if opts.ToolName != "" {
		o.ToolName = opts.ToolName
	}
	if opts.ToolGUID != "" {
		o.ToolGUID = opts.ToolGUID
	}
	if opts.Ref != "" {
		o.Ref = opts.Ref
	}
	if opts.Sort != "" {
		o.Sort = opts.Sort
	}
	if opts.Direction != "" {
		o.Direction = opts.Direction
	}
	return o
}

// UpdateCodeScanningAlertOptions holds options for updating a code scanning alert.
type UpdateCodeScanningAlertOptions struct {
	State            string
	DismissedReason  string
	DismissedComment string
}

// toGitHubCodeScanningAlertState converts UpdateCodeScanningAlertOptions to github.CodeScanningAlertState.
func toGitHubCodeScanningAlertState(opts *UpdateCodeScanningAlertOptions) *github.CodeScanningAlertState {
	if opts == nil {
		return nil
	}
	s := &github.CodeScanningAlertState{
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

// ListCodeScanningAlerts lists code scanning alerts for a repository.
func ListCodeScanningAlerts(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListCodeScanningAlertsOptions) ([]*github.Alert, error) {
	alerts, err := g.ListRepoCodeScanningAlerts(ctx, repo.Owner, repo.Name, toGitHubAlertListOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to list code scanning alerts for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return alerts, nil
}

// GetCodeScanningAlert gets a single code scanning alert for a repository.
func GetCodeScanningAlert(ctx context.Context, g *GitHubClient, repo repository.Repository, number int64) (*github.Alert, error) {
	alert, err := g.GetRepoCodeScanningAlert(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get code scanning alert #%d for %s/%s: %w", number, repo.Owner, repo.Name, err)
	}
	return alert, nil
}

// UpdateCodeScanningAlert updates a code scanning alert for a repository.
func UpdateCodeScanningAlert(ctx context.Context, g *GitHubClient, repo repository.Repository, number int64, opts *UpdateCodeScanningAlertOptions) (*github.Alert, error) {
	alert, err := g.UpdateRepoCodeScanningAlert(ctx, repo.Owner, repo.Name, number, toGitHubCodeScanningAlertState(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to update code scanning alert #%d for %s/%s: %w", number, repo.Owner, repo.Name, err)
	}
	return alert, nil
}
