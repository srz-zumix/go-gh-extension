package gh

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gitutil"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// DanglingCommit represents a commit that is not reachable from normal branch/tag refs.
// These typically originate from squash or rebase merged pull requests where the
// original PR commits are not ancestors of the resulting merge commit, or commits
// dropped by force-pushes on the PR head branch.
type DanglingCommit struct {
	SHA           string
	Message       string
	PRNumber      int
	PRURL         string
	TotalBlobSize int
}

// DanglingBlob represents a blob that is referenced only by dangling commits.
// These blobs may exist in GitHub's object store but are not part of any current
// branch or tag tree.
type DanglingBlob struct {
	SHA       string
	Path      string
	Size      int
	CommitSHA string
	PRNumber  int
	PRURL     string
}

// ReachabilityCheckMode specifies the method used to verify that a candidate commit
// is truly not reachable from any branch ref before reporting it as dangling.
// Zero value (empty string) disables the check.
type ReachabilityCheckMode string

const (
	// ReachabilityCheckNone skips reachability verification (default). Candidates
	// from PR history are reported as dangling without additional API/git checks.
	ReachabilityCheckNone ReachabilityCheckMode = "none"
	// ReachabilityCheckDefaultBranch uses the GitHub Compare API to confirm the
	// commit is not reachable from the repository's default branch.
	ReachabilityCheckDefaultBranch ReachabilityCheckMode = "default-branch"
	// ReachabilityCheckAllBranches uses the GitHub Compare API to confirm the
	// commit is not reachable from any branch. More accurate but requires one API
	// call per branch per candidate commit.
	ReachabilityCheckAllBranches ReachabilityCheckMode = "all-branches"
	// ReachabilityCheckLocalObject checks whether the commit object exists in the
	// local git repository after git fetch --all. Fastest check; a missing object
	// means the commit is not reachable from any remote branch. Stale loose objects
	// from previous fetches can cause false negatives.
	ReachabilityCheckLocalObject ReachabilityCheckMode = "local-object"
	// ReachabilityCheckLocalDefault uses git merge-base --is-ancestor to confirm
	// the commit is not an ancestor of the local default remote-tracking branch.
	// Requires git fetch --all to have been run first.
	ReachabilityCheckLocalDefault ReachabilityCheckMode = "local-default"
	// ReachabilityCheckLocalAnyBranch uses git branch -r --contains to confirm the
	// commit is not reachable from any remote-tracking branch.
	// Requires git fetch --all to have been run first.
	ReachabilityCheckLocalAnyBranch ReachabilityCheckMode = "local-any"
)

// ReachabilityCheckModeValues is the ordered list of valid ReachabilityCheckMode
// string values, suitable for use with flag enum helpers.
var ReachabilityCheckModeValues = []string{
	string(ReachabilityCheckNone),
	string(ReachabilityCheckDefaultBranch),
	string(ReachabilityCheckAllBranches),
	string(ReachabilityCheckLocalObject),
	string(ReachabilityCheckLocalDefault),
	string(ReachabilityCheckLocalAnyBranch),
}

// DanglingOptions controls which detection methods are active when searching for
// dangling commits and blobs. Zero value enables all detection methods and skips
// reachability verification.
type DanglingOptions struct {
	DisableSquashRebase bool // if true, skip squash/rebase merged PR commit detection
	DisableForcePush    bool // if true, skip force-push dropped commit detection
	DisableClosed       bool // if true, skip closed unmerged PR detection
	// ReachabilityCheck specifies an optional secondary verification step that
	// confirms each candidate commit is truly not reachable from any branch before
	// it is included in results. Zero value skips the check.
	ReachabilityCheck ReachabilityCheckMode
	// LocalDefaultBranch is the local remote-tracking ref used for
	// ReachabilityCheckLocalDefault (e.g. "origin/main"). Auto-detected via
	// git symbolic-ref when empty.
	LocalDefaultBranch string
}

