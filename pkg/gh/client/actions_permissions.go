package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// GetOrgDefaultWorkflowPermissions gets the default GITHUB_TOKEN workflow permissions for the organization.
func (g *GitHubClient) GetOrgDefaultWorkflowPermissions(ctx context.Context, org string) (*github.DefaultWorkflowPermissionOrganization, error) {
	permissions, _, err := g.client.Actions.GetDefaultWorkflowPermissionsInOrganization(ctx, org)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// UpdateOrgDefaultWorkflowPermissions sets the default GITHUB_TOKEN workflow permissions for the organization.
func (g *GitHubClient) UpdateOrgDefaultWorkflowPermissions(ctx context.Context, org string, permissions github.DefaultWorkflowPermissionOrganization) (*github.DefaultWorkflowPermissionOrganization, error) {
	result, _, err := g.client.Actions.UpdateDefaultWorkflowPermissionsInOrganization(ctx, org, permissions)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetOrgForkPRContributorApprovalPermissions gets the fork PR contributor approval policy for the organization.
func (g *GitHubClient) GetOrgForkPRContributorApprovalPermissions(ctx context.Context, org string) (*github.ContributorApprovalPermissions, error) {
	permissions, _, err := g.client.Actions.GetOrganizationForkPRContributorApprovalPermissions(ctx, org)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// UpdateOrgForkPRContributorApprovalPermissions sets the fork PR contributor approval policy for the organization.
func (g *GitHubClient) UpdateOrgForkPRContributorApprovalPermissions(ctx context.Context, org string, permissions github.ContributorApprovalPermissions) error {
	_, err := g.client.Actions.UpdateOrganizationForkPRContributorApprovalPermissions(ctx, org, permissions)
	return err
}

// GetRepoActionsPermissions gets the GitHub Actions permissions policy for the repository.
func (g *GitHubClient) GetRepoActionsPermissions(ctx context.Context, owner, repo string) (*github.ActionsPermissionsRepository, error) {
	permissions, _, err := g.client.Repositories.GetActionsPermissions(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// UpdateRepoActionsPermissions sets the GitHub Actions permissions policy for the repository.
func (g *GitHubClient) UpdateRepoActionsPermissions(ctx context.Context, owner, repo string, permissions github.ActionsPermissionsRepository) (*github.ActionsPermissionsRepository, error) {
	result, _, err := g.client.Repositories.UpdateActionsPermissions(ctx, owner, repo, permissions)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetRepoDefaultWorkflowPermissions gets the default GITHUB_TOKEN workflow permissions for the repository.
func (g *GitHubClient) GetRepoDefaultWorkflowPermissions(ctx context.Context, owner, repo string) (*github.DefaultWorkflowPermissionRepository, error) {
	permissions, _, err := g.client.Repositories.GetDefaultWorkflowPermissions(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// UpdateRepoDefaultWorkflowPermissions sets the default GITHUB_TOKEN workflow permissions for the repository.
func (g *GitHubClient) UpdateRepoDefaultWorkflowPermissions(ctx context.Context, owner, repo string, permissions github.DefaultWorkflowPermissionRepository) (*github.DefaultWorkflowPermissionRepository, error) {
	result, _, err := g.client.Repositories.UpdateDefaultWorkflowPermissions(ctx, owner, repo, permissions)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetRepoForkPRContributorApprovalPermissions gets the fork PR contributor approval policy for the repository.
func (g *GitHubClient) GetRepoForkPRContributorApprovalPermissions(ctx context.Context, owner, repo string) (*github.ContributorApprovalPermissions, error) {
	permissions, _, err := g.client.Actions.GetForkPRContributorApprovalPermissions(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// UpdateRepoForkPRContributorApprovalPermissions sets the fork PR contributor approval policy for the repository.
func (g *GitHubClient) UpdateRepoForkPRContributorApprovalPermissions(ctx context.Context, owner, repo string, permissions github.ContributorApprovalPermissions) error {
	_, err := g.client.Actions.UpdateForkPRContributorApprovalPermissions(ctx, owner, repo, permissions)
	return err
}
