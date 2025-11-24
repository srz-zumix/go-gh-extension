package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

// GetWorkflowJobByID retrieves a specific workflow job by its ID.
func GetWorkflowJobByID(ctx context.Context, g *GitHubClient, repo repository.Repository, jobID int64) (*github.WorkflowJob, error) {
	return g.GetWorkflowJobByID(ctx, repo.Owner, repo.Name, jobID)
}

// ListWorkflowJobs retrieves all workflow jobs for a specific workflow run.
func ListWorkflowJobs(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, filter *string) ([]*github.WorkflowJob, error) {
	options := &github.ListWorkflowJobsOptions{}
	if filter != nil {
		options.Filter = *filter
	}
	return g.ListWorkflowJobs(ctx, repo.Owner, repo.Name, runID, options)
}

// ListWorkflowJobsAttempt retrieves all workflow jobs for a specific workflow run attempt.
func ListWorkflowJobsAttempt(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64, attemptNumber int64) ([]*github.WorkflowJob, error) {
	return g.ListWorkflowJobsAttempt(ctx, repo.Owner, repo.Name, runID, attemptNumber, nil)
}

// RerunJobByID re-runs a specific workflow job.
func RerunJobByID(ctx context.Context, g *GitHubClient, repo repository.Repository, jobID int64) error {
	return g.RerunJobByID(ctx, repo.Owner, repo.Name, jobID)
}

// RerunFailedJobsByID re-runs all failed jobs in a workflow run.
func RerunFailedJobsByID(ctx context.Context, g *GitHubClient, repo repository.Repository, runID int64) error {
	return g.RerunFailedJobsByID(ctx, repo.Owner, repo.Name, runID)
}

// GetWorkflowJobLogsURL retrieves the logs URL for a workflow job.
// Note: This function attempts to retrieve logs by treating the checkRunID as a workflow job ID.
// In GitHub Actions, check runs are associated with workflow jobs that have the same ID.
// The maxRedirects parameter specifies the maximum number of redirects to follow.
// Setting maxRedirects to 0 will return the redirect URL without following it.
func GetWorkflowJobLogsURL(ctx context.Context, g *GitHubClient, repo repository.Repository, jobID int64, maxRedirects int) (string, error) {
	return g.GetWorkflowJobLogs(ctx, repo.Owner, repo.Name, jobID, maxRedirects)
}

// GetWorkflowJobLogsContent retrieves the logs content for a workflow job.
// This function downloads the logs from the URL obtained by GetWorkflowJobLogsURL.
func GetWorkflowJobLogsContent(ctx context.Context, g *GitHubClient, repo repository.Repository, jobID int64, maxRedirects int) ([]byte, error) {
	return g.GetWorkflowJobLogsContent(ctx, repo.Owner, repo.Name, jobID, maxRedirects)
}
