package client

import (
	"context"

	"github.com/google/go-github/v79/github"
)

// GetWorkflowRunByID retrieves a specific workflow run by its ID.
func (g *GitHubClient) GetWorkflowRunByID(ctx context.Context, owner string, repo string, runID int64) (*github.WorkflowRun, error) {
	run, _, err := g.client.Actions.GetWorkflowRunByID(ctx, owner, repo, runID)
	if err != nil {
		return nil, err
	}
	return run, nil
}

// GetWorkflowRunAttempt retrieves a specific workflow run attempt.
func (g *GitHubClient) GetWorkflowRunAttempt(ctx context.Context, owner string, repo string, runID int64, attemptNumber int, options *github.WorkflowRunAttemptOptions) (*github.WorkflowRun, error) {
	run, _, err := g.client.Actions.GetWorkflowRunAttempt(ctx, owner, repo, runID, attemptNumber, options)
	if err != nil {
		return nil, err
	}
	return run, nil
}

// ListRepositoryWorkflowRuns retrieves all workflow runs for a repository.
func (g *GitHubClient) ListRepositoryWorkflowRuns(ctx context.Context, owner string, repo string, options *github.ListWorkflowRunsOptions) ([]*github.WorkflowRun, error) {
	opt := github.ListWorkflowRunsOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allRuns []*github.WorkflowRun
	for {
		runs, resp, err := g.client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, &opt)
		if err != nil {
			return nil, err
		}
		allRuns = append(allRuns, runs.WorkflowRuns...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRuns, nil
}

// ListWorkflowRunsByID retrieves all workflow runs for a specific workflow by workflow ID.
func (g *GitHubClient) ListWorkflowRunsByID(ctx context.Context, owner string, repo string, workflowID int64, options *github.ListWorkflowRunsOptions) ([]*github.WorkflowRun, error) {
	opt := github.ListWorkflowRunsOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allRuns []*github.WorkflowRun
	for {
		runs, resp, err := g.client.Actions.ListWorkflowRunsByID(ctx, owner, repo, workflowID, &opt)
		if err != nil {
			return nil, err
		}
		allRuns = append(allRuns, runs.WorkflowRuns...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRuns, nil
}

// ListWorkflowRunsByFileName retrieves all workflow runs for a specific workflow by file name.
func (g *GitHubClient) ListWorkflowRunsByFileName(ctx context.Context, owner string, repo string, workflowFileName string, options *github.ListWorkflowRunsOptions) ([]*github.WorkflowRun, error) {
	opt := github.ListWorkflowRunsOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allRuns []*github.WorkflowRun
	for {
		runs, resp, err := g.client.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, workflowFileName, &opt)
		if err != nil {
			return nil, err
		}
		allRuns = append(allRuns, runs.WorkflowRuns...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRuns, nil
}

// GetWorkflowRunUsageByID retrieves the usage statistics for a specific workflow run.
func (g *GitHubClient) GetWorkflowRunUsageByID(ctx context.Context, owner string, repo string, runID int64) (*github.WorkflowRunUsage, error) {
	usage, _, err := g.client.Actions.GetWorkflowRunUsageByID(ctx, owner, repo, runID)
	if err != nil {
		return nil, err
	}
	return usage, nil
}

// GetWorkflowRunLogs retrieves the logs URL for a workflow run.
// The maxRedirects parameter specifies the maximum number of redirects to follow.
// Setting maxRedirects to 0 will return the redirect URL without following it.
func (g *GitHubClient) GetWorkflowRunLogs(ctx context.Context, owner string, repo string, runID int64, maxRedirects int) (string, error) {
	logURL, _, err := g.client.Actions.GetWorkflowRunLogs(ctx, owner, repo, runID, maxRedirects)
	if err != nil {
		return "", err
	}
	return logURL.String(), nil
}

// GetWorkflowRunAttemptLogs retrieves the logs URL for a specific workflow run attempt.
// The maxRedirects parameter specifies the maximum number of redirects to follow.
// Setting maxRedirects to 0 will return the redirect URL without following it.
func (g *GitHubClient) GetWorkflowRunAttemptLogs(ctx context.Context, owner string, repo string, runID int64, attemptNumber int, maxRedirects int) (string, error) {
	logURL, _, err := g.client.Actions.GetWorkflowRunAttemptLogs(ctx, owner, repo, runID, attemptNumber, maxRedirects)
	if err != nil {
		return "", err
	}
	return logURL.String(), nil
}

// DeleteWorkflowRun deletes a specific workflow run.
func (g *GitHubClient) DeleteWorkflowRun(ctx context.Context, owner string, repo string, runID int64) error {
	_, err := g.client.Actions.DeleteWorkflowRun(ctx, owner, repo, runID)
	return err
}

// DeleteWorkflowRunLogs deletes the logs for a specific workflow run.
func (g *GitHubClient) DeleteWorkflowRunLogs(ctx context.Context, owner string, repo string, runID int64) error {
	_, err := g.client.Actions.DeleteWorkflowRunLogs(ctx, owner, repo, runID)
	return err
}

// CancelWorkflowRunByID cancels a specific workflow run.
func (g *GitHubClient) CancelWorkflowRunByID(ctx context.Context, owner string, repo string, runID int64) error {
	_, err := g.client.Actions.CancelWorkflowRunByID(ctx, owner, repo, runID)
	return err
}

// RerunWorkflowByID re-runs a specific workflow run.
func (g *GitHubClient) RerunWorkflowByID(ctx context.Context, owner string, repo string, runID int64) error {
	_, err := g.client.Actions.RerunWorkflowByID(ctx, owner, repo, runID)
	return err
}

// ListWorkflowRunArtifacts retrieves all artifacts for a specific workflow run.
func (g *GitHubClient) ListWorkflowRunArtifacts(ctx context.Context, owner string, repo string, runID int64, options *github.ListOptions) ([]*github.Artifact, error) {
	opt := github.ListOptions{PerPage: defaultPerPage}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allArtifacts []*github.Artifact
	for {
		artifacts, resp, err := g.client.Actions.ListWorkflowRunArtifacts(ctx, owner, repo, runID, &opt)
		if err != nil {
			return nil, err
		}
		allArtifacts = append(allArtifacts, artifacts.Artifacts...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allArtifacts, nil
}

// GetPendingDeployments retrieves pending deployments for a specific workflow run.
func (g *GitHubClient) GetPendingDeployments(ctx context.Context, owner string, repo string, runID int64) ([]*github.PendingDeployment, error) {
	deployments, _, err := g.client.Actions.GetPendingDeployments(ctx, owner, repo, runID)
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

// PendingDeployments reviews and approves or rejects pending deployments for a workflow run.
func (g *GitHubClient) PendingDeployments(ctx context.Context, owner string, repo string, runID int64, request *github.PendingDeploymentsRequest) ([]*github.Deployment, error) {
	deployments, _, err := g.client.Actions.PendingDeployments(ctx, owner, repo, runID, request)
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

// ReviewCustomDeploymentProtectionRule reviews and approves or rejects a custom deployment protection rule.
func (g *GitHubClient) ReviewCustomDeploymentProtectionRule(ctx context.Context, owner string, repo string, runID int64, request *github.ReviewCustomDeploymentProtectionRuleRequest) error {
	_, err := g.client.Actions.ReviewCustomDeploymentProtectionRule(ctx, owner, repo, runID, request)
	return err
}
