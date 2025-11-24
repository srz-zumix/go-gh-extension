package gh

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// CheckRun is a type alias for client.CheckRun that represents a single check run
type CheckRun = client.CheckRun

// ListCheckRunsResults holds the results of listing check runs
type ListCheckRunsResults struct {
	Total     int
	CheckRuns []*CheckRun
}

// https://docs.github.com/ja/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/about-status-checks#check-statuses-and-conclusions
var (
	// ChecksRunStatusCompleted indicates that the check run has completed.
	ChecksRunStatusCompleted = "completed"
	// ChecksRunStatusExpected indicates that the check run is expected but has not started.
	ChecksRunStatusExpected = "expected"
	// ChecksRunStatusFailure indicates that the check run has failed.
	ChecksRunStatusFailure = "failure"
	// ChecksRunStatusInProgress indicates that the check run is currently in progress.
	ChecksRunStatusInProgress = "in_progress"
	// ChecksRunStatusPending indicates that the check run is pending and has not started.
	ChecksRunStatusPending = "pending"
	// ChecksRunStatusQueued indicates that the check run is queued to start.
	ChecksRunStatusQueued = "queued"
	// ChecksRunStatusRequested indicates that the check run has been requested.
	ChecksRunStatusRequested = "requested"
	// ChecksRunStatusStartupFailure indicates that the check run failed to start.
	ChecksRunStatusStartupFailure = "startup_failure"
	// ChecksRunStatusWaiting indicates that the check run is waiting for resources or dependencies.
	ChecksRunStatusWaiting = "waiting"
)

var (
	// ChecksRunConclusionActionRequired indicates that further action is required for the check run.
	ChecksRunConclusionActionRequired = "action_required"
	// ChecksRunConclusionCancelled indicates that the check run was cancelled.
	ChecksRunConclusionCancelled = "cancelled"
	// ChecksRunConclusionFailure indicates that the check run concluded with a failure.
	ChecksRunConclusionFailure = "failure"
	// ChecksRunConclusionNeutral indicates that the check run completed with a neutral result.
	ChecksRunConclusionNeutral = "neutral"
	// ChecksRunConclusionSkipped indicates that the check run was skipped.
	ChecksRunConclusionSkipped = "skipped"
	// ChecksRunConclusionStale indicates that the check run is stale and may need to be re-run.
	ChecksRunConclusionStale = "stale"
	// ChecksRunConclusionSuccess indicates that the check run completed successfully.
	ChecksRunConclusionSuccess = "success"
	// ChecksRunConclusionTimedOut indicates that the check run timed out.
	ChecksRunConclusionTimedOut = "timed_out"
)

var (
	// ChecksRunFilterStatuses are valid status values for filtering check runs
	ChecksRunFilterStatuses = []string{
		ChecksRunStatusCompleted,
		ChecksRunStatusInProgress,
		ChecksRunStatusPending,
	}
	// ChecksRunFilterConclusions are valid conclusion values for filtering check runs
	ChecksRunFilterConclusions = []string{
		ChecksRunConclusionActionRequired,
		ChecksRunConclusionCancelled,
		ChecksRunConclusionFailure,
		ChecksRunConclusionNeutral,
		ChecksRunConclusionSkipped,
		ChecksRunConclusionStale,
		ChecksRunConclusionSuccess,
		ChecksRunConclusionTimedOut,
	}
)

// ListChecksRunFilterOptions holds filtering options for listing check runs
type ListChecksRunFilterOptions struct {
	CheckName  *string
	Status     *string
	Conclusion *string
	Filter     *string
	AppID      *int64
	Required   *bool
}

// ListCheckRunsForRef retrieves check runs for a specific Git reference and applies filtering options.
func ListCheckRunsForRef(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string, options *ListChecksRunFilterOptions) (*ListCheckRunsResults, error) {
	var opt *github.ListCheckRunsOptions
	if options != nil {
		opt = &github.ListCheckRunsOptions{
			CheckName: options.CheckName,
			Status:    options.Status,
			Filter:    options.Filter,
			AppID:     options.AppID,
		}
	}
	result, err := g.ListCheckRunsForRef(ctx, repo.Owner, repo.Name, ref, opt)
	if err != nil {
		return nil, err
	}

	checkRuns := []*CheckRun{}
	for _, checkRun := range result.CheckRuns {
		checkRuns = append(checkRuns, &CheckRun{CheckRun: *checkRun})
	}

	return filterCheckRuns(checkRuns, options)
}

