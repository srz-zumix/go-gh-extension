package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListCodeQLDatabases lists CodeQL databases for a repository.
func (g *GitHubClient) ListCodeQLDatabases(ctx context.Context, owner, repo string) ([]*github.CodeQLDatabase, error) {
	databases, _, err := g.client.CodeScanning.ListCodeQLDatabases(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return databases, nil
}

// GetCodeQLDatabase gets a CodeQL database by language for a repository.
func (g *GitHubClient) GetCodeQLDatabase(ctx context.Context, owner, repo, language string) (*github.CodeQLDatabase, error) {
	database, _, err := g.client.CodeScanning.GetCodeQLDatabase(ctx, owner, repo, language)
	if err != nil {
		return nil, err
	}
	return database, nil
}
