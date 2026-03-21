package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

func GetOrg(ctx context.Context, g *GitHubClient, orgName string) (*github.Organization, error) {
	org, err := g.GetOrg(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization '%s': %w", orgName, err)
	}
	return org, nil
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

