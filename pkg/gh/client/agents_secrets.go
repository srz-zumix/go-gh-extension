package client

// GitHub Copilot Agents Secrets API functions, backed by go-github's AgentsService.

import (
	"context"

	"github.com/google/go-github/v93/github"
)

// collectAgentsPages calls fetch for each page of results, starting at the first page,
// and passes every page to collect until the last page has been fetched.
func collectAgentsPages[T any](fetch func(opts *github.ListOptions) (T, *github.Response, error), collect func(page T)) error {
	opts := &github.ListOptions{PerPage: defaultPerPage, Page: 1}
	for {
		page, resp, err := fetch(opts)
		if err != nil {
			return err
		}
		collect(page)
		if resp.NextPage == 0 {
			return nil
		}
		opts.Page = resp.NextPage
	}
}

// ListAgentsRepoSecrets lists all Agents secrets in a repository without revealing their encrypted values.
func (g *GitHubClient) ListAgentsRepoSecrets(ctx context.Context, owner, repo string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	err := collectAgentsPages(func(opts *github.ListOptions) (*github.Secrets, *github.Response, error) {
		return g.client.Agents.ListRepoSecrets(ctx, owner, repo, opts)
	}, func(page *github.Secrets) {
		if page != nil {
			allSecrets = append(allSecrets, page.Secrets...)
		}
	})
	if err != nil {
		return nil, err
	}
	return allSecrets, nil
}

// ListAgentsOrgSecrets lists all Agents secrets in an organization without revealing their encrypted values.
func (g *GitHubClient) ListAgentsOrgSecrets(ctx context.Context, org string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	err := collectAgentsPages(func(opts *github.ListOptions) (*github.Secrets, *github.Response, error) {
		return g.client.Agents.ListOrgSecrets(ctx, org, opts)
	}, func(page *github.Secrets) {
		if page != nil {
			allSecrets = append(allSecrets, page.Secrets...)
		}
	})
	if err != nil {
		return nil, err
	}
	return allSecrets, nil
}

// GetAgentsRepoPublicKey gets the public key used to encrypt Agents secrets of a repository.
func (g *GitHubClient) GetAgentsRepoPublicKey(ctx context.Context, owner, repo string) (*github.PublicKey, error) {
	key, _, err := g.client.Agents.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return new(github.PublicKey), nil
	}
	return key, nil
}

// CreateOrUpdateAgentsRepoSecret creates or updates an Agents secret of a repository.
func (g *GitHubClient) CreateOrUpdateAgentsRepoSecret(ctx context.Context, owner, repo string, eSecret *github.EncryptedSecret) error {
	body := github.SecretRequest{KeyID: eSecret.KeyID, EncryptedValue: eSecret.EncryptedValue}
	_, err := g.client.Agents.CreateOrUpdateRepoSecret(ctx, owner, repo, eSecret.Name, body)
	return err
}

// DeleteAgentsRepoSecret deletes an Agents secret of a repository.
func (g *GitHubClient) DeleteAgentsRepoSecret(ctx context.Context, owner, repo, name string) error {
	_, err := g.client.Agents.DeleteRepoSecret(ctx, owner, repo, name)
	return err
}

// ListSelectedReposForAgentsOrgSecret lists all repositories that have access to an organization Agents secret.
func (g *GitHubClient) ListSelectedReposForAgentsOrgSecret(ctx context.Context, org, name string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	err := collectAgentsPages(func(opts *github.ListOptions) (*github.SelectedReposList, *github.Response, error) {
		return g.client.Agents.ListSelectedReposForOrgSecret(ctx, org, name, opts)
	}, func(page *github.SelectedReposList) {
		if page != nil {
			allRepos = append(allRepos, page.Repositories...)
		}
	})
	if err != nil {
		return nil, err
	}
	return allRepos, nil
}

// SetSelectedReposForAgentsOrgSecret sets the repositories that have access to an organization Agents secret.
func (g *GitHubClient) SetSelectedReposForAgentsOrgSecret(ctx context.Context, org, name string, ids github.SelectedRepoIDs) error {
	_, err := g.client.Agents.SetSelectedReposForOrgSecret(ctx, org, name, ids)
	return err
}
