package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/shurcooL/githubv4"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// ProjectV2Role is the permission level granted to a Project v2 collaborator.
type ProjectV2Role = client.ProjectV2Role

// Re-export collaborator role constants.
const (
	ProjectV2RoleAdmin  = client.ProjectV2RoleAdmin
	ProjectV2RoleWriter = client.ProjectV2RoleWriter
	ProjectV2RoleReader = client.ProjectV2RoleReader
	ProjectV2RoleNone   = client.ProjectV2RoleNone
)

// LinkProjectV2ToRepository links a GitHub Project v2 to a repository.
func LinkProjectV2ToRepository(ctx context.Context, g *GitHubClient, projectID string, repo repository.Repository) error {
	repoID, err := GetRepositoryNodeID(ctx, g, repo)
	if err != nil {
		return err
	}
	return g.LinkProjectV2ToRepository(ctx, client.LinkProjectV2ToRepositoryInput{
		ProjectID:    githubv4.ID(projectID),
		RepositoryID: githubv4.ID(repoID),
	})
}

// UnlinkProjectV2FromRepository unlinks a GitHub Project v2 from a repository.
func UnlinkProjectV2FromRepository(ctx context.Context, g *GitHubClient, projectID string, repo repository.Repository) error {
	repoID, err := GetRepositoryNodeID(ctx, g, repo)
	if err != nil {
		return err
	}
	return g.UnlinkProjectV2FromRepository(ctx, client.UnlinkProjectV2FromRepositoryInput{
		ProjectID:    githubv4.ID(projectID),
		RepositoryID: githubv4.ID(repoID),
	})
}

// LinkProjectV2ToTeam links a GitHub Project v2 to a team of the repository's organization.
func LinkProjectV2ToTeam(ctx context.Context, g *GitHubClient, projectID string, repo repository.Repository, teamSlug string) error {
	teamID, err := GetTeamNodeID(ctx, g, repo, teamSlug)
	if err != nil {
		return err
	}
	if teamID == nil {
		return fmt.Errorf("team '%s' not found in '%s'", teamSlug, repo.Owner)
	}
	return g.LinkProjectV2ToTeam(ctx, client.LinkProjectV2ToTeamInput{
		ProjectID: githubv4.ID(projectID),
		TeamID:    githubv4.ID(*teamID),
	})
}

// UnlinkProjectV2FromTeam unlinks a GitHub Project v2 from a team of the repository's organization.
func UnlinkProjectV2FromTeam(ctx context.Context, g *GitHubClient, projectID string, repo repository.Repository, teamSlug string) error {
	teamID, err := GetTeamNodeID(ctx, g, repo, teamSlug)
	if err != nil {
		return err
	}
	if teamID == nil {
		return fmt.Errorf("team '%s' not found in '%s'", teamSlug, repo.Owner)
	}
	return g.UnlinkProjectV2FromTeam(ctx, client.UnlinkProjectV2FromTeamInput{
		ProjectID: githubv4.ID(projectID),
		TeamID:    githubv4.ID(*teamID),
	})
}

// SetProjectV2UserCollaborators grants the given role to the specified users on a GitHub Project v2.
// Use ProjectV2RoleNone to revoke access.
func SetProjectV2UserCollaborators(ctx context.Context, g *GitHubClient, projectID string, userIDs []string, role ProjectV2Role) error {
	collaborators := make([]client.ProjectV2Collaborator, 0, len(userIDs))
	for _, id := range userIDs {
		userID := githubv4.ID(id)
		collaborators = append(collaborators, client.ProjectV2Collaborator{
			UserID: &userID,
			Role:   githubv4.String(role),
		})
	}
	return g.UpdateProjectV2Collaborators(ctx, client.UpdateProjectV2CollaboratorsInput{
		ProjectID:     githubv4.ID(projectID),
		Collaborators: collaborators,
	})
}

// SetProjectV2TeamCollaborators grants the given role to the specified teams on a GitHub Project v2.
// Use ProjectV2RoleNone to revoke access.
func SetProjectV2TeamCollaborators(ctx context.Context, g *GitHubClient, projectID string, teamIDs []string, role ProjectV2Role) error {
	collaborators := make([]client.ProjectV2Collaborator, 0, len(teamIDs))
	for _, id := range teamIDs {
		teamID := githubv4.ID(id)
		collaborators = append(collaborators, client.ProjectV2Collaborator{
			TeamID: &teamID,
			Role:   githubv4.String(role),
		})
	}
	return g.UpdateProjectV2Collaborators(ctx, client.UpdateProjectV2CollaboratorsInput{
		ProjectID:     githubv4.ID(projectID),
		Collaborators: collaborators,
	})
}
