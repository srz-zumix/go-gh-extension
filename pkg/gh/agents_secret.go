package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
)

// ListAgentsRepoSecrets lists all Agents secrets in a repository (wrapper).
func ListAgentsRepoSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	return g.ListAgentsRepoSecrets(ctx, repo.Owner, repo.Name)
}

// ListAgentsOrgSecrets lists all Agents secrets in an organization (wrapper).
func ListAgentsOrgSecrets(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.Secret, error) {
	return g.ListAgentsOrgSecrets(ctx, repo.Owner)
}

// GetAgentsRepoPublicKey gets the public key for encrypting Agents secrets in a repository (wrapper).
func GetAgentsRepoPublicKey(ctx context.Context, g *GitHubClient, repo repository.Repository) (*github.PublicKey, error) {
	return g.GetAgentsRepoPublicKey(ctx, repo.Owner, repo.Name)
}

// DeleteAgentsRepoSecret deletes an Agents secret in a repository (wrapper).
func DeleteAgentsRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) error {
	return g.DeleteAgentsRepoSecret(ctx, repo.Owner, repo.Name, name)
}

// CreateOrUpdateAgentsRepoSecret creates or updates an Agents secret in a repository (wrapper).
func CreateOrUpdateAgentsRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, eSecret *github.EncryptedSecret) error {
	return g.CreateOrUpdateAgentsRepoSecret(ctx, repo.Owner, repo.Name, eSecret)
}

// SetAgentsRepoSecret encrypts a plaintext value with the repository Agents
// public key and stores it as a repository Agents secret.
func SetAgentsRepoSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name, value string) error {
	publicKey, err := GetAgentsRepoPublicKey(ctx, g, repo)
	if err != nil {
		return err
	}
	eSecret, err := EncryptSecret(publicKey, name, value)
	if err != nil {
		return err
	}
	return CreateOrUpdateAgentsRepoSecret(ctx, g, repo, eSecret)
}

// ListSelectedReposForAgentsOrgSecret lists all repositories that have access to an organization Agents secret (wrapper).
func ListSelectedReposForAgentsOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string) ([]*github.Repository, error) {
	return g.ListSelectedReposForAgentsOrgSecret(ctx, repo.Owner, name)
}

// SetSelectedReposForAgentsOrgSecret sets the repositories that have access to an organization Agents secret (wrapper).
func SetSelectedReposForAgentsOrgSecret(ctx context.Context, g *GitHubClient, repo repository.Repository, name string, ids github.SelectedRepoIDs) error {
	return g.SetSelectedReposForAgentsOrgSecret(ctx, repo.Owner, name, ids)
}
