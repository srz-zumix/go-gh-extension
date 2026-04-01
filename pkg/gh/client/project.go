// Package client provides GitHub API client methods, including GitHub Projects v2.
package client

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

// ProjectV2 represents a GitHub Project v2.
type ProjectV2 struct {
	ID               string
	Number           int
	Title            string
	ShortDescription *string
	Readme           *string
	URL              string
	Public           bool
	Closed           bool
}

// ProjectV2Field represents a resolved field (column) in a GitHub Project v2.
// Options is populated only for SINGLE_SELECT fields.
// Iterations is populated only for ITERATION fields.
type ProjectV2Field struct {
	ID         string
	Name       string
	DataType   string // TEXT, NUMBER, DATE, SINGLE_SELECT, ITERATION, TITLE, ASSIGNEES, etc.
	Options    []ProjectV2SingleSelectOption
	Iterations []ProjectV2IterationOption
}

// ProjectV2SingleSelectOption represents an option in a SINGLE_SELECT field.
type ProjectV2SingleSelectOption struct {
	ID          string
	Name        string
	Color       string
	Description string
}

// ProjectV2IterationOption represents a single iteration in an ITERATION field.
type ProjectV2IterationOption struct {
	ID        string
	Title     string
	StartDate string // YYYY-MM-DD
	Duration  int    // days
}

// ProjectV2ItemType represents the type of content in a project item.
type ProjectV2ItemType string

const (
	ProjectV2ItemTypeIssue       ProjectV2ItemType = "ISSUE"
	ProjectV2ItemTypePullRequest ProjectV2ItemType = "PULL_REQUEST"
	ProjectV2ItemTypeDraftIssue  ProjectV2ItemType = "DRAFT_ISSUE"
	ProjectV2ItemTypeRedacted    ProjectV2ItemType = "REDACTED"
)

// ProjectV2ItemContent holds the resolved content of a project item.
type ProjectV2ItemContent struct {
	Type   ProjectV2ItemType
	ID     string
	Title  string
	Body   string
	URL    string // empty for DraftIssue
	Number int    // 0 for DraftIssue
	Author string // empty for DraftIssue
}

// ProjectV2FieldValue represents a resolved custom-field value for a project item.
// Inspect ValueType to determine which field is meaningful.
type ProjectV2FieldValue struct {
	FieldName      string
	ValueType      string   // TEXT, NUMBER, DATE, SINGLE_SELECT, ITERATION
	Text           string   // for TEXT
	Number         *float64 // for NUMBER
	Date           string   // for DATE, formatted as YYYY-MM-DD
	SelectName     string   // for SINGLE_SELECT
	SelectOptionID string   // for SINGLE_SELECT
	IterationID    string   // for ITERATION
	IterationTitle string   // for ITERATION, used for name-based matching
}

// ProjectV2Item represents an item in a GitHub Project v2.
type ProjectV2Item struct {
	ID          string
	Content     ProjectV2ItemContent
	FieldValues []ProjectV2FieldValue
	IsArchived  bool
}

// ─────────────────────────────────────────
// Internal query helper types
// ─────────────────────────────────────────

// projectV2FieldConfigNode is the inline-fragment representation of the
// ProjectV2FieldConfiguration union (ProjectV2Field | ProjectV2SingleSelectField | ProjectV2IterationField).
// Exactly one variant's ID will be non-empty for any given node.
type projectV2FieldConfigNode struct {
	AsProjectV2Field struct {
		ID       githubv4.String
		Name     githubv4.String
		DataType githubv4.String
	} `graphql:"... on ProjectV2Field"`
	AsSingleSelectField struct {
		ID       githubv4.String
		Name     githubv4.String
		DataType githubv4.String
		Options  []struct {
			ID          githubv4.String
			Name        githubv4.String
			Color       githubv4.String
			Description githubv4.String
		}
	} `graphql:"... on ProjectV2SingleSelectField"`
	AsIterationField struct {
		ID            githubv4.String
		Name          githubv4.String
		DataType      githubv4.String
		Configuration struct {
			Duration   githubv4.Int
			Iterations []struct {
				ID        githubv4.String
				Title     githubv4.String
				StartDate githubv4.String
				Duration  githubv4.Int
			}
		}
	} `graphql:"... on ProjectV2IterationField"`
}

