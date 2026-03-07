package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

// PackageTypes is a list of valid package types.
var PackageTypes = []string{"npm", "maven", "rubygems", "docker", "nuget", "container"}

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
