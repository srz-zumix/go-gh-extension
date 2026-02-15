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

func GetRepositoryDependencyGraphSBOMWithSubmodules(ctx context.Context, g *GitHubClient, repo repository.Repository, submodule bool, recursive bool) ([]*github.SBOM, error) {
	sbom, err := GetRepositoryDependencyGraphSBOM(ctx, g, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get SBOM for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}

	if !submodule {
		return []*github.SBOM{sbom}, nil
	}

	submodules, err := GetRepositorySubmodules(ctx, g, repo, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get submodules for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}
	submodules = FlattenRepositorySubmodules(submodules)

	var sboms []*github.SBOM
	sboms = append(sboms, sbom)

	for _, submodule := range submodules {
		submoduleSBOMs, err := GetRepositoryDependencyGraphSBOMWithSubmodules(ctx, g, repository.Repository{
			Host:  repo.Host,
			Owner: submodule.Repository.Owner,
			Name:  submodule.Repository.Name,
		}, recursive, recursive)
		if err != nil {
			return nil, fmt.Errorf("failed to get SBOM for submodule %s/%s: %w", submodule.Repository.Owner, submodule.Repository.Name, err)
		}
		sboms = append(sboms, submoduleSBOMs...)
	}

	return sboms, nil
}

// parseActionRepository extracts the repository from an action package name (e.g. "actions:owner/repo" or "actions:owner/repo/path")
func parseActionRepository(packageName string, host string) (repository.Repository, bool) {
	parts := strings.Split(packageName, ":")
	if len(parts) < 2 {
		return repository.Repository{}, false
	}
	segments := strings.SplitN(parts[1], "/", 3)
	if len(segments) < 2 {
		return repository.Repository{}, false
	}
	return repository.Repository{
		Host:  host,
		Owner: segments[0],
		Name:  segments[1],
	}, true
}

// GetActionsDependenciesRecursive retrieves actions dependencies by traversing referenced action repositories recursively using BFS
func GetActionsDependenciesRecursive(ctx context.Context, g *GitHubClient, repo repository.Repository, recursive bool) ([]*github.SBOM, error) {
	visited := make(map[string]bool)
	repoKey := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	visited[repoKey] = true

	sbom, err := GetRepositoryDependencyGraphSBOM(ctx, g, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get SBOM for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}

	actionsSBOM := FilterSBOMPackage(sbom, "actions")
	var sboms []*github.SBOM
	sboms = append(sboms, actionsSBOM)

	if !recursive {
		return sboms, nil
	}

	// BFS through action dependencies
	var queue []repository.Repository
	for _, pkg := range actionsSBOM.SBOM.Packages {
		if actionRepo, ok := parseActionRepository(*pkg.Name, repo.Host); ok {
			queue = append(queue, actionRepo)
		}
	}

	for len(queue) > 0 {
		actionRepo := queue[0]
		queue = queue[1:]

		key := fmt.Sprintf("%s/%s", actionRepo.Owner, actionRepo.Name)
		if visited[key] {
			continue
		}
		visited[key] = true

		depSBOM, err := GetRepositoryDependencyGraphSBOM(ctx, g, actionRepo)
		if err != nil {
			// skip repos that are not accessible
			continue
		}

		depActionsSBOM := FilterSBOMPackage(depSBOM, "actions")
		sboms = append(sboms, depActionsSBOM)

		for _, pkg := range depActionsSBOM.SBOM.Packages {
			if nextRepo, ok := parseActionRepository(*pkg.Name, repo.Host); ok {
				queue = append(queue, nextRepo)
			}
		}
	}

	return sboms, nil
}

func FlattenSBOMPackages(sboms []*github.SBOM) []*github.RepoDependencies {
	var allPackages []*github.RepoDependencies
	for _, sbom := range sboms {
		allPackages = append(allPackages, sbom.SBOM.Packages...)
	}
	return allPackages
}

func SelectSBOMPackage(deps []*github.RepoDependencies, packageName string) []*github.RepoDependencies {
	var selected []*github.RepoDependencies
	for _, dep := range deps {
		if dep.Name == nil {
			// Skip dependencies without a name to avoid nil pointer dereference
			continue
		}
		parts := strings.Split(*dep.Name, ":")
		if parts[0] == packageName {
			selected = append(selected, dep)
		}
	}
	// Sort by Name
	slices.SortFunc(selected, func(a, b *github.RepoDependencies) int {
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
	return selected
}

func FilterSBOMPackage(sbom *github.SBOM, packageName string) *github.SBOM {
	packages := SelectSBOMPackage(sbom.SBOM.Packages, packageName)
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
			Packages:          packages,
		},
	}
	return filteredSBOM
}

func FilterSBOMsPackage(sboms []*github.SBOM, packageName string) []*github.SBOM {
	var filteredSBOMs []*github.SBOM
	for _, sbom := range sboms {
		filteredSBOM := FilterSBOMPackage(sbom, packageName)
		filteredSBOMs = append(filteredSBOMs, filteredSBOM)
	}
	return filteredSBOMs
}
