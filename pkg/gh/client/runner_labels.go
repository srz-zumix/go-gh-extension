package client

// GitHub Actions Runner Labels API functions
// go-github v90 does not wrap these endpoints, so requests are issued directly.
// See: https://docs.github.com/rest/actions/self-hosted-runners#about-self-hosted-runners-in-github-actions

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/v90/github"
)

// RunnerLabelsResponse represents the response body returned by the runner labels endpoints.
type RunnerLabelsResponse struct {
	TotalCount int                     `json:"total_count"`
	Labels     []*github.RunnerLabels `json:"labels"`
}

// setRunnerLabelsRequest represents the request body for adding or replacing runner labels.
type setRunnerLabelsRequest struct {
	Labels []string `json:"labels"`
}

func (g *GitHubClient) runnerLabelsRequest(ctx context.Context, method, u string, body any) ([]*github.RunnerLabels, error) {
	req, err := g.client.NewRequest(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	result := new(RunnerLabelsResponse)
	if _, err := g.client.Do(req, result); err != nil {
		return nil, err
	}
	return result.Labels, nil
}

// ListRunnerLabels lists all labels for a self-hosted runner in a repository.
func (g *GitHubClient) ListRunnerLabels(ctx context.Context, owner, repo string, runnerID int64) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("repos/%s/%s/actions/runners/%d/labels", owner, repo, runnerID)
	return g.runnerLabelsRequest(ctx, "GET", u, nil)
}

// AddRunnerLabels adds custom labels to a self-hosted runner in a repository.
func (g *GitHubClient) AddRunnerLabels(ctx context.Context, owner, repo string, runnerID int64, labels []string) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("repos/%s/%s/actions/runners/%d/labels", owner, repo, runnerID)
	return g.runnerLabelsRequest(ctx, "POST", u, setRunnerLabelsRequest{Labels: labels})
}

// SetRunnerLabels replaces all custom labels for a self-hosted runner in a repository.
func (g *GitHubClient) SetRunnerLabels(ctx context.Context, owner, repo string, runnerID int64, labels []string) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("repos/%s/%s/actions/runners/%d/labels", owner, repo, runnerID)
	return g.runnerLabelsRequest(ctx, "PUT", u, setRunnerLabelsRequest{Labels: labels})
}

// RemoveAllRunnerLabels removes all custom labels from a self-hosted runner in a repository.
func (g *GitHubClient) RemoveAllRunnerLabels(ctx context.Context, owner, repo string, runnerID int64) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("repos/%s/%s/actions/runners/%d/labels", owner, repo, runnerID)
	return g.runnerLabelsRequest(ctx, "DELETE", u, nil)
}

// RemoveRunnerLabel removes a single custom label from a self-hosted runner in a repository.
func (g *GitHubClient) RemoveRunnerLabel(ctx context.Context, owner, repo string, runnerID int64, name string) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("repos/%s/%s/actions/runners/%d/labels/%s", owner, repo, runnerID, url.PathEscape(name))
	return g.runnerLabelsRequest(ctx, "DELETE", u, nil)
}

// ListOrgRunnerLabels lists all labels for a self-hosted runner in an organization.
func (g *GitHubClient) ListOrgRunnerLabels(ctx context.Context, owner string, runnerID int64) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("orgs/%s/actions/runners/%d/labels", owner, runnerID)
	return g.runnerLabelsRequest(ctx, "GET", u, nil)
}

// AddOrgRunnerLabels adds custom labels to a self-hosted runner in an organization.
func (g *GitHubClient) AddOrgRunnerLabels(ctx context.Context, owner string, runnerID int64, labels []string) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("orgs/%s/actions/runners/%d/labels", owner, runnerID)
	return g.runnerLabelsRequest(ctx, "POST", u, setRunnerLabelsRequest{Labels: labels})
}

// SetOrgRunnerLabels replaces all custom labels for a self-hosted runner in an organization.
func (g *GitHubClient) SetOrgRunnerLabels(ctx context.Context, owner string, runnerID int64, labels []string) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("orgs/%s/actions/runners/%d/labels", owner, runnerID)
	return g.runnerLabelsRequest(ctx, "PUT", u, setRunnerLabelsRequest{Labels: labels})
}

// RemoveAllOrgRunnerLabels removes all custom labels from a self-hosted runner in an organization.
func (g *GitHubClient) RemoveAllOrgRunnerLabels(ctx context.Context, owner string, runnerID int64) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("orgs/%s/actions/runners/%d/labels", owner, runnerID)
	return g.runnerLabelsRequest(ctx, "DELETE", u, nil)
}

// RemoveOrgRunnerLabel removes a single custom label from a self-hosted runner in an organization.
func (g *GitHubClient) RemoveOrgRunnerLabel(ctx context.Context, owner string, runnerID int64, name string) ([]*github.RunnerLabels, error) {
	u := fmt.Sprintf("orgs/%s/actions/runners/%d/labels/%s", owner, runnerID, url.PathEscape(name))
	return g.runnerLabelsRequest(ctx, "DELETE", u, nil)
}
