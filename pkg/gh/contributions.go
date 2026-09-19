package gh

import (
	"context"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// Type aliases for contribution types from the client package.
type ContributionsCollection = client.ContributionsCollection
type RepositoryContributions = client.RepositoryContributions
type ContributionCalendar = client.ContributionCalendar
type ContributionDay = client.ContributionDay

// maxContributionsWindow is the maximum date range GitHub accepts for a single
// contributionsCollection query.
const maxContributionsWindow = 365 * 24 * time.Hour

// GetUserContributions fetches contribution stats for username across [since, until].
// The range is split into at-most-one-year chunks and merged, since GitHub's
// contributionsCollection query rejects a single query spanning more than one year.
//
// When the result spans multiple chunks and any chunk's per-repository breakdown was
// truncated (CommitContributionsByRepositoryTruncated), TotalRepositoriesWithContributedCommits
// is a best-effort lower bound rather than an exact count.
func GetUserContributions(ctx context.Context, g *GitHubClient, username string, since, until time.Time) (*ContributionsCollection, error) {
	var chunks []*ContributionsCollection
	for start := since; start.Before(until); {
		end := start.Add(maxContributionsWindow)
		if end.After(until) {
			end = until
		}
		chunk, err := g.GetUserContributionsCollection(ctx, username, start, end)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
		start = end
	}
	return mergeContributionChunks(chunks), nil
}

// mergeContributionChunks combines per-window contribution results into a single
// collection. Additive totals and calendar days are summed/concatenated, and the
// commit-by-repository breakdown is de-duplicated by NameWithOwner (first-seen order).
//
// TotalRepositoriesWithContributedCommits uses the larger of the distinct-repository
// count and the maximum server-reported per-chunk total. GitHub caps the per-chunk
// repository breakdown at 100 entries and does not paginate it, so len(repoOrder)
// undercounts truncated chunks; for a multi-chunk range the result is a best-effort
// lower bound. CommitContributionsByRepositoryTruncated is set when any source chunk's
// breakdown was incomplete, meaning CommitContributionsByRepository may be partial.
func mergeContributionChunks(chunks []*ContributionsCollection) *ContributionsCollection {
	result := &ContributionsCollection{}
	repoTotals := map[string]int{}
	var repoOrder []string
	var maxServerRepoTotal int

	for _, chunk := range chunks {
		result.TotalCommitContributions += chunk.TotalCommitContributions
		result.TotalIssueContributions += chunk.TotalIssueContributions
		result.TotalPullRequestContributions += chunk.TotalPullRequestContributions
		result.TotalPullRequestReviewContributions += chunk.TotalPullRequestReviewContributions
		result.ContributionCalendar.TotalContributions += chunk.ContributionCalendar.TotalContributions
		result.ContributionCalendar.Days = append(result.ContributionCalendar.Days, chunk.ContributionCalendar.Days...)
		maxServerRepoTotal = max(maxServerRepoTotal, chunk.TotalRepositoriesWithContributedCommits)
		result.CommitContributionsByRepositoryTruncated = result.CommitContributionsByRepositoryTruncated || chunk.CommitContributionsByRepositoryTruncated
		for _, rc := range chunk.CommitContributionsByRepository {
			if _, ok := repoTotals[rc.NameWithOwner]; !ok {
				repoOrder = append(repoOrder, rc.NameWithOwner)
			}
			repoTotals[rc.NameWithOwner] += rc.Contributions
		}
	}

	result.TotalRepositoriesWithContributedCommits = max(len(repoOrder), maxServerRepoTotal)
	for _, name := range repoOrder {
		result.CommitContributionsByRepository = append(result.CommitContributionsByRepository, RepositoryContributions{
			NameWithOwner: name,
			Contributions: repoTotals[name],
		})
	}
	return result
}
