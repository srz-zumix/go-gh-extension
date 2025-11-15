package gh

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

func GetRepositoryDependencyGraphSBOM(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.SBOM, error) {
	sbom, err := g.GetDependencyGraphSBOM(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get SBOM for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return sbom, nil
}

func GetRepositoryDependencyGraphSBOMWithSubmodules(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.SBOM, error) {
	sbom, err := GetRepositoryDependencyGraphSBOM(ctx, g, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get SBOM for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}

	submodules, err := GetRepositorySubmodules(ctx, g, repo, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get submodules for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}
	submodules = FlattenRepositorySubmodules(submodules)

	var sboms []*github.SBOM
	sboms = append(sboms, sbom)

	for _, submodule := range submodules {
		submoduleSBOM, err := g.GetDependencyGraphSBOM(ctx, submodule.Repository.Owner, submodule.Repository.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get SBOM for submodule %s/%s: %w", submodule.Repository.Owner, submodule.Repository.Name, err)
		}
		sboms = append(sboms, submoduleSBOM)
	}

	return sboms, nil
}

func FilterSBOMPackage(sbom *github.SBOM, packageName string) *github.SBOM {
	filteredSBOM := &github.SBOM{
		SBOM: &github.SBOMInfo{
			SPDXID:            sbom.SBOM.SPDXID,
			SPDXVersion:       sbom.SBOM.SPDXVersion,
			CreationInfo:      sbom.SBOM.CreationInfo,
			Name:              sbom.SBOM.Name,
			DataLicense:       sbom.SBOM.DataLicense,
			DocumentDescribes: sbom.SBOM.DocumentDescribes,
			DocumentNamespace: sbom.SBOM.DocumentNamespace,
			Relationships:     sbom.SBOM.Relationships,
		},
	}
	for _, pkg := range sbom.SBOM.Packages {
		parts := strings.Split(*pkg.Name, ":")
		if parts[0] == packageName {
			filteredSBOM.SBOM.Packages = append(filteredSBOM.SBOM.Packages, pkg)
		}
	}
	// Sort by Name
	slices.SortFunc(filteredSBOM.SBOM.Packages, func(a, b *github.RepoDependencies) int {
		if a.Name == nil && b.Name == nil {
			return 0
		}
		if a.Name == nil {
			return -1
		}
		if b.Name == nil {
			return 1
		}
		return strings.Compare(*a.Name, *b.Name)
	})
	return filteredSBOM
}
