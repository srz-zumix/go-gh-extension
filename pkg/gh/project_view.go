package gh

import (
	"context"
	"fmt"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// Type alias for the ProjectV2 view sort criterion from the client package.
type ProjectV2ViewSortBy = client.ProjectV2ViewSortBy

// ProjectV2ViewSortByInput is a sort criterion referencing a field by its REST field ID.
type ProjectV2ViewSortByInput struct {
	FieldID   int64
	Direction string // ASC or DESC
}

// ProjectV2ViewInput describes a view to create in a ProjectV2.
// Layout accepts GraphQL layout names (TABLE_LAYOUT, BOARD_LAYOUT, ROADMAP_LAYOUT) or short forms (TABLE, BOARD, ROADMAP).
// Field references are REST field IDs (ProjectV2Field.DatabaseID) of the target project.
type ProjectV2ViewInput struct {
	Name                    string
	Layout                  string
	Filter                  string
	VisibleFieldIDs         []int64
	SortBy                  []ProjectV2ViewSortByInput
	GroupByFieldIDs         []int64
	VerticalGroupByFieldIDs []int64
}

// restProjectV2ViewLayout converts a GraphQL ProjectV2ViewLayout value into the REST layout name.
func restProjectV2ViewLayout(layout string) (string, error) {
	switch strings.ToUpper(layout) {
	case "TABLE_LAYOUT", "TABLE":
		return "table", nil
	case "BOARD_LAYOUT", "BOARD":
		return "board", nil
	case "ROADMAP_LAYOUT", "ROADMAP":
		return "roadmap", nil
	default:
		return "", fmt.Errorf("unsupported project view layout '%s'", layout)
	}
}

// CreateProjectV2View creates a view in a ProjectV2.
// The GraphQL API has no view creation mutation, so the REST endpoints are used.
func CreateProjectV2View(ctx context.Context, g *GitHubClient, owner string, number int, input ProjectV2ViewInput) (*ProjectV2View, error) {
	layout, err := restProjectV2ViewLayout(input.Layout)
	if err != nil {
		return nil, err
	}
	body := client.CreateProjectV2ViewRequest{
		Name:            input.Name,
		Layout:          layout,
		GroupBy:         input.GroupByFieldIDs,
		VerticalGroupBy: input.VerticalGroupByFieldIDs,
	}
	if layout != "roadmap" {
		// visible_fields is rejected for roadmap layout views.
		body.VisibleFields = input.VisibleFieldIDs
	}
	if input.Filter != "" {
		f := input.Filter
		body.Filter = &f
	}
	for _, s := range input.SortBy {
		dir := strings.ToLower(strings.TrimSpace(s.Direction))
		if dir != "asc" && dir != "desc" {
			return nil, fmt.Errorf("unsupported sort direction '%s' (expected ASC or DESC)", s.Direction)
		}
		body.SortBy = append(body.SortBy, []any{s.FieldID, dir})
	}

	ownerType, err := DetectOwnerType(ctx, g, owner)
	if err != nil {
		return nil, err
	}
	var res *client.CreateProjectV2ViewResponse
	switch ownerType {
	case OwnerTypeOrg:
		res, err = g.CreateOrgProjectV2View(ctx, owner, number, body)
	case OwnerTypeUser:
		user, uerr := g.GetUser(ctx, owner)
		if uerr != nil {
			return nil, uerr
		}
		res, err = g.CreateUserProjectV2View(ctx, user.GetID(), number, body)
	default:
		return nil, fmt.Errorf("unknown owner type for '%s'", owner)
	}
	if err != nil {
		return nil, err
	}
	return &ProjectV2View{
		ID:     res.NodeID,
		Number: res.Number,
		Name:   res.Name,
		Layout: input.Layout,
		Filter: input.Filter,
	}, nil
}
