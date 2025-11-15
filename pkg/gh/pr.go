package gh

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
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

func GetPullRequestNumberFromString(pr string) (int, error) {
	return GetIssueNumberFromString(pr)
}

func GetPullRequestNumber(pull_request any) (int, error) {
	return GetIssueNumber(pull_request)
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
