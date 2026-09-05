package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// GetLatestRelease retrieves the latest published release for a repository.
func (g *GitHubClient) GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, error) {
	release, _, err := g.client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return release, nil
}

// GetReleaseByTag retrieves a release by its tag name.
func (g *GitHubClient) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, error) {
	release, _, err := g.client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		return nil, err
	}
	return release, nil
}
