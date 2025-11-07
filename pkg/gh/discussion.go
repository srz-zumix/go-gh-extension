package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

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

	// Convert GraphQL labels to github.Label
	var labels []*github.Label
	for _, labelNode := range discussion.Labels.Nodes {
		name := string(labelNode.Name)
		color := string(labelNode.Color)
		var desc *string
		if labelNode.Description != nil {
			d := string(*labelNode.Description)
			desc = &d
		}
		labels = append(labels, &github.Label{
			Name:        &name,
			Color:       &color,
			Description: desc,
		})
	}

	return labels, nil
}

// GetDiscussionNumber extracts a number from various types
func GetDiscussionNumber(number any) (int, error) {
	switch t := number.(type) {
	case string:
		return GetNumberFromString(t)
	case int:
		return t, nil
	case *client.Discussion:
		return int(t.Number), nil
	default:
		return 0, fmt.Errorf("unsupported number type: %T", number)
	}
}
