package gh

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// PackageTypes is a list of valid package types.
var PackageTypes = []string{"npm", "maven", "rubygems", "docker", "nuget", "container"}

// ContainerRegistry returns the container registry host for the given GitHub host.
// For github.com, it returns "ghcr.io".
// For GitHub Enterprise Server, it returns "containers.HOSTNAME".
func ContainerRegistry(host string) string {
	if host == "" || host == defaultHost {
		return "ghcr.io"
	}
	return "containers." + host
}

// ContainerImageBase returns the base image path for an OCI image: "registry/owner/package".
// Owner and package are lowercased to comply with the OCI Distribution Spec.
func ContainerImageBase(host, owner, pkg string) string {
	return ContainerRegistry(host) + "/" + strings.ToLower(owner) + "/" + strings.ToLower(pkg)
}

// NuGetRegistryBase returns the NuGet registry base URL for the given GitHub host and owner.
// For github.com, it returns "https://nuget.pkg.github.com/<owner>".
// For GitHub Enterprise Server, it returns "https://<host>/_registry/nuget/<owner>".
func NuGetRegistryBase(host, owner string) string {
	return client.NuGetRegistryBase(host, owner)
}

// NuGetDownloadURL returns the URL to download a .nupkg file from the GitHub NuGet registry.
// Package name is lowercased to comply with NuGet V3 API conventions.
func NuGetDownloadURL(host, owner, packageName, version string) string {
	return client.NuGetDownloadURL(host, owner, packageName, version)
}

// NuGetPushURL returns the URL to push a .nupkg file to the GitHub NuGet registry.
func NuGetPushURL(host, owner string) string {
	return client.NuGetPushURL(host, owner)
}

// RewriteNuPkgRepository rewrites the <repository> element in the .nuspec inside the src
// .nupkg file to use the given repository URL.
// If destPath is empty, a new temporary file is created and returned.
// If destPath equals src.Name(), the file is rewritten in place (truncate + overwrite).
// Otherwise the result is written to a new file at destPath.
// The caller is responsible for closing and removing the returned file.
func RewriteNuPkgRepository(src *os.File, repoURL, destPath string) (_ *os.File, retErr error) {
	info, err := src.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat nupkg: %w", err)
	}

	// Always write to a temp file first to handle the same-file case safely.
	tmp, err := os.CreateTemp("", "nupkg-rewritten-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for rewrite: %w", err)
	}

	if err := client.RewriteNuPkgRepository(src, info.Size(), tmp, repoURL); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return nil, err
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return nil, fmt.Errorf("failed to seek rewritten nupkg: %w", err)
	}

	if destPath == "" {
		// Return the temp file directly; caller is responsible for cleanup.
		return tmp, nil
	}

	// Copy temp content to destPath, then close and remove temp regardless.
	defer func() {
		if err := tmp.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("failed to close temp nupkg: %w", err)
		}
		_ = os.Remove(tmp.Name())
	}()

	var dst *os.File
	if destPath == src.Name() {
		// In-place rewrite: truncate src and overwrite its content.
		if err := src.Truncate(0); err != nil {
			return nil, fmt.Errorf("failed to truncate %s: %w", destPath, err)
		}
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek %s: %w", destPath, err)
		}
		dst = src
	} else {
		var createErr error
		dst, createErr = os.Create(destPath)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create file %s for rewrite: %w", destPath, createErr)
		}
		defer func() {
			if retErr != nil {
				_ = dst.Close()
				_ = os.Remove(dst.Name())
			}
		}()
	}

	if _, err := io.Copy(dst, tmp); err != nil {
		return nil, fmt.Errorf("failed to write rewritten nupkg to %s: %w", destPath, err)
	}
	if _, err := dst.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek %s: %w", destPath, err)
	}
	return dst, nil
}

// DownloadNuGetPackage downloads a .nupkg file from the GitHub NuGet registry.
// If destPath is empty, a temporary file is created; otherwise the file at destPath is
// created (or truncated). The caller is responsible for closing the returned file and for
// removing it when destPath is empty (temporary file). For a non-empty destPath, the file
// is created at the specified location and is not treated as a temporary file by default.
func DownloadNuGetPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, packageName, version, destPath string) (_ *os.File, retErr error) {
	var (tmp *os.File; err error)
	if destPath == "" {
		tmp, err = os.CreateTemp("", "nupkg-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file for download: %w", err)
		}
	} else {
		tmp, err = os.Create(destPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s for download: %w", destPath, err)
		}
	}
	defer func() {
		if retErr != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()
	if err := g.DownloadNuGetPackage(ctx, repo.Owner, packageName, version, tmp); err != nil {
		return nil, fmt.Errorf("failed to download NuGet package '%s' version '%s': %w", packageName, version, err)
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek downloaded nupkg: %w", err)
	}
	return tmp, nil
}

