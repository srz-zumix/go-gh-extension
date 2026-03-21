package client

// GitHub Actions Variables API functions
// See: https://docs.github.com/rest/actions/variables

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListRepoVariables lists all variables in a repository.
func (g *GitHubClient) ListRepoVariables(ctx context.Context, owner, repo string) ([]*github.ActionsVariable, error) {
	var all []*github.ActionsVariable
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		vars, resp, err := g.client.Actions.ListRepoVariables(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, vars.Variables...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return all, nil
}

// ListOrgVariables lists all variables in an organization.
func (g *GitHubClient) ListOrgVariables(ctx context.Context, org string) ([]*github.ActionsVariable, error) {
	var all []*github.ActionsVariable
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		vars, resp, err := g.client.Actions.ListOrgVariables(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, vars.Variables...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return all, nil
}

// GetRepoVariable gets a single repository variable.
func (g *GitHubClient) GetRepoVariable(ctx context.Context, owner, repo, name string) (*github.ActionsVariable, error) {
	v, _, err := g.client.Actions.GetRepoVariable(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetOrgVariable gets a single organization variable.
func (g *GitHubClient) GetOrgVariable(ctx context.Context, org, name string) (*github.ActionsVariable, error) {
	v, _, err := g.client.Actions.GetOrgVariable(ctx, org, name)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// CreateRepoVariable creates a repository variable.
func (g *GitHubClient) CreateRepoVariable(ctx context.Context, owner, repo string, variable *github.ActionsVariable) error {
	_, err := g.client.Actions.CreateRepoVariable(ctx, owner, repo, variable)
	return err
}

// UpdateRepoVariable updates a repository variable.
func (g *GitHubClient) UpdateRepoVariable(ctx context.Context, owner, repo string, variable *github.ActionsVariable) error {
	_, err := g.client.Actions.UpdateRepoVariable(ctx, owner, repo, variable)
	return err
}

// CreateOrgVariable creates an organization variable.
func (g *GitHubClient) CreateOrgVariable(ctx context.Context, org string, variable *github.ActionsVariable) error {
	_, err := g.client.Actions.CreateOrgVariable(ctx, org, variable)
	return err
}

// UpdateOrgVariable updates an organization variable.
func (g *GitHubClient) UpdateOrgVariable(ctx context.Context, org string, variable *github.ActionsVariable) error {
	_, err := g.client.Actions.UpdateOrgVariable(ctx, org, variable)
	return err
}

// ListEnvVariables lists all variables in an environment.
func (g *GitHubClient) ListEnvVariables(ctx context.Context, owner, repo, env string) ([]*github.ActionsVariable, error) {
	var all []*github.ActionsVariable
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		vars, resp, err := g.client.Actions.ListEnvVariables(ctx, owner, repo, env, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, vars.Variables...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return all, nil
}

// GetEnvVariable gets a single environment variable.
func (g *GitHubClient) GetEnvVariable(ctx context.Context, owner, repo, env, name string) (*github.ActionsVariable, error) {
	v, _, err := g.client.Actions.GetEnvVariable(ctx, owner, repo, env, name)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// CreateEnvVariable creates an environment variable.
func (g *GitHubClient) CreateEnvVariable(ctx context.Context, owner, repo, env string, variable *github.ActionsVariable) error {
	_, err := g.client.Actions.CreateEnvVariable(ctx, owner, repo, env, variable)
	return err
}

// UpdateEnvVariable updates an environment variable.
func (g *GitHubClient) UpdateEnvVariable(ctx context.Context, owner, repo, env string, variable *github.ActionsVariable) error {
	_, err := g.client.Actions.UpdateEnvVariable(ctx, owner, repo, env, variable)
	return err
}