// isCommitDanglingByReachability returns true if sha should be included in dangling
// results according to the configured ReachabilityCheck.
// Returns (true, nil) when no check is configured, meaning the commit is assumed dangling.
func isCommitDanglingByReachability(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string, opts DanglingOptions) (bool, error) {
	switch opts.ReachabilityCheck {
	case ReachabilityCheckNone, "":
		return true, nil
	case ReachabilityCheckDefaultBranch:
		reachable, err := IsCommitReachableFromDefaultBranch(ctx, g, repo, sha)
		if err != nil {
			return false, err
		}
		return !reachable, nil
	case ReachabilityCheckAllBranches:
		reachable, err := IsCommitReachableFromAnyBranch(ctx, g, repo, sha)
		if err != nil {
			return false, err
		}
		return !reachable, nil
	case ReachabilityCheckLocalObject:
		exists, err := gitutil.IsCommitObjectExists(ctx, sha)
		if err != nil {
			return false, err
		}
		return !exists, nil
	case ReachabilityCheckLocalDefault:
		ref := opts.LocalDefaultBranch
		if ref == "" {
			ref = gitutil.ResolveDefaultBranch(ctx)
		}
		reachable, err := gitutil.IsCommitReachableFromDefaultBranch(ctx, sha, ref)
		if err != nil {
			return false, err
		}
		return !reachable, nil
	case ReachabilityCheckLocalAnyBranch:
		reachable, err := gitutil.IsCommitReachableFromAnyBranch(ctx, sha)
		if err != nil {
			return false, err
		}
		return !reachable, nil
	default:
		return false, fmt.Errorf("unknown reachability check mode %q", opts.ReachabilityCheck)
	}
}

// isSquashOrRebaseMerge returns true when the merge commit does NOT have the PR head
// SHA as a direct parent. This indicates a squash or rebase merge strategy, which
// leaves the original PR commits unreachable from normal branch refs.
func isSquashOrRebaseMerge(mergeCommit *github.RepositoryCommit, prHeadSHA string) bool {
	for _, parent := range mergeCommit.Parents {
		if parent.GetSHA() == prHeadSHA {
			return false
		}
	}
	return true
}

// appendUniqueCommitsBySHA appends commits whose SHA was not seen yet.
func appendUniqueCommitsBySHA(dst []*github.RepositoryCommit, seen map[string]bool, src []*github.RepositoryCommit) []*github.RepositoryCommit {
	for _, c := range src {
		sha := c.GetSHA()
		if sha == "" || seen[sha] {
			continue
		}
		seen[sha] = true
		dst = append(dst, c)
	}
	return dst
}

// listForcePushedOutPRCommits returns commits that became unreachable from a PR
// head branch due to head_ref_force_pushed timeline events.
func listForcePushedOutPRCommits(ctx context.Context, g *GitHubClient, repo repository.Repository, prNumber int) ([]*github.RepositoryCommit, error) {
	events, err := g.ListPullRequestHeadRefForcePushEvents(ctx, repo.Owner, repo.Name, prNumber)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var result []*github.RepositoryCommit
	for _, e := range events {
		if e.BeforeSHA == "" || e.AfterSHA == "" {
			continue
		}
		logger.Debug("processing force-push event", "pr", prNumber, "before", e.BeforeSHA, "after", e.AfterSHA)

		// Compare after...before to enumerate commits that existed before the force-push
		// and are no longer reachable from the updated head.
		comp, err := g.CompareCommits(ctx, repo.Owner, repo.Name, e.AfterSHA, e.BeforeSHA)
		if err != nil {
			logger.Debug("skipping force-push event: failed to compare commits", "pr", prNumber, "before", e.BeforeSHA, "after", e.AfterSHA, "error", err)
			continue
		}

		result = appendUniqueCommitsBySHA(result, seen, comp.Commits)
	}

	return result, nil
}

