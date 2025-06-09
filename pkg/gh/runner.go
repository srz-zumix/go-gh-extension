package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v71/github"
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
