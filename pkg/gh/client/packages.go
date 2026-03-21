package client

// GitHub Packages API functions
// See: https://docs.github.com/rest/packages/packages

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListOrgPackages lists packages in an organization.
func (g *GitHubClient) ListOrgPackages(ctx context.Context, org string, opts *github.PackageListOptions) ([]*github.Package, error) {
	var allPackages []*github.Package
	if opts == nil {
		opts = &github.PackageListOptions{}
	}
	if opts.PerPage == 0 {
		opts.PerPage = defaultPerPage
	}

	for {
		packages, resp, err := g.client.Organizations.ListPackages(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		allPackages = append(allPackages, packages...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPackages, nil
}

// GetOrgPackage gets a specific package in an organization.
func (g *GitHubClient) GetOrgPackage(ctx context.Context, org, packageType, packageName string) (*github.Package, error) {
	pkg, _, err := g.client.Organizations.GetPackage(ctx, org, packageType, packageName)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

// DeleteOrgPackage deletes an entire package in an organization.
func (g *GitHubClient) DeleteOrgPackage(ctx context.Context, org, packageType, packageName string) error {
	_, err := g.client.Organizations.DeletePackage(ctx, org, packageType, packageName)
	return err
}

// RestoreOrgPackage restores an entire package in an organization.
func (g *GitHubClient) RestoreOrgPackage(ctx context.Context, org, packageType, packageName string) error {
	_, err := g.client.Organizations.RestorePackage(ctx, org, packageType, packageName)
	return err
}

// ListOrgPackageVersions lists package versions for a package owned by an organization.
func (g *GitHubClient) ListOrgPackageVersions(ctx context.Context, org, packageType, packageName string, opts *github.PackageListOptions) ([]*github.PackageVersion, error) {
	var allVersions []*github.PackageVersion
	if opts == nil {
		opts = &github.PackageListOptions{}
	}
	if opts.PerPage == 0 {
		opts.PerPage = defaultPerPage
	}

	for {
		versions, resp, err := g.client.Organizations.PackageGetAllVersions(ctx, org, packageType, packageName, opts)
		if err != nil {
			return nil, err
		}
		allVersions = append(allVersions, versions...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allVersions, nil
}

// GetOrgPackageVersion gets a specific package version in an organization.
func (g *GitHubClient) GetOrgPackageVersion(ctx context.Context, org, packageType, packageName string, packageVersionID int64) (*github.PackageVersion, error) {
	version, _, err := g.client.Organizations.PackageGetVersion(ctx, org, packageType, packageName, packageVersionID)
	if err != nil {
		return nil, err
	}
	return version, nil
}

// DeleteOrgPackageVersion deletes a specific package version in an organization.
func (g *GitHubClient) DeleteOrgPackageVersion(ctx context.Context, org, packageType, packageName string, packageVersionID int64) error {
	_, err := g.client.Organizations.PackageDeleteVersion(ctx, org, packageType, packageName, packageVersionID)
	return err
}

// RestoreOrgPackageVersion restores a specific package version in an organization.
func (g *GitHubClient) RestoreOrgPackageVersion(ctx context.Context, org, packageType, packageName string, packageVersionID int64) error {
	_, err := g.client.Organizations.PackageRestoreVersion(ctx, org, packageType, packageName, packageVersionID)
	return err
}

// ListUserPackages lists packages for a user.
func (g *GitHubClient) ListUserPackages(ctx context.Context, user string, opts *github.PackageListOptions) ([]*github.Package, error) {
	var allPackages []*github.Package
	if opts == nil {
		opts = &github.PackageListOptions{}
	}
	if opts.PerPage == 0 {
		opts.PerPage = defaultPerPage
	}

	for {
		packages, resp, err := g.client.Users.ListPackages(ctx, user, opts)
		if err != nil {
			return nil, err
		}
		allPackages = append(allPackages, packages...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPackages, nil
}

// GetUserPackage gets a specific package for a user.
func (g *GitHubClient) GetUserPackage(ctx context.Context, user, packageType, packageName string) (*github.Package, error) {
	pkg, _, err := g.client.Users.GetPackage(ctx, user, packageType, packageName)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

// DeleteUserPackage deletes an entire package for a user.
func (g *GitHubClient) DeleteUserPackage(ctx context.Context, user, packageType, packageName string) error {
	_, err := g.client.Users.DeletePackage(ctx, user, packageType, packageName)
	return err
}

// RestoreUserPackage restores an entire package for a user.
func (g *GitHubClient) RestoreUserPackage(ctx context.Context, user, packageType, packageName string) error {
	_, err := g.client.Users.RestorePackage(ctx, user, packageType, packageName)
	return err
}

// ListUserPackageVersions lists package versions for a package owned by a user.
func (g *GitHubClient) ListUserPackageVersions(ctx context.Context, user, packageType, packageName string, _ *github.PackageListOptions) ([]*github.PackageVersion, error) {
	versions, _, err := g.client.Users.ListUserPackageVersions(ctx, user, packageType, packageName)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// GetUserPackageVersion gets a specific package version for a user.
func (g *GitHubClient) GetUserPackageVersion(ctx context.Context, user, packageType, packageName string, packageVersionID int64) (*github.PackageVersion, error) {
	version, _, err := g.client.Users.PackageGetVersion(ctx, user, packageType, packageName, packageVersionID)
	if err != nil {
		return nil, err
	}
	return version, nil
}

// DeleteUserPackageVersion deletes a specific package version for a user.
func (g *GitHubClient) DeleteUserPackageVersion(ctx context.Context, user, packageType, packageName string, packageVersionID int64) error {
	_, err := g.client.Users.PackageDeleteVersion(ctx, user, packageType, packageName, packageVersionID)
	return err
}

// RestoreUserPackageVersion restores a specific package version for a user.
func (g *GitHubClient) RestoreUserPackageVersion(ctx context.Context, user, packageType, packageName string, packageVersionID int64) error {
	_, err := g.client.Users.PackageRestoreVersion(ctx, user, packageType, packageName, packageVersionID)
	return err
}