// listDanglingCommitCandidatesForPR collects commit candidates that may be
// dangling for a merged PR. It includes commits from squash/rebase merges and
// commits dropped by force-push on the PR head branch.
func listDanglingCommitCandidatesForPR(ctx context.Context, g *GitHubClient, repo repository.Repository, pr *github.PullRequest, opts DanglingOptions) ([]*github.RepositoryCommit, error) {
	seen := make(map[string]bool)
	var result []*github.RepositoryCommit

	if !opts.DisableSquashRebase {
		mergeCommitSHA := pr.GetMergeCommitSHA()
		head := pr.GetHead()
		includePRCommits := false

		if mergeCommitSHA == "" {
			logger.Debug("skipping squash/rebase check: no merge commit SHA", "pr", pr.GetNumber())
		} else if head == nil || head.GetSHA() == "" {
			logger.Debug("skipping squash/rebase check: no head SHA", "pr", pr.GetNumber())
		} else {
			mergeCommit, err := g.GetCommit(ctx, repo.Owner, repo.Name, mergeCommitSHA)
			if err != nil {
				logger.Debug("skipping squash/rebase check: failed to get merge commit", "pr", pr.GetNumber(), "sha", mergeCommitSHA, "error", err)
			} else if isSquashOrRebaseMerge(mergeCommit, head.GetSHA()) {
				includePRCommits = true
				logger.Debug("found squash/rebase merged PR", "pr", pr.GetNumber())
			} else {
				logger.Debug("PR is regular merge (not squash/rebase)", "pr", pr.GetNumber())
			}
		}

		if includePRCommits {
			prCommits, err := g.ListPullRequestCommits(ctx, repo.Owner, repo.Name, pr.GetNumber())
			if err != nil {
				return nil, fmt.Errorf("failed to list commits for PR #%d: %w", pr.GetNumber(), err)
			}
			result = appendUniqueCommitsBySHA(result, seen, prCommits)
		}
	}

	if !opts.DisableForcePush {
		forcePushedOut, err := listForcePushedOutPRCommits(ctx, g, repo, pr.GetNumber())
		if err != nil {
			logger.Debug("skipping force-push based commit collection", "pr", pr.GetNumber(), "error", err)
		} else {
			result = appendUniqueCommitsBySHA(result, seen, forcePushedOut)
		}
	}

	return result, nil
}

// listDanglingCommitCandidatesForClosedUnmergedPR collects commit candidates for
// a closed, unmerged PR. Commits are only dangling if the PR head branch no longer
// exists; if the branch is still present its commits are reachable from that ref
// and are not dangling. Commits dropped by force-push on the PR head branch are
// also included regardless of branch existence.
func listDanglingCommitCandidatesForClosedUnmergedPR(ctx context.Context, g *GitHubClient, repo repository.Repository, pr *github.PullRequest, opts DanglingOptions) ([]*github.RepositoryCommit, error) {
	seen := make(map[string]bool)
	var result []*github.RepositoryCommit

	// Check whether the head branch still exists. If it does, the commits are
	// reachable from that branch ref and are not dangling.
	headRef := pr.GetHead().GetRef()
	if headRef != "" {
		_, err := g.GetBranch(ctx, repo.Owner, repo.Name, headRef)
		if err == nil {
			logger.Debug("skipping closed PR: head branch still exists", "pr", pr.GetNumber(), "branch", headRef)
			// Force-pushed-out commits are still dangling even if the branch exists.
			if !opts.DisableForcePush {
				forcePushedOut, err := listForcePushedOutPRCommits(ctx, g, repo, pr.GetNumber())
				if err != nil {
					logger.Debug("skipping force-push based commit collection", "pr", pr.GetNumber(), "error", err)
				} else {
					result = appendUniqueCommitsBySHA(result, seen, forcePushedOut)
				}
			}
			return result, nil
		}
		logger.Debug("head branch not found, treating PR commits as dangling", "pr", pr.GetNumber(), "branch", headRef)
	}

	// Branch is gone: all PR commits are dangling candidates.
	prCommits, err := g.ListPullRequestCommits(ctx, repo.Owner, repo.Name, pr.GetNumber())
	if err != nil {
		return nil, fmt.Errorf("failed to list commits for PR #%d: %w", pr.GetNumber(), err)
	}
	result = appendUniqueCommitsBySHA(result, seen, prCommits)

	if !opts.DisableForcePush {
		// Also collect commits dropped by force-pushes during the PR lifetime.
		forcePushedOut, err := listForcePushedOutPRCommits(ctx, g, repo, pr.GetNumber())
		if err != nil {
			logger.Debug("skipping force-push based commit collection", "pr", pr.GetNumber(), "error", err)
		} else {
			result = appendUniqueCommitsBySHA(result, seen, forcePushedOut)
		}
	}

	return result, nil
}

