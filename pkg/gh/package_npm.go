package gh

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// NPMRegistryBase returns the npm registry base URL for the given GitHub host.
func NPMRegistryBase(host string) string {
	return client.NPMRegistryBase(host)
}

// NPMPackageURL returns the URL to download or access an npm package.
func NPMPackageURL(repo repository.Repository, packageName string) string {
	return client.NPMPackageURL(repo.Host, repo.Owner, packageName)
}

// DownloadNPMPackage downloads an npm package tarball from the GitHub npm registry.
// Returns the tarball content as bytes.
func DownloadNPMPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, packageName, version string) ([]byte, error) {
	data, err := g.DownloadNPMPackage(ctx, repo.Owner, packageName, version)
	if err != nil {
		return nil, fmt.Errorf("failed to download npm package: %w", err)
	}
	return data, nil
}

// PushNPMPackage publishes an npm package tarball to the GitHub npm registry.
func PushNPMPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, packageName string, tarballData []byte) error {
	if err := g.PushNPMPackage(ctx, repo.Owner, packageName, tarballData); err != nil {
		return fmt.Errorf("failed to push npm package: %w", err)
	}
	return nil
}

// RewriteNPMPackageJSON rewrites the package.json in an npm tarball to update the repository URL.
// Returns the modified tarball data.
func RewriteNPMPackageJSON(tarballData []byte, repoURL string) ([]byte, error) {
	modified, err := client.RewriteNPMPackageJSON(bytes.NewReader(tarballData), repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to rewrite npm package.json: %w", err)
	}
	return modified, nil
}
