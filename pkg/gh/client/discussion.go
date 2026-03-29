package client

import (
	"context"
	"time"

	"github.com/shurcooL/githubv4"
)

// Discussion represents a GitHub discussion with plain Go types.
type Discussion struct {
	ID     string
	Number int
	Title  string
	Body   string
	Author struct {
		Login string
	}
	Category struct {
		Name string
		Slug string
	}
	Labels struct {
		Nodes []struct {
			ID          string
			Name        string
			Color       string
			Description *string
		}
	}
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ClosedAt    *time.Time
	Locked      bool
	StateReason *string
	URL         string
}

// rawDiscussion is the internal GraphQL representation of a discussion.
type rawDiscussion struct {
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

// toDiscussion converts a rawDiscussion to a plain-type Discussion.
func (r rawDiscussion) toDiscussion() Discussion {
	d := Discussion{
		ID:        string(r.ID),
		Number:    int(r.Number),
		Title:     string(r.Title),
		Body:      string(r.Body),
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
		Locked:    bool(r.Locked),
		URL:       string(r.URL),
	}
	d.Author.Login = string(r.Author.Login)
	d.Category.Name = string(r.Category.Name)
	d.Category.Slug = string(r.Category.Slug)
	if r.ClosedAt != nil {
		t := r.ClosedAt.Time
		d.ClosedAt = &t
	}
	if r.StateReason != nil {
		s := string(*r.StateReason)
		d.StateReason = &s
	}
	for _, l := range r.Labels.Nodes {
		node := struct {
			ID          string
			Name        string
			Color       string
			Description *string
		}{
			ID:    string(l.ID),
			Name:  string(l.Name),
			Color: string(l.Color),
		}
		if l.Description != nil {
			s := string(*l.Description)
			node.Description = &s
		}
		d.Labels.Nodes = append(d.Labels.Nodes, node)
	}
	return d
}

// GetDiscussion retrieves a specific discussion by number using GraphQL
func (g *GitHubClient) GetDiscussion(ctx context.Context, owner string, repo string, number int) (*Discussion, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Repository struct {
			Discussion rawDiscussion `graphql:"discussion(number: $number)"`
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

	d := query.Repository.Discussion.toDiscussion()
	return &d, nil
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
				Nodes    []rawDiscussion
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
		for _, d := range query.Repository.Discussions.Nodes {
			allDiscussions = append(allDiscussions, d.toDiscussion())
		}
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
				Discussion rawDiscussion `graphql:"... on Discussion"`
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
			allDiscussions = append(allDiscussions, node.Discussion.toDiscussion())
		}
		if !searchQuery.Search.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(searchQuery.Search.PageInfo.EndCursor)
	}

	return allDiscussions, nil
}

// DiscussionCategory represents a discussion category with plain Go types.
type DiscussionCategory struct {
	ID   string
	Name string
	Slug string
}

// rawDiscussionCategory is the internal GraphQL representation of a discussion category.
type rawDiscussionCategory struct {
	ID   githubv4.String
	Name githubv4.String
	Slug githubv4.String
}

// toDiscussionCategory converts a rawDiscussionCategory to a plain-type DiscussionCategory.
func (r rawDiscussionCategory) toDiscussionCategory() DiscussionCategory {
	return DiscussionCategory{
		ID:   string(r.ID),
		Name: string(r.Name),
		Slug: string(r.Slug),
	}
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
				Nodes    []rawDiscussionCategory
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
		for _, c := range query.Repository.DiscussionCategories.Nodes {
			allCategories = append(allCategories, c.toDiscussionCategory())
		}
		if !query.Repository.DiscussionCategories.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Repository.DiscussionCategories.PageInfo.EndCursor)
	}

	return allCategories, nil
}

// CreateDiscussionOption is the input for creating a discussion.
type CreateDiscussionOption struct {
	RepositoryID string `json:"repositoryId"`
	CategoryID   string `json:"categoryId"`
	Title        string `json:"title"`
	Body         string `json:"body"`
}

// CreateDiscussion creates a new discussion in a repository using GraphQL mutation.
func (g *GitHubClient) CreateDiscussion(ctx context.Context, input CreateDiscussionOption) (*Discussion, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var mutation struct {
		CreateDiscussion struct {
			Discussion rawDiscussion
		} `graphql:"createDiscussion(input: $input)"`
	}

	type CreateDiscussionInput struct {
		RepositoryID githubv4.ID     `json:"repositoryId"`
		CategoryID   githubv4.ID     `json:"categoryId"`
		Title        githubv4.String `json:"title"`
		Body         githubv4.String `json:"body"`
	}

	gqlInput := CreateDiscussionInput{
		RepositoryID: githubv4.ID(input.RepositoryID),
		CategoryID:   githubv4.ID(input.CategoryID),
		Title:        githubv4.String(input.Title),
		Body:         githubv4.String(input.Body),
	}

	if err := graphql.Mutate(ctx, &mutation, gqlInput, nil); err != nil {
		return nil, err
	}

	d := mutation.CreateDiscussion.Discussion.toDiscussion()
	return &d, nil
}

// DeleteDiscussion deletes a discussion by its node ID.
func (g *GitHubClient) DeleteDiscussion(ctx context.Context, discussionID string) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var mutation struct {
		DeleteDiscussion struct {
			ClientMutationID githubv4.String
		} `graphql:"deleteDiscussion(input: $input)"`
	}

	type DeleteDiscussionInput struct {
		ID githubv4.ID `json:"id"`
	}

	input := DeleteDiscussionInput{ID: githubv4.ID(discussionID)}
	return graphql.Mutate(ctx, &mutation, input, nil)
}

