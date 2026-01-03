package gh

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

type ReviewersRequest struct {
	Reviewers     []string
	TeamReviewers []string
}

type ListPullRequestsOption interface {
	apply(*github.PullRequestListOptions)
}

type ListPullRequestsOptionHead struct {
	Head string
}

func (o *ListPullRequestsOptionHead) apply(opts *github.PullRequestListOptions) {
	opts.Head = o.Head
}

type ListPullRequestsOptionBase struct {
	Base string
}

func (o *ListPullRequestsOptionBase) apply(opts *github.PullRequestListOptions) {
	opts.Base = o.Base
}

type ListPullRequestsOptionState struct {
	State string
}

func (o *ListPullRequestsOptionState) apply(opts *github.PullRequestListOptions) {
	opts.State = o.State
}

func ListPullRequestsOptionStateOpen() *ListPullRequestsOptionState {
	return &ListPullRequestsOptionState{State: "open"}
}

func ListPullRequestsOptionStateClosed() *ListPullRequestsOptionState {
	return &ListPullRequestsOptionState{State: "closed"}
}

func ListPullRequestsOptionStateAll() *ListPullRequestsOptionState {
	return &ListPullRequestsOptionState{State: "all"}
}

type ListPullRequestsOptionSort struct {
	Sort string
}

func (o *ListPullRequestsOptionSort) apply(opts *github.PullRequestListOptions) {
	opts.Sort = o.Sort
}

func ListPullRequestsOptionSortCreated() *ListPullRequestsOptionSort {
	return &ListPullRequestsOptionSort{Sort: "created"}
}

func ListPullRequestsOptionSortUpdated() *ListPullRequestsOptionSort {
	return &ListPullRequestsOptionSort{Sort: "updated"}
}

func ListPullRequestsOptionSortPopularity() *ListPullRequestsOptionSort {
	return &ListPullRequestsOptionSort{Sort: "popularity"}
}

func ListPullRequestsOptionSortLongRunning() *ListPullRequestsOptionSort {
	return &ListPullRequestsOptionSort{Sort: "long-running"}
}

type ListPullRequestsOptionDirection struct {
	Direction string
}

func (o *ListPullRequestsOptionDirection) apply(opts *github.PullRequestListOptions) {
	opts.Direction = o.Direction
}

func ListPullRequestsOptionDirectionAscending() *ListPullRequestsOptionDirection {
	return &ListPullRequestsOptionDirection{Direction: "asc"}
}

func ListPullRequestsOptionDirectionDescending() *ListPullRequestsOptionDirection {
	return &ListPullRequestsOptionDirection{Direction: "desc"}
}

var PullRequestReviewStateApproved = "APPROVED"
var PullRequestReviewStateChangesRequested = "CHANGES_REQUESTED"
var PullRequestReviewStateCommented = "COMMENTED"
var PullRequestReviewStateDismissed = "DISMISSED"
var PullRequestReviewStatePending = "PENDING"

func ListPullRequests(ctx context.Context, g *GitHubClient, repo repository.Repository, opts ...ListPullRequestsOption) ([]*github.PullRequest, error) {
	var ghOpts *github.PullRequestListOptions
	if opts != nil {
		ghOpts = &github.PullRequestListOptions{}
		for _, opt := range opts {
			opt.apply(ghOpts)
		}
	}
	return g.ListPullRequests(ctx, repo.Owner, repo.Name, ghOpts, -1)
}

func FindPullRequest(ctx context.Context, g *GitHubClient, repo repository.Repository, opts ...ListPullRequestsOption) (*github.PullRequest, error) {
	var ghOpts *github.PullRequestListOptions
	if opts != nil {
		ghOpts = &github.PullRequestListOptions{}
		for _, opt := range opts {
			opt.apply(ghOpts)
		}
	}
	prs, err := g.ListPullRequests(ctx, repo.Owner, repo.Name, ghOpts, 1)
	if err != nil {
		return nil, err
	}
	if len(prs) == 0 {
		return nil, fmt.Errorf("no pull request found")
	}
	return prs[0], nil
}

func GetRequestedReviewers(reviewers []string) ReviewersRequest {
	reviewersRequest := ReviewersRequest{}
	for _, reviewer := range reviewers {
		if reviewer[0] == '@' {
			reviewer = reviewer[1:]
		}
		if strings.Contains(reviewer, "/") {
			reviewersRequest.TeamReviewers = append(reviewersRequest.TeamReviewers, reviewer)
		} else {
			reviewersRequest.Reviewers = append(reviewersRequest.Reviewers, reviewer)
		}
	}
	return reviewersRequest
}

