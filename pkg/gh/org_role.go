package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

// CustomOrgRoles is an alias for github.CustomOrgRoles, exposed so callers do not need to import the upstream package directly.
type CustomOrgRoles = github.CustomOrgRoles

// ListOrgRoles retrieves all roles available in the specified organization.
func ListOrgRoles(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*CustomOrgRoles, error) {
	roles, err := g.ListOrgRoles(ctx, repo.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles in organization '%s': %w", repo.Owner, err)
	}
	return roles.CustomRepoRoles, nil
}

// ListOrgRolesBySource retrieves roles in the organization filtered by source.
// An empty sources slice returns all roles.
func ListOrgRolesBySource(ctx context.Context, g *GitHubClient, repo repository.Repository, sources []string) ([]*CustomOrgRoles, error) {
	all, err := ListOrgRoles(ctx, g, repo)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return all, nil
	}
	var filtered []*github.CustomOrgRoles
	for _, role := range all {
		for _, s := range sources {
			if role.GetSource() == s {
				filtered = append(filtered, role)
				break
			}
		}
	}
	return filtered, nil
}

// GetRoleIDByName retrieves the RoleID for a given Role name in the specified organization.
func GetRoleIDByName(ctx context.Context, g *GitHubClient, repo repository.Repository, roleName string) (*int64, error) {
	roles, err := ListOrgRoles(ctx, g, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve roles in organization '%s': %w", repo.Owner, err)
	}

	for _, role := range roles {
		if *role.Name == roleName {
			return role.ID, nil
		}
	}

	return nil, fmt.Errorf("role '%s' not found in organization '%s'", roleName, repo.Owner)
}

// ListUsersAssignedToOrgRole retrieves users assigned to a specific organization role.
func ListUsersAssignedToOrgRole(ctx context.Context, g *GitHubClient, repo repository.Repository, roleName string) ([]*GitHubUser, error) {
	roles, err := ListOrgRoles(ctx, g, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve roles in organization '%s': %w", repo.Owner, err)
	}

	var users []*github.User
	for _, role := range roles {
		if roleName == "" || *role.Name == roleName {
			usersFromRole, err := g.ListUsersAssignedToOrgRole(ctx, repo.Owner, *role.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to list users for role '%s' in organization '%s': %w", *role.Name, repo.Owner, err)
			}
			for _, user := range usersFromRole {
				user.RoleName = role.Name
			}
			users = append(users, usersFromRole...)
		}
	}

	return users, nil
}

// AssignOrgRoleToTeam assigns a specific organization role to a team.
func AssignOrgRoleToTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, role string) error {
	roleID, err := GetRoleIDByName(ctx, g, repo, role)
	if err != nil {
		return fmt.Errorf("failed to get role ID for role '%s' in organization '%s': %w", role, repo.Owner, err)
	}
	return g.AssignOrgRoleToTeam(ctx, repo.Owner, teamSlug, *roleID)
}

func RemoveOrgRoleFromTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, role string) error {
	roleID, err := GetRoleIDByName(ctx, g, repo, role)
	if err != nil {
		return fmt.Errorf("failed to get role ID for role '%s' in organization '%s': %w", role, repo.Owner, err)
	}
	return g.RemoveOrgRoleFromTeam(ctx, repo.Owner, teamSlug, *roleID)
}

func AssignOrgRoleToUser(ctx context.Context, g *GitHubClient, repo repository.Repository, username string, role string) error {
	roleID, err := GetRoleIDByName(ctx, g, repo, role)
	if err != nil {
		return fmt.Errorf("failed to get role ID for role '%s' in organization '%s': %w", role, repo.Owner, err)
	}
	return g.AssignOrgRoleToUser(ctx, repo.Owner, username, *roleID)
}

func RemoveOrgRoleFromUser(ctx context.Context, g *GitHubClient, repo repository.Repository, username string, role string) error {
	roleID, err := GetRoleIDByName(ctx, g, repo, role)
	if err != nil {
		return fmt.Errorf("failed to get role ID for role '%s' in organization '%s': %w", role, repo.Owner, err)
	}
	return g.RemoveOrgRoleFromUser(ctx, repo.Owner, username, *roleID)
}

// CreateOrUpdateOrgRole creates or updates a custom organization role by name.
// Only Organization-sourced (user-defined) roles can be created or updated.
// If a role with the given name already exists, it is updated; otherwise a new role is created.
func CreateOrUpdateOrgRole(ctx context.Context, g *GitHubClient, repo repository.Repository, name, description, baseRole string, permissions []string) (*CustomOrgRoles, error) {
	existing, err := ListOrgRoles(ctx, g, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to list org roles in organization '%s': %w", repo.Owner, err)
	}
	opts := &github.CreateOrUpdateOrgRoleOptions{
		Name:        github.Ptr(name),
		Description: github.Ptr(description),
		BaseRole:    github.Ptr(baseRole),
		Permissions: permissions,
	}
	for _, r := range existing {
		if r.Name != nil && *r.Name == name {
			updated, err := g.UpdateCustomOrgRole(ctx, repo.Owner, *r.ID, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to update org role '%s' in organization '%s': %w", name, repo.Owner, err)
			}
			return updated, nil
		}
	}
	created, err := g.CreateCustomOrgRole(ctx, repo.Owner, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create org role '%s' in organization '%s': %w", name, repo.Owner, err)
	}
	return created, nil
}

// TeamOrgRoleEntry holds a team and the list of custom org roles assigned to it.
type TeamOrgRoleEntry struct {
	Team  *github.Team
	Roles []*CustomOrgRoles
}

// BuildTeamOrgRoleMap returns a map from team slug to a TeamOrgRoleEntry containing the team and its assigned custom org roles.
// Only Organization-sourced (user-defined) roles are included.
func BuildTeamOrgRoleMap(ctx context.Context, g *GitHubClient, repo repository.Repository) (map[string]*TeamOrgRoleEntry, error) {
	roles, err := ListOrgRoles(ctx, g, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to list org roles in organization '%s': %w", repo.Owner, err)
	}
	result := make(map[string]*TeamOrgRoleEntry)
	for _, role := range roles {
		if role.Source == nil || *role.Source != "Organization" {
			continue
		}
		teams, err := g.ListTeamsAssignedToOrgRole(ctx, repo.Owner, *role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list teams for role '%s' in organization '%s': %w", *role.Name, repo.Owner, err)
		}
		for _, team := range teams {
			slug := team.GetSlug()
			if _, ok := result[slug]; !ok {
				result[slug] = &TeamOrgRoleEntry{Team: team}
			}
			result[slug].Roles = append(result[slug].Roles, role)
		}
	}
	return result, nil
}
