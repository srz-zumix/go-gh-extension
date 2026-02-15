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

// GetActionsDependencyGraph retrieves actions dependency graph with SBOMs and edges by traversing referenced action repositories recursively
func GetActionsDependencyGraph(ctx context.Context, g *GitHubClient, repo repository.Repository, recursive bool) ([]*github.SBOM, []GraphEdge, error) {
	visited := make(map[string]bool)
	return getActionsDependencyGraphInternal(ctx, g, repo, recursive, visited)
}

// getActionsDependencyGraphInternal is an internal helper that tracks visited repositories to avoid cycles
func getActionsDependencyGraphInternal(ctx context.Context, g *GitHubClient, repo repository.Repository, recursive bool, visited map[string]bool) ([]*github.SBOM, []GraphEdge, error) {
	repoKey := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	if visited[repoKey] {
		return nil, nil, nil
	}
	visited[repoKey] = true

	sbom, err := GetRepositoryDependencyGraphSBOM(ctx, g, repo)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get SBOM for repository %s/%s: %w", repo.Owner, repo.Name, err)
	}

	actionsSBOM := FilterSBOMPackage(sbom, "actions")
	var sboms []*github.SBOM
	sboms = append(sboms, actionsSBOM)

	var edges []GraphEdge

	for _, pkg := range actionsSBOM.SBOM.Packages {
		if actionRepo, ok := parseActionRepository(*pkg.Name, repo.Host); ok {
			edges = append(edges, GraphEdge{
				From: repo,
				To:   pkg,
			})
			if GetObjectName(pkg) != GetObjectName(actionRepo) {
				edges = append(edges, GraphEdge{
					From: pkg,
					To:   actionRepo,
				})
			}
			if recursive {
				subSBOMs, subEdges, err := getActionsDependencyGraphInternal(ctx, g, actionRepo, recursive, visited)
				if err != nil {
					// skip repos that are not accessible
					continue
				}
				sboms = append(sboms, subSBOMs...)
				edges = append(edges, subEdges...)
			}
		}
	}

	return sboms, edges, nil
}

// FilterDependencyGraphEdges filters the dependency graph edges based on the provided package name
func FlattenSBOMPackages(sboms []*github.SBOM) []*github.RepoDependencies {
	var allPackages []*github.RepoDependencies
	for _, sbom := range sboms {
		allPackages = append(allPackages, sbom.SBOM.Packages...)
	}
	return allPackages
}

// SelectDependencyGraphEdges filters the dependency graph edges based on the provided package name
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

// FilterSBOMPackage filters the SBOM to include only packages that match the specified package name
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

// FilterSBOMsPackage filters multiple SBOMs to include only packages that match the specified package name
func FilterSBOMsPackage(sboms []*github.SBOM, packageName string) []*github.SBOM {
	var filteredSBOMs []*github.SBOM
	for _, sbom := range sboms {
		filteredSBOM := FilterSBOMPackage(sbom, packageName)
		filteredSBOMs = append(filteredSBOMs, filteredSBOM)
	}
	return filteredSBOMs
}
