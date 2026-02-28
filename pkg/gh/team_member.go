package gh

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

var (
	TeamMembershipRoleAll        = "all"
	TeamMembershipRoleMember     = "member"
	TeamMembershipRoleMaintainer = "maintainer"
)

// GetTeamMembership retrieves the membership details of a user in a specific team.
func GetTeamMembership(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, username string) (*github.Membership, error) {
	return g.GetTeamMembership(ctx, repo.Owner, teamSlug, username)
}

// FindTeamMembership retrieves the membership details of a user in a specific team.
func FindTeamMembership(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, username string) (*github.Membership, error) {
	return g.FindTeamMembership(ctx, repo.Owner, teamSlug, username)
}

// ListTeamMembers retrieves all members of a specific team in the organization and filters them by roles if provided.
func ListTeamMembers(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, roles []string, membership bool) ([]*github.User, error) {
	members, err := g.ListTeamMembers(ctx, repo.Owner, teamSlug, GetTeamMembershipFilter(roles))
	if err != nil {
		return nil, err
	}

	if membership {
		for _, member := range members {
			membership, err := g.GetTeamMembership(ctx, repo.Owner, teamSlug, *member.Login)
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

// AddTeamMember is a wrapper function to add or update a team member.
func AddTeamMember(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, username string, role string, allowNonOrganizationMember bool) (*github.Membership, error) {
	if !allowNonOrganizationMember {
		membership, err := g.FindOrgMembership(ctx, repo.Owner, username)
		if err != nil {
			return nil, err
		}
		if membership == nil {
			return nil, fmt.Errorf("user '%s' is not a member of the organization '%s'", username, repo.Owner)
		}
	}
	return g.AddOrUpdateTeamMembership(ctx, repo.Owner, teamSlug, username, role)
}

func AddTeamMembers(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, usernames []string, role string, allowNonOrganizationMember bool) ([]*github.Membership, error) {
	var memberships []*github.Membership
	errorList := []error{}
	for _, username := range usernames {
		membership, err := AddTeamMember(ctx, g, repo, teamSlug, username, role, allowNonOrganizationMember)
		if err != nil {
			errorList = append(errorList, err)
			continue
		}
		memberships = append(memberships, membership)
	}
	if len(errorList) > 0 {
		return memberships, fmt.Errorf("encountered errors while adding team members: %v", errorList)
	}
	return memberships, nil
}

func UpdateTeamMemberRole(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, username string, role string) (*github.Membership, error) {
	membership, err := g.FindTeamMembership(ctx, repo.Owner, teamSlug, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find membership for '%s' in team '%s': %w", username, teamSlug, err)
	}
	if membership == nil {
		return nil, fmt.Errorf("user '%s' is not a member of team '%s'", username, teamSlug)
	}
	return g.AddOrUpdateTeamMembership(ctx, repo.Owner, teamSlug, username, role)
}

// RemoveTeamMember is a wrapper function to remove a user from a team.
func RemoveTeamMember(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, username string) error {
	return g.RemoveTeamMember(ctx, repo.Owner, teamSlug, username)
}

func RemoveTeamMembers(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, usernames []string) error {
	errorList := []error{}
	for _, username := range usernames {
		err := RemoveTeamMember(ctx, g, repo, teamSlug, username)
		if err != nil {
			errorList = append(errorList, err)
		}
	}
	if len(errorList) > 0 {
		return fmt.Errorf("encountered errors while removing team members: %v", errorList)
	}
	return nil
}

func RemoveTeamMembersOther(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, excludeUsernames []string) error {
	members, err := ListTeamMembers(ctx, g, repo, teamSlug, nil, false)
	if err != nil {
		return fmt.Errorf("failed to fetch members from team: %w", err)
	}
	excludeMap := make(map[string]struct{})
	for _, username := range excludeUsernames {
		excludeMap[username] = struct{}{}
	}
	for _, m := range members {
		if m.Login == nil {
			continue
		}
		username := *m.Login
		if _, exists := excludeMap[username]; !exists {
			err := RemoveTeamMember(ctx, g, repo, teamSlug, username)
			if err != nil {
				return fmt.Errorf("failed to remove member %s from team: %w", username, err)
			}
		}
	}
	return nil
}

// ListAnyTeamMembers lists all teams in the organization or repository and returns the union of all team members.
// Membership information is organization membership, not team membership, since a user may have different roles in different teams.
func ListAnyTeamMembers(ctx context.Context, g *GitHubClient, repo repository.Repository, roles []string, membership bool) ([]*github.User, error) {
	teams, err := ListTeams(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	userMap := make(map[int64]*github.User)
	for _, team := range teams {
		members, err := ListTeamMembers(ctx, g, repo, team.GetSlug(), roles, false)
		if err != nil {
			return nil, fmt.Errorf("failed to list members of team '%s': %w", team.GetSlug(), err)
		}
		for _, member := range members {
			if _, exists := userMap[member.GetID()]; !exists {
				userMap[member.GetID()] = member
			}
		}
	}

	members := slices.Collect(maps.Values(userMap))
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

// ListOnlyTeamMembers returns members who belong exclusively to the specified team
// and are not members of any other team in the organization.
func ListOnlyTeamMembers(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, roles []string, membership bool) ([]*github.User, error) {
	targetMembers, err := ListTeamMembers(ctx, g, repo, teamSlug, roles, false)
	if err != nil {
		return nil, err
	}

	if len(targetMembers) == 0 {
		return []*github.User{}, nil
	}

	teams, err := ListTeams(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	// Build a set of target member IDs for tracking candidates
	candidateIDs := make(map[int64]struct{}, len(targetMembers))
	for _, member := range targetMembers {
		candidateIDs[member.GetID()] = struct{}{}
	}

	// Iterate other teams and remove candidates found in them; stop early if none remain
	for _, team := range teams {
		if team.GetSlug() == teamSlug {
			continue
		}
		members, err := ListTeamMembers(ctx, g, repo, team.GetSlug(), nil, false)
		if err != nil {
			return nil, fmt.Errorf("failed to list members of team '%s': %w", team.GetSlug(), err)
		}
		for _, member := range members {
			delete(candidateIDs, member.GetID())
		}
		if len(candidateIDs) == 0 {
			return []*github.User{}, nil
		}
	}

	// Filter to only members still in the candidate set
	members := make([]*github.User, 0, len(candidateIDs))
	for _, member := range targetMembers {
		if _, exists := candidateIDs[member.GetID()]; exists {
			members = append(members, member)
		}
	}
	if membership {
		for _, member := range members {
			membership, err := g.GetTeamMembership(ctx, repo.Owner, teamSlug, *member.Login)
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

const (
	// TeamSpecAny is a special team slug that represents the union of all team members across all teams in the organization or repository.
	TeamSpecAny = "@any"
	// TeamSpecAll is a special team slug that represents all organization or repository members.
	TeamSpecAll = "@all"
)

// ListMembersByTeamSpec lists team members based on the provided team specification.
// TeamSpecAny (@any): returns the union of all team members across all teams in the organization or repository.
// TeamSpecAll (@all): returns all repository collaborators if repo.Name is set, otherwise all organization members.
// Otherwise: returns members of the specified team.
// Membership information is only available when listing by specific team, not for TeamSpecAny or TeamSpecAll.
func ListMembersByTeamSpec(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSpec string, roles []string, membership bool) ([]*github.User, error) {
	switch teamSpec {
	case TeamSpecAny:
		return ListAnyTeamMembers(ctx, g, repo, roles, false)
	case TeamSpecAll:
		if repo.Name != "" {
			return ListRepositoryCollaborators(ctx, g, repo, nil, roles)
		}
		org_roles := []string{}
		for _, role := range roles {
			switch role {
			case TeamMembershipRoleAll:
				org_roles = []string{"admin", "member"}
			case TeamMembershipRoleMember:
				org_roles = append(org_roles, "member")
			case TeamMembershipRoleMaintainer:
				org_roles = append(org_roles, "admin")
			}
		}

		return ListOrgMembers(ctx, g, repo, org_roles, false)
	default:
		return ListTeamMembers(ctx, g, repo, teamSpec, roles, membership)
	}
}

// CopyTeamMembers copies members from the source team to the destination team (add only).
func CopyTeamMembers(ctx context.Context, srcClient *GitHubClient, srcRepo repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dstRepo repository.Repository, dstTeamSlug string) error {
	srcMembers, err := ListTeamMembers(ctx, srcClient, srcRepo, srcTeamSlug, nil, false)
	if err != nil {
		return fmt.Errorf("failed to fetch members from source team: %w", err)
	}
	dstMembers, err := ListTeamMembers(ctx, dstClient, dstRepo, dstTeamSlug, nil, false)
	if err != nil {
		return fmt.Errorf("failed to fetch members from destination team: %w", err)
	}
	dstMemberMap := make(map[string]struct{})
	for _, m := range dstMembers {
		if m.Login != nil {
			dstMemberMap[*m.Login] = struct{}{}
		}
	}
	for _, m := range srcMembers {
		if m.Login == nil {
			continue
		}
		username := *m.Login
		if _, exists := dstMemberMap[username]; !exists {
			_, err := AddTeamMember(ctx, dstClient, dstRepo, dstTeamSlug, username, *m.RoleName, false)
			if err != nil {
				return fmt.Errorf("failed to add member %s to destination team: %w", username, err)
			}
		}
	}
	return nil
}

// SyncTeamMembers syncs members from the source team to the destination team.
func SyncTeamMembers(ctx context.Context, srcClient *GitHubClient, srcRepo repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dstRepo repository.Repository, dstTeamSlug string) error {
	srcMembers, err := ListTeamMembers(ctx, srcClient, srcRepo, srcTeamSlug, nil, false)
	if err != nil {
		return fmt.Errorf("failed to fetch members from source team: %w", err)
	}
	dstMembers, err := ListTeamMembers(ctx, dstClient, dstRepo, dstTeamSlug, nil, false)
	if err != nil {
		return fmt.Errorf("failed to fetch members from destination team: %w", err)
	}
	dstMemberMap := make(map[string]struct{})
	for _, m := range dstMembers {
		if m.Login != nil {
			dstMemberMap[*m.Login] = struct{}{}
		}
	}
	srcMemberMap := make(map[string]struct{})
	for _, m := range srcMembers {
		if m.Login != nil {
			srcMemberMap[*m.Login] = struct{}{}
		}
	}
	// Add or update members in destination team
	for _, m := range srcMembers {
		if m.Login == nil {
			continue
		}
		username := *m.Login
		_, err := AddTeamMember(ctx, dstClient, dstRepo, dstTeamSlug, username, *m.RoleName, false)
		if err != nil {
			return fmt.Errorf("failed to add member %s to destination team: %w", username, err)
		}
	}
	// Remove members from destination team that are not in source team
	for _, m := range dstMembers {
		if m.Login == nil {
			continue
		}
		username := *m.Login
		if _, exists := srcMemberMap[username]; !exists {
			err := RemoveTeamMember(ctx, dstClient, dstRepo, dstTeamSlug, username)
			if err != nil {
				return fmt.Errorf("failed to remove member %s from destination team: %w", username, err)
			}
		}
	}
	return nil
}
