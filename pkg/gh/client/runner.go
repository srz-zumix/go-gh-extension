package client

// GitHub Actions Runner API functions
// See: https://pkg.go.dev/github.com/google/go-github/v84/github#ActionsService

import (
	"context"

	"github.com/google/go-github/v84/github"
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

// CreateRegistrationToken creates a registration token for a repository
func (g *GitHubClient) CreateRegistrationToken(ctx context.Context, owner, repo string) (*github.RegistrationToken, error) {
	token, _, err := g.client.Actions.CreateRegistrationToken(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// CreateOrgRegistrationToken creates a registration token for an organization
func (g *GitHubClient) CreateOrgRegistrationToken(ctx context.Context, owner string) (*github.RegistrationToken, error) {
	token, _, err := g.client.Actions.CreateOrganizationRegistrationToken(ctx, owner)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// RemoveRunner removes a self-hosted runner from a repository
func (g *GitHubClient) RemoveRunner(ctx context.Context, owner, repo string, runnerID int64) error {
	_, err := g.client.Actions.RemoveRunner(ctx, owner, repo, runnerID)
	return err
}

// RemoveOrgRunner removes a self-hosted runner from an organization
func (g *GitHubClient) RemoveOrgRunner(ctx context.Context, owner string, runnerID int64) error {
	_, err := g.client.Actions.RemoveOrganizationRunner(ctx, owner, runnerID)
	return err
}

// GetOrgRunnerGroup gets a single organization runner group by ID
func (g *GitHubClient) GetOrgRunnerGroup(ctx context.Context, owner string, groupID int64) (*github.RunnerGroup, error) {
	group, _, err := g.client.Actions.GetOrganizationRunnerGroup(ctx, owner, groupID)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// CreateOrgRunnerGroup creates a new organization runner group
func (g *GitHubClient) CreateOrgRunnerGroup(ctx context.Context, owner string, name string) (*github.RunnerGroup, error) {
	group, _, err := g.client.Actions.CreateOrganizationRunnerGroup(ctx, owner, github.CreateRunnerGroupRequest{
		Name: github.Ptr(name),
	})
	if err != nil {
		return nil, err
	}
	return group, nil
}

// ListOrgRunnerGroups lists all organization runner groups
func (g *GitHubClient) ListOrgRunnerGroups(ctx context.Context, owner string) ([]*github.RunnerGroup, error) {
	allGroups := []*github.RunnerGroup{}
	opt := &github.ListOrgRunnerGroupOptions{
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}
	for {
		groups, resp, err := g.client.Actions.ListOrganizationRunnerGroups(ctx, owner, opt)
		if err != nil {
			return nil, err
		}
		allGroups = append(allGroups, groups.RunnerGroups...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allGroups, nil
}

// FindOrgRunnerGroup finds an organization runner group by name
func (g *GitHubClient) FindOrgRunnerGroup(ctx context.Context, owner string, groupName string) (*github.RunnerGroup, error) {
	groups, err := g.ListOrgRunnerGroups(ctx, owner)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if group.GetName() == groupName {
			return group, nil
		}
	}
	return nil, nil
}

// DeleteOrgRunnerGroup deletes an organization runner group by ID
func (g *GitHubClient) DeleteOrgRunnerGroup(ctx context.Context, owner string, groupID int64) error {
	_, err := g.client.Actions.DeleteOrganizationRunnerGroup(ctx, owner, groupID)
	return err
}
