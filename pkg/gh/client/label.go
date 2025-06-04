package client

import (
	"context"

	"github.com/google/go-github/v71/github"
)

func (g *GitHubClient) GetLabel(ctx context.Context, owner, repo, name string) (*github.Label, error) {
	label, _, err := g.client.Issues.GetLabel(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}
	return label, nil
}

func (g *GitHubClient) CreateLabel(ctx context.Context, owner, repo string, label *github.Label) (*github.Label, error) {
	label, _, err := g.client.Issues.CreateLabel(ctx, owner, repo, label)
	if err != nil {
		return nil, err
	}
	return label, nil
}

func (g *GitHubClient) DeleteLabel(ctx context.Context, owner, repo, name string) error {
	_, err := g.client.Issues.DeleteLabel(ctx, owner, repo, name)
	if err != nil {
		return err
	}
	return nil
}

func (g *GitHubClient) EditLabel(ctx context.Context, owner, repo, name string, label *github.Label) (*github.Label, error) {
	label, _, err := g.client.Issues.EditLabel(ctx, owner, repo, name, label)
	if err != nil {
		return nil, err
	}
	return label, nil
}

func (g *GitHubClient) ListLabels(ctx context.Context, owner, repo string) ([]*github.Label, error) {
	var allLabels []*github.Label
	opt := &github.ListOptions{PerPage: 50}

	for {
		labels, resp, err := g.client.Issues.ListLabels(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allLabels = append(allLabels, labels...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allLabels, nil
}
