package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

// GetLatestRelease gets the latest published release for a repository (wrapper)
func GetLatestRelease(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.RepositoryRelease, error) {
	return g.GetLatestRelease(ctx, repo.Owner, repo.Name)
}

// GetReleaseByTag gets a release by its tag name for a repository (wrapper)
func GetReleaseByTag(ctx context.Context, g *GitHubClient, repo repository.Repository, tag string) (*github.RepositoryRelease, error) {
	return g.GetReleaseByTag(ctx, repo.Owner, repo.Name, tag)
}
