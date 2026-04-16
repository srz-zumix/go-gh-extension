package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// ListIDPGroups lists all IDP groups available in an organization.
func ListIDPGroups(ctx context.Context, g *GitHubClient, repo repository.Repository, query string) ([]*github.IDPGroup, error) {
	return g.ListIDPGroupsInOrganization(ctx, repo.Owner, query)
}

// ListIDPGroupsForTeam lists IDP groups connected to a team in an organization.
func ListIDPGroupsForTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) ([]*github.IDPGroup, error) {
	return g.ListIDPGroupsForTeamBySlug(ctx, repo.Owner, teamSlug)
}

// CreateOrUpdateIDPGroupConnections creates, updates, or removes IDP group connections for a team.
func CreateOrUpdateIDPGroupConnections(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, groups []*github.IDPGroup) ([]*github.IDPGroup, error) {
	return g.CreateOrUpdateIDPGroupConnectionsBySlug(ctx, repo.Owner, teamSlug, groups)
}

// ListExternalGroups lists all external groups available in an organization (EMU).
func ListExternalGroups(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.ExternalGroup, error) {
	return g.ListExternalGroupsInOrganization(ctx, repo.Owner, "")
}

// GetExternalGroupDetails fetches detailed info for each group in the list by calling GetExternalGroup per entry.
func GetExternalGroupDetails(ctx context.Context, g *GitHubClient, repo repository.Repository, groups []*github.ExternalGroup) ([]*github.ExternalGroup, error) {
	detailed := make([]*github.ExternalGroup, 0, len(groups))
	for _, grp := range groups {
		if grp.GroupID == nil {
			detailed = append(detailed, grp)
			continue
		}
		d, err := g.GetExternalGroup(ctx, repo.Owner, grp.GetGroupID())
		if err != nil {
			return nil, err
		}
		detailed = append(detailed, d)
	}
	return detailed, nil
}

// HasExternalGroupsInOrganization returns true if the organization has any external groups (EMU).
// Returns false on 404/403/400 (for example when EMU is not configured) or when the result is empty.
func HasExternalGroupsInOrganization(ctx context.Context, g *GitHubClient, repo repository.Repository) (bool, error) {
	groups, err := g.ListExternalGroupsInOrganization(ctx, repo.Owner, "")
	if err != nil {
		if IsHTTPNotFound(err) || IsHTTPForbidden(err) || IsHTTPBadRequest(err) {
			return false, nil
		}
		return false, err
	}
	return len(groups) > 0, nil
}

// ListExternalGroupsForTeam lists external groups connected to a team in an organization (EMU).
func ListExternalGroupsForTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) ([]*github.ExternalGroup, error) {
	return g.ListExternalGroupsForTeamBySlug(ctx, repo.Owner, teamSlug)
}

// SearchExternalGroups searches external groups (EMU) in an organization.
// When teamSlug is non-empty, it returns external groups connected to the specified team.
// When teamSlug is empty, it searches all external groups in the organization matching displayName.
// displayName and teamSlug are mutually exclusive; specifying both results in an error.
func SearchExternalGroups(ctx context.Context, g *GitHubClient, repo repository.Repository, displayName, teamSlug string) ([]*github.ExternalGroup, error) {
	if displayName != "" && teamSlug != "" {
		return nil, fmt.Errorf("cannot specify both display name and team slug")
	}
	if teamSlug != "" {
		return g.ListExternalGroupsForTeamBySlug(ctx, repo.Owner, teamSlug)
	}
	return g.ListExternalGroupsInOrganization(ctx, repo.Owner, displayName)
}

