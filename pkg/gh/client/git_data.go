package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// GetGitCommit returns the git commit object for the given SHA.
// This uses the Git Data API, which is lighter than the Repositories API
// (no file patch data). The returned Commit includes tree SHA and parent SHAs.
func (g *GitHubClient) GetGitCommit(ctx context.Context, owner, repo, sha string) (*github.Commit, error) {
	commit, _, err := g.client.Git.GetCommit(ctx, owner, repo, sha)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

// GetGitTree returns the git tree for the given SHA.
// If recursive is true, all nested subtrees are included in Entries.
// Note: GitHub truncates recursive trees with more than 100,000 entries;
// check Tree.Truncated to detect this condition.
func (g *GitHubClient) GetGitTree(ctx context.Context, owner, repo, sha string, recursive bool) (*github.Tree, error) {
	tree, _, err := g.client.Git.GetTree(ctx, owner, repo, sha, recursive)
	if err != nil {
		return nil, err
	}
	return tree, nil
}
