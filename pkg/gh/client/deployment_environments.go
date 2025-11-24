package client

import (
	"context"

	"github.com/google/go-github/v79/github"
)

// ListEnvironments retrieves all environments for a repository.
func (g *GitHubClient) ListEnvironments(ctx context.Context, owner string, repo string, options *github.EnvironmentListOptions) ([]*github.Environment, error) {
	opt := github.EnvironmentListOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allEnvironments []*github.Environment
	for {
		envs, resp, err := g.client.Repositories.ListEnvironments(ctx, owner, repo, &opt)
		if err != nil {
			return nil, err
		}
		allEnvironments = append(allEnvironments, envs.Environments...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allEnvironments, nil
}

// GetEnvironment retrieves a specific environment by name.
func (g *GitHubClient) GetEnvironment(ctx context.Context, owner string, repo string, name string) (*github.Environment, error) {
	env, _, err := g.client.Repositories.GetEnvironment(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}
	return env, nil
}

// CreateUpdateEnvironment creates or updates an environment.
func (g *GitHubClient) CreateUpdateEnvironment(ctx context.Context, owner string, repo string, name string, environment *github.CreateUpdateEnvironment) (*github.Environment, error) {
	env, _, err := g.client.Repositories.CreateUpdateEnvironment(ctx, owner, repo, name, environment)
	if err != nil {
		return nil, err
	}
	return env, nil
}

// DeleteEnvironment deletes a specific environment.
func (g *GitHubClient) DeleteEnvironment(ctx context.Context, owner string, repo string, name string) error {
	_, err := g.client.Repositories.DeleteEnvironment(ctx, owner, repo, name)
	return err
}

// ListDeploymentBranchPolicies retrieves all deployment branch policies for an environment.
func (g *GitHubClient) ListDeploymentBranchPolicies(ctx context.Context, owner string, repo string, environment string) ([]*github.DeploymentBranchPolicy, error) {
	policies, _, err := g.client.Repositories.ListDeploymentBranchPolicies(ctx, owner, repo, environment)
	if err != nil {
		return nil, err
	}
	return policies.BranchPolicies, nil
}

// GetDeploymentBranchPolicy retrieves a specific deployment branch policy.
func (g *GitHubClient) GetDeploymentBranchPolicy(ctx context.Context, owner string, repo string, environment string, branchPolicyID int64) (*github.DeploymentBranchPolicy, error) {
	policy, _, err := g.client.Repositories.GetDeploymentBranchPolicy(ctx, owner, repo, environment, branchPolicyID)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// CreateDeploymentBranchPolicy creates a deployment branch policy for an environment.
func (g *GitHubClient) CreateDeploymentBranchPolicy(ctx context.Context, owner string, repo string, environment string, request *github.DeploymentBranchPolicyRequest) (*github.DeploymentBranchPolicy, error) {
	policy, _, err := g.client.Repositories.CreateDeploymentBranchPolicy(ctx, owner, repo, environment, request)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// UpdateDeploymentBranchPolicy updates a deployment branch policy.
func (g *GitHubClient) UpdateDeploymentBranchPolicy(ctx context.Context, owner string, repo string, environment string, branchPolicyID int64, request *github.DeploymentBranchPolicyRequest) (*github.DeploymentBranchPolicy, error) {
	policy, _, err := g.client.Repositories.UpdateDeploymentBranchPolicy(ctx, owner, repo, environment, branchPolicyID, request)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// DeleteDeploymentBranchPolicy deletes a deployment branch policy.
func (g *GitHubClient) DeleteDeploymentBranchPolicy(ctx context.Context, owner string, repo string, environment string, branchPolicyID int64) error {
	_, err := g.client.Repositories.DeleteDeploymentBranchPolicy(ctx, owner, repo, environment, branchPolicyID)
	return err
}

// GetAllDeploymentProtectionRules retrieves all deployment protection rules for an environment.
func (g *GitHubClient) GetAllDeploymentProtectionRules(ctx context.Context, owner string, repo string, environment string) (*github.ListDeploymentProtectionRuleResponse, error) {
	rules, _, err := g.client.Repositories.GetAllDeploymentProtectionRules(ctx, owner, repo, environment)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// CreateCustomDeploymentProtectionRule creates a custom deployment protection rule.
func (g *GitHubClient) CreateCustomDeploymentProtectionRule(ctx context.Context, owner string, repo string, environment string, request *github.CustomDeploymentProtectionRuleRequest) (*github.CustomDeploymentProtectionRule, error) {
	rule, _, err := g.client.Repositories.CreateCustomDeploymentProtectionRule(ctx, owner, repo, environment, request)
	if err != nil {
		return nil, err
	}
	return rule, nil
}

// ListCustomDeploymentRuleIntegrations retrieves all custom deployment rule integrations.
func (g *GitHubClient) ListCustomDeploymentRuleIntegrations(ctx context.Context, owner string, repo string, environment string) (*github.ListCustomDeploymentRuleIntegrationsResponse, error) {
	integrations, _, err := g.client.Repositories.ListCustomDeploymentRuleIntegrations(ctx, owner, repo, environment)
	if err != nil {
		return nil, err
	}
	return integrations, nil
}

// GetCustomDeploymentProtectionRule retrieves a specific custom deployment protection rule.
func (g *GitHubClient) GetCustomDeploymentProtectionRule(ctx context.Context, owner string, repo string, environment string, protectionRuleID int64) (*github.CustomDeploymentProtectionRule, error) {
	rule, _, err := g.client.Repositories.GetCustomDeploymentProtectionRule(ctx, owner, repo, environment, protectionRuleID)
	if err != nil {
		return nil, err
	}
	return rule, nil
}

// DisableCustomDeploymentProtectionRule disables a custom deployment protection rule.
func (g *GitHubClient) DisableCustomDeploymentProtectionRule(ctx context.Context, owner string, repo string, environment string, protectionRuleID int64) error {
	_, err := g.client.Repositories.DisableCustomDeploymentProtectionRule(ctx, owner, repo, environment, protectionRuleID)
	return err
}
