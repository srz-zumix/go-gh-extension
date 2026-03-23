package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

func GetOrg(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.Organization, error) {
	org, err := g.GetOrg(ctx, repo.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization '%s': %w", repo.Owner, err)
	}
	return org, nil
}

func EditOrg(ctx context.Context, g *GitHubClient, repo repository.Repository, input *github.Organization) (*github.Organization, error) {
	org, err := g.EditOrg(ctx, repo.Owner, input)
	if err != nil {
		return nil, fmt.Errorf("failed to edit organization '%s': %w", repo.Owner, err)
	}
	return org, nil
}

// SetOrgDeployKeysEnabled enables or disables deploy keys for all repositories in the organization.
func SetOrgDeployKeysEnabled(ctx context.Context, g *GitHubClient, repo repository.Repository, enabled bool) (*github.Organization, error) {
	org, err := g.SetOrgDeployKeysEnabled(ctx, repo.Owner, enabled)
	if err != nil {
		return nil, fmt.Errorf("failed to set deploy keys enabled for organization '%s': %w", repo.Owner, err)
	}
	return org, nil
}

// GetOrgDeployKeysEnabled returns whether deploy keys are enabled for all repositories in the organization.
func GetOrgDeployKeysEnabled(ctx context.Context, g *GitHubClient, repo repository.Repository) (bool, error) {
	enabled, err := g.GetOrgDeployKeysEnabled(ctx, repo.Owner)
	if err != nil {
		return false, fmt.Errorf("failed to get deploy keys enabled for organization '%s': %w", repo.Owner, err)
	}
	return enabled, nil
}

// GetOrgMembership retrieves the membership details of a user in the specified organization.
func GetOrgMembership(ctx context.Context, g *GitHubClient, repo repository.Repository, username string) (*github.Membership, error) {
	return g.GetOrgMembership(ctx, repo.Owner, username)
}

// FindOrgMembership retrieves the membership details of a user in the specified organization.
func FindOrgMembership(ctx context.Context, g *GitHubClient, repo repository.Repository, username string) (*github.Membership, error) {
	return g.FindOrgMembership(ctx, repo.Owner, username)
}

// ListOrgMembers wraps the GitHubClient's ListOrgMembers function.
func ListOrgMembers(ctx context.Context, g *GitHubClient, repo repository.Repository, roles []string, membership bool) ([]*github.User, error) {
	roleFilter := GetOrgMembershipFilter(roles)
	members, err := g.ListOrgMembers(ctx, repo.Owner, roleFilter)
	if err != nil {
		return nil, err
	}

	if membership {
		for _, member := range members {
			membership, err := g.GetOrgMembership(ctx, repo.Owner, *member.Login)
			if err != nil {
				return nil, err
			}
			if membership != nil {
				member.RoleName = membership.Role
			}
		}
	}
	return members, nil
}

// RemoveOrgMember removes a member from the specified organization.
func RemoveOrgMember(ctx context.Context, g *GitHubClient, repo repository.Repository, username string) error {
	return g.RemoveOrgMember(ctx, repo.Owner, username)
}

// AddOrUpdateOrgMember adds or updates a member in the specified organization with the given role.
func AddOrUpdateOrgMember(ctx context.Context, g *GitHubClient, repo repository.Repository, username string, role string) (*github.User, error) {
	membership, err := g.AddOrUpdateOrgMembership(ctx, repo.Owner, username, role)
	if err != nil {
		return nil, fmt.Errorf("failed to add or update '%s' in organization '%s': %w", username, repo.Owner, err)
	}
	membership.User.RoleName = membership.Role
	return membership.User, nil
}

func UpdateOrgMemberRole(ctx context.Context, g *GitHubClient, repo repository.Repository, username string, role string) (*github.User, error) {
	membership, err := g.FindOrgMembership(ctx, repo.Owner, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find membership for '%s' in organization '%s': %w", username, repo.Owner, err)
	}
	if membership == nil {
		return nil, fmt.Errorf("user '%s' is not a member of organization '%s'", username, repo.Owner)
	}
	membership, err = g.AddOrUpdateOrgMembership(ctx, repo.Owner, username, role)
	if err != nil {
		return nil, fmt.Errorf("failed to update '%s' role in organization '%s': %w", username, repo.Owner, err)
	}
	membership.User.RoleName = membership.Role
	return membership.User, nil
}

// ListTeamsAssignedToRole retrieves teams assigned to a specific organization role.
func ListTeamsAssignedToRole(ctx context.Context, g *GitHubClient, repo repository.Repository, role string) ([]*github.Team, error) {
	roleID, err := GetRoleIDByName(ctx, g, repo, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get role ID for role '%s' in organization '%s': %w", role, repo.Owner, err)
	}
	teams, err := g.ListTeamsAssignedToOrgRole(ctx, repo.Owner, *roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams for role '%s' in organization '%s': %w", role, repo.Owner, err)
	}
	return teams, nil
}

// GetOrgMemberPrivileges retrieves the member privileges settings for the organization.
func GetOrgMemberPrivileges(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.Organization, error) {
	return GetOrg(ctx, g, repo)
}

// EditOrgMemberPrivileges updates the member privileges settings for the organization.
func EditOrgMemberPrivileges(ctx context.Context, g *GitHubClient, repo repository.Repository, input *github.Organization) (*github.Organization, error) {
	return EditOrg(ctx, g, repo, input)
}
