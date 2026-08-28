package gh

import (
	"context"
	"fmt"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// Type alias for the ProjectV2 workflow type from the client package.
type ProjectV2Workflow = client.ProjectV2Workflow

// ListProjectV2Workflows lists all built-in automations of a ProjectV2.
func ListProjectV2Workflows(ctx context.Context, g *GitHubClient, owner string, number int) ([]ProjectV2Workflow, error) {
	ownerType, err := DetectOwnerType(ctx, g, owner)
	if err != nil {
		return nil, err
	}
	switch ownerType {
	case OwnerTypeOrg:
		return g.ListOrgProjectV2Workflows(ctx, owner, number, 50)
	case OwnerTypeUser:
		return g.ListUserProjectV2Workflows(ctx, owner, number, 50)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", owner)
	}
}
