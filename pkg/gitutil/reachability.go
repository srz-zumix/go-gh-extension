package gitutil

import (
	"context"
	"errors"
	"strings"

	"github.com/cli/cli/v2/git"
)

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

// IsCommitReachableFromAnyTag reports whether sha is reachable from any tag
// using `git tag --contains`.
// The caller must ensure all remote refs have been fetched (e.g. git fetch --all).
func IsCommitReachableFromAnyTag(ctx context.Context, sha string) (bool, error) {
	c := NewClient()
	cmd, err := c.Command(ctx, "tag", "--contains", sha)
	if err != nil {
		return false, err
	}
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// IsCommitReachableFromAnyRef reports whether sha is reachable from any
// remote-tracking branch or any tag.
// This is a combination of IsCommitReachableFromAnyBranch and IsCommitReachableFromAnyTag.
// The caller must ensure all remote refs have been fetched (e.g. git fetch --all).
func IsCommitReachableFromAnyRef(ctx context.Context, sha string) (bool, error) {
	reachable, err := IsCommitReachableFromAnyBranch(ctx, sha)
	if err != nil {
		return false, err
	}
	if reachable {
		return true, nil
	}
	return IsCommitReachableFromAnyTag(ctx, sha)
}
