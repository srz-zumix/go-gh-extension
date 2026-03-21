package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// CreateUpdateEnvironment creates or updates an environment for a repository.
func CreateUpdateEnvironment(ctx context.Context, g *GitHubClient, repo repository.Repository, name string, environment *github.CreateUpdateEnvironment) (*github.Environment, error) {
	return g.CreateUpdateEnvironment(ctx, repo.Owner, repo.Name, name, environment)
}

// EnvironmentToCreateUpdateRequest builds a CreateUpdateEnvironment request from a GET-response Environment.
// Wait timer, reviewers, and prevent-self-review are extracted from ProtectionRules.
func EnvironmentToCreateUpdateRequest(env *github.Environment) *github.CreateUpdateEnvironment {
	req := &github.CreateUpdateEnvironment{
		CanAdminsBypass:        env.CanAdminsBypass,
		DeploymentBranchPolicy: env.DeploymentBranchPolicy,
		Reviewers:              []*github.EnvReviewers{},
	}
	for _, rule := range env.ProtectionRules {
		if rule.Type == nil {
			continue
		}
		switch *rule.Type {
		case "wait_timer":
			req.WaitTimer = rule.WaitTimer
		case "required_reviewers":
			req.PreventSelfReview = rule.PreventSelfReview
			for _, r := range rule.Reviewers {
				if r.Type == nil {
					continue
				}
				reviewer := &github.EnvReviewers{Type: r.Type}
				switch v := r.Reviewer.(type) {
				case *github.User:
					reviewer.ID = v.ID
				case *github.Team:
					reviewer.ID = v.ID
				}
				if reviewer.ID != nil {
					req.Reviewers = append(req.Reviewers, reviewer)
				}
			}
		}
	}
	return req
}

// ListEnvironments retrieves all environments for a repository.
func ListEnvironments(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Environment, error) {
	return g.ListEnvironments(ctx, repo.Owner, repo.Name, nil)
}

// GetEnvironment retrieves a specific environment by name.
func GetEnvironment(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.Environment, error) {
	return g.GetEnvironment(ctx, repo.Owner, repo.Name, name)
}

// GetEnvironmentID retrieves the ID of an environment given its name or ID.
func GetEnvironmentID(ctx context.Context, g *GitHubClient, repo repository.Repository, environments any) (*int64, error) {
	switch v := environments.(type) {
	case string:
		env, err := g.GetEnvironment(ctx, repo.Owner, repo.Name, v)
		if err != nil {
			return nil, err
		}
		return env.ID, nil
	case *string:
		env, err := g.GetEnvironment(ctx, repo.Owner, repo.Name, *v)
		if err != nil {
			return nil, err
		}
		return env.ID, nil
	case int64:
		return &v, nil
	case *int64:
		return v, nil
	default:
		return nil, nil
	}
}

// DeleteEnvironment deletes a specific environment.
func DeleteEnvironment(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) error {
	return g.DeleteEnvironment(ctx, repo.Owner, repo.Name, name)
}

// ListDeploymentBranchPolicies retrieves all deployment branch policies for an environment.
func ListDeploymentBranchPolicies(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string) ([]*github.DeploymentBranchPolicy, error) {
	return g.ListDeploymentBranchPolicies(ctx, repo.Owner, repo.Name, environment)
}

// ListDeploymentCustomBranchPolicies retrieves all custom deployment branch policies for an environment.
func ListDeploymentCustomBranchPolicies(ctx context.Context, g *GitHubClient, repo repository.Repository, environment *github.Environment) ([]*github.DeploymentBranchPolicy, error) {
	if environment == nil {
		return nil, nil
	}

	if environment.DeploymentBranchPolicy != nil &&
		environment.DeploymentBranchPolicy.CustomBranchPolicies != nil &&
		*environment.DeploymentBranchPolicy.CustomBranchPolicies {
		envName := environment.GetName()

		policies, err := ListDeploymentBranchPolicies(ctx, g, repo, envName)
		if err != nil {
			return nil, fmt.Errorf("failed to list deployment branch policies for environment %q: %w", envName, err)
		}
		return policies, nil
	}
	return nil, nil
}

// GetDeploymentBranchPolicy retrieves a specific deployment branch policy.
func GetDeploymentBranchPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string, branchPolicyID int64) (*github.DeploymentBranchPolicy, error) {
	return g.GetDeploymentBranchPolicy(ctx, repo.Owner, repo.Name, environment, branchPolicyID)
}

// CreateDeploymentBranchPolicy creates a deployment branch policy for an environment.
func CreateDeploymentBranchPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string, ref string, refType string) (*github.DeploymentBranchPolicy, error) {
	request := &github.DeploymentBranchPolicyRequest{
		Name: &ref,
		Type: &refType,
	}
	return g.CreateDeploymentBranchPolicy(ctx, repo.Owner, repo.Name, environment, request)
}

// UpdateDeploymentBranchPolicy updates a deployment branch policy.
func UpdateDeploymentBranchPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string, branchPolicyID int64, ref *string, refType *string) (*github.DeploymentBranchPolicy, error) {
	request := &github.DeploymentBranchPolicyRequest{
		Name: ref,
		Type: refType,
	}
	return g.UpdateDeploymentBranchPolicy(ctx, repo.Owner, repo.Name, environment, branchPolicyID, request)
}

// DeleteDeploymentBranchPolicy deletes a deployment branch policy.
func DeleteDeploymentBranchPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string, branchPolicyID int64) error {
	return g.DeleteDeploymentBranchPolicy(ctx, repo.Owner, repo.Name, environment, branchPolicyID)
}

// GetAllDeploymentProtectionRules retrieves all deployment protection rules for an environment.
func GetAllDeploymentProtectionRules(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string) (*github.ListDeploymentProtectionRuleResponse, error) {
	return g.GetAllDeploymentProtectionRules(ctx, repo.Owner, repo.Name, environment)
}

// CreateCustomDeploymentProtectionRule creates a custom deployment protection rule.
func CreateCustomDeploymentProtectionRule(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string, request *github.CustomDeploymentProtectionRuleRequest) (*github.CustomDeploymentProtectionRule, error) {
	return g.CreateCustomDeploymentProtectionRule(ctx, repo.Owner, repo.Name, environment, request)
}

// ListCustomDeploymentRuleIntegrations retrieves all custom deployment rule integrations.
func ListCustomDeploymentRuleIntegrations(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string) ([]*github.CustomDeploymentProtectionRuleApp, error) {
	return g.ListCustomDeploymentRuleIntegrations(ctx, repo.Owner, repo.Name, environment)
}

// GetCustomDeploymentProtectionRule retrieves a specific custom deployment protection rule.
func GetCustomDeploymentProtectionRule(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string, protectionRuleID int64) (*github.CustomDeploymentProtectionRule, error) {
	return g.GetCustomDeploymentProtectionRule(ctx, repo.Owner, repo.Name, environment, protectionRuleID)
}

// DisableCustomDeploymentProtectionRule disables a custom deployment protection rule.
func DisableCustomDeploymentProtectionRule(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string, protectionRuleID int64) error {
	return g.DisableCustomDeploymentProtectionRule(ctx, repo.Owner, repo.Name, environment, protectionRuleID)
}
