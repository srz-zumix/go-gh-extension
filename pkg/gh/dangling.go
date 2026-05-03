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
	// ReachabilityCheckBranches uses the GitHub Compare API to confirm the commit
	// is not reachable from any branch. Requires one API call per branch per
	// candidate commit.
	ReachabilityCheckBranches ReachabilityCheckMode = "branches"
	// ReachabilityCheckRefs uses the GitHub Compare API to confirm the commit is
	// not reachable from any branch or tag. More thorough than branches but
	// requires additional API calls for tags.
	ReachabilityCheckRefs ReachabilityCheckMode = "refs"
	// ReachabilityCheckLocalObject checks whether the commit object exists in the
	// local git repository after git fetch --all. Fastest check; a missing object
	// means the commit is not reachable from any remote ref. Stale loose objects
	// from previous fetches can cause false negatives.
	ReachabilityCheckLocalObject ReachabilityCheckMode = "local-object"
	// ReachabilityCheckLocalRefs uses git branch -r --contains and git tag
	// --contains to confirm the commit is not reachable from any remote-tracking
	// branch or any tag.
	// Requires git fetch --all to have been run first.
	ReachabilityCheckLocalRefs ReachabilityCheckMode = "local-refs"
)

// ReachabilityCheckModeValues is the ordered list of valid ReachabilityCheckMode
// string values, suitable for use with flag enum helpers.
var ReachabilityCheckModeValues = []string{
	string(ReachabilityCheckNone),
	string(ReachabilityCheckDefaultBranch),
	string(ReachabilityCheckBranches),
	string(ReachabilityCheckRefs),
	string(ReachabilityCheckLocalObject),
	string(ReachabilityCheckLocalRefs),
}

// DanglingOptions controls which detection methods are active when searching for
// dangling commits and blobs. Zero value enables all detection methods and skips
// reachability verification.
type DanglingOptions struct {
	DisableSquashRebase bool // if true, skip squash/rebase merged PR commit detection
	DisableForcePush    bool // if true, skip force-push dropped commit detection
	DisableClosed       bool // if true, skip closed unmerged PR detection
	// ReachabilityCheck specifies an optional secondary verification step that
	// confirms each candidate commit is truly not reachable from any branch or tag
	// before it is included in results. Zero value skips the check.
	ReachabilityCheck ReachabilityCheckMode
}

