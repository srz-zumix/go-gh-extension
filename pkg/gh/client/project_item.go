package client

import (
	"context"

	"github.com/shurcooL/githubv4"
)

// UnarchiveProjectV2ItemInput is the input for unarchiving an item in a Project v2.
type UnarchiveProjectV2ItemInput struct {
	ProjectID githubv4.ID `json:"projectId"`
	ItemID    githubv4.ID `json:"itemId"`
}

// ClearProjectV2ItemFieldValueInput is the input for clearing a field value on a project item.
type ClearProjectV2ItemFieldValueInput struct {
	ProjectID githubv4.ID `json:"projectId"`
	ItemID    githubv4.ID `json:"itemId"`
	FieldID   githubv4.ID `json:"fieldId"`
}

// UpdateProjectV2ItemPositionInput is the input for moving an item within a Project v2.
// AfterID is the item the moved item is placed after; nil moves the item to the top.
type UpdateProjectV2ItemPositionInput struct {
	ProjectID githubv4.ID  `json:"projectId"`
	ItemID    githubv4.ID  `json:"itemId"`
	AfterID   *githubv4.ID `json:"afterId,omitempty"`
}

// UpdateProjectV2DraftIssueInput is the input for updating a draft issue in a Project v2.
type UpdateProjectV2DraftIssueInput struct {
	DraftIssueID githubv4.ID      `json:"draftIssueId"`
	Title        *githubv4.String `json:"title,omitempty"`
	Body         *githubv4.String `json:"body,omitempty"`
	AssigneeIDs  []githubv4.ID    `json:"assigneeIds,omitempty"`
}

// ConvertProjectV2DraftIssueItemToIssueInput is the input for converting a draft issue item to an issue.
type ConvertProjectV2DraftIssueItemToIssueInput struct {
	ItemID       githubv4.ID `json:"itemId"`
	RepositoryID githubv4.ID `json:"repositoryId"`
}

// UnarchiveProjectV2Item unarchives an item in a GitHub Project v2.
func (g *GitHubClient) UnarchiveProjectV2Item(ctx context.Context, input UnarchiveProjectV2ItemInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		UnarchiveProjectV2Item struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"unarchiveProjectV2Item(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// ClearProjectV2ItemFieldValue clears the value of a custom field for a project item.
func (g *GitHubClient) ClearProjectV2ItemFieldValue(ctx context.Context, input ClearProjectV2ItemFieldValueInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		ClearProjectV2ItemFieldValue struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"clearProjectV2ItemFieldValue(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// UpdateProjectV2ItemPosition moves an item within a GitHub Project v2.
func (g *GitHubClient) UpdateProjectV2ItemPosition(ctx context.Context, input UpdateProjectV2ItemPositionInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		UpdateProjectV2ItemPosition struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"updateProjectV2ItemPosition(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// UpdateProjectV2DraftIssue updates a draft issue in a GitHub Project v2.
func (g *GitHubClient) UpdateProjectV2DraftIssue(ctx context.Context, input UpdateProjectV2DraftIssueInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		UpdateProjectV2DraftIssue struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"updateProjectV2DraftIssue(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// ConvertProjectV2DraftIssueItemToIssue converts a draft issue item to an issue in the given repository.
// Returns the resulting project item's node ID.
func (g *GitHubClient) ConvertProjectV2DraftIssueItemToIssue(ctx context.Context, input ConvertProjectV2DraftIssueItemToIssueInput) (string, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return "", err
	}
	var mutation struct {
		ConvertProjectV2DraftIssueItemToIssue struct {
			Item struct {
				ID githubv4.String
			}
		} `graphql:"convertProjectV2DraftIssueItemToIssue(input: $input)"`
	}
	if err := gql.Mutate(ctx, &mutation, input, nil); err != nil {
		return "", err
	}
	return string(mutation.ConvertProjectV2DraftIssueItemToIssue.Item.ID), nil
}
