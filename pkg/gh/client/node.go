package client

import (
	"context"

	"github.com/shurcooL/githubv4"
)

// NodeExists reports whether a GitHub GraphQL node with the given global ID exists.
// It returns false (without error) when the node is not found.
func (g *GitHubClient) NodeExists(ctx context.Context, nodeID string) (bool, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return false, err
	}

	var query struct {
		Node *struct {
			ID githubv4.String
		} `graphql:"node(id: $id)"`
	}

	variables := map[string]any{
		"id": githubv4.ID(nodeID),
	}

	if err := graphql.Query(ctx, &query, variables); err != nil {
		return false, err
	}

	return query.Node != nil, nil
}