// ExtractRunIDFromCheckRun extracts the workflow run ID from check run URLs
func ExtractRunIDFromCheckRun(checkRun *CheckRun) string {
	// Try to extract from HTML URL (e.g., https://github.com/owner/repo/runs/12345)
	htmlURL := checkRun.GetHTMLURL()
	if htmlURL != "" {
		if runID := extractRunIDFromURL(htmlURL); runID != "" {
			return runID
		}
	}

	// Try to extract from details URL
	detailsURL := checkRun.GetDetailsURL()
	if detailsURL != "" {
		if runID := extractRunIDFromURL(detailsURL); runID != "" {
			return runID
		}
	}

	return ""
}

// ExtractRunIDFromCheckRunAsInt64 extracts the workflow run ID as int64 from check run URLs
func ExtractRunIDFromCheckRunAsInt64(checkRun *CheckRun) (int64, error) {
	runIDStr := ExtractRunIDFromCheckRun(checkRun)
	if runIDStr == "" {
		return 0, fmt.Errorf("run ID not found")
	}
	return strconv.ParseInt(runIDStr, 10, 64)
}

// extractRunIDFromURL extracts run ID from various GitHub URL patterns
func extractRunIDFromURL(url string) string {
	// Try /runs/ pattern

	parts := strings.Split(url, "/runs/")
	if len(parts) == 2 {
		if runID := extractIDFromPath(parts[1]); runID != "" {
			return runID
		}
	}

	// Try /actions/runs/ pattern
	parts = strings.Split(url, "/actions/runs/")
	if len(parts) == 2 {
		if runID := extractIDFromPath(parts[1]); runID != "" {
			return runID
		}
	}

	return ""
}

// extractIDFromPath extracts the first segment of the path (the ID)
func extractIDFromPath(path string) string {
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return path
}

// ListCheckSuiteFilterOptions holds filtering options for listing check suites
type ListCheckSuiteFilterOptions struct {
	AppID      *int64
	CheckName  *string
	Status     *string
	Conclusion *string
}

// ListCheckSuitesForRef retrieves check suites for a specific Git reference and applies filtering options.
func ListCheckSuitesForRef(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string, options *ListCheckSuiteFilterOptions) (*github.ListCheckSuiteResults, error) {
	var opt *github.ListCheckSuiteOptions
	if options != nil {
		opt = &github.ListCheckSuiteOptions{
			AppID:     options.AppID,
			CheckName: options.CheckName,
		}
	}
	result, err := g.ListCheckSuitesForRef(ctx, repo.Owner, repo.Name, ref, opt)
	if err != nil {
		return nil, err
	}

	// Filter by status and conclusion (not supported by API)
	if options != nil && (options.Status != nil || options.Conclusion != nil) {
		filteredCheckSuites := []*github.CheckSuite{}
		for _, checkSuite := range result.CheckSuites {
			if options.Status != nil && checkSuite.GetStatus() != *options.Status {
				continue
			}
			if options.Conclusion != nil && checkSuite.GetConclusion() != *options.Conclusion {
				continue
			}
			filteredCheckSuites = append(filteredCheckSuites, checkSuite)
		}
		result.CheckSuites = filteredCheckSuites
		total := len(filteredCheckSuites)
		result.Total = &total
	}

	return result, nil
}

// ListCheckRunsForRefWithGraphQL retrieves check runs for a specific Git reference using GraphQL and applies filtering options.
func ListCheckRunsForRefWithGraphQL(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string, prNumber int, options *ListChecksRunFilterOptions) (*ListCheckRunsResults, error) {
	checkRuns, err := g.ListCheckRunsForRefWithGraphQL(ctx, repo.Owner, repo.Name, ref, prNumber)
	if err != nil {
		return nil, err
	}

	return filterCheckRuns(checkRuns, options)
}

