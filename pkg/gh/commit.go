package gh

import (
	"context"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
)

type ListCommitsOptions interface {
	ToCommitListOptions(*github.CommitsListOptions) *github.CommitsListOptions
}

type ListCommitsShaOption struct {
	Sha string
}

func (o ListCommitsShaOption) ToCommitListOptions(opts *github.CommitsListOptions) *github.CommitsListOptions {
	opts.SHA = o.Sha
	return opts
}

type ListCommitsAuthorOption struct {
	Author string
}

func (o ListCommitsAuthorOption) ToCommitListOptions(opts *github.CommitsListOptions) *github.CommitsListOptions {
	opts.Author = o.Author
	return opts
}

type ListCommitsSinceOption struct {
	Since time.Time
}

func (o ListCommitsSinceOption) ToCommitListOptions(opts *github.CommitsListOptions) *github.CommitsListOptions {
	opts.Since = o.Since
	return opts
}

type ListCommitsUntilOption struct {
	Until time.Time
}

func (o ListCommitsUntilOption) ToCommitListOptions(opts *github.CommitsListOptions) *github.CommitsListOptions {
	opts.Until = o.Until
	return opts
}

type ListCommitsPathOption struct {
	Path string
}

func (o ListCommitsPathOption) ToCommitListOptions(opts *github.CommitsListOptions) *github.CommitsListOptions {
	opts.Path = o.Path
	return opts
}

func ListCommits(ctx context.Context, g *GitHubClient, repo repository.Repository, opts ...ListCommitsOptions) ([]*github.RepositoryCommit, error) {
	options := &github.CommitsListOptions{}
	for _, opt := range opts {
		options = opt.ToCommitListOptions(options)
	}
	return g.ListCommits(ctx, repo.Owner, repo.Name, options)
}

func ListBranchesHeadCommit(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) ([]*github.BranchCommit, error) {
	return g.ListBranchesHeadCommit(ctx, repo.Owner, repo.Name, sha)
}

func GetCommit(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) (*github.RepositoryCommit, error) {
	return g.GetCommit(ctx, repo.Owner, repo.Name, sha)
}

func CompareCommits(ctx context.Context, g *GitHubClient, repo repository.Repository, base, head string) (*github.CommitsComparison, error) {
	return g.CompareCommits(ctx, repo.Owner, repo.Name, base, head)
}
