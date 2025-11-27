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
// Expected URL format: https://github.com/owner/repo/pull/123
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
	if len(pathParts) < 4 {
		return nil, fmt.Errorf("invalid GitHub PR URL format: %s", input)
	}

	if pathParts[2] != "pull" {
		return nil, fmt.Errorf("not a pull request URL: %s", input)
	}

	num, err := strconv.Atoi(pathParts[3])
	if err != nil || num <= 0 {
		return nil, fmt.Errorf("invalid PR number in URL: %s", input)
	}

	repo := repository.Repository{
		Host:  parsedURL.Host,
		Owner: pathParts[0],
		Name:  pathParts[1],
	}

	return &PullRequestURL{
		Url:    parsedURL,
		Number: &num,
		Repo:   &repo,
	}, nil
}
