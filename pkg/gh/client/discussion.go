package client

import (
	"context"

	"github.com/shurcooL/githubv4"
)

// Discussion represents a GitHub discussion
type Discussion struct {
	ID     githubv4.String
	Number githubv4.Int
	Title  githubv4.String
	Body   githubv4.String
	Author struct {
		Login githubv4.String
	}
	Category struct {
		Name githubv4.String
		Slug githubv4.String
	}
	Labels struct {
		Nodes []struct {
			ID          githubv4.String
			Name        githubv4.String
			Color       githubv4.String
			Description *githubv4.String
		}
	} `graphql:"labels(first: 100)"`
	CreatedAt   githubv4.DateTime
	UpdatedAt   githubv4.DateTime
	ClosedAt    *githubv4.DateTime
	Locked      githubv4.Boolean
	StateReason *githubv4.String
	URL         githubv4.String
}

// GetDiscussion retrieves a specific discussion by number using GraphQL
func (g *GitHubClient) GetDiscussion(ctx context.Context, owner string, repo string, number int) (*Discussion, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Repository struct {
			Discussion Discussion `graphql:"discussion(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner":  githubv4.String(owner),
		"repo":   githubv4.String(repo),
		"number": githubv4.Int(number),
	}

	if err := graphql.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	return &query.Repository.Discussion, nil
}

// ListDiscussions retrieves all discussions for a repository using GraphQL
func (g *GitHubClient) ListDiscussions(ctx context.Context, owner string, repo string, first int) ([]Discussion, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Repository struct {
			Discussions struct {
				Nodes    []Discussion
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"discussions(first: $first, after: $cursor)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner":  githubv4.String(owner),
		"repo":   githubv4.String(repo),
		"first":  githubv4.Int(first),
		"cursor": (*githubv4.String)(nil), // Null after argument to get first page
	}

	allDiscussions := []Discussion{}
	for {
		if err := graphql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		allDiscussions = append(allDiscussions, query.Repository.Discussions.Nodes...)
		if !query.Repository.Discussions.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Repository.Discussions.PageInfo.EndCursor)
	}

	return allDiscussions, nil
}

// AddLabelsToDiscussion adds labels to a discussion using GraphQL mutation
func (g *GitHubClient) AddLabelsToDiscussion(ctx context.Context, discussionID string, labelIDs []string) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var mutation struct {
		AddLabelsToLabelable struct {
			ClientMutationID githubv4.String
		} `graphql:"addLabelsToLabelable(input: $input)"`
	}

	// Convert string slice to githubv4.ID slice
	var labelIDsGraphQL []githubv4.ID
	for _, id := range labelIDs {
		labelIDsGraphQL = append(labelIDsGraphQL, githubv4.ID(id))
	}

	input := githubv4.AddLabelsToLabelableInput{
		LabelableID: githubv4.ID(discussionID),
		LabelIDs:    labelIDsGraphQL,
	}

	if err := graphql.Mutate(ctx, &mutation, input, nil); err != nil {
		return err
	}

	return nil
}

// RemoveLabelsFromDiscussion removes labels from a discussion using GraphQL mutation
func (g *GitHubClient) RemoveLabelsFromDiscussion(ctx context.Context, discussionID string, labelIDs []string) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var mutation struct {
		RemoveLabelsFromLabelable struct {
			ClientMutationID githubv4.String
		} `graphql:"removeLabelsFromLabelable(input: $input)"`
	}

	// Convert string slice to githubv4.ID slice
	var labelIDsGraphQL []githubv4.ID
	for _, id := range labelIDs {
		labelIDsGraphQL = append(labelIDsGraphQL, githubv4.ID(id))
	}

	input := githubv4.RemoveLabelsFromLabelableInput{
		LabelableID: githubv4.ID(discussionID),
		LabelIDs:    labelIDsGraphQL,
	}

	if err := graphql.Mutate(ctx, &mutation, input, nil); err != nil {
		return err
	}

	return nil
}

// SearchDiscussions searches discussions using GraphQL
func (g *GitHubClient) SearchDiscussions(ctx context.Context, query string, first int) ([]Discussion, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var searchQuery struct {
		Search struct {
			Nodes []struct {
				Discussion Discussion `graphql:"... on Discussion"`
			}
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"search(query: $query, type: DISCUSSION, first: $first, after: $cursor)"`
	}

	variables := map[string]interface{}{
		"query":  githubv4.String(query),
		"first":  githubv4.Int(first),
		"cursor": (*githubv4.String)(nil),
	}

	allDiscussions := []Discussion{}
	for {
		if err := graphql.Query(ctx, &searchQuery, variables); err != nil {
			return nil, err
		}
		for _, node := range searchQuery.Search.Nodes {
			allDiscussions = append(allDiscussions, node.Discussion)
		}
		if !searchQuery.Search.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(searchQuery.Search.PageInfo.EndCursor)
	}

	return allDiscussions, nil
}
