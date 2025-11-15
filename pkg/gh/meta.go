package gh

import (
	"context"

	"github.com/google/go-github/v79/github"
)

func GetMeta(ctx context.Context, g *GitHubClient) (*github.APIMeta, error) {
	return g.GetMeta(ctx)
}

func GetZen(ctx context.Context, g *GitHubClient) (string, error) {
	return g.GetZen(ctx)
}

func GetOctocat(ctx context.Context, g *GitHubClient, message string) (string, error) {
	return g.GetOctocat(ctx, message)
}
