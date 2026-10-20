package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func distinctRunPage(first, count int) string {
	items := make([]string, count)
	for i := range items {
		items[i] = fmt.Sprintf(`{"id":%d}`, first+i)
	}
	return fmt.Sprintf(`{"workflow_runs":[%s]}`, strings.Join(items, ","))
}

func TestConcurrentWorkflowRunEndpointsPreserveOrderAndFilters(t *testing.T) {
	for _, endpoint := range []string{"repository", "workflow ID", "workflow file"} {
		t.Run(endpoint, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			release := make(chan struct{})
			releaseAll := sync.OnceFunc(func() { close(release) })
			defer releaseAll()
			started := make(chan int, 3)
			var active, peak atomic.Int32
			var mu sync.Mutex
			var calls []int
			tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				query := r.URL.Query()
				assert.Equal(t, "feature", query.Get("branch"))
				assert.Equal(t, "push", query.Get("event"))
				assert.Equal(t, "2026-10-01..2026-10-02", query.Get("created"))
				assert.Equal(t, "100", query.Get("per_page"))
				path := "/repos/example/project/actions/runs"
				switch endpoint {
				case "workflow ID":
					path = "/repos/example/project/actions/workflows/123/runs"
				case "workflow file":
					path = "/repos/example/project/actions/workflows/build.yml/runs"
				}
				assert.Equal(t, path, r.URL.Path)
				page, _ := strconv.Atoi(query.Get("page"))
				if page == 0 {
					page = 1
				}
				mu.Lock()
				calls = append(calls, page)
				mu.Unlock()
				if page > 1 {
					n := active.Add(1)
					defer active.Add(-1)
					for previous := peak.Load(); n > previous && !peak.CompareAndSwap(previous, n); previous = peak.Load() {
					}
					started <- page
					select {
					case <-release:
					case <-r.Context().Done():
						return nil, r.Context().Err()
					}
				}
				header := make(http.Header)
				header.Set("Link", fmt.Sprintf(`<%s?page=%d>; rel="next", <%s?page=8>; rel="last"`, path, page+1, path))
				return &http.Response{
					StatusCode: http.StatusOK, Header: header,
					Body: io.NopCloser(strings.NewReader(distinctRunPage((page-1)*100+1, 100))), Request: r,
				}, nil
			})
			client := newTestClient(t, "https://api.github.com/", tr)
			options := &github.ListWorkflowRunsOptions{Branch: "feature", Event: "push", Created: "2026-10-01..2026-10-02"}
			type result struct {
				rows []*github.WorkflowRun
				err  error
			}
			done := make(chan result, 1)
			go func() {
				var rows []*github.WorkflowRun
				var err error
				switch endpoint {
				case "repository":
					rows, err = client.ListRepositoryWorkflowRunsConcurrent(ctx, "example", "project", options, 350, 3)
				case "workflow ID":
					rows, err = client.ListWorkflowRunsByIDConcurrent(ctx, "example", "project", 123, options, 350, 3)
				case "workflow file":
					rows, err = client.ListWorkflowRunsByFileNameConcurrent(ctx, "example", "project", "build.yml", options, 350, 3)
				}
				done <- result{rows, err}
			}()
			for range 3 {
				select {
				case <-started:
				case <-ctx.Done():
					t.Fatal("the advertised pages were not fetched in parallel")
				}
			}
			releaseAll()
			got := <-done
			require.NoError(t, got.err)
			require.Len(t, got.rows, 350)
			for index, run := range got.rows {
				assert.Equal(t, int64(index+1), run.GetID())
			}
			assert.Equal(t, int32(3), peak.Load())
			slices.Sort(calls)
			assert.Equal(t, []int{1, 2, 3, 4}, calls, "do not fetch pages beyond the requested limit")
			assert.Equal(t, 0, options.Page, "request options must not be mutated")
		})
	}
}

