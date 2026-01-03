package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// GitHubURL represents a parsed GitHub URL with common fields
type GitHubURL struct {
	Url       *url.URL
	Repo      *repository.Repository
	PathParts []string
}

// ParseGitHubURL parses a GitHub URL and extracts the repository information and path parts.
// Returns nil, nil for empty strings or non-HTTP(S) URLs.
// Returns an error for malformed URLs or URLs with insufficient path parts.
func ParseGitHubURL(input string) (*GitHubURL, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return nil, nil
	}

	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL format: %s", input)
	}

	repo := repository.Repository{
		Host:  parsedURL.Host,
		Owner: pathParts[0],
		Name:  pathParts[1],
	}

	return &GitHubURL{
		Url:       parsedURL,
		Repo:      &repo,
		PathParts: pathParts,
	}, nil
}

// parseNumberFromPath extracts a positive integer from the path at the specified index.
// Returns the number and true if successful, or 0 and false if the index is out of bounds
// or the value is not a valid positive integer.
func parseNumberFromPath(pathParts []string, index int) (int, bool) {
	if index < 0 || len(pathParts) <= index {
		return 0, false
	}
	num, err := strconv.Atoi(pathParts[index])
	if err != nil || num <= 0 {
		return 0, false
	}
	return num, true
}

// PullRequestURL represents a pull request parsed from a URL
type PullRequestURL struct {
	Url    *url.URL
	Number *int
	Repo   *repository.Repository
}

// ParsePullRequestURL parses a GitHub pull request URL and extracts the PR number and repository information.
// Expected URL formats:
//   - https://github.com/owner/repo/pull/123
//   - https://github.com/owner/repo/actions/runs/123/job/456?pr=789 (from query parameter)
//
// Returns PullRequestURL containing the PR number and repository, or an error if parsing fails.
func ParsePullRequestURL(input string) (*PullRequestURL, error) {
	githubURL, err := ParseGitHubURL(input)
	if err != nil {
		return nil, err
	}
	if githubURL == nil {
		return nil, nil
	}

	// Check if this is a direct PR URL (/owner/repo/pull/123)
	if len(githubURL.PathParts) >= 4 && githubURL.PathParts[2] == "pull" {
		num, ok := parseNumberFromPath(githubURL.PathParts, 3)
		if !ok {
			return nil, fmt.Errorf("invalid PR number in URL: %s", input)
		}

		return &PullRequestURL{
			Url:    githubURL.Url,
			Number: &num,
			Repo:   githubURL.Repo,
		}, nil
	}

	// Check for PR number in query parameters (e.g., ?pr=123)
	if prParam := githubURL.Url.Query().Get("pr"); prParam != "" {
		num, err := strconv.Atoi(prParam)
		if err != nil || num <= 0 {
			return nil, fmt.Errorf("invalid PR number in query parameter: %s", input)
		}

		return &PullRequestURL{
			Url:    githubURL.Url,
			Number: &num,
			Repo:   githubURL.Repo,
		}, nil
	}

	return nil, fmt.Errorf("not a pull request URL: %s", input)
}

// IssueURL represents an issue parsed from a URL
type IssueURL struct {
	Url    *url.URL
	Number *int
	Repo   *repository.Repository
}

// ParseIssueURL parses a GitHub issue or pull request URL and extracts the issue number and repository information.
// Expected URL formats:
//   - https://github.com/owner/repo/issues/123
//   - https://github.com/owner/repo/pull/123
//
// Returns IssueURL containing the issue number and repository, or an error if parsing fails.
func ParseIssueURL(input string) (*IssueURL, error) {
	githubURL, err := ParseGitHubURL(input)
	if err != nil {
		return nil, err
	}
	if githubURL == nil {
		return nil, nil
	}

	// Check if this is a direct issue or pull request URL (/owner/repo/issues/123 or /owner/repo/pull/123)
	if len(githubURL.PathParts) >= 4 && (githubURL.PathParts[2] == "issues" || githubURL.PathParts[2] == "pull") {
		num, ok := parseNumberFromPath(githubURL.PathParts, 3)
		if !ok {
			return nil, fmt.Errorf("invalid issue number in URL: %s", input)
		}

		return &IssueURL{
			Url:    githubURL.Url,
			Number: &num,
			Repo:   githubURL.Repo,
		}, nil
	}

	return nil, fmt.Errorf("not an issue or pull request URL: %s", input)
}

// DiscussionURL represents a discussion parsed from a URL
type DiscussionURL struct {
	Url    *url.URL
	Number *int
	Repo   *repository.Repository
}

// ParseDiscussionURL parses a GitHub discussion URL and extracts the discussion number and repository information.
// Expected URL formats:
//   - https://github.com/owner/repo/discussions/123
//
// Returns DiscussionURL containing the discussion number and repository, or an error if parsing fails.
func ParseDiscussionURL(input string) (*DiscussionURL, error) {
	githubURL, err := ParseGitHubURL(input)
	if err != nil {
		return nil, err
	}
	if githubURL == nil {
		return nil, nil
	}

	// Check if this is a direct discussion URL (/owner/repo/discussions/123)
	if len(githubURL.PathParts) >= 4 && githubURL.PathParts[2] == "discussions" {
		num, ok := parseNumberFromPath(githubURL.PathParts, 3)
		if !ok {
			return nil, fmt.Errorf("invalid discussion number in URL: %s", input)
		}

		return &DiscussionURL{
			Url:    githubURL.Url,
			Number: &num,
			Repo:   githubURL.Repo,
		}, nil
	}

	return nil, fmt.Errorf("not a discussion URL: %s", input)
}
