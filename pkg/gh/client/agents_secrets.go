package client

// GitHub Copilot Agents Secrets API functions.
// go-github does not cover these endpoints yet, so they are called directly.

import (
	"context"
	"fmt"

	"github.com/google/go-github/v90/github"
)

// agentsSecrets is the response of the Agents secrets list endpoints.
type agentsSecrets struct {
	TotalCount int              `json:"total_count"`
	Secrets    []*github.Secret `json:"secrets"`
}

// ListAgentsRepoSecrets lists all Agents secrets in a repository without revealing their encrypted values.
func (g *GitHubClient) ListAgentsRepoSecrets(ctx context.Context, owner, repo string) ([]*github.Secret, error) {
	return g.listAgentsSecrets(ctx, fmt.Sprintf("repos/%s/%s/agents/secrets", owner, repo))
}

// ListAgentsOrgSecrets lists all Agents secrets in an organization without revealing their encrypted values.
func (g *GitHubClient) ListAgentsOrgSecrets(ctx context.Context, org string) ([]*github.Secret, error) {
	return g.listAgentsSecrets(ctx, fmt.Sprintf("orgs/%s/agents/secrets", org))
}

func (g *GitHubClient) listAgentsSecrets(ctx context.Context, path string) ([]*github.Secret, error) {
	var allSecrets []*github.Secret
	page := 1
	for {
		u := fmt.Sprintf("%s?per_page=%d&page=%d", path, defaultPerPage, page)
		req, err := g.client.NewRequest(ctx, "GET", u, nil)
		if err != nil {
			return nil, err
		}
		result := new(agentsSecrets)
		resp, err := g.client.Do(req, result)
		if err != nil {
			return nil, err
		}
		allSecrets = append(allSecrets, result.Secrets...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return allSecrets, nil
}

// GetAgentsRepoPublicKey gets the public key used to encrypt Agents secrets of a repository.
func (g *GitHubClient) GetAgentsRepoPublicKey(ctx context.Context, owner, repo string) (*github.PublicKey, error) {
	u := fmt.Sprintf("repos/%s/%s/agents/secrets/public-key", owner, repo)
	req, err := g.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	key := new(github.PublicKey)
	if _, err := g.client.Do(req, key); err != nil {
		return nil, err
	}
	return key, nil
}

// CreateOrUpdateAgentsRepoSecret creates or updates an Agents secret of a repository.
func (g *GitHubClient) CreateOrUpdateAgentsRepoSecret(ctx context.Context, owner, repo string, eSecret *github.EncryptedSecret) error {
	u := fmt.Sprintf("repos/%s/%s/agents/secrets/%s", owner, repo, eSecret.Name)
	req, err := g.client.NewRequest(ctx, "PUT", u, eSecret)
	if err != nil {
		return err
	}
	_, err = g.client.Do(req, nil)
	return err
}

// DeleteAgentsRepoSecret deletes an Agents secret of a repository.
func (g *GitHubClient) DeleteAgentsRepoSecret(ctx context.Context, owner, repo, name string) error {
	u := fmt.Sprintf("repos/%s/%s/agents/secrets/%s", owner, repo, name)
	req, err := g.client.NewRequest(ctx, "DELETE", u, nil)
	if err != nil {
		return err
	}
	_, err = g.client.Do(req, nil)
	return err
}
