package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListAnalysesForRepo lists code scanning analyses for a repository.
func (g *GitHubClient) ListAnalysesForRepo(ctx context.Context, owner, repo string, opts *github.AnalysesListOptions) ([]*github.ScanningAnalysis, error) {
	var allAnalyses []*github.ScanningAnalysis
	if opts == nil {
		opts = &github.AnalysesListOptions{}
	}
	opts.ListOptions = github.ListOptions{PerPage: defaultPerPage}

	for {
		analyses, resp, err := g.client.CodeScanning.ListAnalysesForRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allAnalyses = append(allAnalyses, analyses...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allAnalyses, nil
}

// GetAnalysis gets a code scanning analysis for a repository.
func (g *GitHubClient) GetAnalysis(ctx context.Context, owner, repo string, id int64) (*github.ScanningAnalysis, error) {
	analysis, _, err := g.client.CodeScanning.GetAnalysis(ctx, owner, repo, id)
	if err != nil {
		return nil, err
	}
	return analysis, nil
}
