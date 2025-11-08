package client

import (
	"context"

	"github.com/google/go-github/v73/github"
)

// GetIssueByNumber retrieves an issue by owner, repo, and issue number
func (g *GitHubClient) GetIssueByNumber(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	issue, _, err := g.client.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

// SetPullRequestLabels sets labels for a pull request
func (g *GitHubClient) ReplaceIssueLabels(ctx context.Context, owner string, repo string, number int, labels []string) ([]*github.Label, error) {
	result, _, err := g.client.Issues.ReplaceLabelsForIssue(ctx, owner, repo, number, labels)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *GitHubClient) AddIssueLabels(ctx context.Context, owner string, repo string, number int, labels []string) ([]*github.Label, error) {
	result, _, err := g.client.Issues.AddLabelsToIssue(ctx, owner, repo, number, labels)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *GitHubClient) RemoveIssueLabel(ctx context.Context, owner string, repo string, number int, label string) error {
	_, err := g.client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label)
	if err != nil {
		return err
	}
	return nil
}

func (g *GitHubClient) ClearIssueLabels(ctx context.Context, owner string, repo string, number int) error {
	_, err := g.client.Issues.RemoveLabelsForIssue(ctx, owner, repo, number)
	if err != nil {
		return err
	}
	return nil
}

func (g *GitHubClient) CreateIssueComment(ctx context.Context, owner string, repo string, number int, body string) (*github.IssueComment, error) {
	comment, _, err := g.client.Issues.CreateComment(ctx, owner, repo, number, &github.IssueComment{Body: &body})
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (g *GitHubClient) DeleteIssueComment(ctx context.Context, owner string, repo string, commentID int64) error {
	_, err := g.client.Issues.DeleteComment(ctx, owner, repo, commentID)
	if err != nil {
		return err
	}
	return nil
}

func (g *GitHubClient) EditIssueComment(ctx context.Context, owner string, repo string, commentID int64, body string) (*github.IssueComment, error) {
	comment, _, err := g.client.Issues.EditComment(ctx, owner, repo, commentID, &github.IssueComment{Body: &body})
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (g *GitHubClient) ListIssueComments(ctx context.Context, owner string, repo string, number int) ([]*github.IssueComment, error) {
	allComments := []*github.IssueComment{}
	opt := &github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	for {
		comments, resp, err := g.client.Issues.ListComments(ctx, owner, repo, number, opt)
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