// computeCommitTotalBlobSize sums blob sizes for a commit by traversing its tree recursively.
func computeCommitTotalBlobSize(ctx context.Context, g *GitHubClient, repo repository.Repository, commitSHA string) int {
	gitCommit, err := g.GetGitCommit(ctx, repo.Owner, repo.Name, commitSHA)
	if err != nil {
		logger.Debug("failed to get git commit for size calculation", "sha", commitSHA, "error", err)
		return 0
	}

	treeSHA := gitCommit.GetTree().GetSHA()
	if treeSHA == "" {
		return 0
	}

	tree, err := g.GetGitTree(ctx, repo.Owner, repo.Name, treeSHA, true)
	if err != nil {
		return 0
	}

	totalBlobSize := 0
	for _, entry := range tree.Entries {
		if entry.GetType() == "blob" {
			totalBlobSize += entry.GetSize()
		}
	}
	return totalBlobSize
}

// ListClosedPRs returns all closed pull requests for the repository, ordered by
// most recently updated. maxPRs limits the number of results; use -1 for unlimited.
func ListClosedPRs(ctx context.Context, g *GitHubClient, repo repository.Repository, maxPRs int) ([]*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State:     "closed",
		Sort:      "updated",
		Direction: "desc",
	}
	prs, err := g.ListPullRequests(ctx, repo.Owner, repo.Name, opts, maxPRs)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return prs, nil
}

// FindDanglingCommits finds commits that are not reachable from any normal branch
// or tag ref. Detection methods are controlled by opts:
//   - Squash/rebase merged PR commits (disabled by opts.DisableSquashRebase)
//   - Commits dropped by force-push on a PR head branch (disabled by opts.DisableForcePush)
//   - All commits from closed unmerged PRs (disabled by opts.DisableClosed)
//
// Note: GitHub retains refs/pull/{number}/head for all PRs, so these commits remain
// accessible via PR refs even after the source branch is deleted.
func FindDanglingCommits(ctx context.Context, g *GitHubClient, repo repository.Repository, prs []*github.PullRequest, opts DanglingOptions) ([]*DanglingCommit, error) {
	logger.Info("inspecting PRs for dangling commits", "total", len(prs))

	var result []*DanglingCommit

	for i, pr := range prs {
		logger.Debug("checking PR", "progress", fmt.Sprintf("%d/%d", i+1, len(prs)), "pr", pr.GetNumber(), "title", pr.GetTitle())

		var prCommits []*github.RepositoryCommit
		var err error

		if pr.MergedAt != nil {
			if opts.DisableSquashRebase && opts.DisableForcePush {
				logger.Debug("skipping merged PR: all merged-PR methods disabled", "pr", pr.GetNumber())
				continue
			}
			prCommits, err = listDanglingCommitCandidatesForPR(ctx, g, repo, pr, opts)
		} else {
			if opts.DisableClosed {
				logger.Debug("skipping closed unmerged PR: closed detection disabled", "pr", pr.GetNumber())
				continue
			}
			prCommits, err = listDanglingCommitCandidatesForClosedUnmergedPR(ctx, g, repo, pr, opts)
		}
		if err != nil {
			return nil, err
		}
		if len(prCommits) == 0 {
			continue
		}

		for _, c := range prCommits {
			sha := c.GetSHA()
			logger.Debug("processing commit", "pr", pr.GetNumber(), "sha", sha)

			dangling, err := isCommitDanglingByReachability(ctx, g, repo, sha, opts)
			if err != nil {
				return nil, fmt.Errorf("check commit reachability for %s: %w", sha, err)
			}
			if !dangling {
				logger.Debug("skipping commit: reachable from branch", "sha", sha)
				continue
			}

			message := ""
			if c.GetCommit() != nil {
				message = c.GetCommit().GetMessage()
			}
			totalBlobSize := computeCommitTotalBlobSize(ctx, g, repo, sha)

			result = append(result, &DanglingCommit{
				SHA:           sha,
				Message:       message,
				PRNumber:      pr.GetNumber(),
				PRURL:         pr.GetHTMLURL(),
				TotalBlobSize: totalBlobSize,
			})
		}
	}

	logger.Info("dangling commit search complete", "found", len(result))
	return result, nil
}

