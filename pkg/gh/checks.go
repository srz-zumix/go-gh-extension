package gh

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

type CheckRun = client.CheckRun

type ListCheckRunsResults struct {
	Total     int
	CheckRuns []*CheckRun
}

// https://docs.github.com/ja/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/about-status-checks#check-statuses-and-conclusions
var (
	ChecksRunStatusCompleted      = "completed"
	ChecksRunStatusExpected       = "expected"
	ChecksRunStatusFailure        = "failure"
	ChecksRunStatusInProgress     = "in_progress"
	ChecksRunStatusPending        = "pending"
	ChecksRunStatusQueued         = "queued"
	ChecksRunStatusRequested      = "requested"
	ChecksRunStatusStartupFailure = "startup_failure"
	ChecksRunStatusWaiting        = "waiting"
)

var (
	ChecksRunConclusionActionRequired = "action_required"
	ChecksRunConclusionCancelled      = "cancelled"
	ChecksRunConclusionFailure        = "failure"
	ChecksRunConclusionNeutral        = "neutral"
	ChecksRunConclusionSkipped        = "skipped"
	ChecksRunConclusionStale          = "stale"
	ChecksRunConclusionSuccess        = "success"
	ChecksRunConclusionTimedOut       = "timed_out"
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

type ListChecksRunFilterOptions struct {
	CheckName  *string
	Status     *string
	Conclusion *string
	Filter     *string
	AppID      *int64
	Required   *bool
}

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
	if parts := splitURL(url, "/runs/"); len(parts) == 2 {
		if runID := extractIDFromPath(parts[1]); runID != "" {
			return runID
		}
	}

	// Try /actions/runs/ pattern
	if parts := splitURL(url, "/actions/runs/"); len(parts) == 2 {
		if runID := extractIDFromPath(parts[1]); runID != "" {
			return runID
		}
	}

	return ""
}

// splitURL splits URL by delimiter
func splitURL(url, delimiter string) []string {
	parts := []string{}
	if idx := findString(url, delimiter); idx >= 0 {
		parts = append(parts, url[:idx])
		parts = append(parts, url[idx+len(delimiter):])
	}
	return parts
}

// findString returns the index of the first occurrence of delimiter in s, or -1 if not found
func findString(s, delimiter string) int {
	for i := 0; i+len(delimiter) <= len(s); i++ {
		if s[i:i+len(delimiter)] == delimiter {
			return i
		}
	}
	return -1
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

type ListCheckSuiteFilterOptions struct {
	AppID      *int64
	CheckName  *string
	Status     *string
	Conclusion *string
}

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

func ListCheckRunsForRefWithGraphQL(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string, prNumber int, options *ListChecksRunFilterOptions) (*ListCheckRunsResults, error) {
	checkRuns, err := g.ListCheckRunsForRefWithGraphQL(ctx, repo.Owner, repo.Name, ref, prNumber)
	if err != nil {
		return nil, err
	}

	return filterCheckRuns(checkRuns, options)
}

func SortCheckRunsByName(checkRuns []*CheckRun) {
	// Simple bubble sort for demonstration; consider using sort.Slice for production code
	n := len(checkRuns)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if checkRuns[j].GetFullName() > checkRuns[j+1].GetFullName() {
				checkRuns[j], checkRuns[j+1] = checkRuns[j+1], checkRuns[j]
			}
		}
	}
}

func SortCheckRunsByRunID(checkRuns []*CheckRun) {
	// Simple bubble sort for demonstration; consider using sort.Slice for production code
	n := len(checkRuns)
	for i := 0; i < n; i++ {
		for j := 0; j < n-i-1; j++ {
			if ExtractRunIDFromCheckRun(checkRuns[j]) > ExtractRunIDFromCheckRun(checkRuns[j+1]) {
				checkRuns[j], checkRuns[j+1] = checkRuns[j+1], checkRuns[j]
			}
		}
	}
}

func filterCheckRuns(checkRuns []*CheckRun, options *ListChecksRunFilterOptions) (*ListCheckRunsResults, error) {
	if options == nil || options.Filter == nil || *options.Filter == "latest" {
		latestMap := map[string]*CheckRun{}
		for _, checkRun := range checkRuns {
			name := checkRun.CheckRun.GetName()
			if existing, ok := latestMap[name]; ok {
				if checkRun.CheckRun.GetCompletedAt().After(existing.CheckRun.GetCompletedAt().Time) {
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
			if options.CheckName != nil && checkRun.CheckRun.GetName() != *options.CheckName {
				continue
			}
			if options.Status != nil && checkRun.CheckRun.GetStatus() != *options.Status {
				continue
			}
			if options.Conclusion != nil && checkRun.CheckRun.GetConclusion() != *options.Conclusion {
				continue
			}
			if options.AppID != nil {
				appID := checkRun.CheckRun.GetApp().GetID()
				if appID != *options.AppID {
					continue
				}
			}
			if options.Required != nil {
				if *checkRun.IsRequired != *options.Required {
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
