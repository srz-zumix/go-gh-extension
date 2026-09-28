package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

// Allowed values for the default_workflow_permissions field (organization and repository level).
const (
	DefaultWorkflowPermissionsRead  = "read"
	DefaultWorkflowPermissionsWrite = "write"
)

// Allowed values for the fork PR contributor approval_policy field (organization and repository level).
const (
	ForkPRApprovalAllExternalContributors          = "all_external_contributors"
	ForkPRApprovalFirstTimeContributors            = "first_time_contributors"
	ForkPRApprovalFirstTimeContributorsNewToGitHub = "first_time_contributors_new_to_github"
)

// GetOrgDefaultWorkflowPermissions retrieves the default GITHUB_TOKEN workflow permissions for the organization.
func GetOrgDefaultWorkflowPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.DefaultWorkflowPermissionOrganization, error) {
	permissions, err := g.GetOrgDefaultWorkflowPermissions(ctx, repo.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get default workflow permissions for organization '%s': %w", repo.Owner, err)
	}
	return permissions, nil
}

// UpdateOrgDefaultWorkflowPermissions updates the default GITHUB_TOKEN workflow permissions for the organization.
func UpdateOrgDefaultWorkflowPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository, permissions github.DefaultWorkflowPermissionOrganization) (*github.DefaultWorkflowPermissionOrganization, error) {
	result, err := g.UpdateOrgDefaultWorkflowPermissions(ctx, repo.Owner, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to update default workflow permissions for organization '%s': %w", repo.Owner, err)
	}
	return result, nil
}

// SetOrgDefaultWorkflowPermissions sets the default GITHUB_TOKEN permission level for the organization.
// permissions must be one of the DefaultWorkflowPermissions* constants.
func SetOrgDefaultWorkflowPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository, permissions string) (*github.DefaultWorkflowPermissionOrganization, error) {
	return UpdateOrgDefaultWorkflowPermissions(ctx, g, repo, github.DefaultWorkflowPermissionOrganization{
		DefaultWorkflowPermissions: &permissions,
	})
}

// SetOrgActionsCanApprovePullRequestReviews sets whether GitHub Actions can approve pull requests for the organization.
func SetOrgActionsCanApprovePullRequestReviews(ctx context.Context, g *GitHubClient, repo repository.Repository, enabled bool) (*github.DefaultWorkflowPermissionOrganization, error) {
	return UpdateOrgDefaultWorkflowPermissions(ctx, g, repo, github.DefaultWorkflowPermissionOrganization{
		CanApprovePullRequestReviews: &enabled,
	})
}

// GetOrgForkPRContributorApprovalPermissions retrieves the fork PR contributor approval policy for the organization.
func GetOrgForkPRContributorApprovalPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.ContributorApprovalPermissions, error) {
	permissions, err := g.GetOrgForkPRContributorApprovalPermissions(ctx, repo.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get fork PR contributor approval permissions for organization '%s': %w", repo.Owner, err)
	}
	return permissions, nil
}

// SetOrgForkPRContributorApprovalPolicy sets the fork PR contributor approval policy for the organization.
// approvalPolicy must be one of the ForkPRApproval* constants.
func SetOrgForkPRContributorApprovalPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository, approvalPolicy string) error {
	if err := g.UpdateOrgForkPRContributorApprovalPermissions(ctx, repo.Owner, github.ContributorApprovalPermissions{ApprovalPolicy: approvalPolicy}); err != nil {
		return fmt.Errorf("failed to update fork PR contributor approval policy for organization '%s': %w", repo.Owner, err)
	}
	return nil
}

// GetRepoActionsPermissions retrieves the GitHub Actions permissions policy for the repository.
func GetRepoActionsPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.ActionsPermissionsRepository, error) {
	permissions, err := g.GetRepoActionsPermissions(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub Actions permissions for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return permissions, nil
}

// UpdateRepoActionsPermissions updates the GitHub Actions permissions policy for the repository.
func UpdateRepoActionsPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository, permissions github.ActionsPermissionsRepository) (*github.ActionsPermissionsRepository, error) {
	result, err := g.UpdateRepoActionsPermissions(ctx, repo.Owner, repo.Name, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to update GitHub Actions permissions for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return result, nil
}

// SetRepoAllowedActions sets which GitHub Actions are allowed to run in the repository.
// allowedActions must be one of the OrgAllowedActions* constants (the allowed values are shared
// between the organization and repository level allowed_actions policy).
func SetRepoAllowedActions(ctx context.Context, g *GitHubClient, repo repository.Repository, allowedActions string) (*github.ActionsPermissionsRepository, error) {
	enabled := true
	return UpdateRepoActionsPermissions(ctx, g, repo, github.ActionsPermissionsRepository{
		Enabled:        &enabled,
		AllowedActions: &allowedActions,
	})
}

// GetRepoDefaultWorkflowPermissions retrieves the default GITHUB_TOKEN workflow permissions for the repository.
func GetRepoDefaultWorkflowPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.DefaultWorkflowPermissionRepository, error) {
	permissions, err := g.GetRepoDefaultWorkflowPermissions(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get default workflow permissions for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return permissions, nil
}

// UpdateRepoDefaultWorkflowPermissions updates the default GITHUB_TOKEN workflow permissions for the repository.
func UpdateRepoDefaultWorkflowPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository, permissions github.DefaultWorkflowPermissionRepository) (*github.DefaultWorkflowPermissionRepository, error) {
	result, err := g.UpdateRepoDefaultWorkflowPermissions(ctx, repo.Owner, repo.Name, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to update default workflow permissions for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return result, nil
}

// SetRepoDefaultWorkflowPermissions sets the default GITHUB_TOKEN permission level for the repository.
// permissions must be one of the DefaultWorkflowPermissions* constants.
func SetRepoDefaultWorkflowPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository, permissions string) (*github.DefaultWorkflowPermissionRepository, error) {
	return UpdateRepoDefaultWorkflowPermissions(ctx, g, repo, github.DefaultWorkflowPermissionRepository{
		DefaultWorkflowPermissions: &permissions,
	})
}

// SetRepoActionsCanApprovePullRequestReviews sets whether GitHub Actions can approve pull requests for the repository.
func SetRepoActionsCanApprovePullRequestReviews(ctx context.Context, g *GitHubClient, repo repository.Repository, enabled bool) (*github.DefaultWorkflowPermissionRepository, error) {
	return UpdateRepoDefaultWorkflowPermissions(ctx, g, repo, github.DefaultWorkflowPermissionRepository{
		CanApprovePullRequestReviews: &enabled,
	})
}

// GetRepoForkPRContributorApprovalPermissions retrieves the fork PR contributor approval policy for the repository.
func GetRepoForkPRContributorApprovalPermissions(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.ContributorApprovalPermissions, error) {
	permissions, err := g.GetRepoForkPRContributorApprovalPermissions(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get fork PR contributor approval permissions for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return permissions, nil
}

// SetRepoForkPRContributorApprovalPolicy sets the fork PR contributor approval policy for the repository.
// approvalPolicy must be one of the ForkPRApproval* constants.
func SetRepoForkPRContributorApprovalPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository, approvalPolicy string) error {
	if err := g.UpdateRepoForkPRContributorApprovalPermissions(ctx, repo.Owner, repo.Name, github.ContributorApprovalPermissions{ApprovalPolicy: approvalPolicy}); err != nil {
		return fmt.Errorf("failed to update fork PR contributor approval policy for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return nil
}
