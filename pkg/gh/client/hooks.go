package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

func (g *GitHubClient) ListOrgHooks(ctx context.Context, org string) ([]*github.Hook, error) {
	var all []*github.Hook
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		hooks, resp, err := g.client.Organizations.ListHooks(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, hooks...)
		if resp.NextPage == 0 {
			return all, nil
		}
		opt.Page = resp.NextPage
	}
}

func (g *GitHubClient) ListRepoHooks(ctx context.Context, owner, repo string) ([]*github.Hook, error) {
	var all []*github.Hook
	opt := &github.ListOptions{PerPage: defaultPerPage}
	for {
		hooks, resp, err := g.client.Repositories.ListHooks(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		all = append(all, hooks...)
		if resp.NextPage == 0 {
			return all, nil
		}
		opt.Page = resp.NextPage
	}
}
