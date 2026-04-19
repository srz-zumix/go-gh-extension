package gh

import (
	"context"
	"fmt"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// DependabotDefaultLevels is the list of valid default level values for Dependabot repository access.
var DependabotDefaultLevels = []string{
	"public",
	"internal",
}

// ListOrgDependabotRepositoryAccess lists repositories that Dependabot can access in an organization.
func ListOrgDependabotRepositoryAccess(ctx context.Context, g *GitHubClient, org string) (*client.DependabotRepositoryAccess, error) {
	access, err := g.ListOrgDependabotRepositoryAccess(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to list Dependabot repository access for org %s: %w", org, err)
	}
	return access, nil
}

// UpdateOrgDependabotRepositoryAccess updates the Dependabot repository access list for an organization.
func UpdateOrgDependabotRepositoryAccess(ctx context.Context, g *GitHubClient, org string, addIDs, removeIDs []int64) error {
	update := &client.DependabotRepositoryAccessUpdate{
		RepositoryIDsToAdd:    addIDs,
		RepositoryIDsToRemove: removeIDs,
	}
	err := g.UpdateOrgDependabotRepositoryAccess(ctx, org, update)
	if err != nil {
		return fmt.Errorf("failed to update Dependabot repository access for org %s: %w", org, err)
	}
	return nil
}

// SetOrgDependabotDefaultLevel sets the default repository access level for Dependabot in an organization.
func SetOrgDependabotDefaultLevel(ctx context.Context, g *GitHubClient, org string, level string) error {
	err := g.SetOrgDependabotDefaultLevel(ctx, org, level)
	if err != nil {
		return fmt.Errorf("failed to set Dependabot default level for org %s: %w", org, err)
	}
	return nil
}
