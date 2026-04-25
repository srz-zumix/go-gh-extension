package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// UploadSarifOptions holds the parameters for uploading SARIF data.
type UploadSarifOptions struct {
	CommitSHA   string
	Ref         string
	Sarif       string
	CheckoutURI string
	StartedAt   string
	ToolName    string
}

// toGitHubSarifAnalysis converts UploadSarifOptions to github.SarifAnalysis.
func toGitHubSarifAnalysis(opts *UploadSarifOptions) *github.SarifAnalysis {
	if opts == nil {
		return nil
	}
	s := &github.SarifAnalysis{
		CommitSHA: &opts.CommitSHA,
		Ref:       &opts.Ref,
		Sarif:     &opts.Sarif,
	}
	if opts.CheckoutURI != "" {
		s.CheckoutURI = &opts.CheckoutURI
	}
	if opts.StartedAt != "" {
		t, err := time.Parse(time.RFC3339, opts.StartedAt)
		if err == nil {
			s.StartedAt = &github.Timestamp{Time: t}
		}
	}
	if opts.ToolName != "" {
		s.ToolName = &opts.ToolName
	}
	return s
}

// UploadSarif uploads SARIF data to a repository.
func UploadSarif(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *UploadSarifOptions) (*github.SarifID, error) {
	sarifID, err := g.UploadSarif(ctx, repo.Owner, repo.Name, toGitHubSarifAnalysis(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to upload SARIF for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return sarifID, nil
}

// GetSarif gets information about a SARIF upload.
func GetSarif(ctx context.Context, g *GitHubClient, repo repository.Repository, sarifID string) (*github.SARIFUpload, error) {
	upload, err := g.GetSARIF(ctx, repo.Owner, repo.Name, sarifID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SARIF upload %q for %s/%s: %w", sarifID, repo.Owner, repo.Name, err)
	}
	return upload, nil
}
