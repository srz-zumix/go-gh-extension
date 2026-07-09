package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/v88/github"
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

// DeleteAnalysis deletes a code scanning analysis from a repository.
// If confirmDelete is true, deletion of the last analysis in a set is allowed.
func (g *GitHubClient) DeleteAnalysis(ctx context.Context, owner, repo string, id int64, confirmDelete bool) (*github.DeleteAnalysis, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/analyses/%v", owner, repo, id)
	if confirmDelete {
		params := url.Values{}
		params.Set("confirm_delete", "true")
		u = fmt.Sprintf("%s?%s", u, params.Encode())
	}

	req, err := g.client.NewRequest(ctx, "DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	result := new(github.DeleteAnalysis)
	_, err = g.client.Do(req, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
