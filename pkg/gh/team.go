package gh

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

type Team struct {
	Team  *github.Team
	Child []Team
}

func (t *Team) Flatten() []*github.Team {
	var teams []*github.Team
	if t.Team != nil {
		teams = append(teams, t.Team)
	}
	for _, child := range t.Child {
		teams = append(teams, child.Flatten()...)
	}
	return teams
}

func TeamByOwner(ctx context.Context, g *GitHubClient, repo repository.Repository, recursive bool) (Team, error) {
	var t Team
	if repo.Owner == "" {
		return t, nil
	}
	teams, err := g.ListTeams(ctx, repo.Owner)
	if err != nil {
		return t, err
	}
	for _, team := range teams {
		if team.Slug != nil && team.Parent == nil {
			c, err := TeamByName(ctx, g, repo, *team.Slug, false, recursive)
			if err != nil {
				return t, err
			}
			t.Child = append(t.Child, c)
		}
	}
	return t, nil
}

func TeamByName(ctx context.Context, g *GitHubClient, repo repository.Repository, teamName string, child bool, recursive bool) (Team, error) {
	var t Team
	if teamName == "" {
		return t, nil
	}
	if child {
		teams, err := g.ListChildTeams(ctx, repo.Owner, teamName)
		if err != nil {
			return t, err
		}
		for _, childTeam := range teams {
			if childTeam.Slug != nil {
				if recursive {
					recursiveTeams, err := TeamByName(ctx, g, repo, *childTeam.Slug, child, recursive)
					if err != nil {
						return t, err
					}
					recursiveTeams.Team = childTeam
					t.Child = append(t.Child, recursiveTeams)
				} else {
					t.Child = append(t.Child, Team{Team: childTeam})
				}
			}
		}
	} else {
		team, err := g.GetTeamBySlug(ctx, repo.Owner, teamName)
		if err != nil {
			return t, err
		}
		t.Team = team
		if recursive {
			teams, err := g.ListChildTeams(ctx, repo.Owner, teamName)
			if err != nil {
				return t, err
			}
			for _, childTeam := range teams {
				if childTeam.Slug != nil {
					recursiveTeams, err := TeamByName(ctx, g, repo, *childTeam.Slug, child, recursive)
					if err != nil {
						return t, err
					}
					t.Child = append(t.Child, recursiveTeams)
				}
			}
		}
	}
	return t, nil
}

// ListTeams is a wrapper function that uses a Repository object to call either ListTeams or ListTeamsByRepo.
func ListTeams(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Team, error) {
	if repo.Name != "" {
		return g.ListRepositoryTeams(ctx, repo.Owner, repo.Name)
	}
	return g.ListTeams(ctx, repo.Owner)
}

// GetTeamBySlug retrieves a team by its name.
func GetTeamBySlug(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) (*github.Team, error) {
	return g.GetTeamBySlug(ctx, repo.Owner, teamSlug)
}

// FindTeamBySlug retrieves a team by its name.
func FindTeamBySlug(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) (*github.Team, error) {
	return g.FindTeamBySlug(ctx, repo.Owner, teamSlug)
}

// IsExistsTeam checks if a team exists by its slug.
func IsExistsTeam(ctx context.Context, client *GitHubClient, repository repository.Repository, teamSlug string) (bool, error) {
	t, err := FindTeamBySlug(ctx, client, repository, teamSlug)
	if err != nil {
		return false, err
	}
	return t != nil, nil
}

// ListChildTeams is a wrapper function that calls the ListChildTeams API.
func ListChildTeams(ctx context.Context, g *GitHubClient, repo repository.Repository, parentSlug string) ([]*github.Team, error) {
	return g.ListChildTeams(ctx, repo.Owner, parentSlug)
}

// HasChildTeams checks if a team has child teams.
func HasChildTeams(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) (bool, error) {
	childTeams, err := ListChildTeams(ctx, g, repo, teamSlug)
	if err != nil {
		return false, err
	}
	return len(childTeams) > 0, nil
}

