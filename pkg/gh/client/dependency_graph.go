package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/v88/github"
)

// GetDependencyGraphSBOM retrieves the SBOM (Software Bill of Materials) for a repository using the dependency-graph/sbom API.
func (g *GitHubClient) GetDependencyGraphSBOM(ctx context.Context, owner, repo string) (*github.SBOM, error) {
	sbom, _, err := g.client.DependencyGraph.GetSBOM(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return sbom, nil
}

// SBOMReportGeneration is the response of requesting generation of an SBOM report.
type SBOMReportGeneration struct {
	SBOMURL string `json:"sbom_url"`
}

// GenerateDependencyGraphSBOMReport requests generation of an SBOM report for a repository
// using the dependency-graph/sbom/generate-report API. GitHub schedules a background job;
// the returned SBOMURL can be polled via FetchDependencyGraphSBOMReport once ready.
func (g *GitHubClient) GenerateDependencyGraphSBOMReport(ctx context.Context, owner, repo string) (*SBOMReportGeneration, error) {
	u := fmt.Sprintf("repos/%v/%v/dependency-graph/sbom/generate-report", owner, repo)
	req, err := g.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	result := new(SBOMReportGeneration)
	_, err = g.client.Do(req, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SBOMReport is the result of fetching a previously requested SBOM report.
// Pending is true when the report is still being generated (HTTP 202); SBOM is
// populated once the report is ready.
type SBOMReport struct {
	Pending bool
	SBOM    *github.SBOMInfo
}

// FetchDependencyGraphSBOMReport fetches a previously generated SBOM report using the
// dependency-graph/sbom/fetch-report/{sbom_uuid} API.
//
// When the report is ready, GitHub responds with a redirect to a temporary download URL
// that is not hosted on the GitHub API host. Redirects are followed manually here so the
// GitHub authentication token is never sent to that third-party host.
func (g *GitHubClient) FetchDependencyGraphSBOMReport(ctx context.Context, owner, repo, sbomUUID string) (*SBOMReport, error) {
	u := fmt.Sprintf("repos/%v/%v/dependency-graph/sbom/fetch-report/%v", owner, repo, url.PathEscape(sbomUUID))
	req, err := g.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	// Reuse the same authenticated transport as the main client, but disable
	// automatic redirect following so we can decide per-hop whether to forward
	// the authentication header.
	authClient := &http.Client{
		Transport: g.client.Client().Transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := authClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusAccepted:
		return &SBOMReport{Pending: true}, nil
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		loc := resp.Header.Get("Location")
		if loc == "" {
			return nil, fmt.Errorf("redirect response missing Location header")
		}
		parsedLoc, err := url.Parse(loc)
		if err != nil {
			return nil, fmt.Errorf("invalid redirect Location header: %w", err)
		}
		downloadURL := resp.Request.URL.ResolveReference(parsedLoc)

		// The download URL is a temporary, pre-signed link; fetch it without
		// forwarding any GitHub credentials.
		downloadReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL.String(), nil)
		if err != nil {
			return nil, err
		}
		downloadResp, err := http.DefaultClient.Do(downloadReq)
		if err != nil {
			return nil, err
		}
		defer func() { _ = downloadResp.Body.Close() }()
		// The download host is not the GitHub API, so validate the status code
		// directly instead of using github.CheckResponse, which expects GitHub's
		// error response format and would produce misleading errors here.
		if downloadResp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status downloading SBOM report: %s", downloadResp.Status)
		}
		sbomInfo := new(github.SBOMInfo)
		if err := json.NewDecoder(downloadResp.Body).Decode(sbomInfo); err != nil {
			return nil, fmt.Errorf("failed to decode SBOM report: %w", err)
		}
		return &SBOMReport{SBOM: sbomInfo}, nil
	default:
		if err := github.CheckResponse(resp); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unexpected status %d fetching SBOM report", resp.StatusCode)
	}
}

// DependencyVulnerability represents a vulnerability associated with a dependency change.
type DependencyVulnerability struct {
	Severity       string `json:"severity"`
	AdvisoryGHSAID string `json:"advisory_ghsa_id"`
	AdvisoryTitle  string `json:"advisory_summary"`
	AdvisoryURL    string `json:"advisory_url"`
}

// DependencyChange represents a single dependency change in a dependency diff.
type DependencyChange struct {
	ChangeType      string                    `json:"change_type"`
	Manifest        string                    `json:"manifest"`
	Ecosystem       string                    `json:"ecosystem"`
	Name            string                    `json:"name"`
	Version         string                    `json:"version"`
	PackageURL      string                    `json:"package_url"`
	License         string                    `json:"license"`
	SourceRepoURL   string                    `json:"source_repository_url"`
	Vulnerabilities []DependencyVulnerability `json:"vulnerabilities"`
	Scope           string                    `json:"scope"`
}

// GetDependencyGraphDiff retrieves the dependency diff between two commits or branches using the dependency-graph/compare API.
func (g *GitHubClient) GetDependencyGraphDiff(ctx context.Context, owner, repo, basehead string) ([]*DependencyChange, error) {
	u := fmt.Sprintf("repos/%v/%v/dependency-graph/compare/%v", owner, repo, url.PathEscape(basehead))
	req, err := g.client.NewRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	var changes []*DependencyChange
	_, err = g.client.Do(req, &changes)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// CreateDependencyGraphSnapshot creates a new snapshot of a repository's dependencies.
func (g *GitHubClient) CreateDependencyGraphSnapshot(ctx context.Context, owner, repo string, snapshot *github.DependencyGraphSnapshot) (*github.DependencyGraphSnapshotCreationData, error) {
	result, _, err := g.client.DependencyGraph.CreateSnapshot(ctx, owner, repo, snapshot)
	if err != nil {
		return nil, err
	}
	return result, nil
}
