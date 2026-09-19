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
func GetUserContributions(ctx context.Context, g *GitHubClient, username string, since, until time.Time) (*ContributionsCollection, error) {
	result := &ContributionsCollection{}
	repoTotals := map[string]int{}
	var repoOrder []string

	for start := since; start.Before(until); {
		end := start.Add(maxContributionsWindow)
		if end.After(until) {
			end = until
		}
		chunk, err := g.GetUserContributionsCollection(ctx, username, start, end)
		if err != nil {
			return nil, err
		}
		result.TotalCommitContributions += chunk.TotalCommitContributions
		result.TotalIssueContributions += chunk.TotalIssueContributions
		result.TotalPullRequestContributions += chunk.TotalPullRequestContributions
		result.TotalPullRequestReviewContributions += chunk.TotalPullRequestReviewContributions
		result.ContributionCalendar.TotalContributions += chunk.ContributionCalendar.TotalContributions
		result.ContributionCalendar.Days = append(result.ContributionCalendar.Days, chunk.ContributionCalendar.Days...)
		for _, rc := range chunk.CommitContributionsByRepository {
			if _, ok := repoTotals[rc.NameWithOwner]; !ok {
				repoOrder = append(repoOrder, rc.NameWithOwner)
			}
			repoTotals[rc.NameWithOwner] += rc.Contributions
		}
		start = end
	}

	result.TotalRepositoriesWithContributedCommits = len(repoOrder)
	for _, name := range repoOrder {
		result.CommitContributionsByRepository = append(result.CommitContributionsByRepository, RepositoryContributions{
			NameWithOwner: name,
			Contributions: repoTotals[name],
		})
	}
	return result, nil
}