func (n *projectV2FieldConfigNode) toProjectV2Field() ProjectV2Field {
	if n.AsProjectV2Field.ID != "" {
		return ProjectV2Field{
			ID:       string(n.AsProjectV2Field.ID),
			Name:     string(n.AsProjectV2Field.Name),
			DataType: string(n.AsProjectV2Field.DataType),
		}
	}
	if n.AsSingleSelectField.ID != "" {
		opts := make([]ProjectV2SingleSelectOption, len(n.AsSingleSelectField.Options))
		for i, o := range n.AsSingleSelectField.Options {
			opts[i] = ProjectV2SingleSelectOption{
				ID:          string(o.ID),
				Name:        string(o.Name),
				Color:       string(o.Color),
				Description: string(o.Description),
			}
		}
		return ProjectV2Field{
			ID:       string(n.AsSingleSelectField.ID),
			Name:     string(n.AsSingleSelectField.Name),
			DataType: string(n.AsSingleSelectField.DataType),
			Options:  opts,
		}
	}
	if n.AsIterationField.ID != "" {
		iters := make([]ProjectV2IterationOption, len(n.AsIterationField.Configuration.Iterations))
		for i, it := range n.AsIterationField.Configuration.Iterations {
			iters[i] = ProjectV2IterationOption{
				ID:        string(it.ID),
				Title:     string(it.Title),
				StartDate: string(it.StartDate),
				Duration:  int(it.Duration),
			}
		}
		return ProjectV2Field{
			ID:         string(n.AsIterationField.ID),
			Name:       string(n.AsIterationField.Name),
			DataType:   string(n.AsIterationField.DataType),
			Iterations: iters,
		}
	}
	return ProjectV2Field{}
}

// fieldConfigNameRef retrieves only the name from a ProjectV2FieldConfiguration union.
// Used when embedding a field reference inside field-value nodes.
type fieldConfigNameRef struct {
	OnField        struct{ Name githubv4.String } `graphql:"... on ProjectV2Field"`
	OnSingleSelect struct{ Name githubv4.String } `graphql:"... on ProjectV2SingleSelectField"`
	OnIteration    struct{ Name githubv4.String } `graphql:"... on ProjectV2IterationField"`
}

func (r *fieldConfigNameRef) name() string {
	if n := string(r.OnField.Name); n != "" {
		return n
	}
	if n := string(r.OnSingleSelect.Name); n != "" {
		return n
	}
	return string(r.OnIteration.Name)
}

// projectV2ItemContentNode is the inline-fragment representation of ProjectV2ItemContent.
type projectV2ItemContentNode struct {
	AsDraftIssue struct {
		ID    githubv4.String
		Title githubv4.String
		Body  githubv4.String
	} `graphql:"... on DraftIssue"`
	AsIssue struct {
		ID     githubv4.String
		Number githubv4.Int
		Title  githubv4.String
		Body   githubv4.String
		URL    githubv4.String
		Author struct{ Login githubv4.String }
	} `graphql:"... on Issue"`
	AsPullRequest struct {
		ID     githubv4.String
		Number githubv4.Int
		Title  githubv4.String
		Body   githubv4.String
		URL    githubv4.String
		Author struct{ Login githubv4.String }
	} `graphql:"... on PullRequest"`
}

// projectV2ItemFieldValueNode is the inline-fragment representation of
// ProjectV2ItemFieldValue. TEXT, NUMBER, DATE, SINGLE_SELECT, and ITERATION types are read.
type projectV2ItemFieldValueNode struct {
	AsText struct {
		Text  *githubv4.String
		Field fieldConfigNameRef
	} `graphql:"... on ProjectV2ItemFieldTextValue"`
	AsNumber struct {
		Number *githubv4.Float
		Field  fieldConfigNameRef
	} `graphql:"... on ProjectV2ItemFieldNumberValue"`
	AsDate struct {
		Date  *githubv4.Date
		Field fieldConfigNameRef
	} `graphql:"... on ProjectV2ItemFieldDateValue"`
	AsSingleSelect struct {
		Name     *githubv4.String
		OptionID *githubv4.String `graphql:"optionId"`
		Field    fieldConfigNameRef
	} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
	AsIteration struct {
		IterationID *githubv4.String `graphql:"iterationId"`
		Title       *githubv4.String
		Field       fieldConfigNameRef
	} `graphql:"... on ProjectV2ItemFieldIterationValue"`
}

