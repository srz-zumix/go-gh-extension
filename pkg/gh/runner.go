package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// ListRunners lists all self-hosted runners for a repository (wrapper)
func ListRunners(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Runner, error) {
	if repo.Name == "" {
		return ListOrgRunners(ctx, g, repo)
	}
	return g.ListRunners(ctx, repo.Owner, repo.Name)
}

// FindRunner finds a self-hosted runner by name for a repository (wrapper)
func FindRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerName string) (*github.Runner, error) {
	if repo.Name == "" {
		return FindOrgRunner(ctx, g, repo, runnerName)
	}
	return g.FindRunner(ctx, repo.Owner, repo.Name, runnerName)
}

// GetRunner gets a single self-hosted runner for a repository (wrapper)
func GetRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64) (*github.Runner, error) {
	if repo.Name == "" {
		return GetRunner(ctx, g, repo, runnerID)
	}
	return g.GetRunner(ctx, repo.Owner, repo.Name, runnerID)
}

// ListOrgRunners lists all self-hosted runners for an organization (wrapper)
func ListOrgRunners(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Runner, error) {
	return g.ListOrgRunners(ctx, repo.Owner)
}

// FindOrgRunner finds a self-hosted runner by name for an organization (wrapper)
func FindOrgRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerName string) (*github.Runner, error) {
	return g.FindOrgRunner(ctx, repo.Owner, runnerName)
}

// GetOrgRunner gets a single self-hosted runner for an organization (wrapper)
func GetOrgRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64) (*github.Runner, error) {
	return g.GetOrgRunner(ctx, repo.Owner, runnerID)
}

// CreateRegistrationToken creates a registration token for a repository or organization (wrapper)
func CreateRegistrationToken(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.RegistrationToken, error) {
	if repo.Name == "" {
		return g.CreateOrgRegistrationToken(ctx, repo.Owner)
	}
	return g.CreateRegistrationToken(ctx, repo.Owner, repo.Name)
}

// RemoveRunner removes a self-hosted runner from a repository or organization (wrapper)
func RemoveRunner(ctx context.Context, g *GitHubClient, repo repository.Repository, runnerID int64) error {
	if repo.Name == "" {
		return g.RemoveOrgRunner(ctx, repo.Owner, runnerID)
	}
	return g.RemoveRunner(ctx, repo.Owner, repo.Name, runnerID)
}

// ListOrgRunnerGroups lists all organization runner groups (wrapper)
func ListOrgRunnerGroups(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.RunnerGroup, error) {
	return g.ListOrgRunnerGroups(ctx, repo.Owner)
}

// FindOrgRunnerGroupByName finds an organization runner group by name (wrapper)
func FindOrgRunnerGroupByName(ctx context.Context, g *GitHubClient, repo repository.Repository, groupName string) (*github.RunnerGroup, error) {
	groups, err := g.ListOrgRunnerGroups(ctx, repo.Owner)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if group.GetName() == groupName {
			return group, nil
		}
	}
	return nil, nil // Group not found
}

// CreateOrgRunnerGroup creates a new organization runner group (wrapper)
func CreateOrgRunnerGroup(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.RunnerGroup, error) {
	return g.CreateOrgRunnerGroup(ctx, repo.Owner, name)
}

// DeleteOrgRunnerGroup deletes an organization runner group by ID (wrapper)
func DeleteOrgRunnerGroup(ctx context.Context, g *GitHubClient, repo repository.Repository, groupID int64) error {
	return g.DeleteOrgRunnerGroup(ctx, repo.Owner, groupID)
}
