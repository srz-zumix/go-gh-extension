package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v93/github"
)

// GetGitCommit retrieves a Git commit object.
func GetGitCommit(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) (*github.Commit, error) {
	return g.GetGitCommit(ctx, repo.Owner, repo.Name, sha)
}

// GetGitTree retrieves a Git tree object.
func GetGitTree(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string, recursive bool) (*github.Tree, error) {
	return g.GetGitTree(ctx, repo.Owner, repo.Name, sha, recursive)
}

// GetGitTreeRecursive retrieves all entries in a Git tree.
func GetGitTreeRecursive(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) (*github.Tree, error) {
	return g.GetGitTreeRecursive(ctx, repo.Owner, repo.Name, sha)
}

// GetGitBlob retrieves a Git blob object.
func GetGitBlob(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) (*github.Blob, error) {
	return g.GetGitBlob(ctx, repo.Owner, repo.Name, sha)
}

// CreateGitTree creates a Git tree from the specified entries.
func CreateGitTree(ctx context.Context, g *GitHubClient, repo repository.Repository, baseTree string, entries []*github.TreeEntry) (*github.Tree, error) {
	return g.CreateGitTree(ctx, repo.Owner, repo.Name, baseTree, entries)
}

// CreateGitCommit creates a Git commit.
func CreateGitCommit(ctx context.Context, g *GitHubClient, repo repository.Repository, commit github.Commit) (*github.Commit, error) {
	return g.CreateGitCommit(ctx, repo.Owner, repo.Name, commit)
}