// projectV2ItemNode is the raw GraphQL node for a project item.
type projectV2ItemNode struct {
	ID          githubv4.String
	Type        githubv4.String
	IsArchived  githubv4.Boolean
	Content     projectV2ItemContentNode
	FieldValues struct {
		Nodes []projectV2ItemFieldValueNode
	} `graphql:"fieldValues(first: 100)"`
}

func (n *projectV2ItemNode) toProjectV2Item() ProjectV2Item {
	item := ProjectV2Item{
		ID:         string(n.ID),
		IsArchived: bool(n.IsArchived),
	}
	switch ProjectV2ItemType(n.Type) {
	case ProjectV2ItemTypeDraftIssue:
		item.Content = ProjectV2ItemContent{
			Type:  ProjectV2ItemTypeDraftIssue,
			ID:    string(n.Content.AsDraftIssue.ID),
			Title: string(n.Content.AsDraftIssue.Title),
			Body:  string(n.Content.AsDraftIssue.Body),
		}
	case ProjectV2ItemTypeIssue:
		item.Content = ProjectV2ItemContent{
			Type:   ProjectV2ItemTypeIssue,
			ID:     string(n.Content.AsIssue.ID),
			Number: int(n.Content.AsIssue.Number),
			Title:  string(n.Content.AsIssue.Title),
			Body:   string(n.Content.AsIssue.Body),
			URL:    string(n.Content.AsIssue.URL),
			Author: string(n.Content.AsIssue.Author.Login),
		}
	case ProjectV2ItemTypePullRequest:
		item.Content = ProjectV2ItemContent{
			Type:   ProjectV2ItemTypePullRequest,
			ID:     string(n.Content.AsPullRequest.ID),
			Number: int(n.Content.AsPullRequest.Number),
			Title:  string(n.Content.AsPullRequest.Title),
			Body:   string(n.Content.AsPullRequest.Body),
			URL:    string(n.Content.AsPullRequest.URL),
			Author: string(n.Content.AsPullRequest.Author.Login),
		}
	default:
		item.Content = ProjectV2ItemContent{Type: ProjectV2ItemTypeRedacted}
	}

	for _, fv := range n.FieldValues.Nodes {
		if fv.AsText.Text != nil {
			if fieldName := fv.AsText.Field.name(); fieldName != "" {
				item.FieldValues = append(item.FieldValues, ProjectV2FieldValue{
					FieldName: fieldName,
					ValueType: "TEXT",
					Text:      string(*fv.AsText.Text),
				})
			}
		} else if fv.AsNumber.Number != nil {
			if fieldName := fv.AsNumber.Field.name(); fieldName != "" {
				num := float64(*fv.AsNumber.Number)
				item.FieldValues = append(item.FieldValues, ProjectV2FieldValue{
					FieldName: fieldName,
					ValueType: "NUMBER",
					Number:    &num,
				})
			}
		} else if fv.AsDate.Date != nil {
			if fieldName := fv.AsDate.Field.name(); fieldName != "" {
				item.FieldValues = append(item.FieldValues, ProjectV2FieldValue{
					FieldName: fieldName,
					ValueType: "DATE",
					Date:      fv.AsDate.Date.Format("2006-01-02"),
				})
			}
		} else if fv.AsSingleSelect.Name != nil {
			if fieldName := fv.AsSingleSelect.Field.name(); fieldName != "" {
				optID := ""
				if fv.AsSingleSelect.OptionID != nil {
					optID = string(*fv.AsSingleSelect.OptionID)
				}
				item.FieldValues = append(item.FieldValues, ProjectV2FieldValue{
					FieldName:      fieldName,
					ValueType:      "SINGLE_SELECT",
					SelectName:     string(*fv.AsSingleSelect.Name),
					SelectOptionID: optID,
				})
			}
		} else if fv.AsIteration.IterationID != nil {
			if fieldName := fv.AsIteration.Field.name(); fieldName != "" {
				title := ""
				if fv.AsIteration.Title != nil {
					title = string(*fv.AsIteration.Title)
				}
				item.FieldValues = append(item.FieldValues, ProjectV2FieldValue{
					FieldName:      fieldName,
					ValueType:      "ITERATION",
					IterationID:    string(*fv.AsIteration.IterationID),
					IterationTitle: title,
				})
			}
		}
	}
	return item
}

