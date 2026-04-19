package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// DependabotRepositoryAccess represents the response from the Dependabot repository access endpoint.
type DependabotRepositoryAccess struct {
	DefaultLevel           string               `json:"default_level"`
	AccessibleRepositories []*github.Repository `json:"accessible_repositories"`
}

// DependabotRepositoryAccessUpdate represents the request body for updating Dependabot repository access.
type DependabotRepositoryAccessUpdate struct {
	RepositoryIDsToAdd    []int64 `json:"repository_ids_to_add,omitempty"`
	RepositoryIDsToRemove []int64 `json:"repository_ids_to_remove,omitempty"`
}

// DependabotDefaultLevel represents the request body for setting the default repository access level.
type DependabotDefaultLevel struct {
	DefaultLevel string `json:"default_level"`
}

// ListOrgDependabotRepositoryAccess lists repositories that organization admins have allowed Dependabot to access.
func (g *GitHubClient) ListOrgDependabotRepositoryAccess(ctx context.Context, org string) (*DependabotRepositoryAccess, error) {
	u := "orgs/" + org + "/dependabot/repository-access"
	req, err := g.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	result := new(DependabotRepositoryAccess)
	_, err = g.client.Do(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateOrgDependabotRepositoryAccess updates the repository access list for Dependabot in an organization.
func (g *GitHubClient) UpdateOrgDependabotRepositoryAccess(ctx context.Context, org string, update *DependabotRepositoryAccessUpdate) error {
	u := "orgs/" + org + "/dependabot/repository-access"
	req, err := g.client.NewRequest("PATCH", u, update)
	if err != nil {
		return err
	}

	_, err = g.client.Do(ctx, req, nil)
	return err
}

// SetOrgDependabotDefaultLevel sets the default repository access level for Dependabot in an organization.
func (g *GitHubClient) SetOrgDependabotDefaultLevel(ctx context.Context, org string, level string) error {
	u := "orgs/" + org + "/dependabot/repository-access/default-level"
	body := &DependabotDefaultLevel{DefaultLevel: level}
	req, err := g.client.NewRequest("PUT", u, body)
	if err != nil {
		return err
	}

	_, err = g.client.Do(ctx, req, nil)
	return err
}
