// Package gh provides wrapper functions for GitHub Projects v2 API operations.
package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// Type aliases for ProjectV2 types from the client package.
type ProjectV2 = client.ProjectV2
type ProjectV2Field = client.ProjectV2Field
type ProjectV2SingleSelectOption = client.ProjectV2SingleSelectOption
type ProjectV2ItemType = client.ProjectV2ItemType
type ProjectV2ItemContent = client.ProjectV2ItemContent
type ProjectV2FieldValue = client.ProjectV2FieldValue
type ProjectV2Item = client.ProjectV2Item

// Re-export item type constants.
const (
	ProjectV2ItemTypeIssue       = client.ProjectV2ItemTypeIssue
	ProjectV2ItemTypePullRequest = client.ProjectV2ItemTypePullRequest
	ProjectV2ItemTypeDraftIssue  = client.ProjectV2ItemTypeDraftIssue
	ProjectV2ItemTypeRedacted    = client.ProjectV2ItemTypeRedacted
)

// GetProjectV2ByNumber retrieves a ProjectV2 by owner and project number.
// It auto-detects whether the owner is a user or organization.
func GetProjectV2ByNumber(ctx context.Context, g *GitHubClient, owner string, number int) (*ProjectV2, error) {
	ownerType, err := DetectOwnerType(ctx, g, owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.GetOrgProjectV2ByNumber(ctx, owner, number)
	case OwnerTypeUser:
		return g.GetUserProjectV2ByNumber(ctx, owner, number)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", owner)
	}
}

// ListProjectsV2 lists all ProjectV2s for a user or organization.
func ListProjectsV2(ctx context.Context, g *GitHubClient, owner string) ([]ProjectV2, error) {
	ownerType, err := DetectOwnerType(ctx, g, owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.ListOrgProjectsV2(ctx, owner, 100)
	case OwnerTypeUser:
		return g.ListUserProjectsV2(ctx, owner, 100)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", owner)
	}
}

// ListProjectV2Fields lists all fields for a ProjectV2.
// Built-in fields (TITLE, ASSIGNEES, LABELS, etc.) are included in results.
func ListProjectV2Fields(ctx context.Context, g *GitHubClient, owner string, number int) ([]ProjectV2Field, error) {
	ownerType, err := DetectOwnerType(ctx, g, owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.ListOrgProjectV2Fields(ctx, owner, number, 100)
	case OwnerTypeUser:
		return g.ListUserProjectV2Fields(ctx, owner, number, 100)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", owner)
	}
}

// ListProjectV2Items lists all items for a ProjectV2, including content and field values.
func ListProjectV2Items(ctx context.Context, g *GitHubClient, owner string, number int) ([]ProjectV2Item, error) {
	ownerType, err := DetectOwnerType(ctx, g, owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.ListOrgProjectV2Items(ctx, owner, number, 50)
	case OwnerTypeUser:
		return g.ListUserProjectV2Items(ctx, owner, number, 50)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", owner)
	}
}

// GetOwnerNodeID returns the GraphQL node ID for a user or organization login.
func GetOwnerNodeID(ctx context.Context, g *GitHubClient, login string) (githubv4.ID, error) {
	return g.GetOwnerNodeID(ctx, login)
}

// CreateProjectV2 creates a new GitHub Project v2 owned by the given owner.
func CreateProjectV2(ctx context.Context, g *GitHubClient, ownerID githubv4.ID, title string) (*ProjectV2, error) {
	return g.CreateProjectV2(ctx, client.CreateProjectV2Input{
		OwnerID: ownerID,
		Title:   githubv4.String(title),
	})
}

// UpdateProjectV2Metadata updates the description, readme, and visibility of a ProjectV2.
// Pass nil for any field that should not be updated.
func UpdateProjectV2Metadata(ctx context.Context, g *GitHubClient, projectID string, shortDesc *string, readme *string, public *bool) (*ProjectV2, error) {
	input := client.UpdateProjectV2Input{
		ProjectID: githubv4.ID(projectID),
	}
	if shortDesc != nil {
		s := githubv4.String(*shortDesc)
		input.ShortDescription = &s
	}
	if readme != nil {
		s := githubv4.String(*readme)
		input.Readme = &s
	}
	if public != nil {
		b := githubv4.Boolean(*public)
		input.Public = &b
	}
	return g.UpdateProjectV2(ctx, input)
}

// DeleteProjectV2 deletes a GitHub Project v2 by its node ID.
func DeleteProjectV2(ctx context.Context, g *GitHubClient, projectID string) error {
	return g.DeleteProjectV2(ctx, client.DeleteProjectV2Input{
		ProjectID: githubv4.ID(projectID),
	})
}

