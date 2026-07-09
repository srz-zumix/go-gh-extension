package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// CreateCodeQLVariantAnalysisOptions holds options for creating a CodeQL variant analysis.
// Exactly one of Repositories, RepositoryOwners or RepositoryLists must be set.
type CreateCodeQLVariantAnalysisOptions struct {
	Language         string
	QueryPack        string
	Repositories     []string
	RepositoryOwners []string
	RepositoryLists  []string
}

// toClientCreateCodeQLVariantAnalysisOptions converts CreateCodeQLVariantAnalysisOptions
// to client.CreateCodeQLVariantAnalysisOptions.
func toClientCreateCodeQLVariantAnalysisOptions(opts *CreateCodeQLVariantAnalysisOptions) *client.CreateCodeQLVariantAnalysisOptions {
	return &client.CreateCodeQLVariantAnalysisOptions{
		Language:         opts.Language,
		QueryPack:        opts.QueryPack,
		Repositories:     opts.Repositories,
		RepositoryOwners: opts.RepositoryOwners,
		RepositoryLists:  opts.RepositoryLists,
	}
}

// CreateCodeQLVariantAnalysis creates a new CodeQL variant analysis, running a CodeQL query
// against one or more repositories. The controller repo is specified by repo.
func CreateCodeQLVariantAnalysis(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *CreateCodeQLVariantAnalysisOptions) (*client.CodeQLVariantAnalysis, error) {
	result, err := g.CreateCodeQLVariantAnalysis(ctx, repo.Owner, repo.Name, toClientCreateCodeQLVariantAnalysisOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to create CodeQL variant analysis for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return result, nil
}

// GetCodeQLVariantAnalysis gets the summary of a CodeQL variant analysis.
func GetCodeQLVariantAnalysis(ctx context.Context, g *GitHubClient, repo repository.Repository, id int64) (*client.CodeQLVariantAnalysis, error) {
	result, err := g.GetCodeQLVariantAnalysis(ctx, repo.Owner, repo.Name, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get CodeQL variant analysis #%d for %s/%s: %w", id, repo.Owner, repo.Name, err)
	}
	return result, nil
}

// GetCodeQLVariantAnalysisRepoStatus gets the analysis status of a repository in a CodeQL variant analysis.
func GetCodeQLVariantAnalysisRepoStatus(ctx context.Context, g *GitHubClient, repo repository.Repository, id int64, targetRepo repository.Repository) (*client.CodeQLVariantAnalysisRepoStatus, error) {
	result, err := g.GetCodeQLVariantAnalysisRepoStatus(ctx, repo.Owner, repo.Name, id, targetRepo.Owner, targetRepo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository status for CodeQL variant analysis #%d in %s/%s: %w", id, repo.Owner, repo.Name, err)
	}
	return result, nil
}
