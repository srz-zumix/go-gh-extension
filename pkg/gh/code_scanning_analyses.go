package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// ListAnalysesOptions holds filter options for listing code scanning analyses.
type ListAnalysesOptions struct {
	SarifID string
	Ref     string
}

// toGitHubAnalysesListOptions converts ListAnalysesOptions to github.AnalysesListOptions.
func toGitHubAnalysesListOptions(opts *ListAnalysesOptions) *github.AnalysesListOptions {
	if opts == nil {
		return nil
	}
	o := &github.AnalysesListOptions{}
	if opts.SarifID != "" {
		o.SarifID = &opts.SarifID
	}
	if opts.Ref != "" {
		o.Ref = &opts.Ref
	}
	return o
}

// ListAnalyses lists code scanning analyses for a repository.
func ListAnalyses(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListAnalysesOptions) ([]*github.ScanningAnalysis, error) {
	analyses, err := g.ListAnalysesForRepo(ctx, repo.Owner, repo.Name, toGitHubAnalysesListOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to list code scanning analyses for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return analyses, nil
}

// GetAnalysis gets a code scanning analysis by ID for a repository.
func GetAnalysis(ctx context.Context, g *GitHubClient, repo repository.Repository, id int64) (*github.ScanningAnalysis, error) {
	analysis, err := g.GetAnalysis(ctx, repo.Owner, repo.Name, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get code scanning analysis #%d for %s/%s: %w", id, repo.Owner, repo.Name, err)
	}
	return analysis, nil
}
