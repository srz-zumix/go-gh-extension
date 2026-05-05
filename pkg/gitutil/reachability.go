package gitutil

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cli/cli/v2/git"
)

// IsCommitObjectExists reports whether the commit object for sha exists in the local
// git object store using `git cat-file -e`.
//
// After `git fetch --all --tags` on a freshly cloned repository this is effectively
// equivalent to "reachable from some remote ref", because only objects reachable
// from remote branches and tags are fetched. In long-lived repositories that have not had
// `git gc --prune` run, unreachable loose objects from previous fetches may still
// be present, so this check can produce false negatives for dangling detection.
// Use IsCommitReachableFromAnyRef for a precise reachability check across branches and tags.
//
// Note: `git fetch --all` does not fetch all tags by default; tags reachable only from
// non-fetched commits may be missing. Use `git fetch --all --tags` to ensure all tags are
// present before calling this function.
// The caller must ensure all remote refs have been fetched (e.g. git fetch --all --tags).
func IsCommitObjectExists(ctx context.Context, sha string) (bool, error) {
	c := NewClient()
	cmd, err := c.Command(ctx, "cat-file", "-e", sha)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err != nil {
		if _, ok := errors.AsType[*git.GitError](err); ok {
			// git cat-file -e exits non-zero (1 or 128) whenever the object is
			// absent; any *git.GitError here means "not found".
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsCommitReachableFromAnyBranch reports whether sha is reachable from any
// remote-tracking branch using `git branch -r --contains`.
// The caller must ensure all remote refs have been fetched (e.g. git fetch --all --tags).
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
// Note: `git fetch --all` does not fetch all tags by default. Run `git fetch --all --tags`
// to ensure all remote tags are present locally.
// The caller must ensure all remote tags have been fetched (e.g. git fetch --all --tags).
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
// Note: `git fetch --all` does not fetch all tags by default; tags reachable only from
// commits not on any branch may be missing. Run `git fetch --all --tags` to ensure
// all remote branches and tags are present locally.
// The caller must ensure all remote refs and tags have been fetched (e.g. git fetch --all --tags).
//
// If the commit object is not present in the local object store (pruned or
// never fetched), the function returns (false, nil) without error, because
// git branch/tag --contains exits non-zero for unknown objects and there is
// nothing to check reachability against.
func IsCommitReachableFromAnyRef(ctx context.Context, sha string) (bool, error) {
	// If the object doesn't exist locally it cannot be reachable from any local
	// ref, and git branch/tag --contains would exit non-zero for this SHA.
	exists, err := IsCommitObjectExists(ctx, sha)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	reachable, err := IsCommitReachableFromAnyBranch(ctx, sha)
	if err != nil {
		return false, err
	}
	if reachable {
		return true, nil
	}
	return IsCommitReachableFromAnyTag(ctx, sha)
}

// IsBlobReachableFromAnyRef reports whether the given blob SHA is referenced by
// any commit reachable from any local ref (branches, tags, remotes).
//
// Git blobs are content-addressed: the same SHA can appear in multiple commits
// across different branches. This function returns true if ANY commit reachable
// from any local ref references the blob, meaning git will not garbage-collect it.
//
// Implementation uses git log --all --find-object which traverses local ref
// history. Requires git fetch --all --tags to have been run first to include
// all remote-tracking refs and tags. Note: git fetch --all alone does not fetch
// tags that are not reachable from any fetched branch.
func IsBlobReachableFromAnyRef(ctx context.Context, sha string) (bool, error) {
	c := NewClient()
	// --find-object=<sha> searches commits that reference the object anywhere
	// in their diff (added or removed). -1 stops after the first match.
	cmd, err := c.Command(ctx, "log", "--all", "--find-object="+sha, "--format=%H", "-1")
	if err != nil {
		return false, err
	}
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git log --find-object: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// IsCommitObjectExistsInDir is like IsCommitObjectExists but operates in the given git directory.
func IsCommitObjectExistsInDir(ctx context.Context, dir string, sha string) (bool, error) {
	c := NewClientWithDir(dir)
	cmd, err := c.Command(ctx, "cat-file", "-e", sha)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err != nil {
		var ge *git.GitError
		if errors.As(err, &ge) {
			// git cat-file -e exits non-zero (1 or 128) whenever the object is
			// absent; any *git.GitError here means "not found".
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsCommitReachableFromAnyRefInDir is like IsCommitReachableFromAnyRef but operates in the given git directory.
func IsCommitReachableFromAnyRefInDir(ctx context.Context, dir string, sha string) (bool, error) {
	exists, err := IsCommitObjectExistsInDir(ctx, dir, sha)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	c := NewClientWithDir(dir)
	cmd, err := c.Command(ctx, "branch", "-r", "--contains", sha)
	if err != nil {
		return false, err
	}
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(string(out)) != "" {
		return true, nil
	}

	cmd, err = c.Command(ctx, "tag", "--contains", sha)
	if err != nil {
		return false, err
	}
	out, err = cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// IsBlobReachableFromAnyRefInDir is like IsBlobReachableFromAnyRef but operates in the given git directory.
func IsBlobReachableFromAnyRefInDir(ctx context.Context, dir string, sha string) (bool, error) {
	c := NewClientWithDir(dir)
	cmd, err := c.Command(ctx, "log", "--all", "--find-object="+sha, "--format=%H", "-1")
	if err != nil {
		return false, err
	}
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git log --find-object: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}
