package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// CodeQualitySetupStates is the list of valid state values for code quality setup.
var CodeQualitySetupStates = []string{
	"configured",
	"not-configured",
}

// CodeQualitySetupRunnerTypes is the list of valid runner_type values.
var CodeQualitySetupRunnerTypes = []string{
	"standard",
	"labeled",
}

// CodeQualitySetupLanguages is the list of supported languages for code quality setup.
var CodeQualitySetupLanguages = []string{
	"csharp",
	"go",
	"java-kotlin",
	"javascript-typescript",
	"python",
	"ruby",
}

// GetCodeQualitySetup gets the code quality setup configuration for a repository.
func GetCodeQualitySetup(ctx context.Context, g *GitHubClient, repo repository.Repository) (*client.CodeQualitySetup, error) {
	setup, err := g.GetCodeQualitySetup(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get code quality setup for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return setup, nil
}

// UpdateCodeQualitySetupOptions holds the parameters for updating a code quality setup configuration.
type UpdateCodeQualitySetupOptions struct {
	State       string
	RunnerType  string
	RunnerLabel *string
	Languages   []string
}

// UpdateCodeQualitySetup updates the code quality setup configuration for a repository.
func UpdateCodeQualitySetup(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *UpdateCodeQualitySetupOptions) error {
	update := &client.CodeQualitySetupUpdate{
		State:       opts.State,
		RunnerType:  opts.RunnerType,
		RunnerLabel: opts.RunnerLabel,
		Languages:   opts.Languages,
	}
	err := g.UpdateCodeQualitySetup(ctx, repo.Owner, repo.Name, update)
	if err != nil {
		return fmt.Errorf("failed to update code quality setup for %s/%s: %w", repo.Owner, repo.Name, err)
	}
	return nil
}
