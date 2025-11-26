package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/shurcooL/githubv4"
)

func (g *GitHubClient) GetPullRequest(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := g.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (g *GitHubClient) ListPullRequestCommits(ctx context.Context, owner string, repo string, number int) ([]*github.RepositoryCommit, error) {
	allCommits := []*github.RepositoryCommit{}
	opt := &github.ListOptions{PerPage: defaultPerPage}

	for {
		commits, resp, err := g.client.PullRequests.ListCommits(ctx, owner, repo, number, opt)
		if err != nil {
			return nil, err
		}
		allCommits = append(allCommits, commits...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allCommits, nil
}

// ListFiles lists files for a pull request
func (g *GitHubClient) ListPullRequestFiles(ctx context.Context, owner string, repo string, number int) ([]*github.CommitFile, error) {
	allCommitFiles := []*github.CommitFile{}
	opt := &github.ListOptions{PerPage: defaultPerPage}

	for {
		files, resp, err := g.client.PullRequests.ListFiles(ctx, owner, repo, number, opt)
		if err != nil {
			return nil, err
		}
		allCommitFiles = append(allCommitFiles, files...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allCommitFiles, nil
}

func (g *GitHubClient) RequestReviewers(ctx context.Context, owner string, repo string, number int, reviewers []string, teamReviewers []string) (*github.PullRequest, error) {
	request := github.ReviewersRequest{
		Reviewers:     reviewers,
		TeamReviewers: teamReviewers,
	}
	pr, _, err := g.client.PullRequests.RequestReviewers(ctx, owner, repo, number, request)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (g *GitHubClient) RemoveReviewers(ctx context.Context, owner string, repo string, number int, reviewers []string, teamReviewers []string) error {
	request := github.ReviewersRequest{
		Reviewers:     reviewers,
		TeamReviewers: teamReviewers,
	}
	_, err := g.client.PullRequests.RemoveReviewers(ctx, owner, repo, number, request)
	return err
}

func (g *GitHubClient) ListRequestedReviewers(ctx context.Context, owner string, repo string, number int) (*github.Reviewers, error) {
	allReviewers := &github.Reviewers{}
	opt := &github.ListOptions{PerPage: defaultPerPage}

	for {
		reviewers, resp, err := g.client.PullRequests.ListReviewers(ctx, owner, repo, number, opt)
		if err != nil {
			return nil, err
		}
		allReviewers.Users = append(allReviewers.Users, reviewers.Users...)
		allReviewers.Teams = append(allReviewers.Teams, reviewers.Teams...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allReviewers, nil
}

func (g *GitHubClient) GetPullRequestReviews(ctx context.Context, owner string, repo string, number int) ([]*github.PullRequestReview, error) {
	allReviews := []*github.PullRequestReview{}
	opt := &github.ListOptions{PerPage: defaultPerPage}

	for {
		reviews, resp, err := g.client.PullRequests.ListReviews(ctx, owner, repo, number, opt)
		if err != nil {
			return nil, err
		}
		allReviews = append(allReviews, reviews...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allReviews, nil
}

func (g *GitHubClient) CreatePullRequestComment(ctx context.Context, owner string, repo string, number int, comment *github.PullRequestComment) (*github.PullRequestComment, error) {
	comment, _, err := g.client.PullRequests.CreateComment(ctx, owner, repo, number, comment)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (g *GitHubClient) DeletePullRequestComment(ctx context.Context, owner string, repo string, commentID int64) error {
	_, err := g.client.PullRequests.DeleteComment(ctx, owner, repo, commentID)
	if err != nil {
		return err
	}
	return nil
}

func (g *GitHubClient) EditPullRequestComment(ctx context.Context, owner string, repo string, commentID int64, body string) (*github.PullRequestComment, error) {
	comment := &github.PullRequestComment{
		Body: &body,
	}
	comment, _, err := g.client.PullRequests.EditComment(ctx, owner, repo, commentID, comment)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (g *GitHubClient) ListPullRequestReviewComments(ctx context.Context, owner string, repo string, number int) ([]*github.PullRequestComment, error) {
	allComments := []*github.PullRequestComment{}
	opt := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}

	for {
		comments, resp, err := g.client.PullRequests.ListComments(ctx, owner, repo, number, opt)
		if err != nil {
			return nil, err
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allComments, nil
}

func (g *GitHubClient) ResolveReviewThread(ctx context.Context, owner string, repo string, threadID string) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var m struct {
		ResolveReviewThread struct {
			Thread struct {
				ID         githubv4.String
				IsResolved githubv4.Boolean
			}
			ClientMutationID githubv4.String
		} `graphql:"resolveReviewThread(input: $input)"`
	}
	input := githubv4.ResolveReviewThreadInput{
		ThreadID: githubv4.String(threadID),
	}
	return graphql.Mutate(ctx, &m, input, nil)
}

func (g *GitHubClient) UnresolveReviewThread(ctx context.Context, owner string, repo string, threadID string) error {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}

	var m struct {
		UnresolveReviewThread struct {
			Thread struct {
				ID         githubv4.String
				IsResolved githubv4.Boolean
			}
			ClientMutationID githubv4.String
		} `graphql:"unresolveReviewThread(input: $input)"`
	}
	input := githubv4.UnresolveReviewThreadInput{
		ThreadID: githubv4.String(threadID),
	}
	return graphql.Mutate(ctx, &m, input, nil)
}

func (g *GitHubClient) GetPullRequestCommentThreadID(ctx context.Context, owner string, repo string, number int, commentID int64) (string, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return "", err
	}

	var q struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					Nodes []struct {
						ID       githubv4.String
						Comments struct {
							Nodes []struct {
								DatabaseID githubv4.Float
							}
						} `graphql:"comments(first: 10)"`
					}
				} `graphql:"reviewThreads(first: 100)"`
			} `graphql:"pullRequest(number: $pr)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	vars := map[string]interface{}{
		"owner": githubv4.String(owner),
		"repo":  githubv4.String(repo),
		"pr":    githubv4.Int(number),
	}
	if err := graphql.Query(ctx, &q, vars); err != nil {
		return "", err
	}
	for _, thread := range q.Repository.PullRequest.ReviewThreads.Nodes {
		for _, comment := range thread.Comments.Nodes {
			if comment.DatabaseID == githubv4.Float(commentID) {
				return string(thread.ID), nil
			}
		}
	}
	return "", fmt.Errorf("failed to find thread ID for comment %d", commentID)
}

// ListPullRequests retrieves all pull requests for a specific repository.
func (g *GitHubClient) ListPullRequests(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions, maxCount int) ([]*github.PullRequest, error) {
	var allPullRequests []*github.PullRequest
	perPage := defaultPerPage
	if maxCount > 0 {
		if maxCount < perPage {
			perPage = maxCount
		}
	}
	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: perPage},
	}
	if opts != nil {
		opt.State = opts.State
		opt.Head = opts.Head
		opt.Base = opts.Base
		opt.Sort = opts.Sort
		opt.Direction = opts.Direction
	}

	for {
		prs, resp, err := g.client.PullRequests.List(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allPullRequests = append(allPullRequests, prs...)
		if resp.NextPage == 0 || maxCount < 0 || len(allPullRequests) >= maxCount {
			break
		}
		opt.Page = resp.NextPage
	}

	return allPullRequests, nil
}

// associatedPullRequestNode represents a pull request from GraphQL
type associatedPullRequestNode struct {
	Number           githubv4.Int
	Title            githubv4.String
	State            githubv4.String
	URL              githubv4.URI
	Body             githubv4.String
	CreatedAt        githubv4.DateTime
	UpdatedAt        githubv4.DateTime
	ClosedAt         githubv4.DateTime
	MergedAt         githubv4.DateTime
	Mergeable        githubv4.String
	IsDraft          githubv4.Boolean
	Locked           githubv4.Boolean
	ActiveLockReason githubv4.String
	Author           struct {
		Login githubv4.String
	}
	HeadRef struct {
		Name githubv4.String
	}
	HeadRefOid githubv4.String
	BaseRef    struct {
		Name githubv4.String
	}
	BaseRefOid githubv4.String
	Repository struct {
		Name  githubv4.String
		Owner struct {
			Login githubv4.String
		}
	}
}

type AssociatedPullRequestsOption struct {
	States  []string
	OrderBy *GraphQLOrderByOption
}

// GetAssociatedPullRequestsForRef retrieves pull requests associated with a ref using GraphQL
func (g *GitHubClient) GetAssociatedPullRequestsForRef(ctx context.Context, owner string, repo string, ref string, opts *AssociatedPullRequestsOption) ([]*github.PullRequest, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	states := []githubv4.PullRequestState{}
	for _, s := range opts.States {
		switch strings.ToLower(s) {
		case "open":
			states = append(states, githubv4.PullRequestStateOpen)
		case "closed":
			states = append(states, githubv4.PullRequestStateClosed)
		case "merged":
			states = append(states, githubv4.PullRequestStateMerged)
		}
	}

	// Ensure ref has refs/heads/ prefix for branch names
	qualifiedRef := ref
	if !strings.HasPrefix(ref, "refs/") {
		qualifiedRef = "refs/heads/" + ref
	}

	var query struct {
		Repository struct {
			Ref struct {
				AssociatedPullRequests struct {
					Nodes    []associatedPullRequestNode
					PageInfo struct {
						HasNextPage githubv4.Boolean
						EndCursor   githubv4.String
					}
				} `graphql:"associatedPullRequests(first: 100, states: $states, orderBy: $orderBy)"`
			} `graphql:"ref(qualifiedName: $ref)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":   githubv4.String(owner),
		"name":    githubv4.String(repo),
		"ref":     githubv4.String(qualifiedRef),
		"states":  states,
		"orderBy": opts.OrderBy.ToIssueOrder(),
	}

	err = graphql.Query(ctx, &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query GraphQL: %w", err)
	}

	var result []*github.PullRequest
	for i := range query.Repository.Ref.AssociatedPullRequests.Nodes {
		result = append(result, convertAssociatedPRToGitHubPR(&query.Repository.Ref.AssociatedPullRequests.Nodes[i]))
	}

	return result, nil
}

// convertAssociatedPRToGitHubPR converts associatedPullRequestNode to github.PullRequest
func convertAssociatedPRToGitHubPR(node *associatedPullRequestNode) *github.PullRequest {
	number := int(node.Number)
	title := string(node.Title)
	state := strings.ToLower(string(node.State))
	htmlURL := node.URL.String()
	body := string(node.Body)
	draft := bool(node.IsDraft)
	locked := bool(node.Locked)
	authorLogin := string(node.Author.Login)
	headRef := string(node.HeadRef.Name)
	headSHA := string(node.HeadRefOid)
	baseRef := string(node.BaseRef.Name)
	baseSHA := string(node.BaseRefOid)
	repoName := string(node.Repository.Name)
	repoOwner := string(node.Repository.Owner.Login)

	createdAt := node.CreatedAt.Time
	updatedAt := node.UpdatedAt.Time

	pr := &github.PullRequest{
		Number:    &number,
		Title:     &title,
		State:     &state,
		HTMLURL:   &htmlURL,
		Body:      &body,
		Draft:     &draft,
		Locked:    &locked,
		CreatedAt: &github.Timestamp{Time: createdAt},
		UpdatedAt: &github.Timestamp{Time: updatedAt},
		User: &github.User{
			Login: &authorLogin,
		},
		Head: &github.PullRequestBranch{
			Ref: &headRef,
			SHA: &headSHA,
			Repo: &github.Repository{
				Name: &repoName,
				Owner: &github.User{
					Login: &repoOwner,
				},
			},
		},
		Base: &github.PullRequestBranch{
			Ref: &baseRef,
			SHA: &baseSHA,
			Repo: &github.Repository{
				Name: &repoName,
				Owner: &github.User{
					Login: &repoOwner,
				},
			},
		},
	}

	if !node.ClosedAt.IsZero() {
		closedAt := node.ClosedAt.Time
		pr.ClosedAt = &github.Timestamp{Time: closedAt}
	}

	if !node.MergedAt.IsZero() {
		mergedAt := node.MergedAt.Time
		pr.MergedAt = &github.Timestamp{Time: mergedAt}
		merged := true
		pr.Merged = &merged
	}

	if node.Mergeable != "" {
		mergeableState := string(node.Mergeable)
		pr.MergeableState = &mergeableState
		mergeable := mergeableState == "MERGEABLE"
		pr.Mergeable = &mergeable
	}

	if node.ActiveLockReason != "" {
		lockReason := string(node.ActiveLockReason)
		pr.ActiveLockReason = &lockReason
	}

	return pr
}
