package gh

import (
	"context"
	"encoding/json"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// ListRepositoryActionsPolicies lists repository Actions policies and applicable parent policies.
func ListRepositoryActionsPolicies(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]json.RawMessage, error) {
	return g.ListRepositoryActionsPolicies(ctx, repo.Owner, repo.Name)
}

// ListOrgActionsPolicies lists organization Actions policies and applicable parent policies.
func ListOrgActionsPolicies(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]json.RawMessage, error) {
	return g.ListOrgActionsPolicies(ctx, repo.Owner)
}

// GetRepositoryActionsPolicy retrieves a repository Actions policy by ID.
func GetRepositoryActionsPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository, policyID int64) (json.RawMessage, error) {
	return g.GetRepositoryActionsPolicy(ctx, repo.Owner, repo.Name, policyID)
}

// GetOrgActionsPolicy retrieves an organization Actions policy by ID.
func GetOrgActionsPolicy(ctx context.Context, g *GitHubClient, repo repository.Repository, policyID int64) (json.RawMessage, error) {
	return g.GetOrgActionsPolicy(ctx, repo.Owner, policyID)
}
