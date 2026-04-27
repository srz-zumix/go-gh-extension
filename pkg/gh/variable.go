package gh

import (
	"context"
	"errors"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// ListRepoVariables lists all variables in a repository (wrapper).
func ListRepoVariables(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.ActionsVariable, error) {
	return g.ListRepoVariables(ctx, repo.Owner, repo.Name)
}

// ListOrgVariables lists all variables in an organization (wrapper).
func ListOrgVariables(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.ActionsVariable, error) {
	return g.ListOrgVariables(ctx, repo.Owner)
}

// ListVariables lists variables for a repository or organization depending on whether repo name is set (wrapper).
func ListVariables(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.ActionsVariable, error) {
	if repo.Name == "" {
		return ListOrgVariables(ctx, g, repo)
	}
	return ListRepoVariables(ctx, g, repo)
}

// GetRepoVariable gets a single repository variable (wrapper).
func GetRepoVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.ActionsVariable, error) {
	return g.GetRepoVariable(ctx, repo.Owner, repo.Name, name)
}

// GetOrgVariable gets a single organization variable (wrapper).
func GetOrgVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.ActionsVariable, error) {
	return g.GetOrgVariable(ctx, repo.Owner, name)
}

// GetVariable gets a single variable for a repository or organization depending on whether repo name is set (wrapper).
func GetVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.ActionsVariable, error) {
	if repo.Name == "" {
		return GetOrgVariable(ctx, g, repo, name)
	}
	return GetRepoVariable(ctx, g, repo, name)
}

// isVariableNotFound returns true if the error is a GitHub 404 response.
func isVariableNotFound(err error) bool {
	var errResp *github.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusNotFound
}

// IsVariableAlreadyExists returns true if the error is a GitHub 409 "Already exists" response.
func IsVariableAlreadyExists(err error) bool {
	var errResp *github.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusConflict
}

// CreateOrUpdateRepoVariable creates or updates a repository variable.
// If overwrite is true, it attempts to update first and falls back to creation if the variable does not exist.
// If overwrite is false, it attempts to create the variable and returns an error if creation fails.
func CreateOrUpdateRepoVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, variable *github.ActionsVariable, overwrite bool) error {
	if variable == nil || variable.Name == "" {
		return errors.New("variable must not be nil and must have a non-empty name")
	}
	if overwrite {
		// Try updating first; if the variable does not exist, fall back to creation.
		err := g.UpdateRepoVariable(ctx, repo.Owner, repo.Name, variable)
		if err == nil {
			return nil
		}
		if isVariableNotFound(err) {
			return g.CreateRepoVariable(ctx, repo.Owner, repo.Name, variable)
		}
		return err
	}

	return g.CreateRepoVariable(ctx, repo.Owner, repo.Name, variable)
}

// CreateOrUpdateOrgVariable creates or updates an organization variable.
// If overwrite is true, it attempts to update first and falls back to creation if the variable does not exist.
// If overwrite is false, it attempts to create the variable and returns an error if creation fails.
func CreateOrUpdateOrgVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, variable *github.ActionsVariable, overwrite bool) error {
	if variable == nil || variable.Name == "" {
		return errors.New("variable must not be nil and must have a non-empty name")
	}
	if overwrite {
		// Try updating first; if the variable does not exist, fall back to creation.
		err := g.UpdateOrgVariable(ctx, repo.Owner, variable)
		if err == nil {
			return nil
		}
		if isVariableNotFound(err) {
			return g.CreateOrgVariable(ctx, repo.Owner, variable)
		}
		return err
	}

	return g.CreateOrgVariable(ctx, repo.Owner, variable)
}

// CreateOrUpdateVariable creates or updates a variable for a repository or organization depending on whether repo name is set.
// If overwrite is true, it attempts to update first and falls back to creation if the variable does not exist.
// If overwrite is false, it attempts to create the variable and returns an error if creation fails.
func CreateOrUpdateVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, variable *github.ActionsVariable, overwrite bool) error {
	if repo.Name == "" {
		return CreateOrUpdateOrgVariable(ctx, g, repo, variable, overwrite)
	}
	return CreateOrUpdateRepoVariable(ctx, g, repo, variable, overwrite)
}

// ListEnvVariables lists all variables in an environment (wrapper).
func ListEnvVariables(ctx context.Context, g *GitHubClient, repo repository.Repository, env string) ([]*github.ActionsVariable, error) {
	return g.ListEnvVariables(ctx, repo.Owner, repo.Name, env)
}

// GetEnvVariable gets a single environment variable (wrapper).
func GetEnvVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, env, name string) (*github.ActionsVariable, error) {
	return g.GetEnvVariable(ctx, repo.Owner, repo.Name, env, name)
}

// CreateOrUpdateEnvVariable creates or updates an environment variable.
// If overwrite is true, it attempts to update first and falls back to creation if the variable does not exist.
// If overwrite is false, it attempts to create the variable and returns an error if creation fails.
func CreateOrUpdateEnvVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, env string, variable *github.ActionsVariable, overwrite bool) error {
	if variable == nil || variable.Name == "" {
		return errors.New("variable must not be nil and must have a non-empty name")
	}
	if overwrite {
		// Try updating first; if the variable does not exist, fall back to creation.
		err := g.UpdateEnvVariable(ctx, repo.Owner, repo.Name, env, variable)
		if err == nil {
			return nil
		}
		if isVariableNotFound(err) {
			return g.CreateEnvVariable(ctx, repo.Owner, repo.Name, env, variable)
		}
		return err
	}

	return g.CreateEnvVariable(ctx, repo.Owner, repo.Name, env, variable)
}
