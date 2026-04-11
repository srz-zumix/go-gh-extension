package gh

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// RubyGemsRegistryBase returns the RubyGems registry base URL for the given GitHub host and owner.
func RubyGemsRegistryBase(host, owner string) string {
	return client.RubyGemsRegistryBase(host, owner)
}

// RubyGemsDownloadURL returns the URL to download a .gem file from the GitHub RubyGems registry.
func RubyGemsDownloadURL(host, owner, packageName, version string) string {
	return client.RubyGemsDownloadURL(host, owner, packageName, version)
}

// RubyGemsPushURL returns the URL to push a .gem file to the GitHub RubyGems registry.
func RubyGemsPushURL(host, owner string) string {
	return client.RubyGemsPushURL(host, owner)
}

// DownloadRubyGemsPackage downloads a .gem file from the GitHub RubyGems registry.
// Returns the gem content as bytes.
func DownloadRubyGemsPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, packageName, version string) ([]byte, error) {
	data, err := g.DownloadRubyGemsPackage(ctx, repo.Owner, packageName, version)
	if err != nil {
		return nil, fmt.Errorf("failed to download rubygems package: %w", err)
	}
	return data, nil
}

// PushRubyGemsPackage pushes a .gem file to the GitHub RubyGems registry.
func PushRubyGemsPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, gemData []byte) error {
	if err := g.PushRubyGemsPackage(ctx, repo.Owner, bytes.NewReader(gemData)); err != nil {
		return fmt.Errorf("failed to push rubygems package: %w", err)
	}
	return nil
}

// RewriteRubyGemsGitHubRepo rewrites the github_repo metadata field in a .gem file
// to reference the destination repository. This is needed when migrating gems between
// GitHub instances to avoid 404 errors caused by the source github_repo URL.
func RewriteRubyGemsGitHubRepo(gemData []byte, repo repository.Repository) ([]byte, error) {
	return client.RewriteGemGitHubRepo(gemData, repo.Host, repo.Owner, repo.Name)
}
