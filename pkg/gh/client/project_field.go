package client

import (
	"context"

	"github.com/shurcooL/githubv4"
)

// ProjectV2SingleSelectFieldOptionInput is a single-select option used when updating a field.
type ProjectV2SingleSelectFieldOptionInput struct {
	ID          *githubv4.String `json:"id,omitempty"`
	Name        githubv4.String  `json:"name"`
	Color       githubv4.String  `json:"color"`
	Description githubv4.String  `json:"description"`
}

// ProjectV2MultiSelectFieldOptionInput is a multi-select option used when updating a field.
type ProjectV2MultiSelectFieldOptionInput struct {
	ID          *githubv4.String `json:"id,omitempty"`
	Name        githubv4.String  `json:"name"`
	Color       githubv4.String  `json:"color"`
	Description githubv4.String  `json:"description"`
}

// ProjectV2IterationFieldIterationInput is a single iteration used when updating an ITERATION field.
type ProjectV2IterationFieldIterationInput struct {
	ID        *githubv4.String `json:"id,omitempty"`
	Title     githubv4.String  `json:"title"`
	StartDate githubv4.String  `json:"startDate"` // YYYY-MM-DD
	Duration  githubv4.Int     `json:"duration"`  // days
}

// UpdateProjectV2FieldInput is the input for updating a custom field in a Project v2.
type UpdateProjectV2FieldInput struct {
	FieldID                githubv4.ID                             `json:"fieldId"`
	Name                   *githubv4.String                        `json:"name,omitempty"`
	SingleSelectOptions    []ProjectV2SingleSelectFieldOptionInput `json:"singleSelectOptions,omitempty"`
	MultiSelectOptions     []ProjectV2MultiSelectFieldOptionInput  `json:"multiSelectOptions,omitempty"`
	IterationConfiguration *ProjectV2IterationFieldConfigInput     `json:"iterationConfiguration,omitempty"`
}

// DeleteProjectV2FieldInput is the input for deleting a custom field from a Project v2.
type DeleteProjectV2FieldInput struct {
	FieldID githubv4.ID `json:"fieldId"`
}

// UpdateProjectV2Field updates a custom field in a GitHub Project v2.
func (g *GitHubClient) UpdateProjectV2Field(ctx context.Context, input UpdateProjectV2FieldInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	// Request only clientMutationId to avoid schema differences across GitHub versions.
	var mutation struct {
		UpdateProjectV2Field struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"updateProjectV2Field(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// DeleteProjectV2Field deletes a custom field from a GitHub Project v2.
func (g *GitHubClient) DeleteProjectV2Field(ctx context.Context, input DeleteProjectV2FieldInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		DeleteProjectV2Field struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"deleteProjectV2Field(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}