// PushNuGetPackage pushes a .nupkg from the given reader to the GitHub NuGet registry,
// streaming the payload without buffering the full content in memory.
func PushNuGetPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, r io.Reader) error {
	if err := g.PushNuGetPackage(ctx, repo.Owner, r); err != nil {
		return fmt.Errorf("failed to push NuGet package: %w", err)
	}
	return nil
}

// ListOrgPackages lists packages in an organization.
// If packageType is empty, lists packages for all package types.
func ListOrgPackages(ctx context.Context, g *GitHubClient, repo repository.Repository, packageType, visibility string) ([]*github.Package, error) {
	if packageType == "" {
		return listOrgPackagesAllTypes(ctx, g, repo, visibility)
	}
	opts := &github.PackageListOptions{
		PackageType: github.Ptr(packageType),
	}
	if visibility != "" {
		opts.Visibility = github.Ptr(visibility)
	}
	packages, err := g.ListOrgPackages(ctx, repo.Owner, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages for organization '%s': %w", repo.Owner, err)
	}
	return packages, nil
}

func listOrgPackagesAllTypes(ctx context.Context, g *GitHubClient, repo repository.Repository, visibility string) ([]*github.Package, error) {
	var allPackages []*github.Package
	for _, pt := range PackageTypes {
		opts := &github.PackageListOptions{
			PackageType: github.Ptr(pt),
		}
		if visibility != "" {
			opts.Visibility = github.Ptr(visibility)
		}
		packages, err := g.ListOrgPackages(ctx, repo.Owner, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list %s packages for organization '%s': %w", pt, repo.Owner, err)
		}
		allPackages = append(allPackages, packages...)
	}
	return allPackages, nil
}

// GetOrgPackage gets a specific package in an organization.
func GetOrgPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, packageType, packageName string) (*github.Package, error) {
	pkg, err := g.GetOrgPackage(ctx, repo.Owner, packageType, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get package '%s' in organization '%s': %w", packageName, repo.Owner, err)
	}
	return pkg, nil
}

// DeleteOrgPackage deletes an entire package in an organization.
func DeleteOrgPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, packageType, packageName string) error {
	err := g.DeleteOrgPackage(ctx, repo.Owner, packageType, packageName)
	if err != nil {
		return fmt.Errorf("failed to delete package '%s' in organization '%s': %w", packageName, repo.Owner, err)
	}
	return nil
}

// RestoreOrgPackage restores an entire package in an organization.
func RestoreOrgPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, packageType, packageName string) error {
	err := g.RestoreOrgPackage(ctx, repo.Owner, packageType, packageName)
	if err != nil {
		return fmt.Errorf("failed to restore package '%s' in organization '%s': %w", packageName, repo.Owner, err)
	}
	return nil
}

// ListOrgPackageVersions lists package versions for a package owned by an organization.
func ListOrgPackageVersions(ctx context.Context, g *GitHubClient, repo repository.Repository, packageType, packageName string, state string) ([]*github.PackageVersion, error) {
	opts := &github.PackageListOptions{}
	if state != "" {
		opts.State = github.Ptr(state)
	}
	versions, err := g.ListOrgPackageVersions(ctx, repo.Owner, packageType, packageName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions for package '%s' in organization '%s': %w", packageName, repo.Owner, err)
	}
	return versions, nil
}

// GetOrgPackageVersion gets a specific package version in an organization.
func GetOrgPackageVersion(ctx context.Context, g *GitHubClient, repo repository.Repository, packageType, packageName string, versionID int64) (*github.PackageVersion, error) {
	version, err := g.GetOrgPackageVersion(ctx, repo.Owner, packageType, packageName, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %d for package '%s' in organization '%s': %w", versionID, packageName, repo.Owner, err)
	}
	return version, nil
}

// DeleteOrgPackageVersion deletes a specific package version in an organization.
func DeleteOrgPackageVersion(ctx context.Context, g *GitHubClient, repo repository.Repository, packageType, packageName string, versionID int64) error {
	err := g.DeleteOrgPackageVersion(ctx, repo.Owner, packageType, packageName, versionID)
	if err != nil {
		return fmt.Errorf("failed to delete version %d for package '%s' in organization '%s': %w", versionID, packageName, repo.Owner, err)
	}
	return nil
}