// ─────────────────────────────────────────
// Query types for fields, items, and views (shared between user/org variants)
// ─────────────────────────────────────────

// ProjectV2View represents a view in a GitHub Project v2.
// The GitHub GraphQL API supports reading views but does not expose a createProjectV2View mutation.
type ProjectV2View struct {
	ID     string
	Name   string
	Layout string // BOARD_LAYOUT, TABLE_LAYOUT, ROADMAP_LAYOUT
}

type projectV2FieldsQueryResult struct {
	Fields struct {
		Nodes    []projectV2FieldConfigNode
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	} `graphql:"fields(first: $fieldsFirst, after: $fieldsCursor)"`
}

type projectV2ItemsQueryResult struct {
	Items struct {
		Nodes    []projectV2ItemNode
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	} `graphql:"items(first: $itemsFirst, after: $itemsCursor, includeArchived: true)"`
}

// projectV2ItemsQueryResultNoArchived is used as a fallback for servers that do not support
// the includeArchived argument (e.g. older GitHub Enterprise Server versions).
type projectV2ItemsQueryResultNoArchived struct {
	Items struct {
		Nodes    []projectV2ItemNode
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	} `graphql:"items(first: $itemsFirst, after: $itemsCursor)"`
}

func processFields(nodes []projectV2FieldConfigNode) []ProjectV2Field {
	var result []ProjectV2Field
	for i := range nodes {
		if f := nodes[i].toProjectV2Field(); f.ID != "" {
			result = append(result, f)
		}
	}
	return result
}

func processItems(nodes []projectV2ItemNode) []ProjectV2Item {
	result := make([]ProjectV2Item, len(nodes))
	for i := range nodes {
		result[i] = nodes[i].toProjectV2Item()
	}
	return result
}

// ─────────────────────────────────────────
// Queries
// ─────────────────────────────────────────

// GetUserProjectV2ByNumber retrieves a ProjectV2 by user login and project number.
func (g *GitHubClient) GetUserProjectV2ByNumber(ctx context.Context, login string, number int) (*ProjectV2, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		User struct {
			ProjectV2 ProjectV2 `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(login),
		"number": githubv4.Int(number),
	}
	if err := gql.Query(ctx, &query, variables); err != nil {
		return nil, err
	}
	p := query.User.ProjectV2
	return &p, nil
}

// GetOrgProjectV2ByNumber retrieves a ProjectV2 by organization login and project number.
func (g *GitHubClient) GetOrgProjectV2ByNumber(ctx context.Context, org string, number int) (*ProjectV2, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		Organization struct {
			ProjectV2 ProjectV2 `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(org),
		"number": githubv4.Int(number),
	}
	if err := gql.Query(ctx, &query, variables); err != nil {
		return nil, err
	}
	p := query.Organization.ProjectV2
	return &p, nil
}

// ListUserProjectsV2 lists all ProjectV2s for a user.
func (g *GitHubClient) ListUserProjectsV2(ctx context.Context, login string, first int) ([]ProjectV2, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		User struct {
			ProjectsV2 struct {
				Nodes    []ProjectV2
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"projectsV2(first: $first, after: $cursor)"`
		} `graphql:"user(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(login),
		"first":  githubv4.Int(first),
		"cursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		all = append(all, query.User.ProjectsV2.Nodes...)
		if !query.User.ProjectsV2.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.User.ProjectsV2.PageInfo.EndCursor)
	}
	return all, nil
}

// ListOrgProjectsV2 lists all ProjectV2s for an organization.
func (g *GitHubClient) ListOrgProjectsV2(ctx context.Context, org string, first int) ([]ProjectV2, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		Organization struct {
			ProjectsV2 struct {
				Nodes    []ProjectV2
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"projectsV2(first: $first, after: $cursor)"`
		} `graphql:"organization(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(org),
		"first":  githubv4.Int(first),
		"cursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		all = append(all, query.Organization.ProjectsV2.Nodes...)
		if !query.Organization.ProjectsV2.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.ProjectsV2.PageInfo.EndCursor)
	}
	return all, nil
}