// RemoveTeamRepo is a wrapper function to remove a repository from a team.
func RemoveTeamRepo(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) error {
	return g.RemoveTeamRepo(ctx, repo.Owner, teamSlug, repo.Owner, repo.Name)
}

// AddTeamRepo is a wrapper function to add a repository to a team.
func AddTeamRepo(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, permission string) error {
	return g.AddTeamRepo(ctx, repo.Owner, teamSlug, repo.Owner, repo.Name, permission)
}

func ListTeamByName(ctx context.Context, g *GitHubClient, repo repository.Repository, teamNames []string, child bool, recursive bool) ([]*github.Team, error) {
	var teams []*github.Team
	for _, teamName := range teamNames {
		team, err := TeamByName(ctx, g, repo, teamName, child, recursive)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team.Flatten()...)
	}
	return teams, nil
}

// ListRepositoryTeams is a wrapper for the GitHubClient's ListRepositoryTeams method.
func ListRepositoryTeams(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Team, error) {
	return g.ListRepositoryTeams(ctx, repo.Owner, repo.Name)
}

func getNotificationSetting(enableNotification any) *string {
	switch v := enableNotification.(type) {
	case *bool:
		if v == nil {
			return nil
		}
		return getNotificationSetting(*v)
	case bool:
		if v {
			return github.Ptr("notifications_enabled")
		}
		return github.Ptr("notifications_disabled")
	case string:
		return github.Ptr(v)
	default:
		return nil
	}
}

func GetTeamID(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) (*int64, error) {
	if teamSlug == "" {
		return nil, nil
	}
	team, err := g.GetTeamBySlug(ctx, repo.Owner, teamSlug)
	if err != nil {
		return nil, err
	}
	if team == nil || team.ID == nil {
		return nil, fmt.Errorf("team '%s' not found in organization '%s'", teamSlug, repo.Owner)
	}
	return team.ID, nil
}

func GetTeamNodeID(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) (*string, error) {
	if teamSlug == "" {
		return nil, nil
	}
	team, err := g.GetTeamBySlug(ctx, repo.Owner, teamSlug)
	if err != nil {
		return nil, err
	}
	if team == nil || team.NodeID == nil {
		return nil, fmt.Errorf("team '%s' not found in organization '%s'", teamSlug, repo.Owner)
	}
	return team.NodeID, nil
}

// CreateTeam creates a new team in the specified organization.
func CreateTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, name string, description string, privacy string, enableNotification any, parentTeamSlug *string) (*github.Team, error) {
	newTeam := &github.NewTeam{
		Name:                name,
		Description:         &description,
		Privacy:             &privacy,
		NotificationSetting: getNotificationSetting(enableNotification),
		ParentTeamID:        nil, // ParentTeamSlug will be handled differently
	}

	if parentTeamSlug != nil {
		parentTeamID, err := GetTeamID(ctx, g, repo, *parentTeamSlug)
		if err != nil {
			return nil, err
		}
		newTeam.ParentTeamID = parentTeamID
	}

	return g.CreateTeam(ctx, repo.Owner, newTeam)
}

// DeleteTeam deletes a team by its slug in the specified repository.
func DeleteTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) error {
	return g.DeleteTeamBySlug(ctx, repo.Owner, teamSlug)
}

// UpdateTeam updates the details of a team in the specified repository.
func UpdateTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, name *string, description *string, privacy *string, enableNotification any, parentTeamSlug *string) (*github.Team, error) {
	team := &github.NewTeam{
		Name:                teamSlug,
		Description:         description,
		Privacy:             privacy,
		NotificationSetting: getNotificationSetting(enableNotification),
		ParentTeamID:        nil, // ParentTeamSlug will be handled differently
	}

	if name != nil {
		team.Name = *name
	}

	removeParent := false
	if parentTeamSlug != nil {
		if *parentTeamSlug != "" {
			parentTeamID, err := GetTeamID(ctx, g, repo, *parentTeamSlug)
			if err != nil {
				return nil, err
			}
			team.ParentTeamID = parentTeamID
		} else {
			// If parentTeamSlug is empty, remove the parent association
			removeParent = true
		}
	}

	return g.UpdateTeam(ctx, repo.Owner, teamSlug, team, removeParent)
}

func CreateOrUpdateTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, name *string, description string, privacy string, enableNotification any, parentTeamSlug *string) (*github.Team, error) {
	existingTeam, err := FindTeamBySlug(ctx, g, repo, teamSlug)
	if err != nil {
		return nil, err
	}
	if existingTeam != nil {
		return UpdateTeam(ctx, g, repo, teamSlug, name, &description, &privacy, &enableNotification, parentTeamSlug)
	}
	if name != nil {
		teamSlug = *name
	}
	return CreateTeam(ctx, g, repo, teamSlug, description, privacy, enableNotification, parentTeamSlug)
}

// RenameTeam renames a team by its slug in the specified repository.
func RenameTeam(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, newName string) (*github.Team, error) {
	return UpdateTeam(ctx, g, repo, teamSlug, &newName, nil, nil, nil, nil)
}

// CopyRepoTeamsAndPermissions copies teams and permissions from the source repository to the destination repository.
func CopyRepoTeamsAndPermissions(ctx context.Context, srcClient *GitHubClient, src repository.Repository, dstClient *GitHubClient, dst repository.Repository, force bool) error {
	// Fetch teams and permissions from the source repository
	srcTeams, err := srcClient.ListRepositoryTeams(ctx, src.Owner, src.Name)
	if err != nil {
		return fmt.Errorf("failed to fetch teams from source repository: %w", err)
	}

	// Iterate over each team and copy permissions to the destination repository
	for _, team := range srcTeams {
		permission := team.GetPermission()

		// Check if the team already has permissions on the destination repository
		if !force {
			existingRepo, err := dstClient.CheckTeamPermissions(ctx, dst.Owner, team.GetSlug(), dst.Owner, dst.Name)
			if err != nil {
				return fmt.Errorf("failed to check existing permissions for team %s: %w", team.GetSlug(), err)
			}
			if existingRepo != nil {
				existingPermission := GetRepositoryPermissions(existingRepo)
				if existingPermission == permission {
					continue // Skip if the permission is already the same
				}
				return fmt.Errorf("team %s already has %s permissions on the destination repository", team.GetSlug(), existingPermission)
			}
		}

		if err := dstClient.AddTeamRepo(ctx, src.Owner, team.GetSlug(), dst.Owner, dst.Name, permission); err != nil {
			return fmt.Errorf("failed to add team %s to destination repository: %w", team.GetSlug(), err)
		}
	}

	return nil
}

func SyncRepoTeamsAndPermissions(ctx context.Context, srcClient *GitHubClient, src repository.Repository, dstClient *GitHubClient, dst repository.Repository) error {
	// Fetch teams and permissions from the source repository
	srcTeams, err := srcClient.ListRepositoryTeams(ctx, src.Owner, src.Name)
	if err != nil {
		return fmt.Errorf("failed to fetch teams from source repository: %w", err)
	}

	// Fetch existing teams and permissions from the destination repository
	dstTeams, err := dstClient.ListRepositoryTeams(ctx, dst.Owner, dst.Name)
	if err != nil {
		return fmt.Errorf("failed to fetch teams from destination repository: %w", err)
	}

	dstTeamMap := make(map[string]string)
	for _, team := range dstTeams {
		dstTeamMap[team.GetSlug()] = team.GetPermission()
	}

	srcTeamMap := make(map[string]string)
	for _, team := range srcTeams {
		srcTeamMap[team.GetSlug()] = team.GetPermission()
	}

	// Sync teams and permissions
	for _, team := range srcTeams {
		permission := team.GetPermission()

		if existingPermission, exists := dstTeamMap[team.GetSlug()]; exists {
			if existingPermission == permission {
				continue // Skip if the permission is already the same
			}
		}

		if err := dstClient.AddTeamRepo(ctx, src.Owner, team.GetSlug(), dst.Owner, dst.Name, permission); err != nil {
			return fmt.Errorf("failed to sync team %s to destination repository: %w", team.GetSlug(), err)
		}
	}

	// Remove teams from the destination repository that are not in the source repository
	for _, team := range dstTeams {
		if _, exists := srcTeamMap[team.GetSlug()]; !exists {
			if err := dstClient.RemoveTeamRepo(ctx, dst.Owner, team.GetSlug(), dst.Owner, dst.Name); err != nil {
				return fmt.Errorf("failed to remove team %s from destination repository: %w", team.GetSlug(), err)
			}
		}
	}

	return nil
}

