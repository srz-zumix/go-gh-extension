package gh

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// IssueIdentifier represents different ways to identify an issue.
//
// It can identify an issue by:
//   - Number: The issue number (set when identified by number or URL).
//   - URL: The issue URL (set when identified by URL).
//   - Repo: The repository containing the issue (set when identified by URL).
//
// Only one or a subset of these fields may be set depending on how the issue is identified.
type IssueIdentifier struct {
	Number *int                   // Issue number, set when identified by number or URL
	URL    *string                // Issue URL, set when identified by URL
	Repo   *repository.Repository // Repository, set when identified by URL
}

func (ii *IssueIdentifier) String() string {
	repo := ""
	if ii.Repo != nil {
		repo = fmt.Sprintf("%s/%s ", ii.Repo.Owner, ii.Repo.Name)
	}
	if ii.Number != nil {
		return fmt.Sprintf("%s#%d", repo, *ii.Number)
	}
	if ii.URL != nil {
		return *ii.URL
	}
	return "<empty>"
}

// ParseIssueIdentifier parses a string that could be an issue number or URL
func ParseIssueIdentifier(input string) (*IssueIdentifier, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return &IssueIdentifier{}, nil
	}

	// Try to parse as issue number with # prefix (e.g., #123)
	if strings.HasPrefix(input, "#") {
		numStr := strings.TrimPrefix(input, "#")
		if num, err := strconv.Atoi(numStr); err == nil && num > 0 {
			return &IssueIdentifier{Number: &num}, nil
		}
	}

	// Try to parse as issue number
	if num, err := strconv.Atoi(input); err == nil && num > 0 {
		return &IssueIdentifier{Number: &num}, nil
	}

	// Try to parse as URL
	issueURL, err := parser.ParseIssueURL(input)
	if err != nil {
		return nil, err
	}
	if issueURL != nil {
		return &IssueIdentifier{
			Number: issueURL.Number,
			URL:    &input,
			Repo:   issueURL.Repo,
		}, nil
	}

	return nil, fmt.Errorf("unable to parse issue identifier: %s", input)
}

// FindIssueByIdentifier finds an issue by identifier (number or URL)
func FindIssueByIdentifier(ctx context.Context, g *GitHubClient, repo repository.Repository, identifier string) (*github.Issue, error) {
	issueID, err := ParseIssueIdentifier(identifier)
	if err != nil {
		return nil, err
	}
	if issueID.Repo == nil {
		issueID.Repo = &repo
	}

	// If we have an issue number, fetch it directly
	if issueID.Number != nil {
		return GetIssue(ctx, g, *issueID.Repo, *issueID.Number)
	}

	return nil, fmt.Errorf("unable to identify issue from: %s", issueID.String())
}

func GetIssue(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any) (*github.Issue, error) {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return nil, err
	}
	return g.GetIssueByNumber(ctx, repo.Owner, repo.Name, number)
}

func GetIssueNumber(issue any) (int, error) {
	switch t := issue.(type) {
	case string:
		return parser.GetIssueNumberFromString(t)
	case int:
		return t, nil
	case *github.Issue:
		return t.GetNumber(), nil
	case *github.PullRequest:
		return t.GetNumber(), nil
	default:
		return 0, fmt.Errorf("unsupported issue type: %T", issue)
	}
}

func SetIssueLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any, labels []string) ([]*github.Label, error) {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issue number from '%s': %w", issue, err)
	}
	result, err := g.ReplaceIssueLabels(ctx, repo.Owner, repo.Name, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to set labels for issue #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return result, nil
}

func AddIssueLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any, labels []string) ([]*github.Label, error) {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issue number from '%s': %w", issue, err)
	}
	result, err := g.AddIssueLabels(ctx, repo.Owner, repo.Name, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to add labels to issue #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return result, nil
}

func RemoveIssueLabel(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any, label string) error {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return fmt.Errorf("failed to parse issue number from '%s': %w", issue, err)
	}
	err = g.RemoveIssueLabel(ctx, repo.Owner, repo.Name, number, label)
	if err != nil {
		return fmt.Errorf("failed to remove label '%s' from issue #%d in repository '%s/%s': %w", label, number, repo.Owner, repo.Name, err)
	}
	return nil
}

func RemoveIssueLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any, label []string) ([]*github.Label, error) {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issue number from '%s': %w", issue, err)
	}

	i, err := g.GetIssueByNumber(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	var replaceLabels []string
	for _, l := range i.Labels {
		name := l.GetName()
		if slices.Contains(label, name) {
			continue
		}
		replaceLabels = append(replaceLabels, name)
	}
	return g.ReplaceIssueLabels(ctx, repo.Owner, repo.Name, number, replaceLabels)
}

func ClearIssueLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any) error {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return fmt.Errorf("failed to parse issue number from '%s': %w", issue, err)
	}
	err = g.ClearIssueLabels(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return fmt.Errorf("failed to clear labels for issue #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return nil
}

func CreateIssueComment(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any, body string) (*github.IssueComment, error) {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issue number from '%s': %w", issue, err)
	}
	comment, err := g.CreateIssueComment(ctx, repo.Owner, repo.Name, number, body)
	if err != nil {
		return nil, fmt.Errorf("failed to comment on issue #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return comment, nil
}

func DeleteIssueComment(ctx context.Context, g *GitHubClient, repo repository.Repository, comment any) error {
	commentID, err := GetCommentID(comment)
	if err != nil {
		return fmt.Errorf("failed to parse comment ID from '%s': %w", comment, err)
	}
	return g.DeleteIssueComment(ctx, repo.Owner, repo.Name, commentID)
}

func EditIssueComment(ctx context.Context, g *GitHubClient, repo repository.Repository, comment any, body string) (*github.IssueComment, error) {
	commentID, err := GetCommentID(comment)
	if err != nil {
		return nil, fmt.Errorf("failed to parse comment ID from '%s': %w", comment, err)
	}
	return g.EditIssueComment(ctx, repo.Owner, repo.Name, commentID, body)
}

func ListIssueComments(ctx context.Context, g *GitHubClient, repo repository.Repository, issue any) ([]*github.IssueComment, error) {
	number, err := GetIssueNumber(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issue number from '%s': %w", issue, err)
	}
	comments, err := g.ListIssueComments(ctx, repo.Owner, repo.Name, number)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments for issue #%d in repository '%s/%s': %w", number, repo.Owner, repo.Name, err)
	}
	return comments, nil
}

func GetCommentID(comment any) (int64, error) {
	switch c := comment.(type) {
	case *github.IssueComment:
		return c.GetID(), nil
	case *github.PullRequestComment:
		return c.GetID(), nil
	case int64:
		return c, nil
	}
	return 0, fmt.Errorf("failed to get comment ID from '%v'", comment)
}

// SearchIssues searches issues in a repository
func SearchIssues(ctx context.Context, g *GitHubClient, repo repository.Repository, query string) ([]*github.Issue, error) {
	searchQuery := fmt.Sprintf("repo:%s/%s %s", repo.Owner, repo.Name, query)
	issues, err := g.SearchIssues(ctx, searchQuery)
	if err != nil {
		return nil, err
	}
	return issues, nil
}
