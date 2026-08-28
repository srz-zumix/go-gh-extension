package client

import (
	"context"

	"github.com/shurcooL/githubv4"
)

// ProjectV2Workflow represents a built-in automation of a GitHub Project v2.
// The GraphQL API exposes no trigger or action details, so a workflow cannot be recreated.
type ProjectV2Workflow struct {
	ID      string
	Number  int
	Name    string
	Enabled bool
}

// projectV2WorkflowNode is the raw GraphQL node for a project workflow.
type projectV2WorkflowNode struct {
	ID      githubv4.String
	Number  githubv4.Int
	Name    githubv4.String
	Enabled githubv4.Boolean
}

func (n *projectV2WorkflowNode) toProjectV2Workflow() ProjectV2Workflow {
	return ProjectV2Workflow{
		ID:      string(n.ID),
		Number:  int(n.Number),
		Name:    string(n.Name),
		Enabled: bool(n.Enabled),
	}
}

// ListUserProjectV2Workflows lists all workflows for a user's ProjectV2.
func (g *GitHubClient) ListUserProjectV2Workflows(ctx context.Context, login string, number int, first int) ([]ProjectV2Workflow, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		User struct {
			ProjectV2 struct {
				Workflows struct {
					Nodes    []projectV2WorkflowNode
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"workflows(first: $first, after: $cursor)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(login),
		"number": githubv4.Int(number),
		"first":  githubv4.Int(first),
		"cursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2Workflow
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		for i := range query.User.ProjectV2.Workflows.Nodes {
			all = append(all, query.User.ProjectV2.Workflows.Nodes[i].toProjectV2Workflow())
		}
		if !query.User.ProjectV2.Workflows.PageInfo.HasNextPage {
			return all, nil
		}
		variables["cursor"] = githubv4.NewString(query.User.ProjectV2.Workflows.PageInfo.EndCursor)
	}
}

// ListOrgProjectV2Workflows lists all workflows for an org's ProjectV2.
func (g *GitHubClient) ListOrgProjectV2Workflows(ctx context.Context, org string, number int, first int) ([]ProjectV2Workflow, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		Organization struct {
			ProjectV2 struct {
				Workflows struct {
					Nodes    []projectV2WorkflowNode
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"workflows(first: $first, after: $cursor)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(org),
		"number": githubv4.Int(number),
		"first":  githubv4.Int(first),
		"cursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2Workflow
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		for i := range query.Organization.ProjectV2.Workflows.Nodes {
			all = append(all, query.Organization.ProjectV2.Workflows.Nodes[i].toProjectV2Workflow())
		}
		if !query.Organization.ProjectV2.Workflows.PageInfo.HasNextPage {
			return all, nil
		}
		variables["cursor"] = githubv4.NewString(query.Organization.ProjectV2.Workflows.PageInfo.EndCursor)
	}
}
