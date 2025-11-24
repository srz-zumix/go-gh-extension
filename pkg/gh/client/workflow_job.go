package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v79/github"
)

// GetWorkflowJobByID retrieves a specific workflow job by its ID.
func (g *GitHubClient) GetWorkflowJobByID(ctx context.Context, owner string, repo string, jobID int64) (*github.WorkflowJob, error) {
	job, _, err := g.client.Actions.GetWorkflowJobByID(ctx, owner, repo, jobID)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// ListWorkflowJobs retrieves all workflow jobs for a specific workflow run.
func (g *GitHubClient) ListWorkflowJobs(ctx context.Context, owner string, repo string, runID int64, options *github.ListWorkflowJobsOptions) ([]*github.WorkflowJob, error) {
	opt := github.ListWorkflowJobsOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allJobs []*github.WorkflowJob
	for {
		jobs, resp, err := g.client.Actions.ListWorkflowJobs(ctx, owner, repo, runID, &opt)
		if err != nil {
			return nil, err
		}
		allJobs = append(allJobs, jobs.Jobs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allJobs, nil
}

// ListWorkflowJobsAttempt retrieves all workflow jobs for a specific workflow run attempt.
func (g *GitHubClient) ListWorkflowJobsAttempt(ctx context.Context, owner string, repo string, runID int64, attemptNumber int64, options *github.ListOptions) ([]*github.WorkflowJob, error) {
	opt := github.ListOptions{PerPage: defaultPerPage}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allJobs []*github.WorkflowJob
	for {
		jobs, resp, err := g.client.Actions.ListWorkflowJobsAttempt(ctx, owner, repo, runID, attemptNumber, &opt)
		if err != nil {
			return nil, err
		}
		allJobs = append(allJobs, jobs.Jobs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allJobs, nil
}

// RerunJobByID re-runs a specific workflow job.
func (g *GitHubClient) RerunJobByID(ctx context.Context, owner string, repo string, jobID int64) error {
	_, err := g.client.Actions.RerunJobByID(ctx, owner, repo, jobID)
	return err
}

// RerunFailedJobsByID re-runs all failed jobs in a workflow run.
func (g *GitHubClient) RerunFailedJobsByID(ctx context.Context, owner string, repo string, runID int64) error {
	_, err := g.client.Actions.RerunFailedJobsByID(ctx, owner, repo, runID)
	return err
}

// GetWorkflowJobLogs retrieves the logs URL for a workflow job.
// The maxRedirects parameter specifies the maximum number of redirects to follow.
// Setting maxRedirects to 0 will return the redirect URL without following it.
func (g *GitHubClient) GetWorkflowJobLogs(ctx context.Context, owner string, repo string, jobID int64, maxRedirects int) (string, error) {
	logURL, _, err := g.client.Actions.GetWorkflowJobLogs(ctx, owner, repo, jobID, maxRedirects)
	if err != nil {
		return "", err
	}
	return logURL.String(), nil
}

// GetWorkflowJobLogsContent retrieves the logs content for a workflow job.
// This function downloads the logs from the URL and returns the content as bytes.
func (g *GitHubClient) GetWorkflowJobLogsContent(ctx context.Context, owner string, repo string, jobID int64, maxRedirects int) ([]byte, error) {
	logURL, _, err := g.client.Actions.GetWorkflowJobLogs(ctx, owner, repo, jobID, maxRedirects)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Get(logURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download logs: status code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
