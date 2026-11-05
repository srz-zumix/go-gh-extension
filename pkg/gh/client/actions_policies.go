package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/google/go-github/v93/github"
)

const actionsPoliciesAPIVersion = "2026-03-10"

type actionsPolicyPage struct {
	TotalCount *int               `json:"total_count"`
	Policies   []*json.RawMessage `json:"policies"`
}

// listActionsPolicies preserves raw policies because the API can return rule
// types and conditions that are not yet represented by go-github.
func (g *GitHubClient) listActionsPolicies(ctx context.Context, path string) ([]json.RawMessage, error) {
	opts := &github.ActionsPolicyListOptions{
		HasParents: github.Ptr(true),
		ListOptions: github.ListOptions{
			Page: 1,
		},
	}
	expectedTotal := -1
	seenPages := make(map[int]bool)
	policies, err := paginate(ctx, &opts.ListOptions, 0, func(ctx context.Context) ([]*json.RawMessage, *github.Response, error) {
		if seenPages[opts.Page] {
			return nil, nil, fmt.Errorf("repeated Actions policies page %d", opts.Page)
		}
		seenPages[opts.Page] = true
		u, err := addOptions(path, opts)
		if err != nil {
			return nil, nil, err
		}
		req, err := g.client.NewRequest(ctx, "GET", u, nil, github.WithVersion(actionsPoliciesAPIVersion))
		if err != nil {
			return nil, nil, err
		}
		var page actionsPolicyPage
		resp, err := g.client.Do(req, &page)
		if err != nil {
			return nil, resp, err
		}
		if page.TotalCount == nil || *page.TotalCount < 0 || page.Policies == nil {
			return nil, resp, fmt.Errorf("invalid Actions policies response")
		}
		if expectedTotal >= 0 && expectedTotal != *page.TotalCount {
			return nil, resp, fmt.Errorf("actions policies changed during pagination")
		}
		expectedTotal = *page.TotalCount
		return page.Policies, resp, nil
	})
	if err != nil {
		return nil, err
	}
	if len(policies) != expectedTotal {
		return nil, fmt.Errorf("incomplete Actions policies response")
	}
	result := make([]json.RawMessage, 0, len(policies))
	for _, policy := range policies {
		if policy == nil {
			return nil, fmt.Errorf("invalid null Actions policy")
		}
		result = append(result, *policy)
	}
	return result, nil
}

func (g *GitHubClient) getActionsPolicy(ctx context.Context, path string) (json.RawMessage, error) {
	req, err := g.client.NewRequest(ctx, "GET", path, nil, github.WithVersion(actionsPoliciesAPIVersion))
	if err != nil {
		return nil, err
	}
	var policy json.RawMessage
	_, err = g.client.Do(req, &policy)
	if err != nil {
		return nil, err
	}
	if trimmed := bytes.TrimSpace(policy); len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, fmt.Errorf("invalid Actions policy response")
	}
	return policy, nil
}

// ListRepositoryActionsPolicies lists repository Actions policies, including applicable parent policies.
func (g *GitHubClient) ListRepositoryActionsPolicies(ctx context.Context, owner, repo string) ([]json.RawMessage, error) {
	path := fmt.Sprintf("repos/%s/%s/actions/policies", url.PathEscape(owner), url.PathEscape(repo))
	return g.listActionsPolicies(ctx, path)
}

// ListOrgActionsPolicies lists organization Actions policies, including applicable parent policies.
func (g *GitHubClient) ListOrgActionsPolicies(ctx context.Context, org string) ([]json.RawMessage, error) {
	path := fmt.Sprintf("orgs/%s/actions/policies", url.PathEscape(org))
	return g.listActionsPolicies(ctx, path)
}

// GetRepositoryActionsPolicy retrieves a repository Actions policy by ID.
func (g *GitHubClient) GetRepositoryActionsPolicy(ctx context.Context, owner, repo string, policyID int64) (json.RawMessage, error) {
	path := fmt.Sprintf("repos/%s/%s/actions/policies/%d", url.PathEscape(owner), url.PathEscape(repo), policyID)
	return g.getActionsPolicy(ctx, path)
}

// GetOrgActionsPolicy retrieves an organization Actions policy by ID.
func (g *GitHubClient) GetOrgActionsPolicy(ctx context.Context, org string, policyID int64) (json.RawMessage, error) {
	path := fmt.Sprintf("orgs/%s/actions/policies/%d", url.PathEscape(org), policyID)
	return g.getActionsPolicy(ctx, path)
}
