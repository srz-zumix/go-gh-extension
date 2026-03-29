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

// GetNodeReactions retrieves all reactions on any reactionable node (discussion, comment, reply) by its GraphQL node ID.
func (g *GitHubClient) GetNodeReactions(ctx context.Context, nodeID string) ([]Reaction, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Node struct {
			Reactable struct {
				Reactions struct {
					Nodes    []rawReaction
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"reactions(first: $first, after: $cursor)"`
			} `graphql:"... on Reactable"`
		} `graphql:"node(id: $id)"`
	}

	variables := map[string]any{
		"id":     githubv4.ID(nodeID),
		"first":  githubv4.Int(100),
		"cursor": (*githubv4.String)(nil),
	}

	var allReactions []Reaction
	for {
		if err := graphql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		for _, r := range query.Node.Reactable.Reactions.Nodes {
			allReactions = append(allReactions, r.toReaction())
		}
		if !query.Node.Reactable.Reactions.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Node.Reactable.Reactions.PageInfo.EndCursor)
	}
	return allReactions, nil
}
