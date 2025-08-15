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
