package client

import (
	"context"

	"github.com/google/go-github/v73/github"
)

// GetMeta retrieves GitHub meta information
func (g *GitHubClient) GetMeta(ctx context.Context) (*github.APIMeta, error) {
	meta, _, err := g.client.Meta.Get(ctx)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// GetZen retrieves a random Zen quote from GitHub
func (g *GitHubClient) GetZen(ctx context.Context) (string, error) {
	zen, _, err := g.client.Meta.Zen(ctx)
	if err != nil {
		return "", err
	}

	return zen, nil
}

// GetOctocat retrieves ASCII art of the Octocat with an optional message
func (g *GitHubClient) GetOctocat(ctx context.Context, message string) (string, error) {
	octocat, _, err := g.client.Meta.Octocat(ctx, message)
	if err != nil {
		return "", err
	}

	return octocat, nil
}
