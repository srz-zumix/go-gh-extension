package client

// GitHub Actions Runner API functions
// See: https://pkg.go.dev/github.com/google/go-github/v79/github#ActionsService

import (
	"context"

	"github.com/google/go-github/v79/github"
)

// ListRunners lists all self-hosted runners for a repository
func (g *GitHubClient) ListRunners(ctx context.Context, owner, repo string) ([]*github.Runner, error) {
	allRunners := []*github.Runner{}
	opt := &github.ListRunnersOptions{
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}
	for {
		runners, resp, err := g.client.Actions.ListRunners(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allRunners = append(allRunners, runners.Runners...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRunners, nil
}

func (g *GitHubClient) FindRunner(ctx context.Context, owner, repo string, runnerName string) (*github.Runner, error) {
	runners, err := g.ListRunners(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	for _, runner := range runners {
		if runner.GetName() == runnerName {
			return runner, nil
		}
	}
	return nil, nil // Runner not found
}

// GetRunner gets a single self-hosted runner for a repository
func (g *GitHubClient) GetRunner(ctx context.Context, owner, repo string, runnerID int64) (*github.Runner, error) {
	runner, _, err := g.client.Actions.GetRunner(ctx, owner, repo, runnerID)
	if err != nil {
		return nil, err
	}
	return runner, nil
}

// ListRunners lists all self-hosted runners for a organization
func (g *GitHubClient) ListOrgRunners(ctx context.Context, owner string) ([]*github.Runner, error) {
	allRunners := []*github.Runner{}
	opt := &github.ListRunnersOptions{
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}
	for {
		runners, resp, err := g.client.Actions.ListOrganizationRunners(ctx, owner, opt)
		if err != nil {
			return nil, err
		}
		allRunners = append(allRunners, runners.Runners...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRunners, nil
}

func (g *GitHubClient) FindOrgRunner(ctx context.Context, owner string, runnerName string) (*github.Runner, error) {
	runners, err := g.ListOrgRunners(ctx, owner)
	if err != nil {
		return nil, err
	}
	for _, runner := range runners {
		if runner.GetName() == runnerName {
			return runner, nil
		}
	}
	return nil, nil // Runner not found
}

// GetRunner gets a single self-hosted runner for a repository
func (g *GitHubClient) GetOrgRunner(ctx context.Context, owner string, runnerID int64) (*github.Runner, error) {
	runner, _, err := g.client.Actions.GetOrganizationRunner(ctx, owner, runnerID)
	if err != nil {
		return nil, err
	}
	return runner, nil
}