func ExpandTeamReviewers(ctx context.Context, g *GitHubClient, repo repository.Repository, reviewersRequest ReviewersRequest) (ReviewersRequest, error) {
	expanded := ReviewersRequest{
		Reviewers: reviewersRequest.Reviewers,
	}

	for _, teamSlug := range reviewersRequest.TeamReviewers {
		// Extract team slug from org/team format
		parts := strings.Split(teamSlug, "/")
		var slug string
		if len(parts) == 2 {
			slug = parts[1]
		} else {
			slug = teamSlug
		}

		members, err := ListTeamMembers(ctx, g, repo, slug, nil, false)
		if err != nil {
			return ReviewersRequest{}, fmt.Errorf("failed to get members for team '%s': %w", teamSlug, err)
		}

		for _, member := range members {
			expanded.Reviewers = append(expanded.Reviewers, member.GetLogin())
		}
	}

	return expanded, nil
}

func GetPullRequestNumber(pull_request any) (int, error) {
	switch t := pull_request.(type) {
	case string:
		return parser.GetPullRequestNumberFromString(t)
	case int:
		return t, nil
	case *github.PullRequest:
		return t.GetNumber(), nil
	default:
		return 0, fmt.Errorf("unsupported pull request type: %T", pull_request)
	}
}

func GetPullRequest(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) (*github.PullRequest, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	pr, err := g.GetPullRequest(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return pr, nil
}

func ListPullRequestCommits(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) ([]*github.RepositoryCommit, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	commits, err := g.ListPullRequestCommits(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return commits, nil
}

func ListPullRequestFiles(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) ([]*github.CommitFile, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	files, err := g.ListPullRequestFiles(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list files for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return files, nil
}

func RequestPullRequestReviewers(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, reviewersRequest ReviewersRequest) (*github.PullRequest, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	pr, err := g.RequestReviewers(ctx, repo.Owner, repo.Name, number, reviewersRequest.Reviewers, reviewersRequest.TeamReviewers)
	if err != nil {
		return nil, fmt.Errorf("failed to request reviewers for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return pr, nil
}

func RemovePullRequestReviewers(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, reviewersRequest ReviewersRequest) error {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	err = g.RemoveReviewers(ctx, repo.Owner, repo.Name, number, reviewersRequest.Reviewers, reviewersRequest.TeamReviewers)
	if err != nil {
		return fmt.Errorf("failed to remove reviewers for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return nil
}

func ListPullRequestReviewers(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) (*github.Reviewers, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	reviewers, err := g.ListRequestedReviewers(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviewers for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return reviewers, nil
}

func SetPullRequestLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, labels []string) ([]*github.Label, error) {
	if len(labels) == 0 {
		err := ClearIssueLabels(ctx, g, repo, pull_request)
		if err != nil {
			return nil, fmt.Errorf("failed to clear labels for pull request '%s': %w", pull_request, err)
		}
		return []*github.Label{}, nil
	}
	return SetIssueLabels(ctx, g, repo, pull_request, labels)
}

func AddPullRequestLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, labels []string) ([]*github.Label, error) {
	return AddIssueLabels(ctx, g, repo, pull_request, labels)
}

func RemovePullRequestLabel(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, label string) error {
	return RemoveIssueLabel(ctx, g, repo, pull_request, label)
}

func RemovePullRequestLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, label []string) ([]*github.Label, error) {
	return RemoveIssueLabels(ctx, g, repo, pull_request, label)
}

func ClearPullRequestLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) error {
	return ClearIssueLabels(ctx, g, repo, pull_request)
}

func GetPullRequestReviews(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) ([]*github.PullRequestReview, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	reviews, err := g.GetPullRequestReviews(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return reviews, nil
}

func GetPullRequestLatestReviews(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) ([]*github.PullRequestReview, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	reviews, err := g.GetPullRequestReviews(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get review status for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	latestReviews := make(map[string]*github.PullRequestReview)
	for _, review := range reviews {
		latestReviews[review.User.GetLogin()] = review
	}
	return slices.Collect(maps.Values(latestReviews)), nil
}

func GetApprovedReviewers(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) ([]string, error) {
	reviews, err := GetPullRequestLatestReviews(ctx, g, repo, pull_request)
	if err != nil {
		return nil, err
	}

	approvedReviewers := []string{}
	for _, review := range reviews {
		if review.GetState() == PullRequestReviewStateApproved {
			approvedReviewers = append(approvedReviewers, review.User.GetLogin())
		}
	}
	return approvedReviewers, nil
}

func ListPullRequestReviewComments(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) ([]*github.PullRequestComment, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	comments, err := g.ListPullRequestReviewComments(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list review comments for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return comments, nil
}

func CreatePullRequestComment(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, comment *github.PullRequestComment) (*github.PullRequestComment, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	comment, err = g.CreatePullRequestComment(ctx, repo.Owner, repo.Name, number, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return comment, nil
}

func DeletePullRequestComment(ctx context.Context, g *GitHubClient, repo repository.Repository, comment any) error {
	commentID, err := GetCommentID(comment)
	if err != nil {
		return fmt.Errorf("failed to parse comment ID from '%s': %w", comment, err)
	}
	return g.DeletePullRequestComment(ctx, repo.Owner, repo.Name, commentID)
}

func EditPullRequestComment(ctx context.Context, g *GitHubClient, repo repository.Repository, comment any, body string) (*github.PullRequestComment, error) {
	commentID, err := GetCommentID(comment)
	if err != nil {
		return nil, fmt.Errorf("failed to parse comment ID from '%s': %w", comment, err)
	}
	return g.EditPullRequestComment(ctx, repo.Owner, repo.Name, commentID, body)
}

func ResolvePullRequestComment(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, comment any) error {
	threadID, err := GetPullRequestCommentThreadID(ctx, g, repo, pull_request, comment)
	if err != nil {
		return fmt.Errorf("failed to get thread ID from pull request '%s' and comment '%s': %w", pull_request, comment, err)
	}
	return g.ResolveReviewThread(ctx, repo.Owner, repo.Name, threadID)
}

func UnresolvePullRequestComment(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, comment any) error {
	threadID, err := GetPullRequestCommentThreadID(ctx, g, repo, pull_request, comment)
	if err != nil {
		return fmt.Errorf("failed to get thread ID from pull request '%s' and comment '%s': %w", pull_request, comment, err)
	}
	return g.UnresolveReviewThread(ctx, repo.Owner, repo.Name, threadID)
}

func GetPullRequestCommentThreadID(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, comment any) (string, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return "", fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	commentID, err := GetCommentID(comment)
	if err != nil {
		return "", fmt.Errorf("failed to parse comment ID from '%s': %w", comment, err)
	}
	return g.GetPullRequestCommentThreadID(ctx, repo.Owner, repo.Name, number, commentID)
}

// AssociatedPullRequestsOption is an interface for specifying options when querying associated pull requests.
type AssociatedPullRequestsOption interface {
	apply(*client.AssociatedPullRequestsOption)
}

// AssociatedPullRequestsOptionOrderBy configures the ordering of associated pull requests when querying.
type AssociatedPullRequestsOptionOrderBy struct {
	OrderBy client.GraphQLOrderByOption
}

func (o *AssociatedPullRequestsOptionOrderBy) apply(opts *client.AssociatedPullRequestsOption) {
	opts.OrderBy = &o.OrderBy
}

// AssociatedPullRequestsOptionStates configures the states of associated pull requests when querying.
type AssociatedPullRequestsOptionStates struct {
	State string
}

func (o *AssociatedPullRequestsOptionStates) apply(opts *client.AssociatedPullRequestsOption) {
	opts.States = append(opts.States, o.State)
}

// AssociatedPullRequestsOptionStateOpen configures the query to include only open pull requests.
func AssociatedPullRequestsOptionStateOpen() *AssociatedPullRequestsOptionStates {
	return &AssociatedPullRequestsOptionStates{State: "OPEN"}
}

// AssociatedPullRequestsOptionStateClosed configures the query to include only closed pull requests.
func AssociatedPullRequestsOptionStateClosed() *AssociatedPullRequestsOptionStates {
	return &AssociatedPullRequestsOptionStates{State: "CLOSED"}
}

// AssociatedPullRequestsOptionStateMerged configures the query to include only merged pull requests.
func AssociatedPullRequestsOptionStateMerged() *AssociatedPullRequestsOptionStates {
	return &AssociatedPullRequestsOptionStates{State: "MERGED"}
}

// GetAssociatedPullRequestsForRef retrieves pull requests associated with a specific ref.
func GetAssociatedPullRequestsForRef(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string, opts ...AssociatedPullRequestsOption) ([]*github.PullRequest, error) {
	option := &client.AssociatedPullRequestsOption{}
	for _, opt := range opts {
		opt.apply(option)
	}
	return g.GetAssociatedPullRequestsForRef(ctx, repo.Owner, repo.Name, ref, option)
}
