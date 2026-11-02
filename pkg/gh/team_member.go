package gh

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v93/github"
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
			if member.Login == nil {
				continue
			}
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
// and are not members of any other team in the organization or repository.
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
		if strings.EqualFold(team.GetSlug(), teamSlug) {
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

// PlanCopyTeamMembers returns the changes required to add source team members to the destination team.
func PlanCopyTeamMembers(ctx context.Context, srcClient *GitHubClient, srcRepo repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dstRepo repository.Repository, dstTeamSlug string) ([]SyncChange, error) {
	srcMembers, err := ListTeamMembers(ctx, srcClient, srcRepo, srcTeamSlug, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members from source team: %w", err)
	}
	dstMembers, err := ListTeamMembers(ctx, dstClient, dstRepo, dstTeamSlug, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members from destination team: %w", err)
	}
	dstMemberMap := make(map[string]string)
	for _, m := range dstMembers {
		if m.Login != nil {
			role, err := teamMemberRole(m)
			if err != nil {
				return nil, err
			}
			dstMemberMap[*m.Login] = role
		}
	}
	var changes []SyncChange
	for _, m := range srcMembers {
		if m.Login == nil {
			continue
		}
		username := *m.Login
		if _, exists := dstMemberMap[username]; exists {
			continue
		}
		role, err := teamMemberRole(m)
		if err != nil {
			return nil, err
		}
		if err := validateDestinationOrgMembership(ctx, dstClient, dstRepo, username); err != nil {
			return changes, err
		}
		changes = append(changes, SyncChange{
			Destination: syncChangeTeamDestination(dstRepo, dstTeamSlug),
			Action:      SyncChangeAdd,
			Target:      username,
			Permission:  role,
		})
	}
	return changes, nil
}

// PlanSyncTeamMembers returns the changes required to make destination team membership match the source team.
func PlanSyncTeamMembers(ctx context.Context, srcClient *GitHubClient, srcRepo repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dstRepo repository.Repository, dstTeamSlug string) ([]SyncChange, error) {
	srcMembers, err := ListTeamMembers(ctx, srcClient, srcRepo, srcTeamSlug, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members from source team: %w", err)
	}
	dstMembers, err := ListTeamMembers(ctx, dstClient, dstRepo, dstTeamSlug, nil, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members from destination team: %w", err)
	}
	dstMemberMap := make(map[string]string)
	dstRoles := make(map[string]string)
	for _, m := range dstMembers {
		if m.Login != nil {
			role, err := teamMemberRole(m)
			if err != nil {
				return nil, err
			}
			dstMemberMap[*m.Login] = role
			dstRoles[*m.Login] = role
		}
	}
	srcMemberMap := make(map[string]string)
	for _, m := range srcMembers {
		if m.Login != nil {
			role, err := teamMemberRole(m)
			if err != nil {
				return nil, err
			}
			srcMemberMap[*m.Login] = role
		}
	}
	var changes []SyncChange
	for _, m := range srcMembers {
		if m.Login == nil {
			continue
		}
		username := *m.Login
		role := srcMemberMap[username]
		current, exists := dstMemberMap[username]
		action := SyncChangeAdd
		if exists {
			if dstRoles[username] == role {
				if err := validateDestinationOrgMembership(ctx, dstClient, dstRepo, username); err != nil {
					return changes, err
				}
				continue
			}
			action = SyncChangeUpdate
			current = dstRoles[username]
		}
		if err := validateDestinationOrgMembership(ctx, dstClient, dstRepo, username); err != nil {
			return changes, err
		}
		changes = append(changes, SyncChange{
			Destination: syncChangeTeamDestination(dstRepo, dstTeamSlug),
			Action:      action,
			Target:      username,
			Permission:  role,
			Current:     current,
		})
	}
	for _, m := range dstMembers {
		if m.Login == nil {
			continue
		}
		username := *m.Login
		if _, exists := srcMemberMap[username]; exists {
			continue
		}
		changes = append(changes, SyncChange{
			Destination: syncChangeTeamDestination(dstRepo, dstTeamSlug),
			Action:      SyncChangeRemove,
			Target:      username,
			Current:     dstRoles[username],
		})
	}
	return changes, nil
}

// ApplyTeamMemberChanges applies planned team membership changes.
func ApplyTeamMemberChanges(ctx context.Context, dstClient *GitHubClient, dstRepo repository.Repository, dstTeamSlug string, changes []SyncChange) error {
	for _, change := range changes {
		switch change.Action {
		case SyncChangeAdd, SyncChangeUpdate:
			if _, err := AddTeamMember(ctx, dstClient, dstRepo, dstTeamSlug, change.Target, change.Permission, false); err != nil {
				return fmt.Errorf("failed to add member %s to destination team: %w", change.Target, err)
			}
		case SyncChangeRemove:
			if err := RemoveTeamMember(ctx, dstClient, dstRepo, dstTeamSlug, change.Target); err != nil {
				return fmt.Errorf("failed to remove member %s from destination team: %w", change.Target, err)
			}
		default:
			return fmt.Errorf("unsupported team member change action %q", change.Action)
		}
	}
	return nil
}

// CopyTeamMembers copies members from the source team to the destination team (add only).
func CopyTeamMembers(ctx context.Context, srcClient *GitHubClient, srcRepo repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dstRepo repository.Repository, dstTeamSlug string) error {
	changes, err := PlanCopyTeamMembers(ctx, srcClient, srcRepo, srcTeamSlug, dstClient, dstRepo, dstTeamSlug)
	if applyErr := ApplyTeamMemberChanges(ctx, dstClient, dstRepo, dstTeamSlug, changes); applyErr != nil {
		return applyErr
	}
	return err
}

// SyncTeamMembers syncs members from the source team to the destination team.
func SyncTeamMembers(ctx context.Context, srcClient *GitHubClient, srcRepo repository.Repository, srcTeamSlug string, dstClient *GitHubClient, dstRepo repository.Repository, dstTeamSlug string) error {
	changes, err := PlanSyncTeamMembers(ctx, srcClient, srcRepo, srcTeamSlug, dstClient, dstRepo, dstTeamSlug)
	if applyErr := ApplyTeamMemberChanges(ctx, dstClient, dstRepo, dstTeamSlug, changes); applyErr != nil {
		return applyErr
	}
	return err
}

func teamMemberRole(member *github.User) (string, error) {
	if member.RoleName == nil || *member.RoleName == "" {
		return "", fmt.Errorf("team member %s has no membership role", member.GetLogin())
	}
	return *member.RoleName, nil
}

func validateDestinationOrgMembership(ctx context.Context, client *GitHubClient, repo repository.Repository, username string) error {
	membership, err := client.FindOrgMembership(ctx, repo.Owner, username)
	if err != nil {
		return fmt.Errorf("failed to check organization membership for user %s: %w", username, err)
	}
	if membership == nil {
		return fmt.Errorf("user '%s' is not a member of the organization '%s'", username, repo.Owner)
	}
	return nil
}
