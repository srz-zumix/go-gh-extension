package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// ListCodeQLDatabases lists CodeQL databases for a repository.
func ListCodeQLDatabases(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.CodeQLDatabase, error) {
	databases, err := g.ListCodeQLDatabases(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list CodeQL databases for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return databases, nil
}

// GetCodeQLDatabase gets a CodeQL database by language for a repository.
func GetCodeQLDatabase(ctx context.Context, g *GitHubClient, repo repository.Repository, language string) (*github.CodeQLDatabase, error) {
	database, err := g.GetCodeQLDatabase(ctx, repo.Owner, repo.Name, language)
	if err != nil {
		return nil, fmt.Errorf("failed to get CodeQL database for language %q in %s/%s: %w", language, repo.Owner, repo.Name, err)
	}
	return database, nil
}
