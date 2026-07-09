package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v88/github"
)

// CreateCodeQLVariantAnalysisOptions specifies parameters for creating a CodeQL variant analysis.
// Exactly one of Repositories, RepositoryOwners or RepositoryLists must be set.
type CreateCodeQLVariantAnalysisOptions struct {
	Language         string   `json:"language"`
	QueryPack        string   `json:"query_pack"`
	Repositories     []string `json:"repositories,omitempty"`
	RepositoryOwners []string `json:"repository_owners,omitempty"`
	RepositoryLists  []string `json:"repository_lists,omitempty"`
}

// CodeQLVariantAnalysisScannedRepository represents a repository scanned as part of a CodeQL variant analysis.
type CodeQLVariantAnalysisScannedRepository struct {
	Repository          *github.Repository `json:"repository,omitempty"`
	AnalysisStatus      string             `json:"analysis_status"`
	ResultCount         *int               `json:"result_count,omitempty"`
	ArtifactSizeInBytes *int               `json:"artifact_size_in_bytes,omitempty"`
	FailureMessage      *string            `json:"failure_message,omitempty"`
}

// CodeQLVariantAnalysis represents a CodeQL variant analysis (multi-repository CodeQL query run).
type CodeQLVariantAnalysis struct {
	ID                   int64                                    `json:"id"`
	ControllerRepo       *github.Repository                       `json:"controller_repo,omitempty"`
	Actor                *github.User                             `json:"actor,omitempty"`
	QueryLanguage        string                                   `json:"query_language"`
	QueryPackURL         string                                   `json:"query_pack_url"`
	CreatedAt            *github.Timestamp                        `json:"created_at,omitempty"`
	UpdatedAt            *github.Timestamp                        `json:"updated_at,omitempty"`
	CompletedAt          *github.Timestamp                        `json:"completed_at,omitempty"`
	Status               string                                   `json:"status"`
	ActionsWorkflowRunID *int64                                   `json:"actions_workflow_run_id,omitempty"`
	FailureReason        *string                                  `json:"failure_reason,omitempty"`
	ScannedRepositories  []CodeQLVariantAnalysisScannedRepository `json:"scanned_repositories,omitempty"`
	SkippedRepositories  json.RawMessage                          `json:"skipped_repositories,omitempty"`
}

// CodeQLVariantAnalysisRepoStatus represents the analysis status of a single repository in a CodeQL variant analysis.
type CodeQLVariantAnalysisRepoStatus struct {
	Repository           *github.Repository `json:"repository,omitempty"`
	AnalysisStatus       string             `json:"analysis_status"`
	ArtifactSizeInBytes  *int               `json:"artifact_size_in_bytes,omitempty"`
	ResultCount          *int               `json:"result_count,omitempty"`
	FailureMessage       *string            `json:"failure_message,omitempty"`
	DatabaseCommitSHA    *string            `json:"database_commit_sha,omitempty"`
	SourceLocationPrefix *string            `json:"source_location_prefix,omitempty"`
	ArtifactURL          *string            `json:"artifact_url,omitempty"`
}

// CreateCodeQLVariantAnalysis creates a new CodeQL variant analysis, running a CodeQL query
// against one or more repositories. The controller repo is specified by owner/repo.
func (g *GitHubClient) CreateCodeQLVariantAnalysis(ctx context.Context, owner, repo string, opts *CreateCodeQLVariantAnalysisOptions) (*CodeQLVariantAnalysis, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/codeql/variant-analyses", owner, repo)
	req, err := g.client.NewRequest(ctx, "POST", u, opts)
	if err != nil {
		return nil, err
	}
	result := new(CodeQLVariantAnalysis)
	_, err = g.client.Do(req, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetCodeQLVariantAnalysis gets the summary of a CodeQL variant analysis.
func (g *GitHubClient) GetCodeQLVariantAnalysis(ctx context.Context, owner, repo string, id int64) (*CodeQLVariantAnalysis, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/codeql/variant-analyses/%v", owner, repo, id)
	req, err := g.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	result := new(CodeQLVariantAnalysis)
	_, err = g.client.Do(req, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetCodeQLVariantAnalysisRepoStatus gets the analysis status of a repository in a CodeQL variant analysis.
func (g *GitHubClient) GetCodeQLVariantAnalysisRepoStatus(ctx context.Context, owner, repo string, id int64, repoOwner, repoName string) (*CodeQLVariantAnalysisRepoStatus, error) {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/codeql/variant-analyses/%v/repos/%v/%v", owner, repo, id, repoOwner, repoName)
	req, err := g.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	result := new(CodeQLVariantAnalysisRepoStatus)
	_, err = g.client.Do(req, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
