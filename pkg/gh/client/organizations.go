package client

import (
	"context"

	"github.com/google/go-github/v88/github"
)

func (g *GitHubClient) GetOrg(ctx context.Context, org string) (*github.Organization, error) {
	organization, _, err := g.client.Organizations.Get(ctx, org)
	if err != nil {
		return nil, err
	}
	return organization, nil
}

// EditOrg updates the organization settings using the GitHub API.
func (g *GitHubClient) EditOrg(ctx context.Context, org string, input *github.Organization) (*github.Organization, error) {
	organization, _, err := g.client.Organizations.Edit(ctx, org, input)
	if err != nil {
		return nil, err
	}
	return organization, nil
}

// ListOrgMembers retrieves all members of the specified organization.
func (g *GitHubClient) ListOrgMembers(ctx context.Context, org string, role string) ([]*github.User, error) {
	var allMembers []*github.User
	opt := &github.ListMembersOptions{
		Role:        role,
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}

	for {
		members, resp, err := g.client.Organizations.ListMembers(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allMembers, nil
}

// GetOrgMembership retrieves the membership details of a user in the organization.
func (g *GitHubClient) GetOrgMembership(ctx context.Context, owner string, username string) (*github.Membership, error) {
	membership, _, err := g.client.Organizations.GetOrgMembership(ctx, username, owner)
	if err != nil {
		return nil, err
	}

	return membership, nil
}

// FindOrgMembership retrieves the membership details of a user in the organization.
// If the user is not a member, it returns nil without an error.
func (g *GitHubClient) FindOrgMembership(ctx context.Context, owner string, username string) (*github.Membership, error) {
	membership, resp, err := g.client.Organizations.GetOrgMembership(ctx, username, owner)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, nil // User is not a member
		}
		return nil, err // Other errors
	}

	return membership, nil
}

// AddOrUpdateOrgMembership sets the membership details of a user in the organization.
func (g *GitHubClient) AddOrUpdateOrgMembership(ctx context.Context, org string, username string, role string) (*github.Membership, error) {
	membership, _, err := g.client.Organizations.EditOrgMembership(ctx, username, org, &github.Membership{Role: &role})
	if err != nil {
		return nil, err
	}
	return membership, nil
}

// RemoveOrgMember removes a member from the specified organization.
func (g *GitHubClient) RemoveOrgMember(ctx context.Context, org string, username string) error {
	_, err := g.client.Organizations.RemoveMember(ctx, org, username)
	if err != nil {
		return err
	}
	return nil
}

// ListTeamsAssignedToOrgRole retrieves teams assigned to a specific organization role by roleID.
func (g *GitHubClient) ListTeamsAssignedToOrgRole(ctx context.Context, org string, roleID int64) ([]*github.Team, error) {
	var allTeams []*github.Team
	opt := &github.ListOptions{PerPage: defaultPerPage}

	for {
		teams, resp, err := g.client.Organizations.ListTeamsAssignedToOrgRole(ctx, org, roleID, opt)
		if err != nil {
			return nil, err
		}
		allTeams = append(allTeams, teams...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allTeams, nil
}

// ListUsersAssignedToOrgRole retrieves users assigned to a specific organization role.
func (g *GitHubClient) ListUsersAssignedToOrgRole(ctx context.Context, org string, roleID int64) ([]*github.User, error) {
	var allUsers []*github.User
	opt := &github.ListOptions{
		PerPage: 50,
	}

	for {
		users, resp, err := g.client.Organizations.ListUsersAssignedToOrgRole(ctx, org, roleID, opt)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allUsers, nil
}

// ListOrgRoles retrieves all custom roles available in the specified organization.
func (g *GitHubClient) ListOrgRoles(ctx context.Context, org string) (*github.OrganizationCustomRoles, error) {
	roles, _, err := g.client.Organizations.ListRoles(ctx, org)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// AssignOrgRoleToTeam assigns a specific organization role to a team using the GitHub API.
func (g *GitHubClient) AssignOrgRoleToTeam(ctx context.Context, org string, teamSlug string, roleID int64) error {
	_, err := g.client.Organizations.AssignOrgRoleToTeam(ctx, org, teamSlug, roleID)
	return err
}

func (g *GitHubClient) RemoveOrgRoleFromTeam(ctx context.Context, org string, teamSlug string, roleID int64) error {
	_, err := g.client.Organizations.RemoveOrgRoleFromTeam(ctx, org, teamSlug, roleID)
	return err
}

func (g *GitHubClient) AssignOrgRoleToUser(ctx context.Context, org string, username string, roleID int64) error {
	_, err := g.client.Organizations.AssignOrgRoleToUser(ctx, org, username, roleID)
	return err
}

func (g *GitHubClient) RemoveOrgRoleFromUser(ctx context.Context, org string, username string, roleID int64) error {
	_, err := g.client.Organizations.RemoveOrgRoleFromUser(ctx, org, username, roleID)
	return err
}

// CreateCustomOrgRole creates a new custom organization role.
func (g *GitHubClient) CreateCustomOrgRole(ctx context.Context, org string, request github.CreateCustomOrgRoleRequest) (*github.CustomOrgRole, error) {
	role, _, err := g.client.Organizations.CreateCustomOrgRole(ctx, org, request)
	return role, err
}

// UpdateCustomOrgRole updates an existing custom organization role by ID.
func (g *GitHubClient) UpdateCustomOrgRole(ctx context.Context, org string, roleID int64, request github.UpdateCustomOrgRoleRequest) (*github.CustomOrgRole, error) {
	role, _, err := g.client.Organizations.UpdateCustomOrgRole(ctx, org, roleID, request)
	return role, err
}

// SetOrgDeployKeysEnabled enables or disables the ability to use deploy keys across
// all repositories in the organization. The deploy_keys_enabled_for_repositories
// field is natively supported by go-github (Organization struct + Organizations.Edit).
func (g *GitHubClient) SetOrgDeployKeysEnabled(ctx context.Context, org string, enabled bool) (*github.Organization, error) {
	result, _, err := g.client.Organizations.Edit(ctx, org, &github.Organization{
		DeployKeysEnabledForRepositories: github.Ptr(enabled),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetOrgDeployKeysEnabled returns whether deploy keys are enabled for repositories in the organization.
func (g *GitHubClient) GetOrgDeployKeysEnabled(ctx context.Context, org string) (bool, error) {
	result, _, err := g.client.Organizations.Get(ctx, org)
	if err != nil {
		return false, err
	}
	return result.GetDeployKeysEnabledForRepositories(), nil
}
