package client

import (
	"context"
	"fmt"

	"github.com/google/go-github/v84/github"
)

// GetDependencyGraphSBOM retrieves the SBOM (Software Bill of Materials) for a repository using the dependency-graph/sbom API.
func (g *GitHubClient) GetDependencyGraphSBOM(ctx context.Context, owner, repo string) (*github.SBOM, error) {
	sbom, _, err := g.client.DependencyGraph.GetSBOM(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return sbom, nil
}

// DependencyVulnerability represents a vulnerability associated with a dependency change.
type DependencyVulnerability struct {
	Severity       string `json:"severity"`
	AdvisoryGHSAID string `json:"advisory_ghsa_id"`
	AdvisoryTitle  string `json:"advisory_summary"`
	AdvisoryURL    string `json:"advisory_url"`
}

// DependencyChange represents a single dependency change in a dependency diff.
type DependencyChange struct {
	ChangeType      string                    `json:"change_type"`
	Manifest        string                    `json:"manifest"`
	Ecosystem       string                    `json:"ecosystem"`
	Name            string                    `json:"name"`
	Version         string                    `json:"version"`
	PackageURL      string                    `json:"package_url"`
	License         string                    `json:"license"`
	SourceRepoURL   string                    `json:"source_repository_url"`
	Vulnerabilities []DependencyVulnerability `json:"vulnerabilities"`
	Scope           string                    `json:"scope"`
}

// GetDependencyGraphDiff retrieves the dependency diff between two commits or branches using the dependency-graph/compare API.
func (g *GitHubClient) GetDependencyGraphDiff(ctx context.Context, owner, repo, basehead string) ([]*DependencyChange, error) {
	u := fmt.Sprintf("repos/%v/%v/dependency-graph/compare/%v", owner, repo, url.PathEscape(basehead))
	req, err := g.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	var changes []*DependencyChange
	_, err = g.client.Do(ctx, req, &changes)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// CreateDependencyGraphSnapshot creates a new snapshot of a repository's dependencies.
func (g *GitHubClient) CreateDependencyGraphSnapshot(ctx context.Context, owner, repo string, snapshot *github.DependencyGraphSnapshot) (*github.DependencyGraphSnapshotCreationData, error) {
	result, _, err := g.client.DependencyGraph.CreateSnapshot(ctx, owner, repo, snapshot)
	if err != nil {
		return nil, err
	}
	return result, nil
}
