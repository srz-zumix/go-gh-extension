package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

// ListCodespacesRepoMachineTypes lists the machine types available for a repository (wrapper).
// ref is optional and selects the branch or commit the availability is checked for.
func ListCodespacesRepoMachineTypes(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string) ([]*github.CodespacesMachine, error) {
	var opts *github.ListRepoMachineTypesOptions
	if ref != "" {
		opts = &github.ListRepoMachineTypesOptions{Ref: github.Ptr(ref)}
	}
	return g.ListCodespacesRepoMachineTypes(ctx, repo.Owner, repo.Name, opts)
}
