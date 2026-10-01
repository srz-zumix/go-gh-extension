package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// GetRepoImmutableReleases gets whether immutable releases are enabled for the repository.
func (g *GitHubClient) GetRepoImmutableReleases(ctx context.Context, owner, repo string) (*github.RepoImmutableReleasesStatus, error) {
	status, _, err := g.client.Repositories.AreImmutableReleasesEnabled(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// EnableRepoImmutableReleases enables immutable releases for the repository.
func (g *GitHubClient) EnableRepoImmutableReleases(ctx context.Context, owner, repo string) error {
	_, err := g.client.Repositories.EnableImmutableReleases(ctx, owner, repo)
	return err
}

// GetOrgImmutableReleasesSettings gets the immutable releases enforcement settings for the organization.
func (g *GitHubClient) GetOrgImmutableReleasesSettings(ctx context.Context, org string) (*github.ImmutableReleaseSettings, error) {
	settings, _, err := g.client.Organizations.GetImmutableReleasesSettings(ctx, org)
	if err != nil {
		return nil, err
	}
	return settings, nil
}
