package client

import (
	"context"
	"fmt"

	"github.com/google/go-github/v84/github"
)

// ListOrgRepositorySecurityAdvisories lists all repository security advisories for an organization.
func (g *GitHubClient) ListOrgRepositorySecurityAdvisories(ctx context.Context, org string, opts *github.ListRepositorySecurityAdvisoriesOptions) ([]*github.SecurityAdvisory, error) {
	var all []*github.SecurityAdvisory
	if opts == nil {
		opts = &github.ListRepositorySecurityAdvisoriesOptions{}
	}
	opts.ListCursorOptions = github.ListCursorOptions{PerPage: defaultPerPage}
	for {
		advisories, resp, err := g.client.SecurityAdvisories.ListRepositorySecurityAdvisoriesForOrg(ctx, org, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, advisories...)
		if resp.Cursor == "" {
			break
		}
		opts.Cursor = resp.Cursor
	}
	return all, nil
}

// ListRepoSecurityAdvisories lists all repository security advisories for a repository.
func (g *GitHubClient) ListRepoSecurityAdvisories(ctx context.Context, owner, repo string, opts *github.ListRepositorySecurityAdvisoriesOptions) ([]*github.SecurityAdvisory, error) {
	var all []*github.SecurityAdvisory
	if opts == nil {
		opts = &github.ListRepositorySecurityAdvisoriesOptions{}
	}
	opts.ListCursorOptions = github.ListCursorOptions{PerPage: defaultPerPage}
	for {
		advisories, resp, err := g.client.SecurityAdvisories.ListRepositorySecurityAdvisories(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, advisories...)
		if resp.Cursor == "" {
			break
		}
		opts.Cursor = resp.Cursor
	}
	return all, nil
}

// GetRepositorySecurityAdvisory gets a repository security advisory by GHSA ID.
func (g *GitHubClient) GetRepositorySecurityAdvisory(ctx context.Context, owner, repo, ghsaID string) (*github.SecurityAdvisory, error) {
	url := fmt.Sprintf("repos/%v/%v/security-advisories/%v", owner, repo, ghsaID)
	req, err := g.client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	advisory := new(github.SecurityAdvisory)
	_, err = g.client.Do(ctx, req, advisory)
	if err != nil {
		return nil, err
	}
	return advisory, nil
}

// RepositorySecurityAdvisoryUpdateOptions specifies the parameters to update a repository security advisory.
type RepositorySecurityAdvisoryUpdateOptions struct {
	Summary            *string  `json:"summary,omitempty"`
	Description        *string  `json:"description,omitempty"`
	CVEID              *string  `json:"cve_id,omitempty"`
	Severity           *string  `json:"severity,omitempty"`
	CVSSVectorString   *string  `json:"cvss_vector_string,omitempty"`
	State              *string  `json:"state,omitempty"`
	CollaboratingUsers []string `json:"collaborating_users,omitempty"`
	CollaboratingTeams []string `json:"collaborating_teams,omitempty"`
}

// UpdateRepositorySecurityAdvisory updates a repository security advisory by GHSA ID.
func (g *GitHubClient) UpdateRepositorySecurityAdvisory(ctx context.Context, owner, repo, ghsaID string, opts *RepositorySecurityAdvisoryUpdateOptions) (*github.SecurityAdvisory, error) {
	url := fmt.Sprintf("repos/%v/%v/security-advisories/%v", owner, repo, ghsaID)
	req, err := g.client.NewRequest("PATCH", url, opts)
	if err != nil {
		return nil, err
	}
	advisory := new(github.SecurityAdvisory)
	_, err = g.client.Do(ctx, req, advisory)
	if err != nil {
		return nil, err
	}
	return advisory, nil
}

// RequestRepositorySecurityAdvisoryCVE requests a CVE for a repository security advisory.
func (g *GitHubClient) RequestRepositorySecurityAdvisoryCVE(ctx context.Context, owner, repo, ghsaID string) error {
	_, err := g.client.SecurityAdvisories.RequestCVE(ctx, owner, repo, ghsaID)
	return err
}

// CreateRepositorySecurityAdvisoryFork creates a temporary private fork for a repository security advisory.
func (g *GitHubClient) CreateRepositorySecurityAdvisoryFork(ctx context.Context, owner, repo, ghsaID string) (*github.Repository, error) {
	fork, _, err := g.client.SecurityAdvisories.CreateTemporaryPrivateFork(ctx, owner, repo, ghsaID)
	if err != nil {
		return nil, err
	}
	return fork, nil
}
