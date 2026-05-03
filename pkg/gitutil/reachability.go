package gitutil

import (
	"context"
	"errors"
	"strings"

	"github.com/cli/cli/v2/git"
)

// ResolveDefaultBranch returns the local remote-tracking ref for the default branch
// (e.g. "origin/main") by reading git symbolic-ref refs/remotes/origin/HEAD.
// Falls back to "origin/main" if the symbolic ref is not configured.
// The caller must ensure git fetch --all has been run so that the symbolic ref is set.
func ResolveDefaultBranch(ctx context.Context) string {
	c := NewClient()
	cmd, err := c.Command(ctx, "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	if err == nil {
		out, err := cmd.Output()
		if err == nil {
			if ref := strings.TrimSpace(string(out)); ref != "" {
				return ref
			}
		}
	}
	return "origin/main"
}

// IsCommitObjectExists reports whether the commit object for sha exists in the local
// git object store using `git cat-file -e`.
//
// After `git fetch --all` on a freshly cloned repository this is effectively
// equivalent to "reachable from some remote ref", because only objects reachable
// from remote branches are fetched. In long-lived repositories that have not had
// `git gc --prune` run, unreachable loose objects from previous fetches may still
// be present, so this check can produce false negatives for dangling detection.
// Use IsCommitReachableFromAnyBranch for a precise reachability check.
//
// The caller must ensure all remote refs have been fetched (e.g. git fetch --all).
func IsCommitObjectExists(ctx context.Context, sha string) (bool, error) {
	c := NewClient()
	cmd, err := c.Command(ctx, "cat-file", "-e", sha)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err != nil {
		var ge *git.GitError
		if errors.As(err, &ge) && ge.ExitCode == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsCommitReachableFromDefaultBranch reports whether sha is an ancestor of the
// given defaultBranch remote-tracking ref (e.g. "origin/main") using
// `git merge-base --is-ancestor`.
// The caller must ensure all remote refs have been fetched (e.g. git fetch --all).
func IsCommitReachableFromDefaultBranch(ctx context.Context, sha, defaultBranch string) (bool, error) {
	c := NewClient()
	cmd, err := c.Command(ctx, "merge-base", "--is-ancestor", sha, defaultBranch)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err != nil {
		// Exit code 1 means "not an ancestor".
		var ge *git.GitError
		if errors.As(err, &ge) && ge.ExitCode == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsCommitReachableFromAnyBranch reports whether sha is reachable from any
// remote-tracking branch using `git branch -r --contains`.
// The caller must ensure all remote refs have been fetched (e.g. git fetch --all).
func IsCommitReachableFromAnyBranch(ctx context.Context, sha string) (bool, error) {
	c := NewClient()
	cmd, err := c.Command(ctx, "branch", "-r", "--contains", sha)
	if err != nil {
		return false, err
	}
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}
