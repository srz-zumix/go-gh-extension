package client

import (
	"context"

	"github.com/google/go-github/v79/github"
)

// GetDependencyGraphSBOM retrieves the SBOM (Software Bill of Materials) for a repository using the dependency-graph/sbom API.
func (g *GitHubClient) GetDependencyGraphSBOM(ctx context.Context, owner, repo string) (*github.SBOM, error) {
	sbom, _, err := g.client.DependencyGraph.GetSBOM(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return sbom, nil
}
