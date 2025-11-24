package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

type ListWorkflowRunsOptions struct {
	Actor               string
	Branch              string
	Event               string
	Status              string
	Created             string
	HeadSHA             string
	ExcludePullRequests bool
	CheckSuiteID        int64
}

// toGitHubListWorkflowRunsOptions converts gh.ListWorkflowRunsOptions to github.ListWorkflowRunsOptions
func toGitHubListWorkflowRunsOptions(options *ListWorkflowRunsOptions) *github.ListWorkflowRunsOptions {
	if options == nil {
		return nil
	}
	opt := &github.ListWorkflowRunsOptions{
		Actor:               options.Actor,
		Branch:              options.Branch,
		Event:               options.Event,
		Status:              options.Status,
		Created:             options.Created,
		HeadSHA:             options.HeadSHA,
		ExcludePullRequests: options.ExcludePullRequests,
		CheckSuiteID:        options.CheckSuiteID,
	}
	return opt
}

// GetWorkflowRunByID retrieves a specific workflow run by its ID.
func GetWorkflowRunByID(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64) (*github.WorkflowRun, error) {
	return g.GetWorkflowRunByID(ctx, repo.Owner, repo.Name, runID)
}

// GetWorkflowRunAttempt retrieves a specific workflow run attempt.
func GetWorkflowRunAttempt(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, attemptNumber int, options *github.WorkflowRunAttemptOptions) (*github.WorkflowRun, error) {
	return g.GetWorkflowRunAttempt(ctx, repo.Owner, repo.Name, runID, attemptNumber, options)
}

// ListRepositoryWorkflowRuns retrieves all workflow runs for a repository.
func ListRepositoryWorkflowRuns(ctx context.Context, g *GitHubClient, repo repository.Repository, options *ListWorkflowRunsOptions) ([]*github.WorkflowRun, error) {
	return g.ListRepositoryWorkflowRuns(ctx, repo.Owner, repo.Name, toGitHubListWorkflowRunsOptions(options))
}

// ListWorkflowRunsByID retrieves all workflow runs for a specific workflow by workflow ID.
func ListWorkflowRunsByID(ctx context.Context, g *GitHubClient, repo repository.Repository, workflowID int64, options *ListWorkflowRunsOptions) ([]*github.WorkflowRun, error) {
	return g.ListWorkflowRunsByID(ctx, repo.Owner, repo.Name, workflowID, toGitHubListWorkflowRunsOptions(options))
}

// ListWorkflowRunsByFileName retrieves all workflow runs for a specific workflow by file name.
func ListWorkflowRunsByFileName(ctx context.Context, g *GitHubClient, repo repository.Repository, workflowFileName string, options *ListWorkflowRunsOptions) ([]*github.WorkflowRun, error) {
	return g.ListWorkflowRunsByFileName(ctx, repo.Owner, repo.Name, workflowFileName, toGitHubListWorkflowRunsOptions(options))
}

// GetWorkflowRunUsageByID retrieves the usage statistics for a specific workflow run.
func GetWorkflowRunUsageByID(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64) (*github.WorkflowRunUsage, error) {
	return g.GetWorkflowRunUsageByID(ctx, repo.Owner, repo.Name, runID)
}

// GetWorkflowRunLogs retrieves the logs URL for a workflow run.
// The maxRedirects parameter specifies the maximum number of redirects to follow.
// Setting maxRedirects to 0 will return the redirect URL without following it.
func GetWorkflowRunLogsURL(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, maxRedirects int) (string, error) {
	return g.GetWorkflowRunLogs(ctx, repo.Owner, repo.Name, runID, maxRedirects)
}

// GetWorkflowRunAttemptLogs retrieves the logs URL for a specific workflow run attempt.
// The maxRedirects parameter specifies the maximum number of redirects to follow.
// Setting maxRedirects to 0 will return the redirect URL without following it.
func GetWorkflowRunAttemptLogsURL(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, attemptNumber int, maxRedirects int) (string, error) {
	if attemptNumber <= 0 {
		return GetWorkflowRunLogsURL(ctx, g, repo, runID, maxRedirects)
	}
	return g.GetWorkflowRunAttemptLogs(ctx, repo.Owner, repo.Name, runID, attemptNumber, maxRedirects)
}

// DeleteWorkflowRun deletes a specific workflow run.
func DeleteWorkflowRun(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64) error {
	return g.DeleteWorkflowRun(ctx, repo.Owner, repo.Name, runID)
}

// DeleteWorkflowRunLogs deletes the logs for a specific workflow run.
func DeleteWorkflowRunLogs(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64) error {
	return g.DeleteWorkflowRunLogs(ctx, repo.Owner, repo.Name, runID)
}

// CancelWorkflowRunByID cancels a specific workflow run.
func CancelWorkflowRunByID(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64) error {
	return g.CancelWorkflowRunByID(ctx, repo.Owner, repo.Name, runID)
}

// RerunWorkflowByID re-runs a specific workflow run.
func RerunWorkflowByID(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64) error {
	return g.RerunWorkflowByID(ctx, repo.Owner, repo.Name, runID)
}

// ListWorkflowRunArtifacts retrieves all artifacts for a specific workflow run.
func ListWorkflowRunArtifacts(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64) ([]*github.Artifact, error) {
	return g.ListWorkflowRunArtifacts(ctx, repo.Owner, repo.Name, runID, nil)
}

// GetPendingDeployments retrieves pending deployments for a specific workflow run.
func GetPendingDeployments(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64) ([]*github.PendingDeployment, error) {
	return g.GetPendingDeployments(ctx, repo.Owner, repo.Name, runID)
}

// PendingDeployments reviews and approves or rejects pending deployments for a workflow run.
func PendingDeployments(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, environments []any, state string, comment string) ([]*github.Deployment, error) {
	environmentIDs := []int64{}
	for _, env := range environments {
		id, err := GetEnvironmentID(ctx, g, repo, env)
		if err != nil {
			return nil, err
		}
		if id != nil {
			environmentIDs = append(environmentIDs, *id)
		}
	}
	request := &github.PendingDeploymentsRequest{
		EnvironmentIDs: environmentIDs,
		State:          state,
		Comment:        comment,
	}
	return g.PendingDeployments(ctx, repo.Owner, repo.Name, runID, request)
}

// ApprovePendingDeployments approves pending deployments for a workflow run.
func ApprovePendingDeployments(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, environments []any, comment string) ([]*github.Deployment, error) {
	return PendingDeployments(ctx, g, repo, runID, environments, "approve", comment)
}

// RejectPendingDeployments rejects pending deployments for a workflow run.
func RejectPendingDeployments(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, environments []any, comment string) ([]*github.Deployment, error) {
	return PendingDeployments(ctx, g, repo, runID, environments, "reject", comment)
}

// ApproveCustomDeploymentProtectionRule reviews and approves a custom deployment protection rule.
func ApproveCustomDeploymentProtectionRule(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, environmentName string, comment string) error {
	request := &github.ReviewCustomDeploymentProtectionRuleRequest{
		EnvironmentName: environmentName,
		Comment:         comment,
		State:           "approve",
	}
	return g.ReviewCustomDeploymentProtectionRule(ctx, repo.Owner, repo.Name, runID, request)
}

func RejectCustomDeploymentProtectionRule(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, environmentName string, comment string) error {
	request := &github.ReviewCustomDeploymentProtectionRuleRequest{
		EnvironmentName: environmentName,
		Comment:         comment,
		State:           "reject",
	}
	return g.ReviewCustomDeploymentProtectionRule(ctx, repo.Owner, repo.Name, runID, request)
}