// UpdateDiscussion updates the body of an existing discussion.
func (g *GitHubClient) UpdateDiscussion(ctx context.Context, discussionID, body string) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var mutation struct {
		UpdateDiscussion struct {
			Discussion struct {
				ID githubv4.String
			}
		} `graphql:"updateDiscussion(input: $input)"`
	}

	type UpdateDiscussionInput struct {
		DiscussionID githubv4.ID     `json:"discussionId"`
		Body         githubv4.String `json:"body"`
	}

	gqlInput := UpdateDiscussionInput{
		DiscussionID: githubv4.ID(discussionID),
		Body:         githubv4.String(body),
	}

	return graphql.Mutate(ctx, &mutation, gqlInput, nil)
}

// DeleteDiscussionComment deletes a discussion comment (or reply) by its node ID.
func (g *GitHubClient) DeleteDiscussionComment(ctx context.Context, commentID string) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var mutation struct {
		DeleteDiscussionComment struct {
			ClientMutationID githubv4.String
		} `graphql:"deleteDiscussionComment(input: $input)"`
	}

	type DeleteDiscussionCommentInput struct {
		ID githubv4.ID `json:"id"`
	}

	input := DeleteDiscussionCommentInput{ID: githubv4.ID(commentID)}
	return graphql.Mutate(ctx, &mutation, input, nil)
}

// Reaction represents a single reaction on a discussion or comment with plain Go types.
type Reaction struct {
	Content string
	User    struct {
		Login string
	}
}

// rawReaction is the internal GraphQL representation of a reaction.
type rawReaction struct {
	Content githubv4.String
	User    struct {
		Login githubv4.String
	}
}

// toReaction converts a rawReaction to a plain-type Reaction.
func (r rawReaction) toReaction() Reaction {
	react := Reaction{Content: string(r.Content)}
	react.User.Login = string(r.User.Login)
	return react
}

// DiscussionCommentReply represents a reply to a top-level discussion comment with plain Go types.
type DiscussionCommentReply struct {
	ID     string
	Body   string
	Author struct {
		Login string
	}
}

// rawDiscussionCommentReply is the internal GraphQL representation of a discussion comment reply.
// Reactions are not included inline to avoid exceeding GraphQL node limits;
// use GetNodeReactions to fetch them separately.
type rawDiscussionCommentReply struct {
	ID     githubv4.String
	Body   githubv4.String
	Author struct {
		Login githubv4.String
	}
}

// toDiscussionCommentReply converts a rawDiscussionCommentReply to a plain-type DiscussionCommentReply.
func (r rawDiscussionCommentReply) toDiscussionCommentReply() DiscussionCommentReply {
	reply := DiscussionCommentReply{
		ID:   string(r.ID),
		Body: string(r.Body),
	}
	reply.Author.Login = string(r.Author.Login)
	return reply
}

// DiscussionComment represents a top-level comment on a discussion with plain Go types.
type DiscussionComment struct {
	ID     string
	Body   string
	Author struct {
		Login string
	}
	Reactions struct {
		Nodes []Reaction
	}
	Replies struct {
		Nodes []DiscussionCommentReply
	}
}

// rawDiscussionComment is the internal GraphQL representation of a discussion comment.
type rawDiscussionComment struct {
	ID     githubv4.String
	Body   githubv4.String
	Author struct {
		Login githubv4.String
	}
	Reactions struct {
		Nodes []rawReaction
	} `graphql:"reactions(first: 100)"`
	Replies struct {
		Nodes []rawDiscussionCommentReply
	} `graphql:"replies(first: 100)"`
}

// toDiscussionComment converts a rawDiscussionComment to a plain-type DiscussionComment.
func (r rawDiscussionComment) toDiscussionComment() DiscussionComment {
	c := DiscussionComment{
		ID:   string(r.ID),
		Body: string(r.Body),
	}
	c.Author.Login = string(r.Author.Login)
	for _, reaction := range r.Reactions.Nodes {
		c.Reactions.Nodes = append(c.Reactions.Nodes, reaction.toReaction())
	}
	for _, reply := range r.Replies.Nodes {
		c.Replies.Nodes = append(c.Replies.Nodes, reply.toDiscussionCommentReply())
	}
	return c
}

// AddDiscussionCommentInput is the internal GraphQL input for adding a comment to a discussion.
// Set ReplyToID to add a reply to an existing top-level comment.
type AddDiscussionCommentInput struct {
	DiscussionID githubv4.ID     `json:"discussionId"`
	Body         githubv4.String `json:"body"`
	ReplyToID    *githubv4.ID    `json:"replyToId,omitempty"`
}

// AddReactionInput is the internal GraphQL input for adding a reaction to a subject.
type AddReactionInput struct {
	SubjectID githubv4.ID              `json:"subjectId"`
	Content   githubv4.ReactionContent `json:"content"`
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
					Nodes    []rawReaction
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
		for _, r := range query.Repository.Discussion.Reactions.Nodes {
			allReactions = append(allReactions, r.toReaction())
		}
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
					Nodes    []rawDiscussionComment
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
		for _, c := range query.Repository.Discussion.Comments.Nodes {
			allComments = append(allComments, c.toDiscussionComment())
		}
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
		Content:   githubv4.ReactionContent(content),
	}

	if err := graphql.Mutate(ctx, &mutation, input, nil); err != nil {
		return err
	}

	return nil
}