// FindExternalGroupByTeamSlug returns the external group connected to a team, or nil if none is connected (EMU).
// Returns nil, nil on 404 or 400 (e.g. team has explicit members and cannot be externally managed),
// or when no external group is connected.
func FindExternalGroupByTeamSlug(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) (*github.ExternalGroup, error) {
	groups, err := g.ListExternalGroupsForTeamBySlug(ctx, repo.Owner, teamSlug)
	if err != nil {
		if IsHTTPNotFound(err) || IsHTTPBadRequest(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(groups) == 0 {
		return nil, nil
	}
	return groups[0], nil
}

// FindExternalGroupByName finds an external group by its exact name (EMU).
// Returns nil if no group with that name exists.
func FindExternalGroupByName(ctx context.Context, g *GitHubClient, repo repository.Repository, groupName string) (*github.ExternalGroup, error) {
	groups, err := g.ListExternalGroupsInOrganization(ctx, repo.Owner, groupName)
	if err != nil {
		return nil, err
	}
	for _, grp := range groups {
		if grp.GetGroupName() == groupName {
			return grp, nil
		}
	}
	return nil, nil
}

// GetExternalGroupByName retrieves an external group by name, returning an error if not found (EMU).
func GetExternalGroupByName(ctx context.Context, g *GitHubClient, repo repository.Repository, groupName string) (*github.ExternalGroup, error) {
	group, err := FindExternalGroupByName(ctx, g, repo, groupName)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("external group %q not found in organization %q", groupName, repo.Owner)
	}
	return group, nil
}

// ExternalGroupTeamDetail holds an external group and a connected team.
type ExternalGroupTeamDetail struct {
	Group *github.ExternalGroup
	Team  *github.Team
}

// GetExternalGroupTeams fetches the teams connected to an external group identified by name.
// For each ExternalGroupTeam entry the corresponding github.Team is fetched by slug.
// When group.Teams is nil (for example, when the field is omitted or unpopulated),
// it falls back to ScanExternalGroupTeams.
func GetExternalGroupTeams(ctx context.Context, g *GitHubClient, repo repository.Repository, groupName string) ([]*ExternalGroupTeamDetail, error) {
	group, err := GetExternalGroupByName(ctx, g, repo, groupName)
	if err != nil {
		return nil, err
	}
	if group.Teams == nil {
		return ScanExternalGroupTeams(ctx, g, repo, groupName)
	}
	var details []*ExternalGroupTeamDetail
	for _, t := range group.Teams {
		team, err := g.GetTeamBySlug(ctx, repo.Owner, t.GetTeamName())
		if err != nil {
			return nil, err
		}
		details = append(details, &ExternalGroupTeamDetail{
			Group: group,
			Team:  team,
		})
	}
	return details, nil
}

// ScanExternalGroupTeams finds teams connected to the named external group by
// scanning all top-level teams that have no child teams and checking each one
// via FindExternalGroupByTeamSlug. This is a fallback for when ExternalGroup.Teams is
// nil, omitted, or otherwise unpopulated, for example due to insufficient API permissions.
func ScanExternalGroupTeams(ctx context.Context, g *GitHubClient, repo repository.Repository, groupName string) ([]*ExternalGroupTeamDetail, error) {
	allTeams, err := g.ListTeams(ctx, repo.Owner)
	if err != nil {
		return nil, err
	}

	// Build a set of team IDs that are referenced as a parent by any team.
	parentIDs := make(map[int64]struct{})
	for _, t := range allTeams {
		if t.Parent != nil {
			parentIDs[t.Parent.GetID()] = struct{}{}
		}
	}

	var details []*ExternalGroupTeamDetail
	for _, t := range allTeams {
		// Only consider top-level teams
		if t.Parent != nil {
			continue
		}
		slug := t.GetSlug()
		if slug == "" {
			continue
		}
		// Skip teams that have child teams
		if _, hasChildren := parentIDs[t.GetID()]; hasChildren {
			continue
		}
		// Check if this team is connected to the target external group
		group, err := FindExternalGroupByTeamSlug(ctx, g, repo, slug)
		if err != nil {
			return nil, err
		}
		if group == nil || group.GetGroupName() != groupName {
			continue
		}
		details = append(details, &ExternalGroupTeamDetail{
			Group: group,
			Team:  t,
		})
	}
	return details, nil
}

// SetExternalGroupForTeam connects an external group (identified by name) to a team (EMU).
func SetExternalGroupForTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, groupName string, teamSlug string) (*github.ExternalGroup, error) {
	group, err := GetExternalGroupByName(ctx, g, repo, groupName)
	if err != nil {
		return nil, err
	}
	return g.UpdateConnectedExternalGroup(ctx, repo.Owner, teamSlug, group.GetGroupID())
}

// UnsetExternalGroupForTeam removes the connection between an external group and a team (EMU).
func UnsetExternalGroupForTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) error {
	return g.RemoveConnectedExternalGroup(ctx, repo.Owner, teamSlug)
}