// parentUnreachable reports whether any of the given parent commits has a SHA
// already recorded in unreachableSHAs. Because reachability from any ref is
// closed under ancestry, a commit whose parent is unreachable is also unreachable,
// so no further API or git call is needed.
func parentUnreachable(parents []*github.Commit, unreachableSHAs map[string]bool) bool {
	for _, p := range parents {
		if unreachableSHAs[p.GetSHA()] {
			return true
		}
	}
	return false
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
	case ReachabilityCheckBranches:
		reachable, err := IsCommitReachableFromAnyBranch(ctx, g, repo, sha)
		if err != nil {
			return false, err
		}
		return !reachable, nil
	case ReachabilityCheckRefs:
		reachable, err := IsCommitReachableFromAnyRef(ctx, g, repo, sha)
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
	case ReachabilityCheckLocalRefs:
		reachable, err := gitutil.IsCommitReachableFromAnyRef(ctx, sha)
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

// listSquashRebaseChainCandidates returns the PR commit chain (oldest first)
// when the PR was merged via squash or rebase. Returns nil for regular merges or
// when required metadata is missing. The returned slice is the linear ancestry
// chain leading to the PR head, suitable for the head-first reachability shortcut.
func listSquashRebaseChainCandidates(ctx context.Context, g *GitHubClient, repo repository.Repository, pr *github.PullRequest) ([]*github.RepositoryCommit, error) {
	mergeCommitSHA := pr.GetMergeCommitSHA()
	head := pr.GetHead()
	if mergeCommitSHA == "" {
		logger.Debug("skipping squash/rebase check: no merge commit SHA", "pr", pr.GetNumber())
		return nil, nil
	}
	if head == nil || head.GetSHA() == "" {
		logger.Debug("skipping squash/rebase check: no head SHA", "pr", pr.GetNumber())
		return nil, nil
	}
	mergeCommit, err := g.GetCommit(ctx, repo.Owner, repo.Name, mergeCommitSHA)
	if err != nil {
		logger.Debug("skipping squash/rebase check: failed to get merge commit", "pr", pr.GetNumber(), "sha", mergeCommitSHA, "error", err)
		return nil, nil
	}
	if !isSquashOrRebaseMerge(mergeCommit, head.GetSHA()) {
		logger.Debug("PR is regular merge (not squash/rebase)", "pr", pr.GetNumber())
		return nil, nil
	}
	logger.Debug("found squash/rebase merged PR", "pr", pr.GetNumber())
	prCommits, err := g.ListPullRequestCommits(ctx, repo.Owner, repo.Name, pr.GetNumber())
	if err != nil {
		return nil, fmt.Errorf("failed to list commits for PR #%d: %w", pr.GetNumber(), err)
	}
	return prCommits, nil
}

// listClosedUnmergedChainCandidates returns the PR commit chain (oldest first)
// for a closed, unmerged PR whose head branch is gone in the base repository.
// Returns nil for fork PRs (the head branch lives in another repository) and
// when the head branch still exists. Errors other than 404 from the branch
// existence check are propagated to avoid misclassification on transient failures.
func listClosedUnmergedChainCandidates(ctx context.Context, g *GitHubClient, repo repository.Repository, pr *github.PullRequest) ([]*github.RepositoryCommit, error) {
	headRepo := pr.GetHead().GetRepo()
	baseRepo := pr.GetBase().GetRepo()
	if headRepo != nil && baseRepo != nil && headRepo.GetFullName() != baseRepo.GetFullName() {
		logger.Debug("skipping closed unmerged fork PR", "pr", pr.GetNumber(), "head_repo", headRepo.GetFullName())
		return nil, nil
	}

	headRef := pr.GetHead().GetRef()
	if headRef != "" {
		_, err := g.GetBranch(ctx, repo.Owner, repo.Name, headRef)
		if err == nil {
			logger.Debug("skipping closed PR: head branch still exists", "pr", pr.GetNumber(), "branch", headRef)
			return nil, nil
		}
		if !IsHTTPNotFound(err) {
			return nil, fmt.Errorf("failed to check branch %q for PR #%d: %w", headRef, pr.GetNumber(), err)
		}
		logger.Debug("head branch not found, treating PR commits as dangling", "pr", pr.GetNumber(), "branch", headRef)
	}

	prCommits, err := g.ListPullRequestCommits(ctx, repo.Owner, repo.Name, pr.GetNumber())
	if err != nil {
		return nil, fmt.Errorf("failed to list commits for PR #%d: %w", pr.GetNumber(), err)
	}
	return prCommits, nil
}

// listCandidatesForPR returns dangling commit candidates for a PR, separated by
// source. The chain is the linear PR commit ancestry (oldest first) suitable for
// chain-level shortcuts; forcePushed are independent commits dropped by
// force-push events. Either may be nil.
func listCandidatesForPR(ctx context.Context, g *GitHubClient, repo repository.Repository, pr *github.PullRequest, opts DanglingOptions) (chain, forcePushed []*github.RepositoryCommit, err error) {
	if pr.MergedAt != nil {
		if opts.DisableSquashRebase && opts.DisableForcePush {
			logger.Debug("skipping merged PR: all merged-PR methods disabled", "pr", pr.GetNumber())
			return nil, nil, nil
		}
		if !opts.DisableSquashRebase {
			chain, err = listSquashRebaseChainCandidates(ctx, g, repo, pr)
			if err != nil {
				return nil, nil, err
			}
		}
	} else {
		if opts.DisableClosed {
			logger.Debug("skipping closed unmerged PR: closed detection disabled", "pr", pr.GetNumber())
			return nil, nil, nil
		}
		chain, err = listClosedUnmergedChainCandidates(ctx, g, repo, pr)
		if err != nil {
			return nil, nil, err
		}
	}

	if !opts.DisableForcePush {
		fp, fpErr := listForcePushedOutPRCommits(ctx, g, repo, pr.GetNumber())
		if fpErr != nil {
			logger.Debug("skipping force-push based commit collection", "pr", pr.GetNumber(), "error", fpErr)
		} else {
			forcePushed = fp
		}
	}
	return chain, forcePushed, nil
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
		logger.Debug("failed to get git tree for size calculation", "sha", commitSHA, "tree", treeSHA, "error", err)
		return 0
	}

	if !tree.GetTruncated() {
		totalBlobSize := 0
		for _, entry := range tree.Entries {
			if entry.GetType() == "blob" {
				totalBlobSize += entry.GetSize()
			}
		}
		return totalBlobSize
	}

	logger.Debug("git tree response was truncated, falling back to full tree traversal", "sha", commitSHA, "tree", treeSHA)

	totalBlobSize, err := computeTreeTotalBlobSize(ctx, g, repo, treeSHA)
	if err != nil {
		logger.Debug("failed to traverse truncated git tree for size calculation", "sha", commitSHA, "tree", treeSHA, "error", err)
		return 0
	}

	return totalBlobSize
}

