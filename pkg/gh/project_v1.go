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

// ListProjectsV1 lists all classic projects for an owner (org or user) or repository.
// If repoName is non-empty, the repository's projects are listed.
// Otherwise, the owner's projects are listed with automatic org/user detection.
func ListProjectsV1(ctx context.Context, g *GitHubClient, owner, repoName string) ([]ProjectV1, error) {
	if repoName != "" {
		return g.ListRepoProjectsV1(ctx, owner, repoName)
	}
	ownerType, err := DetectOwnerType(ctx, g, owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.ListOrgProjectsV1(ctx, owner)
	case OwnerTypeUser:
		return g.ListUserProjectsV1(ctx, owner)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", owner)
	}
}

// GetProjectV1ByNumber finds a classic project by owner, optional repository name, and project number.
// It lists all matching projects and returns the first one with the given number.
func GetProjectV1ByNumber(ctx context.Context, g *GitHubClient, owner, repoName string, number int) (*ProjectV1, error) {
	projects, err := ListProjectsV1(ctx, g, owner, repoName)
	if err != nil {
		return nil, err
	}
	for i := range projects {
		if projects[i].Number == number {
			return &projects[i], nil
		}
	}
	return nil, fmt.Errorf("project #%d not found for '%s'", number, owner)
}

// ListProjectV1Columns lists all columns for a classic project.
func ListProjectV1Columns(ctx context.Context, g *GitHubClient, repo repository.Repository, projectID int64) ([]ProjectV1Column, error) {
	return g.ListProjectV1Columns(ctx, projectID)
}

// ListProjectV1Cards lists all cards for a classic project column.
func ListProjectV1Cards(ctx context.Context, g *GitHubClient, repo repository.Repository, columnID int64) ([]ProjectV1Card, error) {
	return g.ListProjectV1Cards(ctx, columnID)
}

// CreateProjectV1 creates a classic project for an owner or repository.
// If repoName is non-empty, the project is created in that repository.
// Otherwise, the owner type is detected and the appropriate endpoint is used.
func CreateProjectV1(ctx context.Context, g *GitHubClient, owner, repoName, name, body string) (*ProjectV1, error) {
	if repoName != "" {
		return g.CreateRepoProjectV1(ctx, owner, repoName, name, body)
	}
	ownerType, err := DetectOwnerType(ctx, g, owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.CreateOrgProjectV1(ctx, owner, name, body)
	case OwnerTypeUser:
		return g.CreateUserProjectV1(ctx, name, body)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", owner)
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
