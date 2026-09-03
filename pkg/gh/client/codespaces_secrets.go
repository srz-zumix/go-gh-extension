package client

// GitHub Codespaces Secrets API functions
// See: https://docs.github.com/rest/codespaces/secrets

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// ListCodespacesRepoSecrets lists all development environment secrets in a repository without revealing their encrypted values.
func (g *GitHubClient) ListCodespacesRepoSecrets(ctx context.Context, owner, repo string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		secrets, resp, err := g.client.Codespaces.ListRepoSecrets(ctx, owner, repo, opt)
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

// ListCodespacesOrgSecrets lists all development environment secrets in an organization without revealing their encrypted values.
func (g *GitHubClient) ListCodespacesOrgSecrets(ctx context.Context, org string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		secrets, resp, err := g.client.Codespaces.ListOrgSecrets(ctx, org, opt)
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

// ListCodespacesUserSecrets lists all development environment secrets of the authenticated user without revealing their encrypted values.
func (g *GitHubClient) ListCodespacesUserSecrets(ctx context.Context) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		secrets, resp, err := g.client.Codespaces.ListUserSecrets(ctx, opt)
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
