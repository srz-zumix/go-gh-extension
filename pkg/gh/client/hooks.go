package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

func (g *GitHubClient) ListOrgHooks(ctx context.Context, org string) ([]*github.Hook, error) {
	opt := &github.ListOptions{}
	return paginate(ctx, opt, 0, func(ctx context.Context) ([]*github.Hook, *github.Response, error) {
		return g.client.Organizations.ListHooks(ctx, org, opt)
	})
}

func (g *GitHubClient) ListRepoHooks(ctx context.Context, owner, repo string) ([]*github.Hook, error) {
	opt := &github.ListOptions{}
	return paginate(ctx, opt, 0, func(ctx context.Context) ([]*github.Hook, *github.Response, error) {
		return g.client.Repositories.ListHooks(ctx, owner, repo, opt)
	})
}
