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
// GitHub treats the contributionsCollection range as inclusive on both ends, so
// consecutive chunks must not share a boundary instant or the boundary day would be
// counted twice. Non-final chunks therefore end one second before the next chunk
// begins, keeping the windows disjoint; only the final chunk includes until.
//
// When the result spans multiple chunks and any chunk's per-repository breakdown was
// truncated (CommitContributionsByRepositoryTruncated), TotalRepositoriesWithContributedCommits
// is a best-effort lower bound rather than an exact count.
func GetUserContributions(ctx context.Context, g *GitHubClient, username string, since, until time.Time) (*ContributionsCollection, error) {
	var chunks []*ContributionsCollection
	for _, w := range contributionWindows(since, until) {
		chunk, err := g.GetUserContributionsCollection(ctx, username, w.from, w.to)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}
	return mergeContributionChunks(chunks), nil
}

// contributionWindow is a single [from, to] query range.
type contributionWindow struct {
	from time.Time
	to   time.Time
}

// contributionWindows splits [since, until] into pairwise-disjoint query ranges no
// longer than maxContributionsWindow. GitHub treats the range as inclusive on both
// ends, so every non-final window stops one second before the next window begins to
// avoid double-counting the shared boundary instant and its calendar day. The final
// window ends exactly at until. An empty slice is returned when until is not after
// since.
func contributionWindows(since, until time.Time) []contributionWindow {
	// GitHub counts contributions at one-second granularity, so normalize the range
	// to whole seconds. This keeps the one-second handoff between windows exact: with
	// sub-second endpoints the boundaries would be offset by a fraction and a
	// contribution could fall into the gap between one window's end and the next
	// window's start.
	since = since.Truncate(time.Second)
	until = until.Truncate(time.Second)

	var windows []contributionWindow
	for start := since; start.Before(until); {
		next := start.Add(maxContributionsWindow)
		to := until
		final := !next.Before(until)
		if !final {
			to = next.Add(-time.Second)
		}
		windows = append(windows, contributionWindow{from: start, to: to})
		if final {
			break
		}
		start = next
	}
	return windows
}

// mergeContributionChunks combines per-window contribution results into a single
// collection. Callers must supply chunks covering disjoint (non-overlapping) instant
// ranges: additive totals and per-repository commit counts are summed, and the
// commit-by-repository breakdown is de-duplicated by NameWithOwner (first-seen order).
//
// The contribution calendar is day-granular in the user's timezone, so a day that
// straddles a chunk boundary can be reported in full by both adjacent chunks. Calendar
// days are therefore de-duplicated by date (keeping the largest reported count), and
// ContributionCalendar.TotalContributions is recomputed from those unique days.
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
	dayCounts := map[string]int{}
	var dayOrder []string

	for _, chunk := range chunks {
		result.TotalCommitContributions += chunk.TotalCommitContributions
		result.TotalIssueContributions += chunk.TotalIssueContributions
		result.TotalPullRequestContributions += chunk.TotalPullRequestContributions
		result.TotalPullRequestReviewContributions += chunk.TotalPullRequestReviewContributions
		maxServerRepoTotal = max(maxServerRepoTotal, chunk.TotalRepositoriesWithContributedCommits)
		result.CommitContributionsByRepositoryTruncated = result.CommitContributionsByRepositoryTruncated || chunk.CommitContributionsByRepositoryTruncated
		for _, rc := range chunk.CommitContributionsByRepository {
			if _, ok := repoTotals[rc.NameWithOwner]; !ok {
				repoOrder = append(repoOrder, rc.NameWithOwner)
			}
			repoTotals[rc.NameWithOwner] += rc.Contributions
		}
		for _, day := range chunk.ContributionCalendar.Days {
			if existing, ok := dayCounts[day.Date]; !ok {
				dayOrder = append(dayOrder, day.Date)
				dayCounts[day.Date] = day.Count
			} else {
				dayCounts[day.Date] = max(existing, day.Count)
			}
		}
	}

	result.TotalRepositoriesWithContributedCommits = max(len(repoOrder), maxServerRepoTotal)
	for _, name := range repoOrder {
		result.CommitContributionsByRepository = append(result.CommitContributionsByRepository, RepositoryContributions{
			NameWithOwner: name,
			Contributions: repoTotals[name],
		})
	}
	for _, date := range dayOrder {
		result.ContributionCalendar.Days = append(result.ContributionCalendar.Days, ContributionDay{Date: date, Count: dayCounts[date]})
		result.ContributionCalendar.TotalContributions += dayCounts[date]
	}
	return result
}