type TeamCodeReviewSettings struct {
	TeamSlug                     string
	NotifyTeam                   bool
	Enabled                      bool
	Algorithm                    string
	TeamMemberCount              int
	ExcludedTeamMembers          []string
	IncludeChildTeamMembers      *bool
	CountMembersAlreadyRequested *bool
	RemoveTeamRequest            *bool
}

var (
	TeamCodeReviewAlgorithmRoundRobin  = "ROUND_ROBIN"
	TeamCodeReviewAlgorithmLoadBalance = "LOAD_BALANCE"
)

var TeamCodeReviewAlgorithm = []string{
	TeamCodeReviewAlgorithmRoundRobin,
	TeamCodeReviewAlgorithmLoadBalance,
}

func GetTeamCodeReviewSettings(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string) (*TeamCodeReviewSettings, error) {
	s, err := g.GetTeamCodeReviewSettings(ctx, repo.Owner, teamSlug)
	if err != nil {
		return nil, err
	}

	expandedExcludedMembers := []string{}
	for _, id := range s.ExcludedTeamMemberIDs {
		userID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user ID %s: %w", id, err)
		}
		user, err := FindUserByID(ctx, g, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to find user by ID %s: %w", id, err)
		}
		if user != nil && user.Login != nil {
			expandedExcludedMembers = append(expandedExcludedMembers, *user.Login)
		}
	}

	expandedSettings := &TeamCodeReviewSettings{
		TeamSlug:                     s.TeamSlug,
		NotifyTeam:                   s.NotifyTeam,
		Enabled:                      s.Enabled,
		Algorithm:                    s.Algorithm,
		TeamMemberCount:              s.TeamMemberCount,
		ExcludedTeamMembers:          expandedExcludedMembers,
		IncludeChildTeamMembers:      s.IncludeChildTeamMembers,
		CountMembersAlreadyRequested: s.CountMembersAlreadyRequested,
		RemoveTeamRequest:            s.RemoveTeamRequest,
	}

	return expandedSettings, nil
}

func SetTeamCodeReviewSettings(ctx context.Context, g *GitHubClient, repo repository.Repository, teamSlug string, settings *TeamCodeReviewSettings) error {
	teamNodeID, err := GetTeamNodeID(ctx, g, repo, teamSlug)
	if err != nil {
		return err
	}
	if teamNodeID == nil {
		return fmt.Errorf("team '%s' not found in organization '%s'", teamSlug, repo.Owner)
	}

	s := &client.TeamCodeReviewSettings{
		TeamSlug:                     settings.TeamSlug,
		NotifyTeam:                   settings.NotifyTeam,
		Enabled:                      settings.Enabled,
		Algorithm:                    settings.Algorithm,
		TeamMemberCount:              settings.TeamMemberCount,
		ExcludedTeamMemberIDs:        nil,
		IncludeChildTeamMembers:      settings.IncludeChildTeamMembers,
		CountMembersAlreadyRequested: settings.CountMembersAlreadyRequested,
		RemoveTeamRequest:            settings.RemoveTeamRequest,
	}
	if settings.ExcludedTeamMembers != nil {
		s.ExcludedTeamMemberIDs = []string{}
		for _, username := range settings.ExcludedTeamMembers {
			user, err := FindUser(ctx, g, username)
			if err != nil {
				return fmt.Errorf("failed to find user '%s': %w", username, err)
			}
			if user == nil || user.NodeID == nil {
				return fmt.Errorf("user '%s' not found", username)
			}

			s.ExcludedTeamMemberIDs = append(s.ExcludedTeamMemberIDs, *user.NodeID)
		}
	}
	return g.SetTeamCodeReviewSettings(ctx, *teamNodeID, s)
}