func TestPaginateConcurrentBoundsAndNextOnlyFallback(t *testing.T) {
	for _, scenario := range []struct {
		name        string
		limit       int
		concurrency int
		start       int
		lastLink    bool
		wantRows    int
		wantPages   []int
	}{
		{"all pages", 0, 6, 0, true, 250, []int{1, 2, 3}},
		{"negative limit", -1, 6, 0, true, 250, []int{1, 2, 3}},
		{"short first page", 1, 6, 0, true, 1, []int{1}},
		{"full first page", 100, 6, 0, true, 100, []int{1}},
		{"partial final page", 150, 6, 0, true, 150, []int{1, 2}},
		{"limit beyond total", 500, 6, 0, true, 250, []int{1, 2, 3}},
		{"next only", 0, 6, 0, false, 250, []int{1, 2, 3}},
		{"sequential", 0, 1, 0, true, 250, []int{1, 2, 3}},
		{"default sequential", 0, 0, 0, true, 250, []int{1, 2, 3}},
		{"starting page", 150, 6, 2, true, 150, []int{2, 3}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var mu sync.Mutex
			var calls []int
			rows, err := paginateConcurrent(t.Context(), github.ListOptions{Page: scenario.start}, scenario.limit, scenario.concurrency,
				func(ctx context.Context, options github.ListOptions) ([]*int, *github.Response, error) {
					page := max(1, options.Page)
					mu.Lock()
					calls = append(calls, page)
					mu.Unlock()
					first := (page-1)*options.PerPage + 1
					items := make([]*int, max(0, min(options.PerPage, 251-first)))
					for index := range items {
						items[index] = new(first + index)
					}
					response := &github.Response{}
					if first+len(items) <= 250 {
						response.NextPage = page + 1
						if scenario.lastLink {
							response.LastPage = (249 / options.PerPage) + 1
						}
					}
					return items, response, nil
				})
			require.NoError(t, err)
			require.Len(t, rows, scenario.wantRows)
			offset := max(0, scenario.start-1) * 100
			for index, row := range rows {
				assert.Equal(t, offset+index+1, *row)
			}
			slices.Sort(calls)
			assert.Equal(t, scenario.wantPages, calls)
		})
	}
}

func TestPaginateConcurrentCancelsPeersAndPreservesRateLimitError(t *testing.T) {
	want := &github.RateLimitError{}
	peerStarted := make(chan struct{})
	peerCanceled := make(chan struct{})
	rows, err := paginateConcurrent(t.Context(), github.ListOptions{}, 0, 2,
		func(ctx context.Context, options github.ListOptions) ([]*int, *github.Response, error) {
			switch options.Page {
			case 0:
				return []*int{new(1)}, &github.Response{NextPage: 2, LastPage: 4}, nil
			case 2:
				select {
				case <-peerStarted:
				case <-ctx.Done():
					return nil, nil, ctx.Err()
				}
				return nil, nil, want
			case 3:
				close(peerStarted)
				<-ctx.Done()
				close(peerCanceled)
				return nil, nil, ctx.Err()
			default:
				t.Error("a page was started after failure")
				return nil, nil, errors.New("unexpected page")
			}
		})
	assert.ErrorIs(t, err, want)
	assert.Nil(t, rows, "failed pagination must not succeed with partial rows")
	select {
	case <-peerCanceled:
	default:
		t.Fatal("the in-flight page was not canceled")
	}
}

func TestPaginateConcurrentFollowsAnAdvertisedGrowingTail(t *testing.T) {
	var mu sync.Mutex
	var calls []int
	rows, err := paginateConcurrent(t.Context(), github.ListOptions{}, 0, 6,
		func(ctx context.Context, options github.ListOptions) ([]*int, *github.Response, error) {
			page := max(1, options.Page)
			mu.Lock()
			calls = append(calls, page)
			mu.Unlock()
			response := &github.Response{}
			if page < 4 {
				response.NextPage = page + 1
				response.LastPage = max(3, page+1)
			}
			return []*int{new(page)}, response, nil
		})
	require.NoError(t, err)
	require.Len(t, rows, 4)
	for index, row := range rows {
		assert.Equal(t, index+1, *row)
	}
	slices.Sort(calls)
	assert.Equal(t, []int{1, 2, 3, 4}, calls)
}
