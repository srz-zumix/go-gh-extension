package client

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func graphQLResponder(body string) http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})
}

func TestGetUserContributionsCollection_TruncatedWhenServerTotalExceedsList(t *testing.T) {
	// Server reports 3 contributing repositories but returns only 2 entries.
	body := `{"data":{"user":{"contributionsCollection":{
		"totalCommitContributions":10,
		"totalIssueContributions":0,
		"totalPullRequestContributions":0,
		"totalPullRequestReviewContributions":0,
		"totalRepositoriesWithContributedCommits":3,
		"commitContributionsByRepository":[
			{"repository":{"nameWithOwner":"o/a"},"contributions":{"totalCount":5}},
			{"repository":{"nameWithOwner":"o/b"},"contributions":{"totalCount":5}}
		],
		"contributionCalendar":{"totalContributions":10,"weeks":[]}
	}}}}`
	g := newTestClient(t, "https://api.github.com/", graphQLResponder(body))

	got, err := g.GetUserContributionsCollection(t.Context(), "octocat", time.Time{}, time.Time{})
	require.NoError(t, err)
	assert.Equal(t, 3, got.TotalRepositoriesWithContributedCommits)
	assert.Len(t, got.CommitContributionsByRepository, 2)
	assert.True(t, got.CommitContributionsByRepositoryTruncated)
}

func TestGetUserContributionsCollection_NotTruncatedWhenListMatchesTotal(t *testing.T) {
	body := `{"data":{"user":{"contributionsCollection":{
		"totalCommitContributions":10,
		"totalIssueContributions":0,
		"totalPullRequestContributions":0,
		"totalPullRequestReviewContributions":0,
		"totalRepositoriesWithContributedCommits":2,
		"commitContributionsByRepository":[
			{"repository":{"nameWithOwner":"o/a"},"contributions":{"totalCount":5}},
			{"repository":{"nameWithOwner":"o/b"},"contributions":{"totalCount":5}}
		],
		"contributionCalendar":{"totalContributions":10,"weeks":[]}
	}}}}`
	g := newTestClient(t, "https://api.github.com/", graphQLResponder(body))

	got, err := g.GetUserContributionsCollection(t.Context(), "octocat", time.Time{}, time.Time{})
	require.NoError(t, err)
	assert.Equal(t, 2, got.TotalRepositoriesWithContributedCommits)
	assert.False(t, got.CommitContributionsByRepositoryTruncated)
}

func TestGetUserContributionsCollection_NullUserReturnsNotFound(t *testing.T) {
	// A protocol-valid null user without a top-level error must not decode to a
	// zero-value success. The method returns an explicit not-found error instead.
	body := `{"data":{"user":null}}`
	g := newTestClient(t, "https://api.github.com/", graphQLResponder(body))

	got, err := g.GetUserContributionsCollection(t.Context(), "ghost", time.Time{}, time.Time{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "ghost")
	assert.Contains(t, err.Error(), "not found")
}