// RestoreOrgPackageVersion restores a specific package version in an organization.
func RestoreOrgPackageVersion(ctx context.Context, g *GitHubClient, repo repository.Repository, packageType, packageName string, versionID int64) error {
	err := g.RestoreOrgPackageVersion(ctx, repo.Owner, packageType, packageName, versionID)
	if err != nil {
		return fmt.Errorf("failed to restore version %d for package '%s' in organization '%s': %w", versionID, packageName, repo.Owner, err)
	}
	return nil
}

// ListUserPackages lists packages for a user.
// If user is empty, lists packages for the authenticated user.
// If packageType is empty, lists packages for all package types.
func ListUserPackages(ctx context.Context, g *GitHubClient, user string, packageType, visibility string) ([]*github.Package, error) {
	if packageType == "" {
		return listUserPackagesAllTypes(ctx, g, user, visibility)
	}
	opts := &github.PackageListOptions{
		PackageType: github.Ptr(packageType),
	}
	if visibility != "" {
		opts.Visibility = github.Ptr(visibility)
	}
	packages, err := g.ListUserPackages(ctx, user, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages for user '%s': %w", user, err)
	}
	return packages, nil
}

func listUserPackagesAllTypes(ctx context.Context, g *GitHubClient, user string, visibility string) ([]*github.Package, error) {
	var allPackages []*github.Package
	for _, pt := range PackageTypes {
		opts := &github.PackageListOptions{
			PackageType: github.Ptr(pt),
		}
		if visibility != "" {
			opts.Visibility = github.Ptr(visibility)
		}
		packages, err := g.ListUserPackages(ctx, user, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list %s packages for user '%s': %w", pt, user, err)
		}
		allPackages = append(allPackages, packages...)
	}
	return allPackages, nil
}

// GetUserPackage gets a specific package for a user.
// If user is empty, gets the package for the authenticated user.
func GetUserPackage(ctx context.Context, g *GitHubClient, user string, packageType, packageName string) (*github.Package, error) {
	pkg, err := g.GetUserPackage(ctx, user, packageType, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get package '%s' for user '%s': %w", packageName, user, err)
	}
	return pkg, nil
}

// DeleteUserPackage deletes an entire package for a user.
// If user is empty, deletes the package for the authenticated user.
func DeleteUserPackage(ctx context.Context, g *GitHubClient, user string, packageType, packageName string) error {
	err := g.DeleteUserPackage(ctx, user, packageType, packageName)
	if err != nil {
		return fmt.Errorf("failed to delete package '%s' for user '%s': %w", packageName, user, err)
	}
	return nil
}

// RestoreUserPackage restores an entire package for a user.
// If user is empty, restores the package for the authenticated user.
func RestoreUserPackage(ctx context.Context, g *GitHubClient, user string, packageType, packageName string) error {
	err := g.RestoreUserPackage(ctx, user, packageType, packageName)
	if err != nil {
		return fmt.Errorf("failed to restore package '%s' for user '%s': %w", packageName, user, err)
	}
	return nil
}

// ListUserPackageVersions lists package versions for a package owned by a user.
// If user is empty, lists versions for the authenticated user.
func ListUserPackageVersions(ctx context.Context, g *GitHubClient, user string, packageType, packageName string, state string) ([]*github.PackageVersion, error) {
	opts := &github.PackageListOptions{}
	if state != "" {
		opts.State = github.Ptr(state)
	}
	versions, err := g.ListUserPackageVersions(ctx, user, packageType, packageName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions for package '%s' for user '%s': %w", packageName, user, err)
	}
	return versions, nil
}

// GetUserPackageVersion gets a specific package version for a user.
// If user is empty, gets the version for the authenticated user.
func GetUserPackageVersion(ctx context.Context, g *GitHubClient, user string, packageType, packageName string, versionID int64) (*github.PackageVersion, error) {
	version, err := g.GetUserPackageVersion(ctx, user, packageType, packageName, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %d for package '%s' for user '%s': %w", versionID, packageName, user, err)
	}
	return version, nil
}

// DeleteUserPackageVersion deletes a specific package version for a user.
// If user is empty, deletes the version for the authenticated user.
func DeleteUserPackageVersion(ctx context.Context, g *GitHubClient, user string, packageType, packageName string, versionID int64) error {
	err := g.DeleteUserPackageVersion(ctx, user, packageType, packageName, versionID)
	if err != nil {
		return fmt.Errorf("failed to delete version %d for package '%s' for user '%s': %w", versionID, packageName, user, err)
	}
	return nil
}

// RestoreUserPackageVersion restores a specific package version for a user.
// If user is empty, restores the version for the authenticated user.
func RestoreUserPackageVersion(ctx context.Context, g *GitHubClient, user string, packageType, packageName string, versionID int64) error {
	err := g.RestoreUserPackageVersion(ctx, user, packageType, packageName, versionID)
	if err != nil {
		return fmt.Errorf("failed to restore version %d for package '%s' for user '%s': %w", versionID, packageName, user, err)
	}
	return nil
}

