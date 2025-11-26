package gh

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/srz-zumix/go-gh-extension/pkg/git"
)

// PRIdentifier represents different ways to identify a pull request.
//
// It can identify a pull request by:
//   - Number: The pull request number (set when identified by number or URL).
//   - Head: The branch name (set when identified by branch name).
//   - URL: The pull request URL (set when identified by URL).
//   - Repo: The repository containing the pull request (set when identified by URL).
//
// Only one or a subset of these fields may be set depending on how the pull request is identified.
type PRIdentifier struct {
	Number *int    // Pull request number, set when identified by number or URL
	Head   *string // Branch name, set when identified by branch name
	URL    *string // Pull request URL, set when identified by URL
	Repo   *repository.Repository // Repository, set when identified by URL
}

func (pri *PRIdentifier) String() string {
	repo := ""
	if pri.Repo != nil {
		repo = fmt.Sprintf("%s/%s ", pri.Repo.Owner, pri.Repo.Name)
	}
	if pri.Number != nil {
		return fmt.Sprintf("%s #%d", repo, *pri.Number)
	}
	if pri.URL != nil {
		return *pri.URL
	}
	if pri.Head != nil {
		return fmt.Sprintf("%s/%s", repo, *pri.Head)
	}
	return "<empty>"
}

// ParsePRIdentifier parses a string that could be a PR number, URL, or branch name
func ParsePRIdentifier(input string) (*PRIdentifier, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return &PRIdentifier{}, nil
	}

	// Try to parse as PR number with # prefix (e.g., #123)
	if strings.HasPrefix(input, "#") {
		numStr := strings.TrimPrefix(input, "#")
		if num, err := strconv.Atoi(numStr); err == nil && num > 0 {
			return &PRIdentifier{Number: &num}, nil
		}
	}

	// Try to parse as PR number
	if num, err := strconv.Atoi(input); err == nil && num > 0 {
		return &PRIdentifier{Number: &num}, nil
	}

	// Try to parse as URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		parsedURL, err := url.Parse(input)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}

		// Extract PR number from URL path
		// Expected format: /owner/repo/pull/123
		pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(pathParts) >= 4 && pathParts[2] == "pull" {
			if num, err := strconv.Atoi(pathParts[3]); err == nil && num > 0 {
				repo, err := repository.Parse(input)
				if err != nil {
					return nil, fmt.Errorf("unable to parse repository from URL: %w", err)
				}
				return &PRIdentifier{
					Number: &num,
					URL:    &input,
					Repo:   &repo,
				}, nil
			}
		}
		return nil, fmt.Errorf("unable to extract PR number from URL: %s", input)
	}

	// Assume it's a branch name
	return &PRIdentifier{Head: &input}, nil
}

// FindPRByIdentifier finds a pull request by identifier (number, URL, or branch name)
func FindPRByIdentifier(ctx context.Context, g *GitHubClient, repo repository.Repository, identifier string) (*github.PullRequest, error) {
	prID, err := ParsePRIdentifier(identifier)
	if err != nil {
		return nil, err
	}
	if prID.Repo == nil {
		prID.Repo = &repo
	}

	// If we have a PR number, fetch it directly
	if prID.Number != nil {
		return GetPullRequest(ctx, g, repo, *prID.Number)
	}

	if prID.Head == nil {
		currentBranch, err := git.GetCurrentBranchIfRepoMatches(ctx, *prID.Repo)
		if err == nil {
			prID.Head = &currentBranch
		}
	}

	// If we have a head branch, search for PRs with that head
	if prID.Head != nil {
		return FindPRByHead(ctx, g, repo, *prID.Head)
	}

	return nil, fmt.Errorf("unable to identify PR from: %s", prID.String())
}

// FindPRByHead finds a pull request by head branch name using GraphQL
func FindPRByHead(ctx context.Context, g *GitHubClient, repo repository.Repository, head string) (*github.PullRequest, error) {
	// Try ListPullRequests with head filter first (faster and works even if ref doesn't exist)
	pr, err := findPRByHeadWithListAPI(ctx, g, repo, head)
	if err == nil && pr != nil {
		return pr, nil
	}

	// Use GraphQL to find PR associated with the ref
	pr, err = findPRByRefWithGraphQL(ctx, g, repo, head)
	if err != nil {
		return nil, err
	}
	if pr != nil {
		return pr, nil
	}

	return nil, fmt.Errorf("pull request not found with head branch: %s", head)
}

// findPRByRefWithGraphQL finds a pull request associated with a ref using GraphQL
func findPRByRefWithGraphQL(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string) (*github.PullRequest, error) {
	// Fallback to GraphQL query for associated PRs
	orderBy := client.GraphQLOrderByOption{}
	orderBy.CreatedAt()
	orderBy.Desc()
	nodes, err := GetAssociatedPullRequestsForRef(ctx, g, repo, ref, &AssociatedPullRequestsOptionOrderBy{OrderBy: orderBy})
	if err != nil {
		return nil, err
	}

	// If no PRs found, return nil
	if len(nodes) == 0 {
		return nil, nil
	}

	// Return the first PR (most recent)
	return nodes[0], nil
}

// findPRByHeadWithListAPI finds a pull request by head branch using REST API
func findPRByHeadWithListAPI(ctx context.Context, g *GitHubClient, repo repository.Repository, head string) (*github.PullRequest, error) {
	if head == "" {
		return nil, fmt.Errorf("head branch name is empty")
	}
	// Format head as "owner:branch" for API
	headRef := fmt.Sprintf("%s:%s", repo.Owner, head)

	// Search for all PRs with matching head (open, closed, merged)
	prs, err := ListPullRequests(ctx, g, repo,
		&ListPullRequestsOptionHead{Head: headRef},
		ListPullRequestsOptionStateAll(),
		ListPullRequestsOptionSortCreated(),
		ListPullRequestsOptionDirectionDescending(),
	)
	if err != nil {
		return nil, err
	}

	if len(prs) == 0 {
		return nil, nil
	}

	// Return the most recent PR
	return prs[0], nil
}
