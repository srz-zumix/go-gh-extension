package client

import (
	"context"

	"github.com/google/go-github/v71/github"
)

func (g *GitHubClient) GetPullRequest(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := g.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

// ListFiles lists files for a pull request
func (g *GitHubClient) ListPullRequestFiles(ctx context.Context, owner string, repo string, number int) ([]*github.CommitFile, error) {
	allCommitFiles := []*github.CommitFile{}
	opt := &github.ListOptions{PerPage: 50}

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
