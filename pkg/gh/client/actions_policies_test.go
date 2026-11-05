package client

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListActionsPoliciesPagination(t *testing.T) {
	tests := []struct {
		name string
		path string
		list func(*GitHubClient) ([]json.RawMessage, error)
	}{
		{
			name: "repository",
			path: "/repos/fixture-org/fixture-repo/actions/policies",
			list: func(g *GitHubClient) ([]json.RawMessage, error) {
				return g.ListRepositoryActionsPolicies(t.Context(), "fixture-org", "fixture-repo")
			},
		},
		{
			name: "organization",
			path: "/orgs/fixture-org/actions/policies",
			list: func(g *GitHubClient) ([]json.RawMessage, error) {
				return g.ListOrgActionsPolicies(t.Context(), "fixture-org")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pages []string
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tt.path, r.URL.Path)
				assert.Equal(t, "true", r.URL.Query().Get("has_parents"))
				assert.Equal(t, "100", r.URL.Query().Get("per_page"))
				assert.Equal(t, actionsPoliciesAPIVersion, r.Header.Get("X-GitHub-Api-Version"))
				page := r.URL.Query().Get("page")
				pages = append(pages, page)
				headers := make(http.Header)
				body := `{"total_count":2,"policies":[{"id":12,"rules":[{"type":"require_lockfile"}]}]}`
				if page == "1" {
					headers.Set("Link", `<https://api.example.test`+tt.path+`?page=2>; rel="next"`)
					body = `{"total_count":2,"policies":[{"id":11,"conditions":{"future_selector":{"include":["fixture"]}}}]}`
				}
				return &http.Response{StatusCode: http.StatusOK, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
			})
			g := newTestClient(t, "https://api.example.test/", transport)
			policies, err := tt.list(g)
			require.NoError(t, err)
			require.Equal(t, []string{"1", "2"}, pages)
			require.Len(t, policies, 2)
			assert.JSONEq(t, `{"id":11,"conditions":{"future_selector":{"include":["fixture"]}}}`, string(policies[0]))
			assert.JSONEq(t, `{"id":12,"rules":[{"type":"require_lockfile"}]}`, string(policies[1]))
		})
	}
}

func TestListActionsPoliciesRejectsIncompleteResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
		link string
	}{
		{"missing count", `{"policies":[]}`, ""},
		{"missing policies", `{"total_count":0}`, ""},
		{"null policies", `{"total_count":0,"policies":null}`, ""},
		{"null policy", `{"total_count":1,"policies":[null]}`, ""},
		{"incomplete page", `{"total_count":2,"policies":[{"id":11}]}`, ""},
		{"repeated page", `{"total_count":1,"policies":[{"id":11}]}`, `<https://api.example.test/repos/fixture-org/fixture-repo/actions/policies?page=1>; rel="next"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := 0
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				requests++
				headers := make(http.Header)
				if tt.link != "" {
					headers.Set("Link", tt.link)
				}
				return &http.Response{StatusCode: http.StatusOK, Header: headers, Body: io.NopCloser(strings.NewReader(tt.body)), Request: r}, nil
			})
			g := newTestClient(t, "https://api.example.test/", transport)
			policies, err := g.ListRepositoryActionsPolicies(t.Context(), "fixture-org", "fixture-repo")
			require.Error(t, err)
			assert.Nil(t, policies)
			assert.Equal(t, 1, requests)
		})
	}
}

func TestListActionsPoliciesPropagatesLaterPageError(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		headers := make(http.Header)
		if r.URL.Query().Get("page") == "1" {
			headers.Set("Link", `<https://api.example.test/repos/fixture-org/fixture-repo/actions/policies?page=2>; rel="next"`)
			return &http.Response{StatusCode: http.StatusOK, Header: headers, Body: io.NopCloser(strings.NewReader(`{"total_count":2,"policies":[{"id":11}]}`)), Request: r}, nil
		}
		return &http.Response{StatusCode: http.StatusForbidden, Header: headers, Body: io.NopCloser(strings.NewReader(`{"message":"Forbidden"}`)), Request: r}, nil
	})
	g := newTestClient(t, "https://api.example.test/", transport)
	policies, err := g.ListRepositoryActionsPolicies(t.Context(), "fixture-org", "fixture-repo")
	require.Error(t, err)
	assert.Nil(t, policies)
	var responseErr *github.ErrorResponse
	require.True(t, errors.As(err, &responseErr))
	assert.Equal(t, http.StatusForbidden, responseErr.Response.StatusCode)
}

func TestListActionsPoliciesRejectsChangedTotal(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		headers := make(http.Header)
		body := `{"total_count":3,"policies":[{"id":12}]}`
		if r.URL.Query().Get("page") == "1" {
			headers.Set("Link", `<https://api.example.test/repos/fixture-org/fixture-repo/actions/policies?page=2>; rel="next"`)
			body = `{"total_count":2,"policies":[{"id":11}]}`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	g := newTestClient(t, "https://api.example.test/", transport)
	policies, err := g.ListRepositoryActionsPolicies(t.Context(), "fixture-org", "fixture-repo")
	require.Error(t, err)
	assert.Nil(t, policies)
}

func TestGetActionsPolicies(t *testing.T) {
	tests := []struct {
		name string
		path string
		get  func(*GitHubClient) (json.RawMessage, error)
	}{
		{
			name: "repository",
			path: "/repos/fixture-org/fixture-repo/actions/policies/11",
			get: func(g *GitHubClient) (json.RawMessage, error) {
				return g.GetRepositoryActionsPolicy(t.Context(), "fixture-org", "fixture-repo", 11)
			},
		},
		{
			name: "organization",
			path: "/orgs/fixture-org/actions/policies/11",
			get: func(g *GitHubClient) (json.RawMessage, error) {
				return g.GetOrgActionsPolicy(t.Context(), "fixture-org", 11)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tt.path, r.URL.Path)
				assert.Equal(t, actionsPoliciesAPIVersion, r.Header.Get("X-GitHub-Api-Version"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"id":11,"rules":[{"type":"require_lockfile","parameters":{"future":true}}]}`)),
					Request:    r,
				}, nil
			})
			g := newTestClient(t, "https://api.example.test/", transport)
			policy, err := tt.get(g)
			require.NoError(t, err)
			assert.JSONEq(t, `{"id":11,"rules":[{"type":"require_lockfile","parameters":{"future":true}}]}`, string(policy))
		})
	}
}
