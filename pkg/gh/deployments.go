package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
)

type ListDeploymentsOptions struct {
	SHA         string
	Ref         string
	Task        string
	Environment string
}

func ListDeployments(ctx context.Context, g *GitHubClient, repo repository.Repository, options *ListDeploymentsOptions) ([]*github.Deployment, error) {
	var opts *github.DeploymentsListOptions
	if options != nil {
		opts = &github.DeploymentsListOptions{
			SHA:         options.SHA,
			Ref:         options.Ref,
			Task:        options.Task,
			Environment: options.Environment,
		}
	}
	return g.ListDeployments(ctx, repo.Owner, repo.Name, opts)
}

func ListEnvrionmentDeployments(ctx context.Context, g *GitHubClient, repo repository.Repository, environment string) ([]*github.Deployment, error) {
	opts := &github.DeploymentsListOptions{
		Environment: environment,
	}
	return g.ListDeployments(ctx, repo.Owner, repo.Name, opts)
}
