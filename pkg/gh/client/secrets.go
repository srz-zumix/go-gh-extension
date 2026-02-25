package client

// GitHub Actions Secrets API functions
// See: https://docs.github.com/rest/actions/secrets

import (
	"context"

	"github.com/google/go-github/v79/github"
)

// GetRepoPublicKey gets the public key for encrypting secrets in a repository.
func (g *GitHubClient) GetRepoPublicKey(ctx context.Context, owner, repo string) (*github.PublicKey, error) {
	key, _, err := g.client.Actions.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GetOrgPublicKey gets the public key for encrypting secrets in an organization.
func (g *GitHubClient) GetOrgPublicKey(ctx context.Context, org string) (*github.PublicKey, error) {
	key, _, err := g.client.Actions.GetOrgPublicKey(ctx, org)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GetEnvPublicKey gets the public key for encrypting secrets in an environment.
func (g *GitHubClient) GetEnvPublicKey(ctx context.Context, repoID int, env string) (*github.PublicKey, error) {
	key, _, err := g.client.Actions.GetEnvPublicKey(ctx, repoID, env)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// ListRepoSecrets lists all secrets in a repository without revealing their encrypted values.
func (g *GitHubClient) ListRepoSecrets(ctx context.Context, owner, repo string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		secrets, resp, err := g.client.Actions.ListRepoSecrets(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allSecrets = append(allSecrets, secrets.Secrets...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allSecrets, nil
}

// ListRepoOrgSecrets lists all organization secrets available in a repository without revealing their encrypted values.
func (g *GitHubClient) ListRepoOrgSecrets(ctx context.Context, owner, repo string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		secrets, resp, err := g.client.Actions.ListRepoOrgSecrets(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allSecrets = append(allSecrets, secrets.Secrets...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allSecrets, nil
}

// ListOrgSecrets lists all secrets in an organization without revealing their encrypted values.
func (g *GitHubClient) ListOrgSecrets(ctx context.Context, org string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		secrets, resp, err := g.client.Actions.ListOrgSecrets(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		allSecrets = append(allSecrets, secrets.Secrets...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allSecrets, nil
}

// ListEnvSecrets lists all secrets in an environment without revealing their encrypted values.
func (g *GitHubClient) ListEnvSecrets(ctx context.Context, repoID int, env string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		secrets, resp, err := g.client.Actions.ListEnvSecrets(ctx, repoID, env, opt)
		if err != nil {
			return nil, err
		}
		allSecrets = append(allSecrets, secrets.Secrets...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allSecrets, nil
}

// GetRepoSecret gets a single repository secret without revealing its encrypted value.
func (g *GitHubClient) GetRepoSecret(ctx context.Context, owner, repo, name string) (*github.Secret, error) {
	secret, _, err := g.client.Actions.GetRepoSecret(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// GetOrgSecret gets a single organization secret without revealing its encrypted value.
func (g *GitHubClient) GetOrgSecret(ctx context.Context, org, name string) (*github.Secret, error) {
	secret, _, err := g.client.Actions.GetOrgSecret(ctx, org, name)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// GetEnvSecret gets a single environment secret without revealing its encrypted value.
func (g *GitHubClient) GetEnvSecret(ctx context.Context, repoID int, env, secretName string) (*github.Secret, error) {
	secret, _, err := g.client.Actions.GetEnvSecret(ctx, repoID, env, secretName)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// CreateOrUpdateRepoSecret creates or updates a repository secret with an encrypted value.
func (g *GitHubClient) CreateOrUpdateRepoSecret(ctx context.Context, owner, repo string, eSecret *github.EncryptedSecret) error {
	_, err := g.client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repo, eSecret)
	return err
}

// CreateOrUpdateOrgSecret creates or updates an organization secret with an encrypted value.
func (g *GitHubClient) CreateOrUpdateOrgSecret(ctx context.Context, org string, eSecret *github.EncryptedSecret) error {
	_, err := g.client.Actions.CreateOrUpdateOrgSecret(ctx, org, eSecret)
	return err
}

// CreateOrUpdateEnvSecret creates or updates an environment secret with an encrypted value.
func (g *GitHubClient) CreateOrUpdateEnvSecret(ctx context.Context, repoID int, env string, eSecret *github.EncryptedSecret) error {
	_, err := g.client.Actions.CreateOrUpdateEnvSecret(ctx, repoID, env, eSecret)
	return err
}

// DeleteRepoSecret deletes a secret in a repository using the secret name.
func (g *GitHubClient) DeleteRepoSecret(ctx context.Context, owner, repo, name string) error {
	_, err := g.client.Actions.DeleteRepoSecret(ctx, owner, repo, name)
	return err
}

// DeleteOrgSecret deletes a secret in an organization using the secret name.
func (g *GitHubClient) DeleteOrgSecret(ctx context.Context, org, name string) error {
	_, err := g.client.Actions.DeleteOrgSecret(ctx, org, name)
	return err
}

// DeleteEnvSecret deletes a secret in an environment using the secret name.
func (g *GitHubClient) DeleteEnvSecret(ctx context.Context, repoID int, env, secretName string) error {
	_, err := g.client.Actions.DeleteEnvSecret(ctx, repoID, env, secretName)
	return err
}

// ListSelectedReposForOrgSecret lists all repositories that have access to an organization secret.
func (g *GitHubClient) ListSelectedReposForOrgSecret(ctx context.Context, org, name string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		result, resp, err := g.client.Actions.ListSelectedReposForOrgSecret(ctx, org, name, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, result.Repositories...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil
}

// SetSelectedReposForOrgSecret sets the repositories that have access to an organization secret.
func (g *GitHubClient) SetSelectedReposForOrgSecret(ctx context.Context, org, name string, ids github.SelectedRepoIDs) error {
	_, err := g.client.Actions.SetSelectedReposForOrgSecret(ctx, org, name, ids)
	return err
}

// AddSelectedRepoToOrgSecret adds a repository to an organization secret.
func (g *GitHubClient) AddSelectedRepoToOrgSecret(ctx context.Context, org, name string, repo *github.Repository) error {
	_, err := g.client.Actions.AddSelectedRepoToOrgSecret(ctx, org, name, repo)
	return err
}

// RemoveSelectedRepoFromOrgSecret removes a repository from an organization secret.
func (g *GitHubClient) RemoveSelectedRepoFromOrgSecret(ctx context.Context, org, name string, repo *github.Repository) error {
	_, err := g.client.Actions.RemoveSelectedRepoFromOrgSecret(ctx, org, name, repo)
	return err
}
