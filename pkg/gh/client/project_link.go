package client

import (
	"context"

	"github.com/shurcooL/githubv4"
)

// ProjectV2Role represents the permission level granted to a Project v2 collaborator.
type ProjectV2Role string

const (
	ProjectV2RoleAdmin  ProjectV2Role = "ADMIN"
	ProjectV2RoleWriter ProjectV2Role = "WRITER"
	ProjectV2RoleReader ProjectV2Role = "READER"
	ProjectV2RoleNone   ProjectV2Role = "NONE"
)

// LinkProjectV2ToRepositoryInput is the input for linking a Project v2 to a repository.
type LinkProjectV2ToRepositoryInput struct {
	ProjectID    githubv4.ID `json:"projectId"`
	RepositoryID githubv4.ID `json:"repositoryId"`
}

// UnlinkProjectV2FromRepositoryInput is the input for unlinking a Project v2 from a repository.
type UnlinkProjectV2FromRepositoryInput struct {
	ProjectID    githubv4.ID `json:"projectId"`
	RepositoryID githubv4.ID `json:"repositoryId"`
}

// LinkProjectV2ToTeamInput is the input for linking a Project v2 to a team.
type LinkProjectV2ToTeamInput struct {
	ProjectID githubv4.ID `json:"projectId"`
	TeamID    githubv4.ID `json:"teamId"`
}

// UnlinkProjectV2FromTeamInput is the input for unlinking a Project v2 from a team.
type UnlinkProjectV2FromTeamInput struct {
	ProjectID githubv4.ID `json:"projectId"`
	TeamID    githubv4.ID `json:"teamId"`
}

// ProjectV2Collaborator is a single collaborator entry for updateProjectV2Collaborators.
// Exactly one of UserID or TeamID must be set.
type ProjectV2Collaborator struct {
	UserID *githubv4.ID    `json:"userId,omitempty"`
	TeamID *githubv4.ID    `json:"teamId,omitempty"`
	Role   githubv4.String `json:"role"`
}

// UpdateProjectV2CollaboratorsInput is the input for updating Project v2 collaborators.
type UpdateProjectV2CollaboratorsInput struct {
	ProjectID     githubv4.ID             `json:"projectId"`
	Collaborators []ProjectV2Collaborator `json:"collaborators"`
}

// LinkProjectV2ToRepository links a GitHub Project v2 to a repository.
func (g *GitHubClient) LinkProjectV2ToRepository(ctx context.Context, input LinkProjectV2ToRepositoryInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		LinkProjectV2ToRepository struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"linkProjectV2ToRepository(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// UnlinkProjectV2FromRepository unlinks a GitHub Project v2 from a repository.
func (g *GitHubClient) UnlinkProjectV2FromRepository(ctx context.Context, input UnlinkProjectV2FromRepositoryInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		UnlinkProjectV2FromRepository struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"unlinkProjectV2FromRepository(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// LinkProjectV2ToTeam links a GitHub Project v2 to a team.
func (g *GitHubClient) LinkProjectV2ToTeam(ctx context.Context, input LinkProjectV2ToTeamInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		LinkProjectV2ToTeam struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"linkProjectV2ToTeam(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// UnlinkProjectV2FromTeam unlinks a GitHub Project v2 from a team.
func (g *GitHubClient) UnlinkProjectV2FromTeam(ctx context.Context, input UnlinkProjectV2FromTeamInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		UnlinkProjectV2FromTeam struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"unlinkProjectV2FromTeam(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// UpdateProjectV2Collaborators updates the collaborators of a GitHub Project v2.
func (g *GitHubClient) UpdateProjectV2Collaborators(ctx context.Context, input UpdateProjectV2CollaboratorsInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		UpdateProjectV2Collaborators struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"updateProjectV2Collaborators(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}
