package client

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

// CreateProjectV2ViewRequest is the request body for the Project views REST endpoints.
// Layout is the REST layout name (table, board, roadmap).
// Field references are REST field IDs, and SortBy entries are [field_id, direction] tuples.
type CreateProjectV2ViewRequest struct {
	Name            string  `json:"name"`
	Layout          string  `json:"layout"`
	Filter          *string `json:"filter,omitempty"`
	VisibleFields   []int64 `json:"visible_fields,omitempty"`
	SortBy          [][]any `json:"sort_by,omitempty"`
	GroupBy         []int64 `json:"group_by,omitempty"`
	VerticalGroupBy []int64 `json:"vertical_group_by,omitempty"`
}

// CreateProjectV2ViewResponse is the response of the Project views create endpoints.
type CreateProjectV2ViewResponse struct {
	ID      int64  `json:"id"`
	NodeID  string `json:"node_id"`
	Number  int    `json:"number"`
	Name    string `json:"name"`
	Layout  string `json:"layout"`
	HTMLURL string `json:"html_url"`
}

// CreateOrgProjectV2View creates a view in an organization-owned project.
// https://docs.github.com/rest/projects/views#create-a-view-for-an-organization-owned-project
func (g *GitHubClient) CreateOrgProjectV2View(ctx context.Context, org string, projectNumber int, body CreateProjectV2ViewRequest) (*CreateProjectV2ViewResponse, error) {
	return g.createProjectV2View(ctx, fmt.Sprintf("orgs/%s/projectsV2/%d/views", org, projectNumber), body)
}

// CreateUserProjectV2View creates a view in a user-owned project.
// The endpoint identifies the user by numeric ID, not by login.
// https://docs.github.com/rest/projects/views#create-a-view-for-a-user-owned-project
func (g *GitHubClient) CreateUserProjectV2View(ctx context.Context, userID int64, projectNumber int, body CreateProjectV2ViewRequest) (*CreateProjectV2ViewResponse, error) {
	return g.createProjectV2View(ctx, fmt.Sprintf("users/%d/projectsV2/%d/views", userID, projectNumber), body)
}

func (g *GitHubClient) createProjectV2View(ctx context.Context, url string, body CreateProjectV2ViewRequest) (*CreateProjectV2ViewResponse, error) {
	req, err := g.client.NewRequest(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	var view CreateProjectV2ViewResponse
	if _, err := g.client.Do(req, &view); err != nil {
		return nil, err
	}
	return &view, nil
}

// DeleteProjectV2ViewInput is the input for deleting a view from a Project v2.
type DeleteProjectV2ViewInput struct {
	ViewID githubv4.ID `json:"viewId"`
}

// DeleteProjectV2View deletes a view from a GitHub Project v2.
func (g *GitHubClient) DeleteProjectV2View(ctx context.Context, input DeleteProjectV2ViewInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		DeleteProjectV2View struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"deleteProjectV2View(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}
