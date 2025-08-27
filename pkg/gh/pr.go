package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
)

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

func SetPullRequestLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, labels []string) ([]*github.Label, error) {
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