// CreateProjectV2Field creates a custom field in a GitHub Project v2.
// For SINGLE_SELECT fields, pass the options; for others, pass nil.
// Returns the created field's ID and name.
func CreateProjectV2Field(ctx context.Context, g *GitHubClient, projectID string, dataType string, name string, options []ProjectV2SingleSelectOption) (string, string, error) {
	input := client.CreateProjectV2FieldInput{
		ProjectID: githubv4.ID(projectID),
		DataType:  githubv4.String(dataType),
		Name:      githubv4.String(name),
	}
	if len(options) > 0 {
		opts := make([]client.CreateProjectV2FieldSingleSelectOptionInput, len(options))
		for i, o := range options {
			opts[i] = client.CreateProjectV2FieldSingleSelectOptionInput{
				Name:        githubv4.String(o.Name),
				Color:       githubv4.String(o.Color),
				Description: githubv4.String(o.Description),
			}
		}
		input.SingleSelectOptions = opts
	}
	return g.CreateProjectV2Field(ctx, input)
}

// AddProjectV2DraftIssue adds a draft issue to a GitHub Project v2.
// Returns the created item's node ID.
func AddProjectV2DraftIssue(ctx context.Context, g *GitHubClient, projectID string, title string, body string) (string, error) {
	b := githubv4.String(body)
	return g.AddProjectV2DraftIssue(ctx, client.AddProjectV2DraftIssueInput{
		ProjectID: githubv4.ID(projectID),
		Title:     githubv4.String(title),
		Body:      &b,
	})
}

// SetProjectV2ItemTextValue sets a TEXT field value for a project item.
func SetProjectV2ItemTextValue(ctx context.Context, g *GitHubClient, projectID, itemID, fieldID, text string) error {
	t := githubv4.String(text)
	return g.UpdateProjectV2ItemFieldValue(ctx, client.UpdateProjectV2ItemFieldValueInput{
		ProjectID: githubv4.ID(projectID),
		ItemID:    githubv4.ID(itemID),
		FieldID:   githubv4.ID(fieldID),
		Value:     client.ProjectV2FieldValueInput{Text: &t},
	})
}

// SetProjectV2ItemNumberValue sets a NUMBER field value for a project item.
func SetProjectV2ItemNumberValue(ctx context.Context, g *GitHubClient, projectID, itemID, fieldID string, number float64) error {
	n := githubv4.Float(number)
	return g.UpdateProjectV2ItemFieldValue(ctx, client.UpdateProjectV2ItemFieldValueInput{
		ProjectID: githubv4.ID(projectID),
		ItemID:    githubv4.ID(itemID),
		FieldID:   githubv4.ID(fieldID),
		Value:     client.ProjectV2FieldValueInput{Number: &n},
	})
}

// SetProjectV2ItemDateValue sets a DATE field value for a project item (YYYY-MM-DD format).
func SetProjectV2ItemDateValue(ctx context.Context, g *GitHubClient, projectID, itemID, fieldID, dateStr string) error {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format %q (expected YYYY-MM-DD): %w", dateStr, err)
	}
	d := githubv4.Date{Time: t}
	return g.UpdateProjectV2ItemFieldValue(ctx, client.UpdateProjectV2ItemFieldValueInput{
		ProjectID: githubv4.ID(projectID),
		ItemID:    githubv4.ID(itemID),
		FieldID:   githubv4.ID(fieldID),
		Value:     client.ProjectV2FieldValueInput{Date: &d},
	})
}

// SetProjectV2ItemSingleSelectValue sets a SINGLE_SELECT field value for a project item.
// optionID is the destination field option's node ID.
func SetProjectV2ItemSingleSelectValue(ctx context.Context, g *GitHubClient, projectID, itemID, fieldID, optionID string) error {
	opt := githubv4.String(optionID)
	return g.UpdateProjectV2ItemFieldValue(ctx, client.UpdateProjectV2ItemFieldValueInput{
		ProjectID: githubv4.ID(projectID),
		ItemID:    githubv4.ID(itemID),
		FieldID:   githubv4.ID(fieldID),
		Value:     client.ProjectV2FieldValueInput{SingleSelectOptionID: &opt},
	})
}

// DeleteProjectV2Item removes an item from a GitHub Project v2.
func DeleteProjectV2Item(ctx context.Context, g *GitHubClient, projectID, itemID string) error {
	return g.DeleteProjectV2Item(ctx, client.DeleteProjectV2ItemInput{
		ProjectID: githubv4.ID(projectID),
		ItemID:    githubv4.ID(itemID),
	})
}
