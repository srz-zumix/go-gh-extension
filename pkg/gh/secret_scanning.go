package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// SecretScanningPushProtectionSettings is the list of valid push protection settings.
var SecretScanningPushProtectionSettings = []string{
	"enabled",
	"disabled",
	"not-set",
}

// ListSecretScanningPatternConfigs lists secret scanning pattern configurations for an organization.
func ListSecretScanningPatternConfigs(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.SecretScanningPatternConfigs, error) {
	return g.ListOrgSecretScanningPatternConfigs(ctx, repo.Owner)
}

// UpdateSecretScanningPatternConfigs updates secret scanning pattern configurations for an organization.
func UpdateSecretScanningPatternConfigs(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *github.SecretScanningPatternConfigsUpdateOptions) (*github.SecretScanningPatternConfigsUpdate, error) {
	return g.UpdateOrgSecretScanningPatternConfigs(ctx, repo.Owner, opts)
}
