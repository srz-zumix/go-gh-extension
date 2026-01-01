package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// PullRequestURL represents a pull request parsed from a URL
type PullRequestURL struct {
	Url    *url.URL
	Number *int
	Repo   *repository.Repository
}

// ParsePRURL parses a GitHub pull request URL and extracts the PR number and repository information.
// Expected URL formats:
//   - https://github.com/owner/repo/pull/123
//   - https://github.com/owner/repo/actions/runs/123/job/456?pr=789 (from query parameter)
//
// Returns PullRequestURL containing the PR number and repository, or an error if parsing fails.
func ParsePullRequestURL(input string) (*PullRequestURL, error) {
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

	// Extract PR number and repository from URL path
	// Expected format: /owner/repo/pull/123
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL format: %s", input)
	}

	// Extract repository information
	repo := repository.Repository{
		Host:  parsedURL.Host,
		Owner: pathParts[0],
		Name:  pathParts[1],
	}

	// Check if this is a direct PR URL (/owner/repo/pull/123)
	if len(pathParts) >= 4 && pathParts[2] == "pull" {
		num, err := strconv.Atoi(pathParts[3])
		if err != nil || num <= 0 {
			return nil, fmt.Errorf("invalid PR number in URL: %s", input)
		}

		return &PullRequestURL{
			Url:    parsedURL,
			Number: &num,
			Repo:   &repo,
		}, nil
	}

	// Check for PR number in query parameters (e.g., ?pr=123)
	if prParam := parsedURL.Query().Get("pr"); prParam != "" {
		num, err := strconv.Atoi(prParam)
		if err != nil || num <= 0 {
			return nil, fmt.Errorf("invalid PR number in query parameter: %s", input)
		}

		return &PullRequestURL{
			Url:    parsedURL,
			Number: &num,
			Repo:   &repo,
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

	// Extract issue number and repository from URL path
	// Expected format: /owner/repo/issues/123 or /owner/repo/pull/123
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL format: %s", input)
	}

	// Extract repository information
	repo := repository.Repository{
		Host:  parsedURL.Host,
		Owner: pathParts[0],
		Name:  pathParts[1],
	}

	// Check if this is a direct issue or pull request URL (/owner/repo/issues/123 or /owner/repo/pull/123)
	if len(pathParts) >= 4 && (pathParts[2] == "issues" || pathParts[2] == "pull") {
		num, err := strconv.Atoi(pathParts[3])
		if err != nil || num <= 0 {
			return nil, fmt.Errorf("invalid issue number in URL: %s", input)
		}

		return &IssueURL{
			Url:    parsedURL,
			Number: &num,
			Repo:   &repo,
		}, nil
	}

	return nil, fmt.Errorf("not an issue or pull request URL: %s", input)
}
