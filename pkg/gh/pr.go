package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v71/github"
)

func GetPullRequest(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request string) (*github.PullRequest, error) {
	number, err := GetPullRequestNumberFromString(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	pr, err := g.GetPullRequest(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return pr, nil
}

func GetPullRequestByNumber(ctx context.Context, g *GitHubClient, repo repository.Repository, number int) (*github.PullRequest, error) {
	pr, err := g.GetPullRequest(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return pr, nil
}

func GetPullRequestNumberFromString(pr string) (int, error) {
	number, err := GetNumberFromString(pr)
	if err != nil {
		return 0, err
	}
	return number, nil
}

func ListPullRequestFiles(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request string) ([]*github.CommitFile, error) {
	number, err := GetPullRequestNumberFromString(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	files, err := g.ListPullRequestFiles(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list files for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return files, nil
}

func ListPullRequestFilesByNumber(ctx context.Context, g *GitHubClient, repo repository.Repository, number int) ([]*github.CommitFile, error) {
	files, err := g.ListPullRequestFiles(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list files for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return files, nil
}

func SetPullRequestLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, pull_request string, labels []string) ([]*github.Label, error) {
	number, err := GetPullRequestNumberFromString(pull_request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pull request number from '%s': %w", pull_request, err)
	}
	result, err := g.SetPullRequestLabels(ctx, repo.Owner, repo.Name, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to set labels for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return result, nil
}

func SetPullRequestLabelsByNumber(ctx context.Context, g *GitHubClient, repo repository.Repository, number int, labels []string) ([]*github.Label, error) {
	result, err := g.SetPullRequestLabels(ctx, repo.Owner, repo.Name, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to set labels for pull request #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return result, nil
}
