package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

// Allowed values for the organization immutable releases enforced_repositories setting.
const (
	ImmutableReleasesEnforcedAll      = "all"
	ImmutableReleasesEnforcedNone     = "none"
	ImmutableReleasesEnforcedSelected = "selected"
)

// GetRepoImmutableReleases retrieves whether immutable releases are enabled for the repository.
func GetRepoImmutableReleases(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.RepoImmutableReleasesStatus, error) {
	status, err := g.GetRepoImmutableReleases(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get immutable releases status for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return status, nil
}

// EnableRepoImmutableReleases enables immutable releases for the repository.
func EnableRepoImmutableReleases(ctx context.Context, g *GitHubClient, repo repository.Repository) error {
	if err := g.EnableRepoImmutableReleases(ctx, repo.Owner, repo.Name); err != nil {
		return fmt.Errorf("failed to enable immutable releases for repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}
	return nil
}

// GetOrgImmutableReleasesSettings retrieves the immutable releases enforcement settings for the organization.
func GetOrgImmutableReleasesSettings(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.ImmutableReleaseSettings, error) {
	settings, err := g.GetOrgImmutableReleasesSettings(ctx, repo.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get immutable releases settings for organization '%s': %w", repo.Owner, err)
	}
	return settings, nil
}
