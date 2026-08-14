package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/shurcooL/githubv4"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// UnarchiveProjectV2Item unarchives an item in a GitHub Project v2.
func UnarchiveProjectV2Item(ctx context.Context, g *GitHubClient, projectID, itemID string) error {
	return g.UnarchiveProjectV2Item(ctx, client.UnarchiveProjectV2ItemInput{
		ProjectID: githubv4.ID(projectID),
		ItemID:    githubv4.ID(itemID),
	})
}

// ClearProjectV2ItemFieldValue clears the value of a custom field for a project item.
func ClearProjectV2ItemFieldValue(ctx context.Context, g *GitHubClient, projectID, itemID, fieldID string) error {
	return g.ClearProjectV2ItemFieldValue(ctx, client.ClearProjectV2ItemFieldValueInput{
		ProjectID: githubv4.ID(projectID),
		ItemID:    githubv4.ID(itemID),
		FieldID:   githubv4.ID(fieldID),
	})
}

// MoveProjectV2Item moves an item within a GitHub Project v2.
// The item is placed after afterItemID; pass nil to move the item to the top.
func MoveProjectV2Item(ctx context.Context, g *GitHubClient, projectID, itemID string, afterItemID *string) error {
	input := client.UpdateProjectV2ItemPositionInput{
		ProjectID: githubv4.ID(projectID),
		ItemID:    githubv4.ID(itemID),
	}
	if afterItemID != nil {
		id := githubv4.ID(*afterItemID)
		input.AfterID = &id
	}
	return g.UpdateProjectV2ItemPosition(ctx, input)
}

// UpdateProjectV2DraftIssue updates the title and/or body of a draft issue in a GitHub Project v2.
// Pass nil for any field that should not be updated.
func UpdateProjectV2DraftIssue(ctx context.Context, g *GitHubClient, draftIssueID string, title *string, body *string) error {
	input := client.UpdateProjectV2DraftIssueInput{
		DraftIssueID: githubv4.ID(draftIssueID),
	}
	if title != nil {
		t := githubv4.String(*title)
		input.Title = &t
	}
	if body != nil {
		b := githubv4.String(*body)
		input.Body = &b
	}
	return g.UpdateProjectV2DraftIssue(ctx, input)
}

// ConvertProjectV2DraftIssueItemToIssue converts a draft issue item into an issue in the given repository.
// Returns the resulting project item's node ID.
func ConvertProjectV2DraftIssueItemToIssue(ctx context.Context, g *GitHubClient, itemID string, repo repository.Repository) (string, error) {
	repoID, err := GetRepositoryNodeID(ctx, g, repo)
	if err != nil {
		return "", err
	}
	return g.ConvertProjectV2DraftIssueItemToIssue(ctx, client.ConvertProjectV2DraftIssueItemToIssueInput{
		ItemID:       githubv4.ID(itemID),
		RepositoryID: githubv4.ID(repoID),
	})
}
