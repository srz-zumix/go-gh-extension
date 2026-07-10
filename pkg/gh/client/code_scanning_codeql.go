package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/v88/github"
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

// DeleteCodeQLDatabase deletes a CodeQL database for a language in a repository.
func (g *GitHubClient) DeleteCodeQLDatabase(ctx context.Context, owner, repo, language string) error {
	u := fmt.Sprintf("repos/%v/%v/code-scanning/codeql/databases/%v", owner, repo, url.PathEscape(language))
	req, err := g.client.NewRequest(ctx, "DELETE", u, nil)
	if err != nil {
		return err
	}
	_, err = g.client.Do(req, nil)
	if err != nil {
		return err
	}
	return nil
}
