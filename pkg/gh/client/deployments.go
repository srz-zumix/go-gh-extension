package client

import (
	"context"

	"github.com/google/go-github/v79/github"
)

// ListDeployments retrieves all deployments for a specific repository
func (g *GitHubClient) ListDeployments(ctx context.Context, owner string, repo string, opts *github.DeploymentsListOptions) ([]*github.Deployment, error) {
	var allDeployments []*github.Deployment

	if opts == nil {
		opts = &github.DeploymentsListOptions{
			ListOptions: github.ListOptions{PerPage: defaultPerPage},
		}
	} else if opts.PerPage == 0 {
		opts.PerPage = defaultPerPage
	}

	for {
		deployments, resp, err := g.client.Repositories.ListDeployments(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allDeployments = append(allDeployments, deployments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allDeployments, nil
}
