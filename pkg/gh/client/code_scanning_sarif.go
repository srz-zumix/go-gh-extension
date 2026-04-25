package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// UploadSarif uploads SARIF data for a repository.
func (g *GitHubClient) UploadSarif(ctx context.Context, owner, repo string, sarif *github.SarifAnalysis) (*github.SarifID, error) {
	sarifID, _, err := g.client.CodeScanning.UploadSarif(ctx, owner, repo, sarif)
	if err != nil {
		return nil, err
	}
	return sarifID, nil
}

// GetSARIF gets information about a SARIF upload.
func (g *GitHubClient) GetSARIF(ctx context.Context, owner, repo, sarifID string) (*github.SARIFUpload, error) {
	upload, _, err := g.client.CodeScanning.GetSARIF(ctx, owner, repo, sarifID)
	if err != nil {
		return nil, err
	}
	return upload, nil
}
