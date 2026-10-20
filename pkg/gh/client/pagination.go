package client

import (
	"context"
	"sync"

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

// paginateConcurrent only parallelizes pages advertised by the first response.
// Servers without a last-page link retain sequential next-page traversal.
func paginateConcurrent[T any](ctx context.Context, options github.ListOptions, limit, concurrency int, fetch func(context.Context, github.ListOptions) ([]*T, *github.Response, error)) ([]*T, error) {
	if concurrency <= 1 {
		return paginate(ctx, &options, limit, func(ctx context.Context) ([]*T, *github.Response, error) {
			return fetch(ctx, options)
		})
	}
	options.PerPage = defaultPerPage
	if limit > 0 && limit < options.PerPage {
		options.PerPage = limit
	}
	first, response, err := fetch(ctx, options)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(first) >= limit {
		return first[:limit], nil
	}
	if response.NextPage == 0 {
		return first, nil
	}
	if response.LastPage < response.NextPage {
		for response.NextPage != 0 {
			options.Page = response.NextPage
			var rows []*T
			rows, response, err = fetch(ctx, options)
			if err != nil {
				return nil, err
			}
			first = append(first, rows...)
			if limit > 0 && len(first) >= limit {
				return first[:limit], nil
			}
		}
		return first, nil
	}
	next, last := response.NextPage, response.LastPage
	if limit > 0 {
		pagesNeeded := (limit-len(first)-1)/options.PerPage + 1
		last = min(last, next+pagesNeeded-1)
	}
	pages := make([][]*T, last-next+1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	work := make(chan int)
	var workers sync.WaitGroup
	var failure sync.Once
	var firstError error
	var tail *github.Response
	for range min(concurrency, len(pages)) {
		workers.Go(func() {
			for page := range work {
				if ctx.Err() != nil {
					return
				}
				pageOptions := options
				pageOptions.Page = page
				rows, response, err := fetch(ctx, pageOptions)
				if err != nil {
					failure.Do(func() {
						firstError = err
						cancel()
					})
					return
				}
				pages[page-next] = rows
				if page == last {
					tail = response
				}
			}
		})
	}
dispatch:
	for page := next; page <= last; page++ {
		select {
		case work <- page:
		case <-ctx.Done():
			break dispatch
		}
	}
	close(work)
	workers.Wait()
	if firstError != nil {
		return nil, firstError
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, rows := range pages {
		first = append(first, rows...)
	}
	// Follow a tail that grew after the initial response without guessing new pages.
	for tail.NextPage != 0 && (limit <= 0 || len(first) < limit) {
		options.Page = tail.NextPage
		var rows []*T
		rows, tail, err = fetch(ctx, options)
		if err != nil {
			return nil, err
		}
		first = append(first, rows...)
	}
	if limit > 0 && len(first) > limit {
		first = first[:limit]
	}
	return first, nil
}
