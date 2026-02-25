package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

// GetRepoPublicKey gets the public key for encrypting secrets in a repository (wrapper).
func GetRepoPublicKey(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.PublicKey, error) {
	return g.GetRepoPublicKey(ctx, repo.Owner, repo.Name)
}

// GetOrgPublicKey gets the public key for encrypting secrets in an organization (wrapper).
func GetOrgPublicKey(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.PublicKey, error) {
	return g.GetOrgPublicKey(ctx, repo.Owner)
}

// GetEnvPublicKey gets the public key for encrypting secrets in an environment (wrapper).
func GetEnvPublicKey(ctx context.Context, g *GitHubClient, repoID int, env string) (*github.PublicKey, error) {
	return g.GetEnvPublicKey(ctx, repoID, env)
}

// ListRepoSecrets lists all secrets in a repository (wrapper).
func ListRepoSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	return g.ListRepoSecrets(ctx, repo.Owner, repo.Name)
}

// ListRepoOrgSecrets lists all organization secrets available in a repository (wrapper).
func ListRepoOrgSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	return g.ListRepoOrgSecrets(ctx, repo.Owner, repo.Name)
}

// ListOrgSecrets lists all secrets in an organization (wrapper).
func ListOrgSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	return g.ListOrgSecrets(ctx, repo.Owner)
}

// ListEnvSecrets lists all secrets in an environment (wrapper).
func ListEnvSecrets(ctx context.Context, g *GitHubClient, repoID int, env string) ([]*github.Secret, error) {
	return g.ListEnvSecrets(ctx, repoID, env)
}

// ListSecrets lists secrets for a repository or organization depending on whether repo name is set (wrapper).
func ListSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	if repo.Name == "" {
		return ListOrgSecrets(ctx, g, repo)
	}
	return ListRepoSecrets(ctx, g, repo)
}

// GetRepoSecret gets a single repository secret (wrapper).
func GetRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.Secret, error) {
	return g.GetRepoSecret(ctx, repo.Owner, repo.Name, name)
}

// GetOrgSecret gets a single organization secret (wrapper).
func GetOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.Secret, error) {
	return g.GetOrgSecret(ctx, repo.Owner, name)
}

// GetEnvSecret gets a single environment secret (wrapper).
func GetEnvSecret(ctx context.Context, g *GitHubClient, repoID int, env, secretName string) (*github.Secret, error) {
	return g.GetEnvSecret(ctx, repoID, env, secretName)
}

// GetSecret gets a secret for a repository or organization depending on whether repo name is set (wrapper).
func GetSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) (*github.Secret, error) {
	if repo.Name == "" {
		return GetOrgSecret(ctx, g, repo, name)
	}
	return GetRepoSecret(ctx, g, repo, name)
}

// CreateOrUpdateRepoSecret creates or updates a repository secret (wrapper).
func CreateOrUpdateRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, eSecret *github.EncryptedSecret) error {
	return g.CreateOrUpdateRepoSecret(ctx, repo.Owner, repo.Name, eSecret)
}

// CreateOrUpdateOrgSecret creates or updates an organization secret (wrapper).
func CreateOrUpdateOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, eSecret *github.EncryptedSecret) error {
	return g.CreateOrUpdateOrgSecret(ctx, repo.Owner, eSecret)
}

// CreateOrUpdateEnvSecret creates or updates an environment secret (wrapper).
func CreateOrUpdateEnvSecret(ctx context.Context, g *GitHubClient, repoID int, env string, eSecret *github.EncryptedSecret) error {
	return g.CreateOrUpdateEnvSecret(ctx, repoID, env, eSecret)
}

// CreateOrUpdateSecret creates or updates a secret for a repository or organization depending on whether repo name is set (wrapper).
func CreateOrUpdateSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, eSecret *github.EncryptedSecret) error {
	if repo.Name == "" {
		return CreateOrUpdateOrgSecret(ctx, g, repo, eSecret)
	}
	return CreateOrUpdateRepoSecret(ctx, g, repo, eSecret)
}

// DeleteRepoSecret deletes a secret in a repository (wrapper).
func DeleteRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) error {
	return g.DeleteRepoSecret(ctx, repo.Owner, repo.Name, name)
}

// DeleteOrgSecret deletes a secret in an organization (wrapper).
func DeleteOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) error {
	return g.DeleteOrgSecret(ctx, repo.Owner, name)
}

// DeleteEnvSecret deletes a secret in an environment (wrapper).
func DeleteEnvSecret(ctx context.Context, g *GitHubClient, repoID int, env, secretName string) error {
	return g.DeleteEnvSecret(ctx, repoID, env, secretName)
}

// DeleteSecret deletes a secret for a repository or organization depending on whether repo name is set (wrapper).
func DeleteSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) error {
	if repo.Name == "" {
		return DeleteOrgSecret(ctx, g, repo, name)
	}
	return DeleteRepoSecret(ctx, g, repo, name)
}

// ListSelectedReposForOrgSecret lists all repositories that have access to an organization secret (wrapper).
func ListSelectedReposForOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) ([]*github.Repository, error) {
	return g.ListSelectedReposForOrgSecret(ctx, repo.Owner, name)
}

// SetSelectedReposForOrgSecret sets the repositories that have access to an organization secret (wrapper).
func SetSelectedReposForOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string, ids github.SelectedRepoIDs) error {
	return g.SetSelectedReposForOrgSecret(ctx, repo.Owner, name, ids)
}

// AddSelectedRepoToOrgSecret adds a repository to an organization secret (wrapper).
func AddSelectedRepoToOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string, targetRepo *github.Repository) error {
	return g.AddSelectedRepoToOrgSecret(ctx, repo.Owner, name, targetRepo)
}

// RemoveSelectedRepoFromOrgSecret removes a repository from an organization secret (wrapper).
func RemoveSelectedRepoFromOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string, targetRepo *github.Repository) error {
	return g.RemoveSelectedRepoFromOrgSecret(ctx, repo.Owner, name, targetRepo)
}
