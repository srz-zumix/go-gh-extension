package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v88/github"
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

// ListRepositorySecurityAdvisoriesOptions holds filter/sort/pagination options for listing repository security advisories.
// All fields correspond directly to the upstream github.ListRepositorySecurityAdvisoriesOptions and its embedded
// github.ListCursorOptions, so every parameter supported by the GitHub API is available to callers.
type ListRepositorySecurityAdvisoriesOptions struct {
	// State filters advisories by state. Possible values: triage, draft, published, closed.
	State string
	// Sort specifies how to sort advisories. Possible values: created, updated, published. Default: created.
	Sort string
	// Direction specifies the sort direction. Possible values: asc, desc. Default: asc.
	Direction string
	// PerPage is the number of results per page (max 100).
	PerPage int
	// Before is a cursor for backward pagination (as given in the Link header).
	Before string
	// After is a cursor for forward pagination (as given in the Link header).
	After string
	// Page is a page cursor for pagination.
	Page string
	// First is the number of results per page starting from the first matching result.
	// Must not be combined with Last.
	First int
	// Last is the number of results per page starting from the last matching result.
	// Must not be combined with First.
	Last int
	// Cursor continues a search from a previous cursor value (as given in the Link header).
	Cursor string
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
		ListCursorOptions: github.ListCursorOptions{
			PerPage: opts.PerPage,
			Before:  opts.Before,
			After:   opts.After,
			Page:    opts.Page,
			First:   opts.First,
			Last:    opts.Last,
			Cursor:  opts.Cursor,
		},
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

// GlobalSecurityAdvisoryTypes is the list of valid type values for filtering global security advisories.
var GlobalSecurityAdvisoryTypes = []string{
	"reviewed",
	"malware",
	"unreviewed",
}

// GlobalSecurityAdvisoryEcosystems is the list of valid ecosystem values for filtering global security advisories.
var GlobalSecurityAdvisoryEcosystems = []string{
	"actions",
	"composer",
	"erlang",
	"go",
	"maven",
	"npm",
	"nuget",
	"other",
	"pip",
	"pub",
	"rubygems",
	"rust",
}

// ListGlobalSecurityAdvisoriesOptions holds filter/sort/pagination options for listing global security advisories.
type ListGlobalSecurityAdvisoriesOptions struct {
	Type      string
	Severity  string
	Ecosystem string
	GHSAID    string
	CVEID     string
}

// toGitHubListGlobalSecurityAdvisoriesOptions converts ListGlobalSecurityAdvisoriesOptions to github.ListGlobalSecurityAdvisoriesOptions.
func toGitHubListGlobalSecurityAdvisoriesOptions(opts *ListGlobalSecurityAdvisoriesOptions) *github.ListGlobalSecurityAdvisoriesOptions {
	if opts == nil {
		return nil
	}
	o := &github.ListGlobalSecurityAdvisoriesOptions{}
	if opts.Type != "" {
		o.Type = &opts.Type
	}
	if opts.Severity != "" {
		o.Severity = &opts.Severity
	}
	if opts.Ecosystem != "" {
		o.Ecosystem = &opts.Ecosystem
	}
	if opts.GHSAID != "" {
		o.GHSAID = &opts.GHSAID
	}
	if opts.CVEID != "" {
		o.CVEID = &opts.CVEID
	}
	return o
}

// ListGlobalSecurityAdvisories lists global security advisories.
func ListGlobalSecurityAdvisories(ctx context.Context, g *GitHubClient, opts *ListGlobalSecurityAdvisoriesOptions) ([]*github.GlobalSecurityAdvisory, error) {
	advisories, err := g.ListGlobalSecurityAdvisories(ctx, toGitHubListGlobalSecurityAdvisoriesOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to list global security advisories: %w", err)
	}
	return advisories, nil
}

// GetGlobalSecurityAdvisory gets a global security advisory by GHSA ID.
func GetGlobalSecurityAdvisory(ctx context.Context, g *GitHubClient, ghsaID string) (*github.GlobalSecurityAdvisory, error) {
	advisory, err := g.GetGlobalSecurityAdvisory(ctx, ghsaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get global security advisory %s: %w", ghsaID, err)
	}
	return advisory, nil
}

// CreateRepositorySecurityAdvisoryOptions holds options for creating a repository security advisory.
type CreateRepositorySecurityAdvisoryOptions struct {
	Summary                string
	Description            string
	CVEID                  string
	Severity               string
	CVSSVectorString       string
	Ecosystem              string
	PackageName            string
	VulnerableVersionRange string
	PatchedVersions        string
	CWEIDs                 []string
	StartPrivateFork       bool
}

// toClientCreateRepositorySecurityAdvisoryOptions converts CreateRepositorySecurityAdvisoryOptions to client.CreateRepositorySecurityAdvisoryOptions.
func toClientCreateRepositorySecurityAdvisoryOptions(opts *CreateRepositorySecurityAdvisoryOptions) *client.CreateRepositorySecurityAdvisoryOptions {
	if opts == nil {
		return nil
	}
	o := &client.CreateRepositorySecurityAdvisoryOptions{
		Summary:     opts.Summary,
		Description: opts.Description,
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
	if len(opts.CWEIDs) > 0 {
		o.CWEIDs = opts.CWEIDs
	}
	if opts.StartPrivateFork {
		b := true
		o.StartPrivateFork = &b
	}
	pkg := &client.VulnerabilityPackageInput{
		Ecosystem: opts.Ecosystem,
	}
	if opts.PackageName != "" {
		pkg.Name = &opts.PackageName
	}
	vuln := client.RepositorySecurityAdvisoryVulnerabilityInput{
		Package: pkg,
	}
	if opts.VulnerableVersionRange != "" {
		vuln.VulnerableVersionRange = &opts.VulnerableVersionRange
	}
	if opts.PatchedVersions != "" {
		vuln.PatchedVersions = &opts.PatchedVersions
	}
	o.Vulnerabilities = []client.RepositorySecurityAdvisoryVulnerabilityInput{vuln}
	return o
}

// CreateRepositorySecurityAdvisory creates a repository security advisory.
func CreateRepositorySecurityAdvisory(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *CreateRepositorySecurityAdvisoryOptions) (*github.SecurityAdvisory, error) {
	advisory, err := g.CreateRepositorySecurityAdvisory(ctx, repo.Owner, repo.Name, toClientCreateRepositorySecurityAdvisoryOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to create repository security advisory: %w", err)
	}
	return advisory, nil
}

// ReportRepositorySecurityAdvisoryOptions holds options for reporting a repository security advisory.
type ReportRepositorySecurityAdvisoryOptions struct {
	Ecosystem        string
	PackageName      string
	Summary          string
	Description      string
	Severity         string
	CVSSVectorString string
	CWEIDs           []string
	StartPrivateFork bool
}

// toClientReportRepositorySecurityAdvisoryOptions converts ReportRepositorySecurityAdvisoryOptions to client.ReportRepositorySecurityAdvisoryOptions.
func toClientReportRepositorySecurityAdvisoryOptions(opts *ReportRepositorySecurityAdvisoryOptions) *client.ReportRepositorySecurityAdvisoryOptions {
	if opts == nil {
		return nil
	}
	pkg := &client.VulnerabilityPackageInput{
		Ecosystem: opts.Ecosystem,
	}
	if opts.PackageName != "" {
		pkg.Name = &opts.PackageName
	}
	o := &client.ReportRepositorySecurityAdvisoryOptions{
		Package:     pkg,
		Summary:     opts.Summary,
		Description: opts.Description,
	}
	if opts.Severity != "" {
		o.Severity = &opts.Severity
	}
	if opts.CVSSVectorString != "" {
		o.CVSSVectorString = &opts.CVSSVectorString
	}
	if len(opts.CWEIDs) > 0 {
		o.CWEIDs = opts.CWEIDs
	}
	if opts.StartPrivateFork {
		b := true
		o.StartPrivateFork = &b
	}
	return o
}

// ReportRepositorySecurityAdvisory reports a vulnerability in a repository.
func ReportRepositorySecurityAdvisory(ctx context.Context, g *GitHubClient, repo repository.Repository, opts *ReportRepositorySecurityAdvisoryOptions) (*github.SecurityAdvisory, error) {
	advisory, err := g.ReportRepositorySecurityAdvisory(ctx, repo.Owner, repo.Name, toClientReportRepositorySecurityAdvisoryOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to report repository security advisory: %w", err)
	}
	return advisory, nil
}
