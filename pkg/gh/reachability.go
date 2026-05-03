package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
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
			// Skip branches that cannot be compared (e.g. empty repos).
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
func isCommitReachableFromRef(ctx context.Context, g *GitHubClient, repo repository.Repository, sha, ref string) (bool, error) {
	comp, err := g.CompareCommits(ctx, repo.Owner, repo.Name, ref, sha)
	if err != nil {
		return false, err
	}
	return comp.GetAheadBy() == 0, nil
}

