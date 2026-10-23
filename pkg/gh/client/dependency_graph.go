package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/v93/github"
)

const (
	sbomPollInterval = 2 * time.Second
	sbomPollTimeout  = 5 * time.Minute
)

// GetDependencyGraphSBOM retrieves the SBOM (Software Bill of Materials) for a repository.
// GitHub builds the SBOM asynchronously, so this requests generation and polls until the
// report is ready. The whole operation is bounded by a timeout.
func (g *GitHubClient) GetDependencyGraphSBOM(ctx context.Context, owner, repo string) (*github.SBOM, error) {
	return g.getDependencyGraphSBOM(ctx, owner, repo, sbomPollInterval, sbomPollTimeout)
}

func (g *GitHubClient) getDependencyGraphSBOM(ctx context.Context, owner, repo string, interval, timeout time.Duration) (*github.SBOM, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	generation, _, err := g.client.DependencyGraph.GenerateSBOM(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	if generation == nil {
		return nil, errors.New("empty SBOM generation response")
	}
	sbomUUID, err := sbomUUIDFromURL(generation.GetSBOMURL())
	if err != nil {
		return nil, err
	}

	for {
		report, err := g.FetchDependencyGraphSBOMReport(ctx, owner, repo, sbomUUID)
		if err != nil {
			return nil, err
		}
		if !report.Pending {
			if report.SBOM == nil {
				return nil, errors.New("SBOM report has no content")
			}
			return &github.SBOM{SBOM: report.SBOM}, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// sbomUUIDFromURL extracts the report UUID from an SBOM fetch URL of the form
// .../dependency-graph/sbom/fetch-report/{uuid}. Query strings are ignored.
func sbomUUIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid SBOM URL: %w", err)
	}
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(segments) < 2 {
		return "", errors.New("unexpected SBOM URL format")
	}
	id := segments[len(segments)-1]
	if segments[len(segments)-2] != "fetch-report" || id == "" || id == "." || id == ".." {
		return "", errors.New("unexpected SBOM URL format")
	}
	return id, nil
}

// SBOMReportGeneration is the response of requesting generation of an SBOM report.
type SBOMReportGeneration struct {
	SBOMURL string `json:"sbom_url"`
}

// GenerateDependencyGraphSBOMReport requests generation of an SBOM report for a repository
// using the dependency-graph/sbom/generate-report API. GitHub schedules a background job;
// the returned SBOMURL can be polled via FetchDependencyGraphSBOMReport once ready.
func (g *GitHubClient) GenerateDependencyGraphSBOMReport(ctx context.Context, owner, repo string) (*SBOMReportGeneration, error) {
	generation, _, err := g.client.DependencyGraph.GenerateSBOM(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return &SBOMReportGeneration{SBOMURL: generation.GetSBOMURL()}, nil
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
// that is not hosted on the GitHub API host. go-github downloads it with http.DefaultClient,
// which carries no credentials, so the GitHub authentication token is never sent to that host.
func (g *GitHubClient) FetchDependencyGraphSBOMReport(ctx context.Context, owner, repo, sbomUUID string) (*SBOMReport, error) {
	sbom, _, resp, err := g.client.DependencyGraph.FetchSBOM(ctx, owner, repo, url.PathEscape(sbomUUID), http.DefaultClient)
	if err != nil {
		// Only a 202 from the GitHub API means "still generating". A 202 from the
		// download host also yields *AcceptedError but comes with the API's 302 response.
		if _, ok := errors.AsType[*github.AcceptedError](err); ok && resp != nil && resp.StatusCode == http.StatusAccepted {
			return &SBOMReport{Pending: true}, nil
		}
		return nil, err
	}
	return &SBOMReport{SBOM: sbom.SBOM}, nil
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
func (g *GitHubClient) CreateDependencyGraphSnapshot(ctx context.Context, owner, repo string, snapshot *github.CreateDependencyGraphSnapshotRequest) (*github.DependencyGraphSnapshotCreationData, error) {
	result, _, err := g.client.DependencyGraph.CreateSnapshot(ctx, owner, repo, *snapshot)
	if err != nil {
		return nil, err
	}
	return result, nil
}
