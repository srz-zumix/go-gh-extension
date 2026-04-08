package gh

import (
	"context"
	"errors"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// MavenRegistryBase returns the Maven registry base URL for the given GitHub host, owner, and repository.
func MavenRegistryBase(host, owner, repo string) string {
	return client.MavenRegistryBase(host, owner, repo)
}

// MavenArtifactURL returns the URL for a Maven artifact file.
func MavenArtifactURL(host, owner, repo, groupID, artifactID, version, classifier, ext string) string {
	return client.MavenArtifactURL(host, owner, repo, groupID, artifactID, version, classifier, ext)
}

// ParseMavenPackageName splits a Maven package name into groupID and artifactID.
func ParseMavenPackageName(packageName string) (groupID, artifactID string, err error) {
	return client.ParseMavenPackageName(packageName)
}

// IsMavenConflictError returns true if err indicates that a Maven artifact already exists (HTTP 409 Conflict).
func IsMavenConflictError(err error) bool {
	var pushErr *client.MavenPushError
	return errors.As(err, &pushErr) && pushErr.IsConflict()
}

// DownloadMavenArtifacts downloads the primary Maven artifacts (.pom and .jar) for a package version.
// The repository name is taken from repo.Name; it must be non-empty for Maven.
func DownloadMavenArtifacts(ctx context.Context, g *GitHubClient, repo repository.Repository, packageName, version string) ([]client.MavenArtifact, error) {
	artifacts, err := g.DownloadMavenArtifacts(ctx, repo.Owner, repo.Name, packageName, version)
	if err != nil {
		return nil, fmt.Errorf("failed to download maven artifacts: %w", err)
	}
	return artifacts, nil
}

// PushMavenArtifact pushes a single Maven artifact file to the registry.
// The repository name is taken from repo.Name; it must be non-empty for Maven.
func PushMavenArtifact(ctx context.Context, g *GitHubClient, repo repository.Repository, packageName, version, classifier, ext string, data []byte) error {
	if err := g.PushMavenArtifact(ctx, repo.Owner, repo.Name, packageName, version, classifier, ext, data); err != nil {
		return fmt.Errorf("failed to push maven artifact: %w", err)
	}
	return nil
}
