package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// UploadSARIFOptions holds the parameters for uploading SARIF data.
type UploadSARIFOptions struct {
	CommitSHA string
	Ref       string
	// SARIF must be a base64-encoded SARIF payload. The GitHub API expects
	// gzip-compressed SARIF content encoded as base64 rather than raw JSON.
	SARIF       string
	CheckoutURI string
	StartedAt   time.Time
	ToolName    string
}

// toGitHubSARIFAnalysis converts UploadSARIFOptions to github.SarifAnalysis.
func toGitHubSARIFAnalysis(opts *UploadSARIFOptions) *github.SarifAnalysis {
	if opts == nil {
		return nil
	}
	s := &github.SarifAnalysis{
		CommitSHA: &opts.CommitSHA,
		Ref:       &opts.Ref,
		Sarif:     &opts.SARIF,
	}
	if opts.CheckoutURI != "" {
		s.CheckoutURI = &opts.CheckoutURI
	}
	if !opts.StartedAt.IsZero() {
		s.StartedAt = &github.Timestamp{Time: opts.StartedAt}
	}
	if opts.ToolName != "" {
		s.ToolName = &opts.ToolName
	}
	return s
}

// UploadSARIF uploads SARIF data to a repository.
//
// The SARIF payload must already be encoded in the format expected by the
// GitHub API: base64-encoded and typically gzip-compressed.
func UploadSARIF(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *UploadSARIFOptions) (*github.SarifID, error) {
	sarifID, err := g.UploadSarif(ctx, repo.Owner, repo.Name, toGitHubSARIFAnalysis(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to upload SARIF for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return sarifID, nil
}

// GetSARIF gets information about a SARIF upload.
func GetSARIF(ctx context.Context, g *GitHubClient, repo repository.Repository, sarifID string) (*github.SARIFUpload, error) {
	upload, err := g.GetSARIF(ctx, repo.Owner, repo.Name, sarifID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SARIF upload %q for %s/%s: %w", sarifID, repo.Owner, repo.Name, err)
	}
	return upload, nil
}
