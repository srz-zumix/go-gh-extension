package client

import (
	"context"

	"github.com/google/go-github/v93/github"
)

// CreateGitTree creates a Git tree from the specified entries.
func (g *GitHubClient) CreateGitTree(ctx context.Context, owner, repo, baseTree string, entries []*github.TreeEntry) (*github.Tree, error) {
	tree, _, err := g.client.Git.CreateTree(ctx, owner, repo, baseTree, entries)
	return tree, err
}

// CreateGitCommit creates a Git commit.
func (g *GitHubClient) CreateGitCommit(ctx context.Context, owner, repo string, commit github.Commit) (*github.Commit, error) {
	result, _, err := g.client.Git.CreateCommit(ctx, owner, repo, commit, nil)
	return result, err
}
