package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

// ListDependabotRepoSecrets lists all Dependabot secrets in a repository (wrapper).
func ListDependabotRepoSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	return g.ListDependabotRepoSecrets(ctx, repo.Owner, repo.Name)
}

// ListDependabotOrgSecrets lists all Dependabot secrets in an organization (wrapper).
func ListDependabotOrgSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	return g.ListDependabotOrgSecrets(ctx, repo.Owner)
}

// GetDependabotRepoPublicKey gets the public key for encrypting Dependabot secrets in a repository (wrapper).
func GetDependabotRepoPublicKey(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.PublicKey, error) {
	return g.GetDependabotRepoPublicKey(ctx, repo.Owner, repo.Name)
}

// CreateOrUpdateDependabotRepoSecret creates or updates a Dependabot secret in a repository (wrapper).
func CreateOrUpdateDependabotRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, eSecret *github.EncryptedSecret) error {
	return g.CreateOrUpdateDependabotRepoSecret(ctx, repo.Owner, repo.Name, eSecret)
}

// SetDependabotRepoSecret encrypts a plaintext value with the repository Dependabot
// public key and stores it as a repository Dependabot secret.
func SetDependabotRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name, value string) error {
	publicKey, err := GetDependabotRepoPublicKey(ctx, g, repo)
	if err != nil {
		return err
	}
	eSecret, err := EncryptSecret(publicKey, name, value)
	if err != nil {
		return err
	}
	return CreateOrUpdateDependabotRepoSecret(ctx, g, repo, eSecret)
}

// DeleteDependabotRepoSecret deletes a Dependabot secret in a repository (wrapper).
func DeleteDependabotRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) error {
	return g.DeleteDependabotRepoSecret(ctx, repo.Owner, repo.Name, name)
}

// ListSelectedReposForDependabotOrgSecret lists all repositories that have access to an organization Dependabot secret (wrapper).
func ListSelectedReposForDependabotOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) ([]*github.Repository, error) {
	return g.ListSelectedReposForDependabotOrgSecret(ctx, repo.Owner, name)
}

// SetSelectedReposForDependabotOrgSecret sets the repositories that have access to an organization Dependabot secret (wrapper).
func SetSelectedReposForDependabotOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string, ids []int64) error {
	return g.SetSelectedReposForDependabotOrgSecret(ctx, repo.Owner, name, ids)
}