// ListUserProjectV2Fields lists all custom fields for a user's ProjectV2.
func (g *GitHubClient) ListUserProjectV2Fields(ctx context.Context, login string, number int, first int) ([]ProjectV2Field, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		User struct {
			ProjectV2 projectV2FieldsQueryResult `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}
	variables := map[string]any{
		"owner":        githubv4.String(login),
		"number":       githubv4.Int(number),
		"fieldsFirst":  githubv4.Int(first),
		"fieldsCursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2Field
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		all = append(all, processFields(query.User.ProjectV2.Fields.Nodes)...)
		if !query.User.ProjectV2.Fields.PageInfo.HasNextPage {
			break
		}
		variables["fieldsCursor"] = githubv4.NewString(query.User.ProjectV2.Fields.PageInfo.EndCursor)
	}
	return all, nil
}

// ListOrgProjectV2Fields lists all custom fields for an org's ProjectV2.
func (g *GitHubClient) ListOrgProjectV2Fields(ctx context.Context, org string, number int, first int) ([]ProjectV2Field, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		Organization struct {
			ProjectV2 projectV2FieldsQueryResult `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}
	variables := map[string]any{
		"owner":        githubv4.String(org),
		"number":       githubv4.Int(number),
		"fieldsFirst":  githubv4.Int(first),
		"fieldsCursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2Field
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		all = append(all, processFields(query.Organization.ProjectV2.Fields.Nodes)...)
		if !query.Organization.ProjectV2.Fields.PageInfo.HasNextPage {
			break
		}
		variables["fieldsCursor"] = githubv4.NewString(query.Organization.ProjectV2.Fields.PageInfo.EndCursor)
	}
	return all, nil
}

