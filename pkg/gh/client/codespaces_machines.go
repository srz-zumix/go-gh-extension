package client

// GitHub Codespaces Machines API functions
// See: https://docs.github.com/rest/codespaces/machines

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// ListCodespacesRepoMachineTypes lists the machine types available for a given repository based on its configuration.
func (g *GitHubClient) ListCodespacesRepoMachineTypes(ctx context.Context, owner, repo string, opts *github.ListRepoMachineTypesOptions) ([]*github.CodespacesMachine, error) {
	machines, _, err := g.client.Codespaces.ListRepositoryMachineTypes(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}
	return machines.Machines, nil
}