// computeTreeTotalBlobSize sums blob sizes by traversing tree objects without using truncated recursive responses.
func computeTreeTotalBlobSize(ctx context.Context, g *GitHubClient, repo repository.Repository, treeSHA string) (int, error) {
	tree, err := g.GetGitTree(ctx, repo.Owner, repo.Name, treeSHA, false)
	if err != nil {
		return 0, err
	}

	totalBlobSize := 0
	for _, entry := range tree.Entries {
		switch entry.GetType() {
		case "blob":
			totalBlobSize += entry.GetSize()
		case "tree":
			childTreeSHA := entry.GetSHA()
			if childTreeSHA == "" {
				continue
			}

			childSize, err := computeTreeTotalBlobSize(ctx, g, repo, childTreeSHA)
			if err != nil {
				return 0, err
			}
			totalBlobSize += childSize
		}
	}

	return totalBlobSize, nil
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

// checkCommitDangling determines whether a single commit should be treated as
// dangling, applying the parent-based shortcut and recording the result in
// unreachableSHAs when applicable. It is the single per-commit reachability
// decision used by both chain and force-push processing.
func checkCommitDangling(ctx context.Context, g *GitHubClient, repo repository.Repository, c *github.RepositoryCommit, opts DanglingOptions, unreachableSHAs map[string]bool) (bool, error) {
	sha := c.GetSHA()
	if unreachableSHAs[sha] {
		return true, nil
	}
	if parentUnreachable(c.Parents, unreachableSHAs) {
		unreachableSHAs[sha] = true
		return true, nil
	}
	dangling, err := isCommitDanglingByReachability(ctx, g, repo, sha, opts)
	if err != nil {
		return false, err
	}
	if dangling {
		unreachableSHAs[sha] = true
	}
	return dangling, nil
}

// processChainCandidates returns the dangling subset of a linear PR commit
// chain (oldest first). Reachability is closed under ancestry:
//   - oldest unreachable → all descendants unreachable via parent shortcut (common case)
//   - oldest reachable → check newest; if newest is also reachable → all chain reachable
//
// Check order: oldest → (if reachable) newest → parent shortcut from oldest.
// For the common "all unreachable" scenario, this costs exactly 1 API call.
func processChainCandidates(ctx context.Context, g *GitHubClient, repo repository.Repository, chain []*github.RepositoryCommit, opts DanglingOptions, unreachableSHAs map[string]bool) ([]*github.RepositoryCommit, error) {
	if len(chain) == 0 {
		return nil, nil
	}

	// Two-endpoint shortcut: only when reachability verification is active and
	// there are at least two commits to make the pre-checks worthwhile.
	if opts.ReachabilityCheck != ReachabilityCheckNone && opts.ReachabilityCheck != "" && len(chain) > 1 {
		oldest := chain[0]
		if !unreachableSHAs[oldest.GetSHA()] && !parentUnreachable(oldest.Parents, unreachableSHAs) {
			oldestDangling, err := isCommitDanglingByReachability(ctx, g, repo, oldest.GetSHA(), opts)
			if err != nil {
				return nil, fmt.Errorf("check chain oldest reachability for %s: %w", oldest.GetSHA(), err)
			}
			if oldestDangling {
				// Oldest unreachable → parent shortcut propagates to every later commit
				// in the chain; no further API/git calls needed for this shortcut.
				unreachableSHAs[oldest.GetSHA()] = true
				logger.Debug("chain oldest unreachable, parent shortcut covers rest", "sha", oldest.GetSHA(), "chain_len", len(chain))
			} else {
				// Oldest reachable → check newest for a full-chain reachable shortcut.
				newest := chain[len(chain)-1]
				if !unreachableSHAs[newest.GetSHA()] && !parentUnreachable(newest.Parents, unreachableSHAs) {
					newestDangling, err := isCommitDanglingByReachability(ctx, g, repo, newest.GetSHA(), opts)
					if err != nil {
						return nil, fmt.Errorf("check chain newest reachability for %s: %w", newest.GetSHA(), err)
					}
					if !newestDangling {
						// Both endpoints reachable → all chain commits reachable; skip.
						logger.Debug("skipping chain: oldest and newest both reachable", "oldest", oldest.GetSHA(), "newest", newest.GetSHA())
						return nil, nil
					}
					unreachableSHAs[newest.GetSHA()] = true
				}
			}
		}
	}

	var result []*github.RepositoryCommit
	for _, c := range chain {
		dangling, err := checkCommitDangling(ctx, g, repo, c, opts, unreachableSHAs)
		if err != nil {
			return nil, fmt.Errorf("check commit reachability for %s: %w", c.GetSHA(), err)
		}
		if dangling {
			result = append(result, c)
		} else {
			logger.Debug("skipping commit: reachable from a ref", "sha", c.GetSHA())
		}
	}
	return result, nil
}

// danglingCommitVisitor is invoked for each PR with the confirmed dangling
// commits found in that PR. The slice combines chain and force-push commits,
// preserving their original collection order.
type danglingCommitVisitor func(pr *github.PullRequest, commits []*github.RepositoryCommit) error

// iterateDanglingCommits walks the PR list, collects candidates, applies all
// reachability shortcuts, and invokes visit for each PR that has at least one
// confirmed dangling commit. It is the shared driver behind FindDanglingCommits
// and FindDanglingBlobs.
func iterateDanglingCommits(ctx context.Context, g *GitHubClient, repo repository.Repository, prs []*github.PullRequest, opts DanglingOptions, visit danglingCommitVisitor) error {
	// unreachableSHAs accumulates commit SHAs confirmed unreachable from any ref.
	// It is shared across PRs because reachability is a property of the commit
	// object itself, not of any individual PR.
	unreachableSHAs := make(map[string]bool)

	for i, pr := range prs {
		logger.Debug("checking PR", "progress", fmt.Sprintf("%d/%d", i+1, len(prs)), "pr", pr.GetNumber(), "title", pr.GetTitle())

		chain, forcePushed, err := listCandidatesForPR(ctx, g, repo, pr, opts)
		if err != nil {
			return err
		}
		if len(chain) == 0 && len(forcePushed) == 0 {
			continue
		}

		chainDangling, err := processChainCandidates(ctx, g, repo, chain, opts, unreachableSHAs)
		if err != nil {
			return err
		}

		var fpDangling []*github.RepositoryCommit
		for _, c := range forcePushed {
			dangling, err := checkCommitDangling(ctx, g, repo, c, opts, unreachableSHAs)
			if err != nil {
				return fmt.Errorf("check commit reachability for %s: %w", c.GetSHA(), err)
			}
			if dangling {
				fpDangling = append(fpDangling, c)
			}
		}

		if len(chainDangling) == 0 && len(fpDangling) == 0 {
			continue
		}
		combined := append(chainDangling, fpDangling...)
		if err := visit(pr, combined); err != nil {
			return err
		}
	}
	return nil
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
	err := iterateDanglingCommits(ctx, g, repo, prs, opts, func(pr *github.PullRequest, commits []*github.RepositoryCommit) error {
		for _, c := range commits {
			sha := c.GetSHA()
			message := ""
			if c.GetCommit() != nil {
				message = c.GetCommit().GetMessage()
			}
			result = append(result, &DanglingCommit{
				SHA:           sha,
				Message:       message,
				PRNumber:      pr.GetNumber(),
				PRURL:         pr.GetHTMLURL(),
				TotalBlobSize: computeCommitTotalBlobSize(ctx, g, repo, sha),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
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
	err := iterateDanglingCommits(ctx, g, repo, prs, opts, func(pr *github.PullRequest, commits []*github.RepositoryCommit) error {
		// Deduplicate blob SHAs within this PR to avoid redundant output.
		seen := make(map[string]bool)
		for _, c := range commits {
			commitSHA := c.GetSHA()
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
		return nil
	})
	if err != nil {
		return nil, err
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
