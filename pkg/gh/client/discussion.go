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

	variables := map[string]any{
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

	variables := map[string]any{
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

	variables := map[string]any{
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

// DiscussionCategory represents a discussion category in a repository.
type DiscussionCategory struct {
	ID   githubv4.String
	Name githubv4.String
	Slug githubv4.String
}

// ListDiscussionCategories retrieves all discussion categories for a repository using GraphQL.
func (g *GitHubClient) ListDiscussionCategories(ctx context.Context, owner string, repo string) ([]DiscussionCategory, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Repository struct {
			DiscussionCategories struct {
				Nodes    []DiscussionCategory
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"discussionCategories(first: $first, after: $cursor)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]any{
		"owner":  githubv4.String(owner),
		"repo":   githubv4.String(repo),
		"first":  githubv4.Int(100),
		"cursor": (*githubv4.String)(nil),
	}

	allCategories := []DiscussionCategory{}
	for {
		if err := graphql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		allCategories = append(allCategories, query.Repository.DiscussionCategories.Nodes...)
		if !query.Repository.DiscussionCategories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Repository.DiscussionCategories.PageInfo.EndCursor)
	}

	return allCategories, nil
}

// CreateDiscussionInput is the input for creating a discussion.
type CreateDiscussionInput struct {
	RepositoryID githubv4.ID     `json:"repositoryId"`
	CategoryID   githubv4.ID     `json:"categoryId"`
	Title        githubv4.String `json:"title"`
	Body         githubv4.String `json:"body"`
}

// CreateDiscussion creates a new discussion in a repository using GraphQL mutation.
func (g *GitHubClient) CreateDiscussion(ctx context.Context, input CreateDiscussionInput) (*Discussion, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var mutation struct {
		CreateDiscussion struct {
			Discussion Discussion
		} `graphql:"createDiscussion(input: $input)"`
	}

	if err := graphql.Mutate(ctx, &mutation, input, nil); err != nil {
		return nil, err
	}

	return &mutation.CreateDiscussion.Discussion, nil
}

// Reaction represents a single reaction on a discussion or comment.
type Reaction struct {
	Content githubv4.String
}

// DiscussionCommentReply represents a reply to a top-level discussion comment.
// Reactions are not included inline to avoid exceeding GraphQL node limits;
// use GetNodeReactions to fetch them separately.
type DiscussionCommentReply struct {
	ID     githubv4.String
	Body   githubv4.String
	Author struct {
		Login githubv4.String
	}
}

// DiscussionComment represents a top-level comment on a discussion.
type DiscussionComment struct {
	ID     githubv4.String
	Body   githubv4.String
	Author struct {
		Login githubv4.String
	}
	Reactions struct {
		Nodes []Reaction
	} `graphql:"reactions(first: 100)"`
	Replies struct {
		Nodes []DiscussionCommentReply
	} `graphql:"replies(first: 100)"`
}

// AddDiscussionCommentInput is the input for adding a comment to a discussion.
// Set ReplyToID to add a reply to an existing top-level comment.
type AddDiscussionCommentInput struct {
	DiscussionID githubv4.ID     `json:"discussionId"`
	Body         githubv4.String `json:"body"`
	ReplyToID    *githubv4.ID    `json:"replyToId,omitempty"`
}

// AddReactionInput is the input for adding a reaction to a subject (discussion or comment).
type AddReactionInput struct {
	SubjectID githubv4.ID              `json:"subjectId"`
	Content   githubv4.ReactionContent `json:"content"`
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
					Nodes    []Reaction
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
		allReactions = append(allReactions, query.Node.Reactable.Reactions.Nodes...)
		if !query.Node.Reactable.Reactions.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Node.Reactable.Reactions.PageInfo.EndCursor)
	}
	return allReactions, nil
}

// GetDiscussionReactions retrieves all reactions on a discussion.
func (g *GitHubClient) GetDiscussionReactions(ctx context.Context, owner, repo string, number int) ([]Reaction, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Repository struct {
			Discussion struct {
				Reactions struct {
					Nodes    []Reaction
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"reactions(first: $first, after: $cursor)"`
			} `graphql:"discussion(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]any{
		"owner":  githubv4.String(owner),
		"repo":   githubv4.String(repo),
		"number": githubv4.Int(number),
		"first":  githubv4.Int(100),
		"cursor": (*githubv4.String)(nil),
	}

	var allReactions []Reaction
	for {
		if err := graphql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		allReactions = append(allReactions, query.Repository.Discussion.Reactions.Nodes...)
		if !query.Repository.Discussion.Reactions.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Repository.Discussion.Reactions.PageInfo.EndCursor)
	}
	return allReactions, nil
}

// ListDiscussionComments retrieves all top-level comments for a discussion.
// Replies and reactions on each comment are included up to the limit configured in the underlying GraphQL query (they are not paginated here).
func (g *GitHubClient) ListDiscussionComments(ctx context.Context, owner, repo string, number int) ([]DiscussionComment, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Repository struct {
			Discussion struct {
				Comments struct {
					Nodes    []DiscussionComment
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"comments(first: $first, after: $cursor)"`
			} `graphql:"discussion(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]any{
		"owner":  githubv4.String(owner),
		"repo":   githubv4.String(repo),
		"number": githubv4.Int(number),
		"first":  githubv4.Int(100),
		"cursor": (*githubv4.String)(nil),
	}

	var allComments []DiscussionComment
	for {
		if err := graphql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		allComments = append(allComments, query.Repository.Discussion.Comments.Nodes...)
		if !query.Repository.Discussion.Comments.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Repository.Discussion.Comments.PageInfo.EndCursor)
	}
	return allComments, nil
}

// CreateDiscussionComment adds a top-level comment to a discussion and returns the new comment's node ID.
func (g *GitHubClient) CreateDiscussionComment(ctx context.Context, discussionID, body string) (string, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return "", err
	}

	var mutation struct {
		AddDiscussionComment struct {
			Comment struct {
				ID githubv4.String
			}
		} `graphql:"addDiscussionComment(input: $input)"`
	}

	input := AddDiscussionCommentInput{
		DiscussionID: githubv4.ID(discussionID),
		Body:         githubv4.String(body),
	}

	if err := graphql.Mutate(ctx, &mutation, input, nil); err != nil {
		return "", err
	}

	return string(mutation.AddDiscussionComment.Comment.ID), nil
}

// AddDiscussionCommentReply adds a reply to an existing discussion comment and returns the new reply's node ID.
func (g *GitHubClient) AddDiscussionCommentReply(ctx context.Context, discussionID, replyToID, body string) (string, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return "", err
	}

	var mutation struct {
		AddDiscussionComment struct {
			Comment struct {
				ID githubv4.String
			}
		} `graphql:"addDiscussionComment(input: $input)"`
	}

	replyToIDVal := githubv4.ID(replyToID)
	input2 := AddDiscussionCommentInput{
		DiscussionID: githubv4.ID(discussionID),
		Body:         githubv4.String(body),
		ReplyToID:    &replyToIDVal,
	}

	if err := graphql.Mutate(ctx, &mutation, input2, nil); err != nil {
		return "", err
	}

	return string(mutation.AddDiscussionComment.Comment.ID), nil
}

// AddReaction adds a reaction to a subject (discussion or comment).
func (g *GitHubClient) AddReaction(ctx context.Context, subjectID, content string) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var mutation struct {
		AddReaction struct {
			Reaction struct {
				Content githubv4.String
			}
		} `graphql:"addReaction(input: $input)"`
	}

	input := AddReactionInput{
		SubjectID: githubv4.ID(subjectID),
		Content:   githubv4.String(content),
	}

	if err := graphql.Mutate(ctx, &mutation, input, nil); err != nil {
		return err
	}

	return nil
}
