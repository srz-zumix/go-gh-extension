package client

// GitHub Dependabot Secrets API functions
// See: https://docs.github.com/rest/dependabot/secrets

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// ListDependabotRepoSecrets lists all Dependabot secrets in a repository without revealing their encrypted values.
func (g *GitHubClient) ListDependabotRepoSecrets(ctx context.Context, owner, repo string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		secrets, resp, err := g.client.Dependabot.ListRepoSecrets(ctx, owner, repo, opt)
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

// ListDependabotOrgSecrets lists all Dependabot secrets in an organization without revealing their encrypted values.
func (g *GitHubClient) ListDependabotOrgSecrets(ctx context.Context, org string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		secrets, resp, err := g.client.Dependabot.ListOrgSecrets(ctx, org, opt)
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

// GetDependabotRepoPublicKey gets the public key for encrypting Dependabot secrets in a repository.
func (g *GitHubClient) GetDependabotRepoPublicKey(ctx context.Context, owner, repo string) (*github.PublicKey, error) {
	key, _, err := g.client.Dependabot.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// CreateOrUpdateDependabotRepoSecret creates or updates a Dependabot secret in a repository with an encrypted value.
func (g *GitHubClient) CreateOrUpdateDependabotRepoSecret(ctx context.Context, owner, repo string, eSecret *github.EncryptedSecret) error {
	body := github.SecretRequest{KeyID: eSecret.KeyID, EncryptedValue: eSecret.EncryptedValue}
	_, err := g.client.Dependabot.CreateOrUpdateRepoSecret(ctx, owner, repo, eSecret.Name, body)
	return err
}

// DeleteDependabotRepoSecret deletes a Dependabot secret in a repository using the secret name.
func (g *GitHubClient) DeleteDependabotRepoSecret(ctx context.Context, owner, repo, name string) error {
	_, err := g.client.Dependabot.DeleteRepoSecret(ctx, owner, repo, name)
	return err
}

// ListSelectedReposForDependabotOrgSecret lists all repositories that have access to an organization Dependabot secret.
func (g *GitHubClient) ListSelectedReposForDependabotOrgSecret(ctx context.Context, org, name string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		result, resp, err := g.client.Dependabot.ListSelectedReposForOrgSecret(ctx, org, name, opt)
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

// SetSelectedReposForDependabotOrgSecret sets the repositories that have access to an organization Dependabot secret.
func (g *GitHubClient) SetSelectedReposForDependabotOrgSecret(ctx context.Context, org, name string, ids []int64) error {
	_, err := g.client.Dependabot.SetSelectedReposForOrgSecret(ctx, org, name, ids)
	return err
}
