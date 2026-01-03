package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

type Discussion = client.Discussion

// GetDiscussion retrieves a specific discussion by number
func GetDiscussion(ctx context.Context, g *GitHubClient, repo repository.Repository, number any) ([]*github.Label, error) {
	discussionNumber, err := GetDiscussionNumber(number)
	if err != nil {
		return nil, fmt.Errorf("failed to parse discussion number: %w", err)
	}

	discussion, err := g.GetDiscussion(ctx, repo.Owner, repo.Name, discussionNumber)
	if err != nil {
		return nil, err
	}

	// Get label names from discussion
	var labelNames []string
	for _, labelNode := range discussion.Labels.Nodes {
		labelNames = append(labelNames, string(labelNode.Name))
	}

	// Get full label information using ListLabels
	return getLabelsFromNames(ctx, g, repo, labelNames)
}

// AddDiscussionLabels adds labels to a discussion
func AddDiscussionLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, number any, labelNames []string) ([]*github.Label, error) {
	discussionNumber, err := GetDiscussionNumber(number)
	if err != nil {
		return nil, fmt.Errorf("failed to parse discussion number: %w", err)
	}

	// Get discussion to retrieve its ID
	discussion, err := g.GetDiscussion(ctx, repo.Owner, repo.Name, discussionNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get discussion #%d in repository '%s/%s': %w", discussionNumber, repo.Owner, repo.Name, err)
	}

	// Get full label information from ListLabels
	labels, err := getLabelsFromNames(ctx, g, repo, labelNames)
	if err != nil {
		return nil, err
	}

	// Check if all requested labels were found and collect their NodeIDs
	var labelIDs []string
	for _, labelName := range labelNames {
		found := false
		for _, label := range labels {
			if label.Name != nil && *label.Name == labelName {
				if label.NodeID == nil {
					return nil, fmt.Errorf("label '%s' has no NodeID in repository '%s/%s'", labelName, repo.Owner, repo.Name)
				}
				labelIDs = append(labelIDs, *label.NodeID)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("label '%s' not found in repository '%s/%s'", labelName, repo.Owner, repo.Name)
		}
	}

	// Add labels to discussion
	discussionID := string(discussion.ID)
	if err := g.AddLabelsToDiscussion(ctx, discussionID, labelIDs); err != nil {
		return nil, fmt.Errorf("failed to add labels to discussion #%d in repository '%s/%s': %w", discussionNumber, repo.Owner, repo.Name, err)
	}

	// Retrieve updated labels using GetDiscussion which now uses ListLabels
	return GetDiscussion(ctx, g, repo, discussionNumber)
}

// RemoveDiscussionLabels removes labels from a discussion
func RemoveDiscussionLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, number any, labelNames []string) ([]*github.Label, error) {
	discussionNumber, err := GetDiscussionNumber(number)
	if err != nil {
		return nil, fmt.Errorf("failed to parse discussion number: %w", err)
	}

	// Get discussion to retrieve its ID and current labels
	discussion, err := g.GetDiscussion(ctx, repo.Owner, repo.Name, discussionNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get discussion #%d in repository '%s/%s': %w", discussionNumber, repo.Owner, repo.Name, err)
	}

	// Get full label information for the labels to remove
	labelsToRemove, err := getLabelsFromNames(ctx, g, repo, labelNames)
	if err != nil {
		return nil, err
	}

	// Check if all requested labels were found and collect their NodeIDs
	var labelIDs []string
	for _, labelName := range labelNames {
		found := false
		for _, label := range labelsToRemove {
			if label.Name != nil && *label.Name == labelName {
				if label.NodeID == nil {
					return nil, fmt.Errorf("label '%s' has no NodeID in repository '%s/%s'", labelName, repo.Owner, repo.Name)
				}
				labelIDs = append(labelIDs, *label.NodeID)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("label '%s' not found in repository '%s/%s'", labelName, repo.Owner, repo.Name)
		}
	}

	// Remove labels from discussion
	discussionID := string(discussion.ID)
	if err := g.RemoveLabelsFromDiscussion(ctx, discussionID, labelIDs); err != nil {
		return nil, fmt.Errorf("failed to remove labels from discussion #%d in repository '%s/%s': %w", discussionNumber, repo.Owner, repo.Name, err)
	}

	// Retrieve updated labels using GetDiscussion which now uses ListLabels
	return GetDiscussion(ctx, g, repo, discussionNumber)
}

// ClearDiscussionLabels removes all labels from a discussion
func ClearDiscussionLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, number any) error {
	discussionNumber, err := GetDiscussionNumber(number)
	if err != nil {
		return fmt.Errorf("failed to parse discussion number: %w", err)
	}

	// Get discussion to retrieve its ID and current labels
	discussion, err := g.GetDiscussion(ctx, repo.Owner, repo.Name, discussionNumber)
	if err != nil {
		return fmt.Errorf("failed to get discussion #%d in repository '%s/%s': %w", discussionNumber, repo.Owner, repo.Name, err)
	}

	// Get current labels from discussion
	var labelIDs []string
	for _, labelNode := range discussion.Labels.Nodes {
		labelIDs = append(labelIDs, string(labelNode.ID))
	}

	// If no labels, nothing to do
	if len(labelIDs) == 0 {
		return nil
	}

	// Remove all labels from discussion
	discussionID := string(discussion.ID)
	if err := g.RemoveLabelsFromDiscussion(ctx, discussionID, labelIDs); err != nil {
		return fmt.Errorf("failed to clear labels from discussion #%d in repository '%s/%s': %w", discussionNumber, repo.Owner, repo.Name, err)
	}

	return nil
}