// SortCheckRunsByName sorts check runs by their full names
func SortCheckRunsByName(checkRuns []*CheckRun) {
	slices.SortFunc(checkRuns, func(a, b *CheckRun) int {
		return strings.Compare(a.GetFullName(), b.GetFullName())
	})
}

// SortCheckRunsByRunID sorts check runs by their associated workflow run IDs
func SortCheckRunsByRunID(checkRuns []*CheckRun) {
	slices.SortFunc(checkRuns, func(a, b *CheckRun) int {
		aRunID := ExtractRunIDFromCheckRun(a)
		bRunID := ExtractRunIDFromCheckRun(b)
		return strings.Compare(aRunID, bRunID)
	})
}

func filterCheckRuns(checkRuns []*CheckRun, options *ListChecksRunFilterOptions) (*ListCheckRunsResults, error) {
	if options == nil || options.Filter == nil || *options.Filter == "latest" {
		latestMap := map[string]*CheckRun{}
		for _, checkRun := range checkRuns {
			name := checkRun.GetName()
			if existing, ok := latestMap[name]; ok {
				// Check for nil before comparing CompletedAt timestamps
				if checkRun.GetCompletedAt() != nil && existing.GetCompletedAt() != nil {
					if checkRun.GetCompletedAt().After(existing.GetCompletedAt().Time) {
						latestMap[name] = checkRun
					}
				} else if checkRun.GetCompletedAt() != nil {
					// Prefer completed checkRun over incomplete existing
					latestMap[name] = checkRun
				}
			} else {
				latestMap[name] = checkRun
			}
		}
		checkRuns = []*CheckRun{}
		for _, checkRun := range latestMap {
			checkRuns = append(checkRuns, checkRun)
		}
	}

	if options != nil {
		filteredCheckRuns := []*CheckRun{}
		for _, checkRun := range checkRuns {
			if options.CheckName != nil && checkRun.GetName() != *options.CheckName {
				continue
			}
			if options.Status != nil && checkRun.GetStatus() != *options.Status {
				continue
			}
			if options.Conclusion != nil && checkRun.GetConclusion() != *options.Conclusion {
				continue
			}
			if options.AppID != nil {
				appID := checkRun.GetApp().GetID()
				if appID != *options.AppID {
					continue
				}
			}
			if options.Required != nil {
				// Check for nil before dereferencing checkRun.IsRequired
				if checkRun.IsRequired == nil || *checkRun.IsRequired != *options.Required {
					continue
				}
			}
			filteredCheckRuns = append(filteredCheckRuns, checkRun)
		}
		checkRuns = filteredCheckRuns
	}
	return &ListCheckRunsResults{
		Total:     len(checkRuns),
		CheckRuns: checkRuns,
	}, nil
}

// GetCheckRun retrieves a specific check run.
func GetCheckRun(ctx context.Context, g *GitHubClient, repo repository.Repository, checkRunID int64) (*CheckRun, error) {
	return g.GetCheckRun(ctx, repo.Owner, repo.Name, checkRunID)
}

// GetCheckRunJobLogsURL retrieves the logs URL for a check run.
// Note: This function attempts to retrieve logs by treating the checkRunID as a workflow job ID.
// In GitHub Actions, check runs are associated with workflow jobs that have the same ID.
// The maxRedirects parameter specifies the maximum number of redirects to follow.
// Setting maxRedirects to 0 will return the redirect URL without following it.
func GetCheckRunJobLogsURL(ctx context.Context, g *GitHubClient, repo repository.Repository, checkRunID int64, maxRedirects int) (string, error) {
	return g.GetWorkflowJobLogs(ctx, repo.Owner, repo.Name, checkRunID, maxRedirects)
}

// GetCheckRunJobLogsContent retrieves the logs content for a check run.
// This function downloads the logs from the URL obtained by GetCheckRunJobLogsURL.
func GetCheckRunJobLogsContent(ctx context.Context, g *GitHubClient, repo repository.Repository, checkRunID int64, maxRedirects int) ([]byte, error) {
	return g.GetWorkflowJobLogsContent(ctx, repo.Owner, repo.Name, checkRunID, maxRedirects)
}