// FindDanglingBlobs finds blobs that are referenced only by dangling commits.
// Detection methods are controlled by opts (same as FindDanglingCommits).
// For each dangling commit the full git tree is traversed recursively.
// Blobs are deduplicated by SHA within each PR.
func FindDanglingBlobs(ctx context.Context, g *GitHubClient, repo repository.Repository, prs []*github.PullRequest, opts DanglingOptions) ([]*DanglingBlob, error) {
	logger.Info("inspecting PRs for dangling blobs", "total", len(prs))

	var result []*DanglingBlob

	for i, pr := range prs {
		logger.Debug("checking PR", "progress", fmt.Sprintf("%d/%d", i+1, len(prs)), "pr", pr.GetNumber(), "title", pr.GetTitle())

		var prCommits []*github.RepositoryCommit
		var err error

		if pr.MergedAt != nil {
			if opts.DisableSquashRebase && opts.DisableForcePush {
				logger.Debug("skipping merged PR: all merged-PR methods disabled", "pr", pr.GetNumber())
				continue
			}
			prCommits, err = listDanglingCommitCandidatesForPR(ctx, g, repo, pr, opts)
		} else {
			if opts.DisableClosed {
				logger.Debug("skipping closed unmerged PR: closed detection disabled", "pr", pr.GetNumber())
				continue
			}
			prCommits, err = listDanglingCommitCandidatesForClosedUnmergedPR(ctx, g, repo, pr, opts)
		}
		if err != nil {
			return nil, err
		}
		if len(prCommits) == 0 {
			continue
		}

		// Deduplicate blob SHAs within this PR to avoid redundant output.
		seen := make(map[string]bool)

		for _, c := range prCommits {
			commitSHA := c.GetSHA()
			logger.Debug("processing commit", "pr", pr.GetNumber(), "sha", commitSHA)

			dangling, err := isCommitDanglingByReachability(ctx, g, repo, commitSHA, opts)
			if err != nil {
				logger.Debug("skipping commit: reachability check failed", "sha", commitSHA, "error", err)
				continue
			}
			if !dangling {
				logger.Debug("skipping commit: reachable from branch", "sha", commitSHA)
				continue
			}

			gitCommit, err := g.GetGitCommit(ctx, repo.Owner, repo.Name, commitSHA)
			if err != nil {
				logger.Debug("skipping commit: failed to get git commit", "sha", commitSHA, "error", err)
				continue
			}
			treeSHA := gitCommit.GetTree().GetSHA()
			if treeSHA == "" {
				logger.Debug("skipping commit: empty tree SHA", "sha", commitSHA)
				continue
			}

			tree, err := g.GetGitTree(ctx, repo.Owner, repo.Name, treeSHA, true)
			if err != nil {
				logger.Debug("skipping commit: failed to get git tree", "sha", commitSHA, "tree", treeSHA, "error", err)
				continue
			}

			for _, entry := range tree.Entries {
				if entry.GetType() != "blob" {
					continue
				}
				blobSHA := entry.GetSHA()
				if seen[blobSHA] {
					continue
				}
				seen[blobSHA] = true
				result = append(result, &DanglingBlob{
					SHA:       blobSHA,
					Path:      entry.GetPath(),
					Size:      entry.GetSize(),
					CommitSHA: commitSHA,
					PRNumber:  pr.GetNumber(),
					PRURL:     pr.GetHTMLURL(),
				})
			}
		}
	}

	logger.Info("dangling blob search complete", "found", len(result))
	return result, nil
}

