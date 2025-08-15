package client

import (
	"context"

	"github.com/google/go-github/v73/github"
)

func (g *GitHubClient) SearchIssues(ctx context.Context, query string) ([]*github.Issue, error) {
	allIssues := []*github.Issue{}
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		result, resp, err := g.client.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, err
		}
		allIssues = append(allIssues, result.Issues...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage

	}
	return allIssues, nil
}