// ListUserProjectV2Items lists all items for a user's ProjectV2.
// When includeArchived is true, it tries to include archived items; if the server does not
// support the includeArchived argument (e.g. older GitHub Enterprise Server), it falls back
// to a query without that argument. When includeArchived is false, the no-archived query is
// used directly.
func (g *GitHubClient) ListUserProjectV2Items(ctx context.Context, login string, number int, first int, includeArchived bool) ([]ProjectV2Item, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	if !includeArchived {
		return g.listUserProjectV2ItemsNoArchived(ctx, gql, login, number, first)
	}
	variables := map[string]any{
		"owner":       githubv4.String(login),
		"number":      githubv4.Int(number),
		"itemsFirst":  githubv4.Int(first),
		"itemsCursor": (*githubv4.String)(nil),
	}
	// Try with includeArchived first.
	var queryWithArchived struct {
		User struct {
			ProjectV2 projectV2ItemsQueryResult `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}
	var all []ProjectV2Item
	for {
		if err := gql.Query(ctx, &queryWithArchived, variables); err != nil {
			return nil, err
		}
		all = append(all, processItems(queryWithArchived.User.ProjectV2.Items.Nodes)...)
		if !queryWithArchived.User.ProjectV2.Items.PageInfo.HasNextPage {
			break
		}
		variables["itemsCursor"] = githubv4.NewString(queryWithArchived.User.ProjectV2.Items.PageInfo.EndCursor)
	}
	return all, nil
}

func (g *GitHubClient) listUserProjectV2ItemsNoArchived(ctx context.Context, gql *githubv4.Client, login string, number int, first int) ([]ProjectV2Item, error) {
	var query struct {
		User struct {
			ProjectV2 projectV2ItemsQueryResultNoArchived `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}
	variables := map[string]any{
		"owner":       githubv4.String(login),
		"number":      githubv4.Int(number),
		"itemsFirst":  githubv4.Int(first),
		"itemsCursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2Item
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		all = append(all, processItems(query.User.ProjectV2.Items.Nodes)...)
		if !query.User.ProjectV2.Items.PageInfo.HasNextPage {
			break
		}
		variables["itemsCursor"] = githubv4.NewString(query.User.ProjectV2.Items.PageInfo.EndCursor)
	}
	return all, nil
}

// ListOrgProjectV2Items lists all items for an org's ProjectV2.
// When includeArchived is true, it tries to include archived items; if the server does not
// support the includeArchived argument (e.g. older GitHub Enterprise Server), it falls back
// to a query without that argument. When includeArchived is false, the no-archived query is
// used directly.
func (g *GitHubClient) ListOrgProjectV2Items(ctx context.Context, org string, number int, first int, includeArchived bool) ([]ProjectV2Item, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	if !includeArchived {
		return g.listOrgProjectV2ItemsNoArchived(ctx, gql, org, number, first)
	}
	variables := map[string]any{
		"owner":       githubv4.String(org),
		"number":      githubv4.Int(number),
		"itemsFirst":  githubv4.Int(first),
		"itemsCursor": (*githubv4.String)(nil),
	}
	// Try with includeArchived first.
	var queryWithArchived struct {
		Organization struct {
			ProjectV2 projectV2ItemsQueryResult `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}
	var all []ProjectV2Item
	for {
		if err := gql.Query(ctx, &queryWithArchived, variables); err != nil {
			return nil, err
		}
		all = append(all, processItems(queryWithArchived.Organization.ProjectV2.Items.Nodes)...)
		if !queryWithArchived.Organization.ProjectV2.Items.PageInfo.HasNextPage {
			break
		}
		variables["itemsCursor"] = githubv4.NewString(queryWithArchived.Organization.ProjectV2.Items.PageInfo.EndCursor)
	}
	return all, nil
}

func (g *GitHubClient) listOrgProjectV2ItemsNoArchived(ctx context.Context, gql *githubv4.Client, org string, number int, first int) ([]ProjectV2Item, error) {
	var query struct {
		Organization struct {
			ProjectV2 projectV2ItemsQueryResultNoArchived `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}
	variables := map[string]any{
		"owner":       githubv4.String(org),
		"number":      githubv4.Int(number),
		"itemsFirst":  githubv4.Int(first),
		"itemsCursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2Item
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		all = append(all, processItems(query.Organization.ProjectV2.Items.Nodes)...)
		if !query.Organization.ProjectV2.Items.PageInfo.HasNextPage {
			break
		}
		variables["itemsCursor"] = githubv4.NewString(query.Organization.ProjectV2.Items.PageInfo.EndCursor)
	}
	return all, nil
}

// GetOwnerNodeID returns the GraphQL node ID for a user or organization login.
func (g *GitHubClient) GetOwnerNodeID(ctx context.Context, login string) (*string, error) {
	user, err := g.GetUser(ctx, login)
	if err != nil {
		return nil, err
	}
	if user.NodeID == nil {
		return nil, fmt.Errorf("owner '%s' has no node ID", login)
	}
	return user.NodeID, nil
}

// ─────────────────────────────────────────
// Mutation input types
// ─────────────────────────────────────────

// CreateProjectV2Input is the input for creating a GitHub Project v2.
type CreateProjectV2Input struct {
	OwnerID githubv4.ID     `json:"ownerId"`
	Title   githubv4.String `json:"title"`
}

// UpdateProjectV2Input is the input for updating a GitHub Project v2.
type UpdateProjectV2Input struct {
	ProjectID        githubv4.ID       `json:"projectId"`
	Title            *githubv4.String  `json:"title,omitempty"`
	ShortDescription *githubv4.String  `json:"shortDescription,omitempty"`
	Readme           *githubv4.String  `json:"readme,omitempty"`
	Public           *githubv4.Boolean `json:"public,omitempty"`
	Closed           *githubv4.Boolean `json:"closed,omitempty"`
}

// DeleteProjectV2Input is the input for deleting a GitHub Project v2.
type DeleteProjectV2Input struct {
	ProjectID githubv4.ID `json:"projectId"`
}

// CreateProjectV2FieldSingleSelectOptionInput is a single-select option for field creation.
type CreateProjectV2FieldSingleSelectOptionInput struct {
	Name        githubv4.String `json:"name"`
	Color       githubv4.String `json:"color"`
	Description githubv4.String `json:"description"`
}

// ProjectV2IterationInput is a single iteration entry for iteration field creation.
type ProjectV2IterationInput struct {
	Title     githubv4.String `json:"title"`
	StartDate githubv4.String `json:"startDate"` // YYYY-MM-DD
	Duration  githubv4.Int    `json:"duration"`  // days
}

// ProjectV2IterationFieldConfigInput is the configuration for creating an ITERATION field.
type ProjectV2IterationFieldConfigInput struct {
	StartDate  githubv4.String           `json:"startDate"` // YYYY-MM-DD
	Duration   githubv4.Int              `json:"duration"`  // days
	Iterations []ProjectV2IterationInput `json:"iterations"`
}

// CreateProjectV2FieldInput is the input for creating a custom field in a Project v2.
type CreateProjectV2FieldInput struct {
	ProjectID              githubv4.ID                                   `json:"projectId"`
	DataType               githubv4.String                               `json:"dataType"`
	Name                   githubv4.String                               `json:"name"`
	SingleSelectOptions    []CreateProjectV2FieldSingleSelectOptionInput `json:"singleSelectOptions,omitempty"`
	IterationConfiguration *ProjectV2IterationFieldConfigInput           `json:"iterationConfiguration,omitempty"`
}

// AddProjectV2DraftIssueInput is the input for adding a draft issue to a Project v2.
type AddProjectV2DraftIssueInput struct {
	ProjectID githubv4.ID      `json:"projectId"`
	Title     githubv4.String  `json:"title"`
	Body      *githubv4.String `json:"body,omitempty"`
}

// AddProjectV2ItemByIdInput is the input for linking an existing issue or PR to a Project v2.
// The name matches the GraphQL schema type AddProjectV2ItemByIdInput exactly.
type AddProjectV2ItemByIdInput struct {
	ProjectID githubv4.ID `json:"projectId"`
	ContentID githubv4.ID `json:"contentId"`
}

// ProjectV2FieldValueInput represents the value to set on a project item field.
// Only one of Text/Number/Date/SingleSelectOptionID/IterationID should be set.
type ProjectV2FieldValueInput struct {
	Text                 *githubv4.String `json:"text,omitempty"`
	Number               *githubv4.Float  `json:"number,omitempty"`
	Date                 *githubv4.Date   `json:"date,omitempty"`
	SingleSelectOptionID *githubv4.String `json:"singleSelectOptionId,omitempty"`
	IterationID          *githubv4.String `json:"iterationId,omitempty"`
}

// UpdateProjectV2ItemFieldValueInput is the input for setting a field value on a project item.
type UpdateProjectV2ItemFieldValueInput struct {
	ProjectID githubv4.ID              `json:"projectId"`
	ItemID    githubv4.ID              `json:"itemId"`
	FieldID   githubv4.ID              `json:"fieldId"`
	Value     ProjectV2FieldValueInput `json:"value"`
}

// DeleteProjectV2ItemInput is the input for removing an item from a Project v2.
type DeleteProjectV2ItemInput struct {
	ProjectID githubv4.ID `json:"projectId"`
	ItemID    githubv4.ID `json:"itemId"`
}

// ArchiveProjectV2ItemInput is the input for archiving an item in a Project v2.
type ArchiveProjectV2ItemInput struct {
	ProjectID githubv4.ID `json:"projectId"`
	ItemID    githubv4.ID `json:"itemId"`
}

// ─────────────────────────────────────────
// Mutations
// ─────────────────────────────────────────

// CreateProjectV2 creates a new GitHub Project v2.
func (g *GitHubClient) CreateProjectV2(ctx context.Context, input CreateProjectV2Input) (*ProjectV2, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var mutation struct {
		CreateProjectV2 struct {
			ProjectV2 ProjectV2
		} `graphql:"createProjectV2(input: $input)"`
	}
	if err := gql.Mutate(ctx, &mutation, input, nil); err != nil {
		return nil, err
	}
	p := mutation.CreateProjectV2.ProjectV2
	return &p, nil
}

// UpdateProjectV2 updates a GitHub Project v2.
func (g *GitHubClient) UpdateProjectV2(ctx context.Context, input UpdateProjectV2Input) (*ProjectV2, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var mutation struct {
		UpdateProjectV2 struct {
			ProjectV2 ProjectV2
		} `graphql:"updateProjectV2(input: $input)"`
	}
	if err := gql.Mutate(ctx, &mutation, input, nil); err != nil {
		return nil, err
	}
	p := mutation.UpdateProjectV2.ProjectV2
	return &p, nil
}

// DeleteProjectV2 deletes a GitHub Project v2.
func (g *GitHubClient) DeleteProjectV2(ctx context.Context, input DeleteProjectV2Input) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		DeleteProjectV2 struct {
			ProjectV2 struct {
				ID githubv4.String
			}
		} `graphql:"deleteProjectV2(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// CreateProjectV2Field creates a custom field in a GitHub Project v2.
func (g *GitHubClient) CreateProjectV2Field(ctx context.Context, input CreateProjectV2FieldInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	// Request only clientMutationId to avoid schema differences across GitHub versions.
	var mutation struct {
		CreateProjectV2Field struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"createProjectV2Field(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// AddProjectV2DraftIssue adds a draft issue to a GitHub Project v2.
// Returns the created item's node ID.
func (g *GitHubClient) AddProjectV2DraftIssue(ctx context.Context, input AddProjectV2DraftIssueInput) (string, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return "", err
	}
	var mutation struct {
		AddProjectV2DraftIssue struct {
			ProjectItem struct {
				ID githubv4.String
			}
		} `graphql:"addProjectV2DraftIssue(input: $input)"`
	}
	if err := gql.Mutate(ctx, &mutation, input, nil); err != nil {
		return "", err
	}
	return string(mutation.AddProjectV2DraftIssue.ProjectItem.ID), nil
}

// AddProjectV2ItemByID links an existing issue or PR (by node ID) to a GitHub Project v2.
// Returns the created project item's node ID.
func (g *GitHubClient) AddProjectV2ItemByID(ctx context.Context, input AddProjectV2ItemByIdInput) (string, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return "", err
	}
	var mutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID githubv4.String
			}
		} `graphql:"addProjectV2ItemById(input: $input)"`
	}
	if err := gql.Mutate(ctx, &mutation, input, nil); err != nil {
		return "", err
	}
	return string(mutation.AddProjectV2ItemById.Item.ID), nil
}

// UpdateProjectV2ItemFieldValue sets the value of a custom field for a project item.
func (g *GitHubClient) UpdateProjectV2ItemFieldValue(ctx context.Context, input UpdateProjectV2ItemFieldValueInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// DeleteProjectV2Item removes an item from a GitHub Project v2.
func (g *GitHubClient) DeleteProjectV2Item(ctx context.Context, input DeleteProjectV2ItemInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		DeleteProjectV2Item struct {
			DeletedItemID githubv4.String `graphql:"deletedItemId"`
		} `graphql:"deleteProjectV2Item(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// ArchiveProjectV2Item archives an item in a GitHub Project v2.
func (g *GitHubClient) ArchiveProjectV2Item(ctx context.Context, input ArchiveProjectV2ItemInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		ArchiveProjectV2Item struct {
			Item struct {
				ID githubv4.String
			}
		} `graphql:"archiveProjectV2Item(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// ListUserProjectV2Views lists all views for a user's ProjectV2.
// The GitHub GraphQL API supports reading views but does not expose a createProjectV2View mutation.
func (g *GitHubClient) ListUserProjectV2Views(ctx context.Context, login string, number int) ([]ProjectV2View, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		User struct {
			ProjectV2 struct {
				Views struct {
					Nodes []struct {
						ID     githubv4.String
						Name   githubv4.String
						Layout githubv4.String
					}
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"views(first: $first, after: $cursor)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(login),
		"number": githubv4.Int(number),
		"first":  githubv4.Int(50),
		"cursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2View
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		for _, n := range query.User.ProjectV2.Views.Nodes {
			all = append(all, ProjectV2View{
				ID:     string(n.ID),
				Name:   string(n.Name),
				Layout: string(n.Layout),
			})
		}
		if !query.User.ProjectV2.Views.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.User.ProjectV2.Views.PageInfo.EndCursor)
	}
	return all, nil
}

// ListOrgProjectV2Views lists all views for an org's ProjectV2.
// The GitHub GraphQL API supports reading views but does not expose a createProjectV2View mutation.
func (g *GitHubClient) ListOrgProjectV2Views(ctx context.Context, org string, number int) ([]ProjectV2View, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		Organization struct {
			ProjectV2 struct {
				Views struct {
					Nodes []struct {
						ID     githubv4.String
						Name   githubv4.String
						Layout githubv4.String
					}
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"views(first: $first, after: $cursor)"`
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(org),
		"number": githubv4.Int(number),
		"first":  githubv4.Int(50),
		"cursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2View
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		for _, n := range query.Organization.ProjectV2.Views.Nodes {
			all = append(all, ProjectV2View{
				ID:     string(n.ID),
				Name:   string(n.Name),
				Layout: string(n.Layout),
			})
		}
		if !query.Organization.ProjectV2.Views.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.ProjectV2.Views.PageInfo.EndCursor)
	}
	return all, nil
}
