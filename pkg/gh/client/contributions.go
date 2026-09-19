package client

import (
	"context"
	"time"

	"github.com/shurcooL/githubv4"
)

// ContributionsCollection summarizes a user's contribution activity over a date range.
type ContributionsCollection struct {
	TotalCommitContributions                int
	TotalIssueContributions                 int
	TotalPullRequestContributions           int
	TotalPullRequestReviewContributions     int
	TotalRepositoriesWithContributedCommits int
	CommitContributionsByRepository         []RepositoryContributions
	ContributionCalendar                    ContributionCalendar
}

// RepositoryContributions is the number of commit contributions made to a single repository.
type RepositoryContributions struct {
	NameWithOwner string
	Contributions int
}

// ContributionCalendar is a user's daily contribution counts over a date range.
type ContributionCalendar struct {
	TotalContributions int
	Days                []ContributionDay
}

// ContributionDay is the number of contributions made on a single day.
type ContributionDay struct {
	Date  string // YYYY-MM-DD
	Count int
}

// GetUserContributionsCollection fetches contribution stats for login over [from, to].
// GitHub restricts the range of a single contributionsCollection query to at most one year;
// callers needing a longer range must issue multiple queries and merge the results.
func (g *GitHubClient) GetUserContributionsCollection(ctx context.Context, login string, from, to time.Time) (*ContributionsCollection, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var q struct {
		User struct {
			ContributionsCollection struct {
				TotalCommitContributions                githubv4.Int
				TotalIssueContributions                 githubv4.Int
				TotalPullRequestContributions           githubv4.Int
				TotalPullRequestReviewContributions     githubv4.Int
				TotalRepositoriesWithContributedCommits githubv4.Int
				CommitContributionsByRepository         []struct {
					Repository struct {
						NameWithOwner githubv4.String
					}
					Contributions struct {
						TotalCount githubv4.Int
					}
				} `graphql:"commitContributionsByRepository(maxRepositories: 100)"`
				ContributionCalendar struct {
					TotalContributions githubv4.Int
					Weeks              []struct {
						ContributionDays []struct {
							Date              githubv4.String
							ContributionCount githubv4.Int
						}
					}
				}
			} `graphql:"contributionsCollection(from: $from, to: $to)"`
		} `graphql:"user(login: $login)"`
	}

	variables := map[string]any{
		"login": githubv4.String(login),
		"from":  githubv4.DateTime{Time: from},
		"to":    githubv4.DateTime{Time: to},
	}

	if err := graphql.Query(ctx, &q, variables); err != nil {
		return nil, err
	}

	cc := q.User.ContributionsCollection
	result := &ContributionsCollection{
		TotalCommitContributions:                int(cc.TotalCommitContributions),
		TotalIssueContributions:                 int(cc.TotalIssueContributions),
		TotalPullRequestContributions:           int(cc.TotalPullRequestContributions),
		TotalPullRequestReviewContributions:     int(cc.TotalPullRequestReviewContributions),
		TotalRepositoriesWithContributedCommits: int(cc.TotalRepositoriesWithContributedCommits),
		ContributionCalendar: ContributionCalendar{
			TotalContributions: int(cc.ContributionCalendar.TotalContributions),
		},
	}
	for _, repoContrib := range cc.CommitContributionsByRepository {
		result.CommitContributionsByRepository = append(result.CommitContributionsByRepository, RepositoryContributions{
			NameWithOwner: string(repoContrib.Repository.NameWithOwner),
			Contributions: int(repoContrib.Contributions.TotalCount),
		})
	}
	for _, week := range cc.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			result.ContributionCalendar.Days = append(result.ContributionCalendar.Days, ContributionDay{
				Date:  string(day.Date),
				Count: int(day.ContributionCount),
			})
		}
	}
	return result, nil
}
