package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// Type aliases for ProjectV2 status update types from the client package.
type ProjectV2StatusUpdate = client.ProjectV2StatusUpdate
type ProjectV2StatusUpdateStatus = client.ProjectV2StatusUpdateStatus

// Re-export status update status constants.
const (
	ProjectV2StatusUpdateStatusInactive = client.ProjectV2StatusUpdateStatusInactive
	ProjectV2StatusUpdateStatusOnTrack  = client.ProjectV2StatusUpdateStatusOnTrack
	ProjectV2StatusUpdateStatusAtRisk   = client.ProjectV2StatusUpdateStatusAtRisk
	ProjectV2StatusUpdateStatusOffTrack = client.ProjectV2StatusUpdateStatusOffTrack
	ProjectV2StatusUpdateStatusComplete = client.ProjectV2StatusUpdateStatusComplete
)

// ListProjectV2StatusUpdates lists all status updates for a ProjectV2.
func ListProjectV2StatusUpdates(ctx context.Context, g *GitHubClient, owner string, number int) ([]ProjectV2StatusUpdate, error) {
	ownerType, err := DetectOwnerType(ctx, g, owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.ListOrgProjectV2StatusUpdates(ctx, owner, number, 50)
	case OwnerTypeUser:
		return g.ListUserProjectV2StatusUpdates(ctx, owner, number, 50)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", owner)
	}
}

// parseProjectV2Date parses a YYYY-MM-DD date string into a githubv4.Date.
func parseProjectV2Date(dateStr string) (*githubv4.Date, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format %q (expected YYYY-MM-DD): %w", dateStr, err)
	}
	return &githubv4.Date{Time: t}, nil
}

// CreateProjectV2StatusUpdate creates a status update on a GitHub Project v2.
// Pass nil for any optional value that should not be set. Dates use the YYYY-MM-DD format.
// Returns the created status update's node ID.
func CreateProjectV2StatusUpdate(ctx context.Context, g *GitHubClient, projectID string, body *string, status *ProjectV2StatusUpdateStatus, startDate *string, targetDate *string) (string, error) {
	input := client.CreateProjectV2StatusUpdateInput{
		ProjectID: githubv4.ID(projectID),
	}
	if body != nil {
		b := githubv4.String(*body)
		input.Body = &b
	}
	if status != nil {
		s := githubv4.String(*status)
		input.Status = &s
	}
	if startDate != nil {
		d, err := parseProjectV2Date(*startDate)
		if err != nil {
			return "", err
		}
		input.StartDate = d
	}
	if targetDate != nil {
		d, err := parseProjectV2Date(*targetDate)
		if err != nil {
			return "", err
		}
		input.TargetDate = d
	}
	return g.CreateProjectV2StatusUpdate(ctx, input)
}

// UpdateProjectV2StatusUpdate updates a status update on a GitHub Project v2.
// Pass nil for any value that should not be updated. Dates use the YYYY-MM-DD format.
func UpdateProjectV2StatusUpdate(ctx context.Context, g *GitHubClient, statusUpdateID string, body *string, status *ProjectV2StatusUpdateStatus, startDate *string, targetDate *string) error {
	input := client.UpdateProjectV2StatusUpdateInput{
		StatusUpdateID: githubv4.ID(statusUpdateID),
	}
	if body != nil {
		b := githubv4.String(*body)
		input.Body = &b
	}
	if status != nil {
		s := githubv4.String(*status)
		input.Status = &s
	}
	if startDate != nil {
		d, err := parseProjectV2Date(*startDate)
		if err != nil {
			return err
		}
		input.StartDate = d
	}
	if targetDate != nil {
		d, err := parseProjectV2Date(*targetDate)
		if err != nil {
			return err
		}
		input.TargetDate = d
	}
	return g.UpdateProjectV2StatusUpdate(ctx, input)
}

// DeleteProjectV2StatusUpdate deletes a status update from a GitHub Project v2.
func DeleteProjectV2StatusUpdate(ctx context.Context, g *GitHubClient, statusUpdateID string) error {
	return g.DeleteProjectV2StatusUpdate(ctx, client.DeleteProjectV2StatusUpdateInput{
		StatusUpdateID: githubv4.ID(statusUpdateID),
	})
}
