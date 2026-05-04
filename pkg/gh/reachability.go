package gh

import (
	"context"
	"errors"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// IsCommitReachableFromDefaultBranch reports whether sha is an ancestor of (or equal
// to) the repository's default branch, using the GitHub Compare API.
// Returns (true, nil) if reachable, (false, nil) if not reachable.
func IsCommitReachableFromDefaultBranch(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) (bool, error) {
	repoInfo, err := g.GetRepository(ctx, repo.Owner, repo.Name)
	if err != nil {
		return false, err
	}
	defaultBranch := repoInfo.GetDefaultBranch()
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	return isCommitReachableFromRef(ctx, g, repo, sha, defaultBranch)
}

// IsCommitReachableFromAnyBranch reports whether sha is an ancestor of any branch
// in the repository, using the GitHub Compare API. Short-circuits on first match.
func IsCommitReachableFromAnyBranch(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) (bool, error) {
	branches, err := g.ListBranches(ctx, repo.Owner, repo.Name, nil)
	if err != nil {
		return false, err
	}
	for _, b := range branches {
		reachable, err := isCommitReachableFromRef(ctx, g, repo, sha, b.GetName())
		if err != nil {
			if !isSkippableCompareError(err) {
				return false, err
			}
			// Skip branches that cannot be compared (e.g. empty repos, no common ancestor).
			continue
		}
		if reachable {
			return true, nil
		}
	}
	return false, nil
}

// IsCommitReachableFromAnyRef reports whether sha is an ancestor of any branch or
// tag in the repository, using the GitHub Compare API. Short-circuits on first match.
func IsCommitReachableFromAnyRef(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) (bool, error) {
	reachable, err := IsCommitReachableFromAnyBranch(ctx, g, repo, sha)
	if err != nil {
		return false, err
	}
	if reachable {
		return true, nil
	}
	tags, err := g.ListTags(ctx, repo.Owner, repo.Name)
	if err != nil {
		return false, err
	}
	for _, t := range tags {
		reachable, err := isCommitReachableFromRef(ctx, g, repo, sha, t.GetName())
		if err != nil {
			if !isSkippableCompareError(err) {
				return false, err
			}
			continue
		}
		if reachable {
			return true, nil
		}
	}
	return false, nil
}

// isCommitReachableFromRef reports whether sha is an ancestor of (or equal to) the
// given ref via the GitHub Compare API.
// compare(ref, sha): ahead_by == 0 means sha has no commits that are not in ref,
// i.e. sha is an ancestor of ref (or they are identical).
// Only metadata is fetched (no commit list pagination).
func isCommitReachableFromRef(ctx context.Context, g *GitHubClient, repo repository.Repository, sha, ref string) (bool, error) {
	comp, err := g.CompareCommitsMeta(ctx, repo.Owner, repo.Name, ref, sha)
	if err != nil {
		return false, err
	}
	return comp.GetAheadBy() == 0, nil
}

// isSkippableCompareError reports whether err is an expected, non-transient error
// from the Compare API that indicates the ref and sha simply cannot be compared
// (e.g. ref not found, no common ancestor). Transient errors such as rate-limit
// or server errors must NOT be skipped so that callers surface them correctly.
func isSkippableCompareError(err error) bool {
	// Rate-limit errors must be propagated, not silently skipped.
	if _, ok := errors.AsType[*github.RateLimitError](err); ok {
		return false
	}
	if _, ok := errors.AsType[*github.AbuseRateLimitError](err); ok {
		return false
	}
	// 404 (ref/sha not found) and 422 (no common ancestor, too divergent) are
	// expected when comparing across unrelated histories and can be skipped.
	if errResp, ok := errors.AsType[*github.ErrorResponse](err); ok {
		switch errResp.Response.StatusCode {
		case http.StatusNotFound, http.StatusUnprocessableEntity:
			return true
		}
	}
	return false
}
