package client

import (
	"context"

	"github.com/google/go-github/v73/github"
)

func (g *GitHubClient) ListCommits(ctx context.Context, owner, repo string, options *github.CommitsListOptions) ([]*github.RepositoryCommit, error) {
	var allCommits []*github.RepositoryCommit
	opt := github.CommitsListOptions{ListOptions: github.ListOptions{PerPage: 50}}
	if options != nil {
		opt = *options
		opt.PerPage = 50
	}

	for {
		commits, resp, err := g.client.Repositories.ListCommits(ctx, owner, repo, &opt)
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

func (g *GitHubClient) ListBranchesHeadCommit(ctx context.Context, owner, repo, sha string) ([]*github.BranchCommit, error) {
	c, _, err := g.client.Repositories.ListBranchesHeadCommit(ctx, owner, repo, sha)
	return c, err
}

func (g *GitHubClient) GetCommit(ctx context.Context, owner, repo, sha string) (commit *github.RepositoryCommit, err error) {
	opt := &github.ListOptions{PerPage: 100}
	for {
		c, resp, err := g.client.Repositories.GetCommit(ctx, owner, repo, sha, opt)
		if err != nil {
			return nil, err
		}
		if commit == nil {
			commit = c
		} else {
			commit.Files = append(commit.Files, c.Files...)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return commit, nil
}

func (g *GitHubClient) GetCommitSHA1(ctx context.Context, owner, repo, ref string, lastSHA string) (string, error) {
	commit, _, err := g.client.Repositories.GetCommitSHA1(ctx, owner, repo, ref, lastSHA)
	if err != nil {
		return "", err
	}
	return commit, nil
}

func (g *GitHubClient) CompareCommits(ctx context.Context, owner, repo, base, head string) (commitComparison *github.CommitsComparison, err error) {
	opt := &github.ListOptions{PerPage: 100}

	for {
		comp, resp, err := g.client.Repositories.CompareCommits(ctx, owner, repo, base, head, opt)
		if err != nil {
			return nil, err
		}
		if commitComparison == nil {
			commitComparison = comp
		} else {
			commitComparison.Commits = append(commitComparison.Commits, comp.Commits...)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return commitComparison, nil
}
