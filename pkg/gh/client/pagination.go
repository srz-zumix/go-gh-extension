package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// paginate collects every page of a list API into a single slice.
// A limit of 0 or less collects all pages; otherwise pagination stops as soon as
// limit items are gathered and the result is truncated to exactly limit items.
// listOpts is mutated to advance pages, so callers must pass the ListOptions that
// fetch reads from.
func paginate[T any](ctx context.Context, listOpts *github.ListOptions, limit int, fetch func(context.Context) ([]*T, *github.Response, error)) ([]*T, error) {
	listOpts.PerPage = defaultPerPage
	if limit > 0 && limit < defaultPerPage {
		listOpts.PerPage = limit
	}

	var all []*T
	for {
		items, resp, err := fetch(ctx)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if limit > 0 && len(all) >= limit {
			return all[:limit], nil
		}
		if resp.NextPage == 0 {
			return all, nil
		}
		listOpts.Page = resp.NextPage
	}
}
