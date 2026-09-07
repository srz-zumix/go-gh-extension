package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runsPage renders a workflow runs response body holding count runs.
func runsPage(count int) string {
	items := make([]string, 0, count)
	for i := range count {
		items = append(items, fmt.Sprintf(`{"id":%d}`, i+1))
	}
	return fmt.Sprintf(`{"total_count":%d,"workflow_runs":[%s]}`, count, strings.Join(items, ","))
}

func TestListRepositoryWorkflowRunsLimit(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		pageSizes     []int
		wantCount     int
		wantPerPages  []string
		wantPageCalls int
	}{
		{
			name:          "no limit collects every page",
			limit:         0,
			pageSizes:     []int{100, 100, 20},
			wantCount:     220,
			wantPerPages:  []string{"100", "100", "100"},
			wantPageCalls: 3,
		},
		{
			name:          "negative limit behaves as no limit",
			limit:         -1,
			pageSizes:     []int{100, 5},
			wantCount:     105,
			wantPerPages:  []string{"100", "100"},
			wantPageCalls: 2,
		},
		{
			name:          "limit below per page requests a single short page",
			limit:         1,
			pageSizes:     []int{1},
			wantCount:     1,
			wantPerPages:  []string{"1"},
			wantPageCalls: 1,
		},
		{
			name:          "limit stops pagination early and truncates",
			limit:         150,
			pageSizes:     []int{100, 100, 100},
			wantCount:     150,
			wantPerPages:  []string{"100", "100"},
			wantPageCalls: 2,
		},
		{
			name:          "limit larger than the total returns everything",
			limit:         500,
			pageSizes:     []int{100, 30},
			wantCount:     130,
			wantPerPages:  []string{"100", "100"},
			wantPageCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPerPages []string
			call := 0
			tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				gotPerPages = append(gotPerPages, r.URL.Query().Get("per_page"))
				require.Less(t, call, len(tt.pageSizes), "unexpected extra request")
				size := tt.pageSizes[call]
				call++

				header := make(http.Header)
				if call < len(tt.pageSizes) {
					header.Set("Link", fmt.Sprintf(`<%s?page=%d>; rel="next"`, r.URL.Path, call+1))
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(strings.NewReader(runsPage(size))),
					Request:    r,
				}, nil
			})
			g := newTestClient(t, "https://api.github.com/", tr)

			runs, err := g.ListRepositoryWorkflowRuns(t.Context(), "owner", "repo", nil, tt.limit)
			require.NoError(t, err)
			assert.Len(t, runs, tt.wantCount)
			assert.Equal(t, tt.wantPageCalls, call)
			assert.Equal(t, tt.wantPerPages, gotPerPages)
		})
	}
}

func TestListRepositoryWorkflowRunsKeepsFilters(t *testing.T) {
	var gotQuery string
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotQuery = r.URL.RawQuery
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(runsPage(1))),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	_, err := g.ListRepositoryWorkflowRuns(t.Context(), "owner", "repo", &github.ListWorkflowRunsOptions{
		Branch:  "main",
		Created: ">=2026-08-31",
	}, 0)
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "branch=main")
	assert.Contains(t, gotQuery, "per_page=100")
}

// jobsPage renders a workflow jobs response body holding jobs with the given IDs.
func jobsPage(totalCount int, ids ...int64) string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, fmt.Sprintf(`{"id":%d}`, id))
	}
	return fmt.Sprintf(`{"total_count":%d,"jobs":[%s]}`, totalCount, strings.Join(items, ","))
}

func TestListWorkflowJobsLimit(t *testing.T) {
	// Two pages with distinct jobs and a consistent total_count so the fixture
	// stays realistic; limit=3 must accumulate across pages and truncate.
	pages := []string{
		jobsPage(4, 1, 2),
		jobsPage(4, 3, 4),
	}
	call := 0
	tr := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		require.Less(t, call, len(pages), "unexpected extra request")
		body := pages[call]
		call++
		header := make(http.Header)
		if call < len(pages) {
			header.Set("Link", fmt.Sprintf(`<%s?page=%d>; rel="next"`, r.URL.Path, call+1))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     header,
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})
	g := newTestClient(t, "https://api.github.com/", tr)

	jobs, err := g.ListWorkflowJobs(t.Context(), "owner", "repo", 7, nil, 3)
	require.NoError(t, err)
	assert.Equal(t, 2, call)

	gotIDs := make([]int64, 0, len(jobs))
	for _, job := range jobs {
		gotIDs = append(gotIDs, job.GetID())
	}
	assert.Equal(t, []int64{1, 2, 3}, gotIDs)
}
