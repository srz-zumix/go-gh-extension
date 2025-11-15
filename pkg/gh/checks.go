package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

type ListChecksRunFilterOptions struct {
	CheckName *string
	Status    *string
	Filter    *string
	AppID     *int64
}

func ListCheckRunsForRef(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string, options *ListChecksRunFilterOptions) (*github.ListCheckRunsResults, error) {
	var opt *github.ListCheckRunsOptions
	if options != nil {
		opt = &github.ListCheckRunsOptions{
			CheckName: options.CheckName,
			Status:    options.Status,
			Filter:    options.Filter,
			AppID:     options.AppID,
		}
	}
	return g.ListCheckRunsForRef(ctx, repo.Owner, repo.Name, ref, opt)
}