// ListPackageVersionsByOwnerType lists package versions using the appropriate API based on owner type.
func ListPackageVersionsByOwnerType(ctx context.Context, g *GitHubClient, ownerType OwnerType, owner, packageType, packageName string) ([]*github.PackageVersion, error) {
	opts := &github.PackageListOptions{}
	switch ownerType {
	case OwnerTypeOrg:
		versions, err := g.ListOrgPackageVersions(ctx, owner, packageType, packageName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list versions for package '%s' in organization '%s': %w", packageName, owner, err)
		}
		return versions, nil
	case OwnerTypeUser:
		versions, err := g.ListUserPackageVersions(ctx, owner, packageType, packageName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list versions for package '%s' for user '%s': %w", packageName, owner, err)
		}
		return versions, nil
	default:
		return nil, fmt.Errorf("unknown owner type: %s", ownerType)
	}
}

// DeletePackageVersionByOwnerType deletes a package version using the appropriate API based on owner type.
func DeletePackageVersionByOwnerType(ctx context.Context, g *GitHubClient, ownerType OwnerType, owner, packageType, packageName string, versionID int64) error {
	switch ownerType {
	case OwnerTypeOrg:
		err := g.DeleteOrgPackageVersion(ctx, owner, packageType, packageName, versionID)
		if err != nil {
			return fmt.Errorf("failed to delete version %d for package '%s' in organization '%s': %w", versionID, packageName, owner, err)
		}
		return nil
	case OwnerTypeUser:
		err := g.DeleteUserPackageVersion(ctx, owner, packageType, packageName, versionID)
		if err != nil {
			return fmt.Errorf("failed to delete version %d for package '%s' for user '%s': %w", versionID, packageName, owner, err)
		}
		return nil
	default:
		return fmt.Errorf("unknown owner type: %s", ownerType)
	}
}

// VersionFilter defines criteria for filtering package versions.
type VersionFilter struct {
	VersionIDs []int64
	Latest     int
	Since      *time.Time
	Until      *time.Time
}

// FilterVersions applies the given filter to a list of package versions.
// Filters are applied as intersection (AND):
// 1. Filter by version IDs (if specified)
// 2. Filter by date range (since/until)
// 3. Sort by creation date descending, then apply latest N
func FilterVersions(versions []*github.PackageVersion, filter VersionFilter) []*github.PackageVersion {
	result := versions
	isNewSlice := false

	// Filter by version IDs
	if len(filter.VersionIDs) > 0 {
		var filtered []*github.PackageVersion
		for _, v := range result {
			if v.ID != nil && slices.Contains(filter.VersionIDs, *v.ID) {
				filtered = append(filtered, v)
			}
		}
		result = filtered
		isNewSlice = true
	}

	// Filter by since
	if filter.Since != nil {
		var filtered []*github.PackageVersion
		for _, v := range result {
			if v.CreatedAt != nil && !v.CreatedAt.Before(*filter.Since) {
				filtered = append(filtered, v)
			}
		}
		result = filtered
		isNewSlice = true
	}

	// Filter by until
	if filter.Until != nil {
		var filtered []*github.PackageVersion
		for _, v := range result {
			if v.CreatedAt != nil && !v.CreatedAt.After(*filter.Until) {
				filtered = append(filtered, v)
			}
		}
		result = filtered
		isNewSlice = true
	}

	// Apply latest N (sort by creation date descending, then truncate)
	if filter.Latest > 0 {
		if !isNewSlice {
			result = slices.Clone(result)
		}
		slices.SortFunc(result, func(a, b *github.PackageVersion) int {
			if a.CreatedAt == nil && b.CreatedAt == nil {
				return 0
			}
			if a.CreatedAt == nil {
				return 1
			}
			if b.CreatedAt == nil {
				return -1
			}
			return b.CreatedAt.Compare(a.CreatedAt.Time)
		})
		if len(result) > filter.Latest {
			result = result[:filter.Latest]
		}
	}

	return result
}

// GetVersionTags extracts tags from a package version's container metadata.
func GetVersionTags(v *github.PackageVersion) []string {
	metadata, ok := v.GetMetadata()
	if !ok || metadata.Container == nil {
		return nil
	}
	return metadata.Container.Tags
}

// GetVersionDigest returns the version name as digest if it looks like a digest.
func GetVersionDigest(v *github.PackageVersion) string {
	n := v.GetName()
	if strings.HasPrefix(n, "sha256:") {
		return n
	}
	return ""
}
