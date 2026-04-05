// Package gh provides wrapper functions for GitHub Projects (classic) v1 API operations.
package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// Type aliases for ProjectV1 types from the client package.
type ProjectV1 = client.ProjectV1
type ProjectV1Column = client.ProjectV1Column
type ProjectV1Card = client.ProjectV1Card

// ListProjectsV1 lists all classic projects for a repository owner or repository.
// If repo.Name is non-empty, the repository's projects are listed.
// Otherwise, the owner's projects are listed with automatic org/user detection.
func ListProjectsV1(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]ProjectV1, error) {
	if repo.Name != "" {
		return g.ListRepoProjectsV1(ctx, repo.Owner, repo.Name)
	}
	ownerType, err := DetectOwnerType(ctx, g, repo.Owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.ListOrgProjectsV1(ctx, repo.Owner)
	case OwnerTypeUser:
		return g.ListUserProjectsV1(ctx, repo.Owner)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", repo.Owner)
	}
}

// GetProjectV1ByNumber finds a classic project by repository (or owner) and project number.
// It lists all matching projects and returns the first one with the given number.
func GetProjectV1ByNumber(ctx context.Context, g *GitHubClient, repo repository.Repository, number int) (*ProjectV1, error) {
	projects, err := ListProjectsV1(ctx, g, repo)
	if err != nil {
		return nil, err
	}
	for i := range projects {
		if projects[i].Number == number {
			return &projects[i], nil
		}
	}
	repoScope := repo.Owner
	if repo.Name != "" {
		repoScope = repo.Owner + "/" + repo.Name
	}
	return nil, fmt.Errorf("project #%d not found for '%s'", number, repoScope)
}

// ListProjectV1Columns lists all columns for a classic project.
func ListProjectV1Columns(ctx context.Context, g *GitHubClient, repo repository.Repository, projectID int64) ([]ProjectV1Column, error) {
	return g.ListProjectV1Columns(ctx, projectID)
}

// ListProjectV1Cards lists all cards for a classic project column.
func ListProjectV1Cards(ctx context.Context, g *GitHubClient, repo repository.Repository, columnID int64) ([]ProjectV1Card, error) {
	return g.ListProjectV1Cards(ctx, columnID)
}

// CreateProjectV1 creates a classic project under repo.
// If repo.Name is non-empty, the project is created in that repository.
// Otherwise, the owner type is detected and the appropriate endpoint is used.
func CreateProjectV1(ctx context.Context, g *GitHubClient, repo repository.Repository, name, body string) (*ProjectV1, error) {
	if repo.Name != "" {
		return g.CreateRepoProjectV1(ctx, repo.Owner, repo.Name, name, body)
	}
	ownerType, err := DetectOwnerType(ctx, g, repo.Owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.CreateOrgProjectV1(ctx, repo.Owner, name, body)
	case OwnerTypeUser:
		return g.CreateUserProjectV1(ctx, name, body)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", repo.Owner)
	}
}

// DeleteProjectV1 deletes a classic project by its ID.
func DeleteProjectV1(ctx context.Context, g *GitHubClient, projectID int64) error {
	return g.DeleteProjectV1(ctx, projectID)
}

// CreateProjectV1Column creates a column in a classic project.
func CreateProjectV1Column(ctx context.Context, g *GitHubClient, projectID int64, name string) (*ProjectV1Column, error) {
	return g.CreateProjectV1Column(ctx, projectID, name)
}

// CreateProjectV1Card creates a note card in a classic project column.
func CreateProjectV1Card(ctx context.Context, g *GitHubClient, columnID int64, note string) (*ProjectV1Card, error) {
	return g.CreateProjectV1Card(ctx, columnID, note)
}
