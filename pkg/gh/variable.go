package gh

import (
	"context"
	"errors"
	"net/http"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
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

// isVariableNotFound returns true if the error is a GitHub 404 response.
func isVariableNotFound(err error) bool {
	var errResp *github.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusNotFound
}

// isVariableConflict returns true if the error is a GitHub 409 response.
func isVariableConflict(err error) bool {
	var errResp *github.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusConflict
}

// CreateOrUpdateRepoVariable creates or updates a repository variable.
// Returns true if the variable was written, false if it was skipped (already exists and overwrite is false).
func CreateOrUpdateRepoVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, variable *github.ActionsVariable, overwrite bool) (bool, error) {
	if overwrite {
		// Try updating first; if the variable does not exist, fall back to creation.
		err := g.UpdateRepoVariable(ctx, repo.Owner, repo.Name, variable)
		if err == nil {
			return true, nil
		}
		if isVariableNotFound(err) {
			return true, g.CreateRepoVariable(ctx, repo.Owner, repo.Name, variable)
		}
		return false, err
	}

	// When not overwriting, try to create and treat conflicts as "already exists".
	err := g.CreateRepoVariable(ctx, repo.Owner, repo.Name, variable)
	if err == nil {
		return true, nil
	}
	if isVariableConflict(err) {
		return false, nil
	}
	return false, err
}

// CreateOrUpdateOrgVariable creates or updates an organization variable.
// Returns true if the variable was written, false if it was skipped (already exists and overwrite is false).
func CreateOrUpdateOrgVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, variable *github.ActionsVariable, overwrite bool) (bool, error) {
	if overwrite {
		// Try updating first; if the variable does not exist, fall back to creation.
		err := g.UpdateOrgVariable(ctx, repo.Owner, variable)
		if err == nil {
			return true, nil
		}
		if isVariableNotFound(err) {
			return true, g.CreateOrgVariable(ctx, repo.Owner, variable)
		}
		return false, err
	}

	// When not overwriting, try to create and treat conflicts as "already exists".
	err := g.CreateOrgVariable(ctx, repo.Owner, variable)
	if err == nil {
		return true, nil
	}
	if isVariableConflict(err) {
		return false, nil
	}
	return false, err
}

// CreateOrUpdateVariable creates or updates a variable for a repository or organization depending on whether repo name is set (wrapper).
// Returns true if the variable was written, false if it was skipped (already exists and overwrite is false).
func CreateOrUpdateVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, variable *github.ActionsVariable, overwrite bool) (bool, error) {
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
// Returns true if the variable was written, false if it was skipped (already exists and overwrite is false).
func CreateOrUpdateEnvVariable(ctx context.Context, g *GitHubClient, repo repository.Repository, env string, variable *github.ActionsVariable, overwrite bool) (bool, error) {
	_, err := GetEnvVariable(ctx, g, repo, env, variable.Name)
	if err != nil {
		if isVariableNotFound(err) {
			return true, g.CreateEnvVariable(ctx, repo.Owner, repo.Name, env, variable)
		}
		return false, err
	}
	if !overwrite {
		return false, nil
	}
	return true, g.UpdateEnvVariable(ctx, repo.Owner, repo.Name, env, variable)
}
