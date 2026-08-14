package client

import (
	"context"
	"time"

	"github.com/shurcooL/githubv4"
)

// ProjectV2StatusUpdateStatus represents the status of a Project v2 status update.
type ProjectV2StatusUpdateStatus string

const (
	ProjectV2StatusUpdateStatusInactive ProjectV2StatusUpdateStatus = "INACTIVE"
	ProjectV2StatusUpdateStatusOnTrack  ProjectV2StatusUpdateStatus = "ON_TRACK"
	ProjectV2StatusUpdateStatusAtRisk   ProjectV2StatusUpdateStatus = "AT_RISK"
	ProjectV2StatusUpdateStatusOffTrack ProjectV2StatusUpdateStatus = "OFF_TRACK"
	ProjectV2StatusUpdateStatusComplete ProjectV2StatusUpdateStatus = "COMPLETE"
)

// ProjectV2StatusUpdate represents a status update posted on a GitHub Project v2.
type ProjectV2StatusUpdate struct {
	ID         string
	Body       string
	Status     ProjectV2StatusUpdateStatus // INACTIVE, ON_TRACK, AT_RISK, OFF_TRACK, COMPLETE
	StartDate  string // YYYY-MM-DD, empty when unset
	TargetDate string // YYYY-MM-DD, empty when unset
	Creator    string
	CreatedAt  string
	UpdatedAt  string
}

// projectV2StatusUpdateNode is the raw GraphQL node for a project status update.
type projectV2StatusUpdateNode struct {
	ID         githubv4.String
	Body       githubv4.String
	Status     githubv4.String
	StartDate  *githubv4.Date
	TargetDate *githubv4.Date
	CreatedAt  githubv4.DateTime
	UpdatedAt  githubv4.DateTime
	Creator    struct {
		Login githubv4.String
	}
}

func (n *projectV2StatusUpdateNode) toProjectV2StatusUpdate() ProjectV2StatusUpdate {
	s := ProjectV2StatusUpdate{
		ID:        string(n.ID),
		Body:      string(n.Body),
		Status:    ProjectV2StatusUpdateStatus(n.Status),
		Creator:   string(n.Creator.Login),
		CreatedAt: n.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: n.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if n.StartDate != nil {
		s.StartDate = n.StartDate.Format("2006-01-02")
	}
	if n.TargetDate != nil {
		s.TargetDate = n.TargetDate.Format("2006-01-02")
	}
	return s
}

// ListUserProjectV2StatusUpdates lists all status updates for a user's ProjectV2.
func (g *GitHubClient) ListUserProjectV2StatusUpdates(ctx context.Context, login string, number int, first int) ([]ProjectV2StatusUpdate, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		User struct {
			ProjectV2 struct {
				StatusUpdates struct {
					Nodes    []projectV2StatusUpdateNode
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"statusUpdates(first: $first, after: $cursor)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(login),
		"number": githubv4.Int(number),
		"first":  githubv4.Int(first),
		"cursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2StatusUpdate
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		for i := range query.User.ProjectV2.StatusUpdates.Nodes {
			all = append(all, query.User.ProjectV2.StatusUpdates.Nodes[i].toProjectV2StatusUpdate())
		}
		if !query.User.ProjectV2.StatusUpdates.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.User.ProjectV2.StatusUpdates.PageInfo.EndCursor)
	}
	return all, nil
}

// ListOrgProjectV2StatusUpdates lists all status updates for an org's ProjectV2.
func (g *GitHubClient) ListOrgProjectV2StatusUpdates(ctx context.Context, org string, number int, first int) ([]ProjectV2StatusUpdate, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		Organization struct {
			ProjectV2 struct {
				StatusUpdates struct {
					Nodes    []projectV2StatusUpdateNode
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"statusUpdates(first: $first, after: $cursor)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(org),
		"number": githubv4.Int(number),
		"first":  githubv4.Int(first),
		"cursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2StatusUpdate
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		for i := range query.Organization.ProjectV2.StatusUpdates.Nodes {
			all = append(all, query.Organization.ProjectV2.StatusUpdates.Nodes[i].toProjectV2StatusUpdate())
		}
		if !query.Organization.ProjectV2.StatusUpdates.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.ProjectV2.StatusUpdates.PageInfo.EndCursor)
	}
	return all, nil
}

// CreateProjectV2StatusUpdateInput is the input for creating a status update on a Project v2.
type CreateProjectV2StatusUpdateInput struct {
	ProjectID  githubv4.ID      `json:"projectId"`
	Body       *githubv4.String `json:"body,omitempty"`
	Status     *githubv4.String `json:"status,omitempty"`
	StartDate  *githubv4.Date   `json:"startDate,omitempty"`
	TargetDate *githubv4.Date   `json:"targetDate,omitempty"`
}

// UpdateProjectV2StatusUpdateInput is the input for updating a status update on a Project v2.
type UpdateProjectV2StatusUpdateInput struct {
	StatusUpdateID githubv4.ID      `json:"statusUpdateId"`
	Body           *githubv4.String `json:"body,omitempty"`
	Status         *githubv4.String `json:"status,omitempty"`
	StartDate      *githubv4.Date   `json:"startDate,omitempty"`
	TargetDate     *githubv4.Date   `json:"targetDate,omitempty"`
}

// DeleteProjectV2StatusUpdateInput is the input for deleting a status update from a Project v2.
type DeleteProjectV2StatusUpdateInput struct {
	StatusUpdateID githubv4.ID `json:"statusUpdateId"`
}

// CreateProjectV2StatusUpdate creates a status update on a GitHub Project v2.
// Returns the created status update's node ID.
func (g *GitHubClient) CreateProjectV2StatusUpdate(ctx context.Context, input CreateProjectV2StatusUpdateInput) (string, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return "", err
	}
	var mutation struct {
		CreateProjectV2StatusUpdate struct {
			StatusUpdate struct {
				ID githubv4.String
			}
		} `graphql:"createProjectV2StatusUpdate(input: $input)"`
	}
	if err := gql.Mutate(ctx, &mutation, input, nil); err != nil {
		return "", err
	}
	return string(mutation.CreateProjectV2StatusUpdate.StatusUpdate.ID), nil
}

// UpdateProjectV2StatusUpdate updates a status update on a GitHub Project v2.
func (g *GitHubClient) UpdateProjectV2StatusUpdate(ctx context.Context, input UpdateProjectV2StatusUpdateInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		UpdateProjectV2StatusUpdate struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"updateProjectV2StatusUpdate(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// DeleteProjectV2StatusUpdate deletes a status update from a GitHub Project v2.
func (g *GitHubClient) DeleteProjectV2StatusUpdate(ctx context.Context, input DeleteProjectV2StatusUpdateInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		DeleteProjectV2StatusUpdate struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"deleteProjectV2StatusUpdate(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}
