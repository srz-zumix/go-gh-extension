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

// GetGitBlob returns the git blob object for the given SHA.
// The returned Blob includes the Size field (in bytes) without fetching content.
func (g *GitHubClient) GetGitBlob(ctx context.Context, owner, repo, sha string) (*github.Blob, error) {
	blob, _, err := g.client.Git.GetBlob(ctx, owner, repo, sha)
	if err != nil {
		return nil, err
	}
	return blob, nil
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

// GetGitTreeRecursive fetches a tree non-recursively and then manually
// traverses into each subtree entry, prefixing child paths with the parent
// directory so that every entry in the returned slice carries a full path
// relative to the root tree (e.g. "dir/file.txt" instead of "file.txt").
// This avoids the GitHub 100,000-entry truncation limit of recursive fetches.
func (g *GitHubClient) GetGitTreeRecursive(ctx context.Context, owner, repo, sha string) (*github.Tree, error) {
	return g.getGitTreeRecursive(ctx, owner, repo, sha, "")
}

// getGitTreeRecursive is the internal implementation that carries the path
// prefix accumulated from parent tree entries.
func (g *GitHubClient) getGitTreeRecursive(ctx context.Context, owner, repo, sha, prefix string) (*github.Tree, error) {
	tree, _, err := g.client.Git.GetTree(ctx, owner, repo, sha, false)
	if err != nil {
		return nil, err
	}

	// Build a new entry slice with corrected paths.
	entries := make([]*github.TreeEntry, 0, len(tree.Entries))
	for _, entry := range tree.Entries {
		fullPath := entry.GetPath()
		if prefix != "" {
			fullPath = prefix + "/" + fullPath
		}
		// Copy the entry with the updated path.
		e := *entry
		e.Path = github.Ptr(fullPath)
		entries = append(entries, &e)

		if entry.GetType() == "tree" {
			subtree, err := g.getGitTreeRecursive(ctx, owner, repo, entry.GetSHA(), fullPath)
			if err != nil {
				return nil, err
			}
			entries = append(entries, subtree.Entries...)
		}
	}

	tree.Entries = entries
	return tree, nil
}