// SetDiscussionLabels sets (replaces) all labels for a discussion
func SetDiscussionLabels(ctx context.Context, g *GitHubClient, repo repository.Repository, number any, labelNames []string) ([]*github.Label, error) {
	discussionNumber, err := GetDiscussionNumber(number)
	if err != nil {
		return nil, fmt.Errorf("failed to parse discussion number: %w", err)
	}

	// Get discussion to retrieve its ID and current labels
	discussion, err := g.GetDiscussion(ctx, repo.Owner, repo.Name, discussionNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get discussion #%d in repository '%s/%s': %w", discussionNumber, repo.Owner, repo.Name, err)
	}

	// Get current label IDs from discussion
	var currentLabelIDs []string
	for _, labelNode := range discussion.Labels.Nodes {
		currentLabelIDs = append(currentLabelIDs, string(labelNode.ID))
	}

	discussionID := string(discussion.ID)

	// Remove all current labels if any exist
	if len(currentLabelIDs) > 0 {
		if err := g.RemoveLabelsFromDiscussion(ctx, discussionID, currentLabelIDs); err != nil {
			return nil, fmt.Errorf("failed to clear labels from discussion #%d in repository '%s/%s': %w", discussionNumber, repo.Owner, repo.Name, err)
		}
	}

	// If no new labels to add, return empty list
	if len(labelNames) == 0 {
		return []*github.Label{}, nil
	}

	// Get full label information for the new labels
	labels, err := getLabelsFromNames(ctx, g, repo, labelNames)
	if err != nil {
		return nil, err
	}

	// Check if all requested labels were found and collect their NodeIDs
	var labelIDs []string
	for _, labelName := range labelNames {
		found := false
		for _, label := range labels {
			if label.Name != nil && *label.Name == labelName {
				if label.NodeID == nil {
					return nil, fmt.Errorf("label '%s' has no NodeID in repository '%s/%s'", labelName, repo.Owner, repo.Name)
				}
				labelIDs = append(labelIDs, *label.NodeID)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("label '%s' not found in repository '%s/%s'", labelName, repo.Owner, repo.Name)
		}
	}

	// Add new labels to discussion
	if err := g.AddLabelsToDiscussion(ctx, discussionID, labelIDs); err != nil {
		return nil, fmt.Errorf("failed to add labels to discussion #%d in repository '%s/%s': %w", discussionNumber, repo.Owner, repo.Name, err)
	}

	// Retrieve updated labels using GetDiscussion which now uses ListLabels
	return GetDiscussion(ctx, g, repo, discussionNumber)
}

// GetDiscussionNumber extracts a number from various types
func GetDiscussionNumber(number any) (int, error) {
	switch t := number.(type) {
	case string:
		return parser.GetDiscussionNumberFromString(t)
	case int:
		return t, nil
	case *Discussion:
		return int(t.Number), nil
	case *github.Discussion:
		return t.GetNumber(), nil
	default:
		return 0, fmt.Errorf("unsupported number type: %T", number)
	}
}

// getLabelsFromNames retrieves full label information from label names using ListLabels
func getLabelsFromNames(ctx context.Context, g *GitHubClient, repo repository.Repository, labelNames []string) ([]*github.Label, error) {
	if len(labelNames) == 0 {
		return []*github.Label{}, nil
	}

	// Get all labels in the repository
	allLabels, err := g.ListLabels(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list labels in repository '%s/%s': %w", repo.Owner, repo.Name, err)
	}

	// Create a map for quick lookup
	labelMap := make(map[string]*github.Label)
	for _, label := range allLabels {
		if label.Name != nil {
			labelMap[*label.Name] = label
		}
	}

	// Build the result list maintaining the order of labelNames
	var result []*github.Label
	for _, name := range labelNames {
		if label, ok := labelMap[name]; ok {
			result = append(result, label)
		}
	}

	return result, nil
}

// SearchDiscussions searches discussions in a repository
func SearchDiscussions(ctx context.Context, g *GitHubClient, repo repository.Repository, query string) ([]Discussion, error) {
	searchQuery := fmt.Sprintf("repo:%s/%s %s", repo.Owner, repo.Name, query)
	discussions, err := g.SearchDiscussions(ctx, searchQuery, 100)
	if err != nil {
		return nil, err
	}
	return discussions, nil
}
