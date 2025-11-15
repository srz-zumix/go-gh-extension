package client

import (
	"context"
	"fmt"

	"github.com/google/go-github/v79/github"
)

// RuleSuite represents a rule suite evaluation
// Note: This is not yet part of the official go-github library
// Based on GitHub REST API: GET /repos/{owner}/{repo}/rulesets/rule-suites
type RuleSuite struct {
	ID               *int64                     `json:"id,omitempty"`
	ActorID          *int64                     `json:"actor_id,omitempty"`
	ActorName        *string                    `json:"actor_name,omitempty"`
	BeforeSHA        *string                    `json:"before_sha,omitempty"`
	AfterSHA         *string                    `json:"after_sha,omitempty"`
	Ref              *string                    `json:"ref,omitempty"`
	RepositoryID     *int64                     `json:"repository_id,omitempty"`
	RepositoryName   *string                    `json:"repository_name,omitempty"`
	PushedAt         *github.Timestamp          `json:"pushed_at,omitempty"`
	Result           *string                    `json:"result,omitempty"`
	EvaluationResult *string                    `json:"evaluation_result,omitempty"`
	RuleEvaluations  []*RuleSuiteRuleEvaluation `json:"rule_evaluations,omitempty"`
}

// RuleSuiteRuleEvaluation represents a single rule evaluation within a rule suite
type RuleSuiteRuleEvaluation struct {
	RuleSource      *RuleSuiteRuleSource `json:"rule_source,omitempty"`
	RuleType        *string              `json:"rule_type,omitempty"`
	Result          *string              `json:"result,omitempty"`
	EnforcementMode *string              `json:"enforcement_mode,omitempty"`
	Details         *string              `json:"details,omitempty"`
}

// RuleSuiteRuleSource represents the source of a rule
type RuleSuiteRuleSource struct {
	Type *string `json:"type,omitempty"`
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// ListRuleSuitesOptions specifies the optional parameters to the
// RepositoriesService.ListRuleSuites method
type ListRuleSuitesOptions struct {
	Ref             string `url:"ref,omitempty"`
	TimePeriod      string `url:"time_period,omitempty"`
	ActorName       string `url:"actor_name,omitempty"`
	RuleSuiteResult string `url:"rule_suite_result,omitempty"`
	github.ListOptions
}

// ListRepositoryRuleSuites retrieves all rule suites for a specific repository
func (g *GitHubClient) ListRepositoryRuleSuites(ctx context.Context, owner string, repo string, opts *ListRuleSuitesOptions) ([]*RuleSuite, error) {
	allRuleSuites := []*RuleSuite{}
	opt := &ListRuleSuitesOptions{
		ListOptions: github.ListOptions{
			PerPage: defaultPerPage,
		},
	}
	if opts != nil {
		opt = opts
		if opt.PerPage == 0 {
			opt.PerPage = defaultPerPage
		}
	}

	for {
		u, err := addOptions(fmt.Sprintf("repos/%s/%s/rulesets/rule-suites", owner, repo), opt)
		if err != nil {
			return nil, err
		}

		req, err := g.client.NewRequest("GET", u, nil)
		if err != nil {
			return nil, err
		}

		var ruleSuites []*RuleSuite
		resp, err := g.client.Do(ctx, req, &ruleSuites)
		if err != nil {
			return nil, err
		}

		allRuleSuites = append(allRuleSuites, ruleSuites...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRuleSuites, nil
}

// GetRepositoryRuleSuite retrieves a single rule suite for a specific repository by rule suite ID
func (g *GitHubClient) GetRepositoryRuleSuite(ctx context.Context, owner string, repo string, ruleSuiteID int64) (*RuleSuite, error) {
	u := fmt.Sprintf("repos/%s/%s/rulesets/rule-suites/%d", owner, repo, ruleSuiteID)

	req, err := g.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var ruleSuite *RuleSuite
	_, err = g.client.Do(ctx, req, &ruleSuite)
	if err != nil {
		return nil, err
	}

	return ruleSuite, nil
}

// ListOrgRuleSuites retrieves all rule suites for a specific organization
func (g *GitHubClient) ListOrgRuleSuites(ctx context.Context, org string, opts *ListRuleSuitesOptions) ([]*RuleSuite, error) {
	allRuleSuites := []*RuleSuite{}
	opt := &ListRuleSuitesOptions{
		ListOptions: github.ListOptions{
			PerPage: defaultPerPage,
		},
	}
	if opts != nil {
		opt = opts
		if opt.PerPage == 0 {
			opt.PerPage = defaultPerPage
		}
	}

	for {
		u, err := addOptions(fmt.Sprintf("orgs/%s/rulesets/rule-suites", org), opt)
		if err != nil {
			return nil, err
		}

		req, err := g.client.NewRequest("GET", u, nil)
		if err != nil {
			return nil, err
		}

		var ruleSuites []*RuleSuite
		resp, err := g.client.Do(ctx, req, &ruleSuites)
		if err != nil {
			return nil, err
		}

		allRuleSuites = append(allRuleSuites, ruleSuites...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRuleSuites, nil
}

// GetOrgRuleSuite retrieves a single rule suite for a specific organization by rule suite ID
func (g *GitHubClient) GetOrgRuleSuite(ctx context.Context, org string, ruleSuiteID int64) (*RuleSuite, error) {
	u := fmt.Sprintf("orgs/%s/rulesets/rule-suites/%d", org, ruleSuiteID)

	req, err := g.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var ruleSuite *RuleSuite
	_, err = g.client.Do(ctx, req, &ruleSuite)
	if err != nil {
		return nil, err
	}

	return ruleSuite, nil
}
