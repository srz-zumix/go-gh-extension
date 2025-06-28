package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v71/github"
)

func GetPullRequestNumberFromString(pr string) (int, error) {
	number, err := GetNumberFromString(pr)
	if err != nil {
		return 0, err
	}
	return number, nil
}

func GetPullRequestNumber(pull_request any) (int, error) {
	switch pr := pull_request.(type) {
	case string:
		return GetPullRequestNumberFromString(pr)
	case int:
		return pr, nil
	case *github.PullRequest:
		return pr.GetNumber(), nil
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
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	result, err := g.ReplaceIssueLabels(ctx, repo.Owner, repo.Name, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to set labels for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return result, nil
}

func AddPullRequestLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, labels []string) ([]*github.Label, error) {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	result, err := g.AddIssueLabels(ctx, repo.Owner, repo.Name, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to add labels to pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return result, nil
}

func RemovePullRequestLabel(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any, label string) error {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	err = g.RemoveIssueLabel(ctx, repo.Owner, repo.Name, number, label)
	if err != nil {
		return fmt.Errorf("failed to remove label '%s' from pull request #%d in repository '%s/%s': %w", label, number, repo.Owner, repo.Name, err)
	}
	return nil
}

func ClearPullRequestLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request any) error {
	number, err := GetPullRequestNumber(pull_request)
	if err != nil {
		return fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	err = g.ClearIssueLabels(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return fmt.Errorf("failed to clear labels for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return nil
}
