package client

// GitHub Deploy Keys API functions
// See: https://docs.github.com/rest/deploy-keys/deploy-keys

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListDeployKeys lists all deploy keys for a repository.
func (g *GitHubClient) ListDeployKeys(ctx context.Context, owner, repo string) ([]*github.Key, error) {
	var allKeys []*github.Key
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		keys, resp, err := g.client.Repositories.ListKeys(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allKeys, nil
}

// GetDeployKey fetches a single deploy key by ID.
func (g *GitHubClient) GetDeployKey(ctx context.Context, owner, repo string, id int64) (*github.Key, error) {
	key, _, err := g.client.Repositories.GetKey(ctx, owner, repo, id)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// CreateDeployKey adds a deploy key to a repository.
func (g *GitHubClient) CreateDeployKey(ctx context.Context, owner, repo string, key *github.Key) (*github.Key, error) {
	k, _, err := g.client.Repositories.CreateKey(ctx, owner, repo, key)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// DeleteDeployKey removes a deploy key from a repository.
func (g *GitHubClient) DeleteDeployKey(ctx context.Context, owner, repo string, id int64) error {
	_, err := g.client.Repositories.DeleteKey(ctx, owner, repo, id)
	return err
}
