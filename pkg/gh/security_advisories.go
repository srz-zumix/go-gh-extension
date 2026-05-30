package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// RepositorySecurityAdvisoryStates is the list of valid states for filtering repository security advisories.
var RepositorySecurityAdvisoryStates = []string{
	"triage",
	"draft",
	"published",
	"closed",
}

// RepositorySecurityAdvisoryUpdateStates is the list of valid states for updating a repository security advisory.
var RepositorySecurityAdvisoryUpdateStates = []string{
	"published",
	"closed",
	"draft",
}

// RepositorySecurityAdvisorySortOptions is the list of valid sort options for repository security advisories.
var RepositorySecurityAdvisorySortOptions = []string{
	"created",
	"updated",
	"published",
}

// RepositorySecurityAdvisoryDirections is the list of valid direction options.
var RepositorySecurityAdvisoryDirections = []string{
	"asc",
	"desc",
}

// RepositorySecurityAdvisorySeverities is the list of valid severity values.
var RepositorySecurityAdvisorySeverities = []string{
	"critical",
	"high",
	"medium",
	"low",
}

// ListRepositorySecurityAdvisoriesOptions holds filter/sort options for listing repository security advisories.
type ListRepositorySecurityAdvisoriesOptions struct {
	State     string
	Sort      string
	Direction string
}

// toGitHubListRepositorySecurityAdvisoriesOptions converts ListRepositorySecurityAdvisoriesOptions to github.ListRepositorySecurityAdvisoriesOptions.
func toGitHubListRepositorySecurityAdvisoriesOptions(opts *ListRepositorySecurityAdvisoriesOptions) *github.ListRepositorySecurityAdvisoriesOptions {
	if opts == nil {
		return nil
	}
	return &github.ListRepositorySecurityAdvisoriesOptions{
		State:     opts.State,
		Sort:      opts.Sort,
		Direction: opts.Direction,
	}
}

// ListRepositorySecurityAdvisories lists repository security advisories.
// If repo.Name is empty, lists org-level advisories; otherwise lists repo-level advisories.
func ListRepositorySecurityAdvisories(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListRepositorySecurityAdvisoriesOptions) ([]*github.SecurityAdvisory, error) {
	if repo.Name == "" {
		return ListOrgRepositorySecurityAdvisories(ctx, g, repo, opts)
	}
	return ListRepoSecurityAdvisories(ctx, g, repo, opts)
}

// ListOrgRepositorySecurityAdvisories lists repository security advisories for an organization.
func ListOrgRepositorySecurityAdvisories(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListRepositorySecurityAdvisoriesOptions) ([]*github.SecurityAdvisory, error) {
	advisories, err := g.ListOrgRepositorySecurityAdvisories(ctx, repo.Owner, toGitHubListRepositorySecurityAdvisoriesOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to list org repository security advisories: %w", err)
	}
	return advisories, nil
}

// ListRepoSecurityAdvisories lists repository security advisories for a repository.
func ListRepoSecurityAdvisories(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ListRepositorySecurityAdvisoriesOptions) ([]*github.SecurityAdvisory, error) {
	advisories, err := g.ListRepoSecurityAdvisories(ctx, repo.Owner, repo.Name, toGitHubListRepositorySecurityAdvisoriesOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to list repository security advisories: %w", err)
	}
	return advisories, nil
}

// GetRepositorySecurityAdvisory gets a repository security advisory by GHSA ID.
func GetRepositorySecurityAdvisory(ctx context.Context, g *GitHubClient, repo repository.Repository, ghsaID string) (*github.SecurityAdvisory, error) {
	advisory, err := g.GetRepositorySecurityAdvisory(ctx, repo.Owner, repo.Name, ghsaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository security advisory %s: %w", ghsaID, err)
	}
	return advisory, nil
}

// RepositorySecurityAdvisoryUpdateOptions holds options for updating a repository security advisory.
type RepositorySecurityAdvisoryUpdateOptions struct {
	Summary            string
	Description        string
	CVEID              string
	Severity           string
	CVSSVectorString   string
	State              string
	CollaboratingUsers []string
	CollaboratingTeams []string
}

// toClientRepositorySecurityAdvisoryUpdateOptions converts RepositorySecurityAdvisoryUpdateOptions to client.RepositorySecurityAdvisoryUpdateOptions.
func toClientRepositorySecurityAdvisoryUpdateOptions(opts *RepositorySecurityAdvisoryUpdateOptions) *client.RepositorySecurityAdvisoryUpdateOptions {
	if opts == nil {
		return nil
	}
	o := &client.RepositorySecurityAdvisoryUpdateOptions{}
	if opts.Summary != "" {
		o.Summary = &opts.Summary
	}
	if opts.Description != "" {
		o.Description = &opts.Description
	}
	if opts.CVEID != "" {
		o.CVEID = &opts.CVEID
	}
	if opts.Severity != "" {
		o.Severity = &opts.Severity
	}
	if opts.CVSSVectorString != "" {
		o.CVSSVectorString = &opts.CVSSVectorString
	}
	if opts.State != "" {
		o.State = &opts.State
	}
	if len(opts.CollaboratingUsers) > 0 {
		o.CollaboratingUsers = opts.CollaboratingUsers
	}
	if len(opts.CollaboratingTeams) > 0 {
		o.CollaboratingTeams = opts.CollaboratingTeams
	}
	return o
}

// UpdateRepositorySecurityAdvisory updates a repository security advisory by GHSA ID.
func UpdateRepositorySecurityAdvisory(ctx context.Context, g *GitHubClient, repo repository.Repository, ghsaID string, opts *RepositorySecurityAdvisoryUpdateOptions) (*github.SecurityAdvisory, error) {
	advisory, err := g.UpdateRepositorySecurityAdvisory(ctx, repo.Owner, repo.Name, ghsaID, toClientRepositorySecurityAdvisoryUpdateOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to update repository security advisory %s: %w", ghsaID, err)
	}
	return advisory, nil
}

// RequestRepositorySecurityAdvisoryCVE requests a CVE for a repository security advisory.
func RequestRepositorySecurityAdvisoryCVE(ctx context.Context, g *GitHubClient, repo repository.Repository, ghsaID string) error {
	err := g.RequestRepositorySecurityAdvisoryCVE(ctx, repo.Owner, repo.Name, ghsaID)
	if err != nil {
		return fmt.Errorf("failed to request CVE for repository security advisory %s: %w", ghsaID, err)
	}
	return nil
}

// CreateRepositorySecurityAdvisoryFork creates a temporary private fork for a repository security advisory.
func CreateRepositorySecurityAdvisoryFork(ctx context.Context, g *GitHubClient, repo repository.Repository, ghsaID string) (*github.Repository, error) {
	fork, err := g.CreateRepositorySecurityAdvisoryFork(ctx, repo.Owner, repo.Name, ghsaID)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary private fork for repository security advisory %s: %w", ghsaID, err)
	}
	return fork, nil
}
