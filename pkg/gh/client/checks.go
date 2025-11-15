package client

import (
	"context"

	"github.com/google/go-github/v79/github"
)

// ListCheckRunsForRef retrieves all check runs for a specific Git reference.
func (g *GitHubClient) ListCheckRunsForRef(ctx context.Context, owner string, repo string, ref string, options *github.ListCheckRunsOptions) (*github.ListCheckRunsResults, error) {
	opt := github.ListCheckRunsOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	if options != nil {
		opt = *options
		opt.PerPage = defaultPerPage
	}

	var allCheckRuns []*github.CheckRun
	for {
		checkRuns, resp, err := g.client.Checks.ListCheckRunsForRef(ctx, owner, repo, ref, &opt)
		if err != nil {
			return nil, err
		}
		allCheckRuns = append(allCheckRuns, checkRuns.CheckRuns...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	total := len(allCheckRuns)
	return &github.ListCheckRunsResults{
		Total:     &total,
		CheckRuns: allCheckRuns,
	}, nil
}