// SortBlobsBy sorts blobs in-place by the given field name (case-insensitive).
// Supported fields: "size", "path", "pr_number".
// desc=true reverses the order.
// Returns an error for unknown field names.
func SortBlobsBy(blobs []*DanglingBlob, field string, desc bool) error {
	var less func(a, b *DanglingBlob) int
	switch strings.ToLower(field) {
	case "size":
		less = func(a, b *DanglingBlob) int { return cmp.Compare(a.Size, b.Size) }
	case "path":
		less = func(a, b *DanglingBlob) int { return cmp.Compare(a.Path, b.Path) }
	case "pr_number":
		less = func(a, b *DanglingBlob) int { return cmp.Compare(a.PRNumber, b.PRNumber) }
	default:
		return fmt.Errorf("unknown sort field %q: valid values are size, path, pr_number", field)
	}
	slices.SortStableFunc(blobs, func(a, b *DanglingBlob) int {
		if desc {
			return less(b, a)
		}
		return less(a, b)
	})
	return nil
}

// SortCommitsBy sorts commits in-place by the given field name (case-insensitive).
// Supported fields: "size", "pr_number".
// desc=true reverses the order.
// Returns an error for unknown field names.
func SortCommitsBy(commits []*DanglingCommit, field string, desc bool) error {
	var less func(a, b *DanglingCommit) int
	switch strings.ToLower(field) {
	case "size":
		less = func(a, b *DanglingCommit) int { return cmp.Compare(a.TotalBlobSize, b.TotalBlobSize) }
	case "pr_number":
		less = func(a, b *DanglingCommit) int { return cmp.Compare(a.PRNumber, b.PRNumber) }
	default:
		return fmt.Errorf("unknown sort field %q: valid values are size, pr_number", field)
	}
	slices.SortStableFunc(commits, func(a, b *DanglingCommit) int {
		if desc {
			return less(b, a)
		}
		return less(a, b)
	})
	return nil
}

// FindLocalDanglingCommitsOnRemote checks which of the given commit SHAs exist on
// the remote GitHub repository. SHAs that return a successful API response are
// returned as DanglingCommit entries with their commit message populated.
// SHAs that are not found on the remote (404) or any other API error are silently
// skipped (logged at debug level).
func FindLocalDanglingCommitsOnRemote(ctx context.Context, g *GitHubClient, repo repository.Repository, shas []string) ([]*DanglingCommit, error) {
	logger.Info("checking local dangling commits against remote", "total", len(shas))
	var result []*DanglingCommit
	for _, sha := range shas {
		commit, err := g.GetCommit(ctx, repo.Owner, repo.Name, sha)
		if err != nil {
			logger.Debug("commit not found on remote", "sha", sha, "error", err)
			continue
		}
		message := ""
		if commit.GetCommit() != nil {
			message = commit.GetCommit().GetMessage()
		}
		result = append(result, &DanglingCommit{
			SHA:     sha,
			Message: message,
		})
	}
	logger.Info("local dangling commit check complete", "found", len(result))
	return result, nil
}
