package gh

import (
	"context"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// CommitAuthor identifies the author or committer of a commit.
// All fields are optional.
type CommitAuthor struct {
	Date  *time.Time
	Name  *string
	Email *string
	Login *string
}

// toGitHubCommitAuthor converts CommitAuthor to github.CommitAuthor.
func toGitHubCommitAuthor(a *CommitAuthor) *github.CommitAuthor {
	if a == nil {
		return nil
	}
	r := &github.CommitAuthor{
		Name:  a.Name,
		Email: a.Email,
		Login: a.Login,
	}
	if a.Date != nil {
		r.Date = &github.Timestamp{Time: *a.Date}
	}
	return r
}

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

// GetLatestCommitForPath returns the most recent commit that touched filePath in the given repository.
// ref is the branch/tag/SHA to start from; an empty string uses the default branch.
func GetLatestCommitForPath(ctx context.Context, g *GitHubClient, repo repository.Repository, filePath, ref string) (*github.RepositoryCommit, error) {
	return g.GetLatestCommitForPath(ctx, repo.Owner, repo.Name, filePath, ref)
}

func ListBranchesHeadCommit(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) ([]*github.BranchCommit, error) {
	return g.ListBranchesHeadCommit(ctx, repo.Owner, repo.Name, sha)
}

func GetCommit(ctx context.Context, g *GitHubClient, repo repository.Repository, sha string) (*github.RepositoryCommit, error) {
	return g.GetCommit(ctx, repo.Owner, repo.Name, sha)
}

func GetCommitSHA1(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string) (string, error) {
	return g.GetCommitSHA1(ctx, repo.Owner, repo.Name, ref, "")
}

func CompareCommits(ctx context.Context, g *GitHubClient, repo repository.Repository, base, head string) (*github.CommitsComparison, error) {
	return g.CompareCommits(ctx, repo.Owner, repo.Name, base, head)
}
