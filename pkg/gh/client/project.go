// Package client provides GitHub API client methods, including GitHub Projects v2.
package client

import (
	"context"
	"fmt"
	"sort"
	"strings"

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
// Options is populated only for SINGLE_SELECT and MULTI_SELECT fields.
// Iterations and CompletedIterations are populated only for ITERATION fields.
type ProjectV2Field struct {
	ID         string
	DatabaseID int64 // REST field ID, used by the Project views REST endpoints
	Name       string
	DataType   string // TEXT, NUMBER, DATE, SINGLE_SELECT, MULTI_SELECT, ITERATION, TITLE, ASSIGNEES, etc.
	Options    []ProjectV2SingleSelectOption
	// Iterations holds the current and upcoming iterations of an ITERATION field.
	Iterations []ProjectV2IterationOption
	// CompletedIterations holds the past iterations of an ITERATION field.
	CompletedIterations []ProjectV2IterationOption
	// IterationDuration is the default iteration length in days of an ITERATION field.
	IterationDuration int
}

// AllIterations returns the completed and current iterations of an ITERATION field
// ordered by start date, oldest first.
func (f *ProjectV2Field) AllIterations() []ProjectV2IterationOption {
	all := make([]ProjectV2IterationOption, 0, len(f.CompletedIterations)+len(f.Iterations))
	all = append(all, f.CompletedIterations...)
	all = append(all, f.Iterations...)
	sort.SliceStable(all, func(i, j int) bool { return all[i].StartDate < all[j].StartDate })
	return all
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
	Type      ProjectV2ItemType
	ID        string
	Title     string
	Body      string
	URL       string // empty for DraftIssue
	Number    int    // 0 for DraftIssue
	Author    string // empty for DraftIssue
	RepoOwner string // owner login of the issue/PR repository, empty for DraftIssue
	RepoName  string // name of the issue/PR repository, empty for DraftIssue
}

// ProjectV2FieldValue represents a resolved custom-field value for a project item.
// Inspect ValueType to determine which field is meaningful.
type ProjectV2FieldValue struct {
	FieldName       string
	ValueType       string   // TEXT, NUMBER, DATE, SINGLE_SELECT, MULTI_SELECT, ITERATION
	Text            string   // for TEXT
	Number          *float64 // for NUMBER
	Date            string   // for DATE, formatted as YYYY-MM-DD
	SelectName      string   // for SINGLE_SELECT
	SelectOptionID  string   // for SINGLE_SELECT
	SelectNames     []string // for MULTI_SELECT
	SelectOptionIDs []string // for MULTI_SELECT
	IterationID     string   // for ITERATION
	IterationTitle  string   // for ITERATION, used for name-based matching
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
// The GraphQL response merges every matching inline fragment into a single flat object, so the
// variant must be selected by __typename instead of by which fragment has a non-empty ID.
type projectV2FieldConfigNode struct {
	Typename         githubv4.String `graphql:"__typename"`
	AsProjectV2Field struct {
		ID         githubv4.String
		DatabaseID githubv4.Int
		Name       githubv4.String
		DataType   githubv4.String
	} `graphql:"... on ProjectV2Field"`
	AsSingleSelectField struct {
		ID         githubv4.String
		DatabaseID githubv4.Int
		Name       githubv4.String
		DataType   githubv4.String
		Options    []struct {
			ID          githubv4.String
			Name        githubv4.String
			Color       githubv4.String
			Description githubv4.String
		}
	} `graphql:"... on ProjectV2SingleSelectField"`
	AsIterationField struct {
		ID            githubv4.String
		DatabaseID    githubv4.Int
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
			CompletedIterations []struct {
				ID        githubv4.String
				Title     githubv4.String
				StartDate githubv4.String
				Duration  githubv4.Int
			}
		}
	} `graphql:"... on ProjectV2IterationField"`
}

func (n projectV2FieldConfigNode) toProjectV2Field() ProjectV2Field {
	switch string(n.Typename) {
	case "ProjectV2SingleSelectField":
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
			ID:         string(n.AsSingleSelectField.ID),
			DatabaseID: int64(n.AsSingleSelectField.DatabaseID),
			Name:       string(n.AsSingleSelectField.Name),
			DataType:   string(n.AsSingleSelectField.DataType),
			Options:    opts,
		}
	case "ProjectV2IterationField":
		iters := make([]ProjectV2IterationOption, len(n.AsIterationField.Configuration.Iterations))
		for i, it := range n.AsIterationField.Configuration.Iterations {
			iters[i] = ProjectV2IterationOption{
				ID:        string(it.ID),
				Title:     string(it.Title),
				StartDate: string(it.StartDate),
				Duration:  int(it.Duration),
			}
		}
		completed := make([]ProjectV2IterationOption, len(n.AsIterationField.Configuration.CompletedIterations))
		for i, it := range n.AsIterationField.Configuration.CompletedIterations {
			completed[i] = ProjectV2IterationOption{
				ID:        string(it.ID),
				Title:     string(it.Title),
				StartDate: string(it.StartDate),
				Duration:  int(it.Duration),
			}
		}
		return ProjectV2Field{
			ID:                  string(n.AsIterationField.ID),
			DatabaseID:          int64(n.AsIterationField.DatabaseID),
			Name:                string(n.AsIterationField.Name),
			DataType:            string(n.AsIterationField.DataType),
			Iterations:          iters,
			CompletedIterations: completed,
			IterationDuration:   int(n.AsIterationField.Configuration.Duration),
		}
	case "ProjectV2Field":
		return ProjectV2Field{
			ID:         string(n.AsProjectV2Field.ID),
			DatabaseID: int64(n.AsProjectV2Field.DatabaseID),
			Name:       string(n.AsProjectV2Field.Name),
			DataType:   string(n.AsProjectV2Field.DataType),
		}
	}
	return ProjectV2Field{}
}

// multiSelectProjectV2FieldConfigNode additionally reads ProjectV2MultiSelectField, which is
// missing from GraphQL schemas that predate multi-select support.
type multiSelectProjectV2FieldConfigNode struct {
	projectV2FieldConfigNode
	AsMultiSelectField struct {
		ID         githubv4.String
		DatabaseID githubv4.Int
		Name       githubv4.String
		DataType   githubv4.String
		Options    []struct {
			ID          githubv4.String
			Name        githubv4.String
			Color       githubv4.String
			Description githubv4.String
		}
	} `graphql:"... on ProjectV2MultiSelectField"`
}

func (n multiSelectProjectV2FieldConfigNode) toProjectV2Field() ProjectV2Field {
	if string(n.Typename) != "ProjectV2MultiSelectField" {
		return n.projectV2FieldConfigNode.toProjectV2Field()
	}
	opts := make([]ProjectV2SingleSelectOption, len(n.AsMultiSelectField.Options))
	for i, o := range n.AsMultiSelectField.Options {
		opts[i] = ProjectV2SingleSelectOption{
			ID:          string(o.ID),
			Name:        string(o.Name),
			Color:       string(o.Color),
			Description: string(o.Description),
		}
	}
	return ProjectV2Field{
		ID:         string(n.AsMultiSelectField.ID),
		DatabaseID: int64(n.AsMultiSelectField.DatabaseID),
		Name:       string(n.AsMultiSelectField.Name),
		DataType:   string(n.AsMultiSelectField.DataType),
		Options:    opts,
	}
}

// fieldConfigNode is implemented by the ProjectV2FieldConfiguration fragments so that queries can
// be instantiated with or without multi-select support.
type fieldConfigNode interface {
	toProjectV2Field() ProjectV2Field
}

// fieldConfigNameRef retrieves only the name from a ProjectV2FieldConfiguration union.
// Used when embedding a field reference inside field-value nodes.
type fieldConfigNameRef struct {
	OnField        struct{ Name githubv4.String } `graphql:"... on ProjectV2Field"`
	OnSingleSelect struct{ Name githubv4.String } `graphql:"... on ProjectV2SingleSelectField"`
	OnIteration    struct{ Name githubv4.String } `graphql:"... on ProjectV2IterationField"`
}

func (r fieldConfigNameRef) name() string {
	if n := string(r.OnField.Name); n != "" {
		return n
	}
	if n := string(r.OnSingleSelect.Name); n != "" {
		return n
	}
	return string(r.OnIteration.Name)
}

// multiSelectFieldConfigNameRef additionally resolves the name of MULTI_SELECT fields, which are
// missing from GraphQL schemas that predate multi-select support.
type multiSelectFieldConfigNameRef struct {
	fieldConfigNameRef
	OnMultiSelect struct{ Name githubv4.String } `graphql:"... on ProjectV2MultiSelectField"`
}

func (r multiSelectFieldConfigNameRef) name() string {
	if n := string(r.OnMultiSelect.Name); n != "" {
		return n
	}
	return r.fieldConfigNameRef.name()
}

// nameRef is implemented by the name-only ProjectV2FieldConfiguration fragments so that queries
// can be instantiated with or without multi-select support.
type nameRef interface {
	name() string
}

// projectV2ItemContentNode is the inline-fragment representation of ProjectV2ItemContent.
type projectV2ItemContentNode struct {
	AsDraftIssue struct {
		ID    githubv4.String
		Title githubv4.String
		Body  githubv4.String
	} `graphql:"... on DraftIssue"`
	AsIssue struct {
		ID         githubv4.String
		Number     githubv4.Int
		Title      githubv4.String
		Body       githubv4.String
		URL        githubv4.String
		Author     struct{ Login githubv4.String }
		Repository repositoryNameRef
	} `graphql:"... on Issue"`
	AsPullRequest struct {
		ID         githubv4.String
		Number     githubv4.Int
		Title      githubv4.String
		Body       githubv4.String
		URL        githubv4.String
		Author     struct{ Login githubv4.String }
		Repository repositoryNameRef
	} `graphql:"... on PullRequest"`
}

// repositoryNameRef identifies the repository that owns an issue or pull request.
type repositoryNameRef struct {
	Name  githubv4.String
	Owner struct{ Login githubv4.String }
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
		Date  *githubv4.String
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

// toFieldValue converts the populated fragment into a ProjectV2FieldValue.
// The boolean result reports whether a fragment carried a usable value.
func (n projectV2ItemFieldValueNode) toFieldValue() (ProjectV2FieldValue, bool) {
	switch {
	case n.AsText.Text != nil:
		if fieldName := n.AsText.Field.name(); fieldName != "" {
			return ProjectV2FieldValue{
				FieldName: fieldName,
				ValueType: "TEXT",
				Text:      string(*n.AsText.Text),
			}, true
		}
	case n.AsNumber.Number != nil:
		if fieldName := n.AsNumber.Field.name(); fieldName != "" {
			num := float64(*n.AsNumber.Number)
			return ProjectV2FieldValue{
				FieldName: fieldName,
				ValueType: "NUMBER",
				Number:    &num,
			}, true
		}
	case n.AsDate.Date != nil:
		if fieldName := n.AsDate.Field.name(); fieldName != "" {
			return ProjectV2FieldValue{
				FieldName: fieldName,
				ValueType: "DATE",
				Date:      normalizeDateScalar(string(*n.AsDate.Date)),
			}, true
		}
	case n.AsSingleSelect.Name != nil:
		if fieldName := n.AsSingleSelect.Field.name(); fieldName != "" {
			optID := ""
			if n.AsSingleSelect.OptionID != nil {
				optID = string(*n.AsSingleSelect.OptionID)
			}
			return ProjectV2FieldValue{
				FieldName:      fieldName,
				ValueType:      "SINGLE_SELECT",
				SelectName:     string(*n.AsSingleSelect.Name),
				SelectOptionID: optID,
			}, true
		}
	case n.AsIteration.IterationID != nil:
		if fieldName := n.AsIteration.Field.name(); fieldName != "" {
			title := ""
			if n.AsIteration.Title != nil {
				title = string(*n.AsIteration.Title)
			}
			return ProjectV2FieldValue{
				FieldName:      fieldName,
				ValueType:      "ITERATION",
				IterationID:    string(*n.AsIteration.IterationID),
				IterationTitle: title,
			}, true
		}
	}
	return ProjectV2FieldValue{}, false
}

// multiSelectItemFieldValueNode additionally reads ProjectV2ItemFieldMultiSelectValue, which is
// missing from GraphQL schemas that predate multi-select support.
type multiSelectItemFieldValueNode struct {
	projectV2ItemFieldValueNode
	AsMultiSelect struct {
		Options []struct {
			ID   githubv4.String
			Name githubv4.String
		}
		Field multiSelectFieldConfigNameRef
	} `graphql:"... on ProjectV2ItemFieldMultiSelectValue"`
}

func (n multiSelectItemFieldValueNode) toFieldValue() (ProjectV2FieldValue, bool) {
	if len(n.AsMultiSelect.Options) > 0 {
		if fieldName := n.AsMultiSelect.Field.name(); fieldName != "" {
			names := make([]string, len(n.AsMultiSelect.Options))
			ids := make([]string, len(n.AsMultiSelect.Options))
			for i, o := range n.AsMultiSelect.Options {
				names[i] = string(o.Name)
				ids[i] = string(o.ID)
			}
			return ProjectV2FieldValue{
				FieldName:       fieldName,
				ValueType:       "MULTI_SELECT",
				SelectNames:     names,
				SelectOptionIDs: ids,
			}, true
		}
	}
	return n.projectV2ItemFieldValueNode.toFieldValue()
}

// itemFieldValueNode is implemented by the ProjectV2ItemFieldValue fragments so that item queries
// can be instantiated with or without multi-select support.
type itemFieldValueNode interface {
	toFieldValue() (ProjectV2FieldValue, bool)
}

// projectV2ItemNode is the raw GraphQL node for a project item.
type projectV2ItemNode[FV itemFieldValueNode] struct {
	ID          githubv4.String
	Type        githubv4.String
	IsArchived  githubv4.Boolean
	Content     projectV2ItemContentNode
	FieldValues struct {
		Nodes []FV
	} `graphql:"fieldValues(first: 100)"`
}

func (n projectV2ItemNode[FV]) toProjectV2Item() ProjectV2Item {
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
			Type:      ProjectV2ItemTypeIssue,
			ID:        string(n.Content.AsIssue.ID),
			Number:    int(n.Content.AsIssue.Number),
			Title:     string(n.Content.AsIssue.Title),
			Body:      string(n.Content.AsIssue.Body),
			URL:       string(n.Content.AsIssue.URL),
			Author:    string(n.Content.AsIssue.Author.Login),
			RepoOwner: string(n.Content.AsIssue.Repository.Owner.Login),
			RepoName:  string(n.Content.AsIssue.Repository.Name),
		}
	case ProjectV2ItemTypePullRequest:
		item.Content = ProjectV2ItemContent{
			Type:      ProjectV2ItemTypePullRequest,
			ID:        string(n.Content.AsPullRequest.ID),
			Number:    int(n.Content.AsPullRequest.Number),
			Title:     string(n.Content.AsPullRequest.Title),
			Body:      string(n.Content.AsPullRequest.Body),
			URL:       string(n.Content.AsPullRequest.URL),
			Author:    string(n.Content.AsPullRequest.Author.Login),
			RepoOwner: string(n.Content.AsPullRequest.Repository.Owner.Login),
			RepoName:  string(n.Content.AsPullRequest.Repository.Name),
		}
	default:
		item.Content = ProjectV2ItemContent{Type: ProjectV2ItemTypeRedacted}
	}

	for _, fv := range n.FieldValues.Nodes {
		if v, ok := fv.toFieldValue(); ok {
			item.FieldValues = append(item.FieldValues, v)
		}
	}
	return item
}

// ─────────────────────────────────────────
// Query types for fields, items, and views (shared between user/org variants)
// ─────────────────────────────────────────

// ProjectV2View represents a view in a GitHub Project v2.
// The GraphQL API supports reading views; creation is only available through the REST API.
type ProjectV2View struct {
	ID     string
	Number int
	Name   string
	Layout string // BOARD_LAYOUT, TABLE_LAYOUT, ROADMAP_LAYOUT
	Filter string
	// VisibleFields holds the visible field names in display order.
	VisibleFields         []string
	GroupByFields         []string
	VerticalGroupByFields []string
	SortBy                []ProjectV2ViewSortBy
}

// ProjectV2ViewSortBy represents a single sort criterion of a project view.
type ProjectV2ViewSortBy struct {
	FieldName string
	Direction string // ASC or DESC
}

// projectV2ViewNode is the raw GraphQL node for a project view.
type projectV2ViewNode[R nameRef] struct {
	ID     githubv4.String
	Number githubv4.Int
	Name   githubv4.String
	Layout githubv4.String
	Filter githubv4.String
	Fields struct {
		Nodes []R
	} `graphql:"fields(first: 50)"`
	GroupByFields struct {
		Nodes []R
	} `graphql:"groupByFields(first: 10)"`
	VerticalGroupByFields struct {
		Nodes []R
	} `graphql:"verticalGroupByFields(first: 10)"`
	SortByFields struct {
		Nodes []struct {
			Direction githubv4.String
			Field     R
		}
	} `graphql:"sortByFields(first: 10)"`
}

func (n projectV2ViewNode[R]) toProjectV2View() ProjectV2View {
	v := ProjectV2View{
		ID:     string(n.ID),
		Number: int(n.Number),
		Name:   string(n.Name),
		Layout: string(n.Layout),
		Filter: string(n.Filter),
	}
	for i := range n.Fields.Nodes {
		v.VisibleFields = append(v.VisibleFields, n.Fields.Nodes[i].name())
	}
	for i := range n.GroupByFields.Nodes {
		v.GroupByFields = append(v.GroupByFields, n.GroupByFields.Nodes[i].name())
	}
	for i := range n.VerticalGroupByFields.Nodes {
		v.VerticalGroupByFields = append(v.VerticalGroupByFields, n.VerticalGroupByFields.Nodes[i].name())
	}
	for i := range n.SortByFields.Nodes {
		v.SortBy = append(v.SortBy, ProjectV2ViewSortBy{
			FieldName: n.SortByFields.Nodes[i].Field.name(),
			Direction: string(n.SortByFields.Nodes[i].Direction),
		})
	}
	return v
}

type projectV2FieldsQueryResult[N fieldConfigNode] struct {
	Fields struct {
		Nodes    []N
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
	} `graphql:"fields(first: $fieldsFirst, after: $fieldsCursor)"`
}

type projectV2ItemsConnection[FV itemFieldValueNode] struct {
	Nodes    []projectV2ItemNode[FV]
	PageInfo struct {
		EndCursor   githubv4.String
		HasNextPage bool
	}
}

type projectV2ItemsQueryResult[FV itemFieldValueNode] struct {
	Items projectV2ItemsConnection[FV] `graphql:"items(first: $itemsFirst, after: $itemsCursor)"`
}

// projectV2AllItemsQueryResult selects archived items as well. The items connection defaults to
// archivedStates: [NOT_ARCHIVED], and the argument is missing on older GitHub Enterprise Server.
type projectV2AllItemsQueryResult[FV itemFieldValueNode] struct {
	Items projectV2ItemsConnection[FV] `graphql:"items(first: $itemsFirst, after: $itemsCursor, archivedStates: [ARCHIVED, NOT_ARCHIVED])"`
}

// unsupportedArchivedStates reports whether the GraphQL API rejected the archivedStates argument.
func unsupportedArchivedStates(err error) bool {
	return err != nil && strings.Contains(err.Error(), "archivedStates")
}

// unsupportedMultiSelect reports whether the GraphQL API rejected the multi-select fragments,
// which are absent from schemas that predate multi-select fields.
func unsupportedMultiSelect(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "ProjectV2MultiSelectField") ||
		strings.Contains(msg, "ProjectV2ItemFieldMultiSelectValue")
}

// normalizeDateScalar returns the YYYY-MM-DD part of a GraphQL Date scalar. The scalar must be read
// as a string because githubv4.Date embeds time.Time, which only unmarshals RFC 3339 timestamps.
func normalizeDateScalar(s string) string {
	if i := strings.IndexByte(s, 'T'); i > 0 {
		return s[:i]
	}
	return s
}

// paginateProjectV2Items runs query until every item page is consumed. conn must point at the
// items connection inside query so that each page is read after the query is re-executed.
func paginateProjectV2Items[FV itemFieldValueNode](ctx context.Context, gql *githubv4.Client, query any, variables map[string]any, conn *projectV2ItemsConnection[FV]) ([]ProjectV2Item, error) {
	var all []ProjectV2Item
	for {
		if err := gql.Query(ctx, query, variables); err != nil {
			return nil, err
		}
		all = append(all, processItems(conn.Nodes)...)
		if !conn.PageInfo.HasNextPage {
			return all, nil
		}
		variables["itemsCursor"] = githubv4.NewString(conn.PageInfo.EndCursor)
	}
}

func processFields[N fieldConfigNode](nodes []N) []ProjectV2Field {
	var result []ProjectV2Field
	for i := range nodes {
		if f := nodes[i].toProjectV2Field(); f.ID != "" {
			result = append(result, f)
		}
	}
	return result
}

func processItems[FV itemFieldValueNode](nodes []projectV2ItemNode[FV]) []ProjectV2Item {
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
	fields, err := listUserProjectV2Fields[multiSelectProjectV2FieldConfigNode](ctx, gql, login, number, first)
	if !unsupportedMultiSelect(err) {
		return fields, err
	}
	return listUserProjectV2Fields[projectV2FieldConfigNode](ctx, gql, login, number, first)
}

func listUserProjectV2Fields[N fieldConfigNode](ctx context.Context, gql *githubv4.Client, login string, number int, first int) ([]ProjectV2Field, error) {
	var query struct {
		User struct {
			ProjectV2 projectV2FieldsQueryResult[N] `graphql:"projectV2(number: $number)"`
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
	fields, err := listOrgProjectV2Fields[multiSelectProjectV2FieldConfigNode](ctx, gql, org, number, first)
	if !unsupportedMultiSelect(err) {
		return fields, err
	}
	return listOrgProjectV2Fields[projectV2FieldConfigNode](ctx, gql, org, number, first)
}

func listOrgProjectV2Fields[N fieldConfigNode](ctx context.Context, gql *githubv4.Client, org string, number int, first int) ([]ProjectV2Field, error) {
	var query struct {
		Organization struct {
			ProjectV2 projectV2FieldsQueryResult[N] `graphql:"projectV2(number: $number)"`
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

// ListUserProjectV2Items lists all items, including archived ones, for a user's ProjectV2.
func (g *GitHubClient) ListUserProjectV2Items(ctx context.Context, login string, number int, first int) ([]ProjectV2Item, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	items, err := listUserProjectV2Items[multiSelectItemFieldValueNode](ctx, gql, login, number, first)
	if !unsupportedMultiSelect(err) {
		return items, err
	}
	return listUserProjectV2Items[projectV2ItemFieldValueNode](ctx, gql, login, number, first)
}

func listUserProjectV2Items[FV itemFieldValueNode](ctx context.Context, gql *githubv4.Client, login string, number int, first int) ([]ProjectV2Item, error) {
	newVariables := func() map[string]any {
		return map[string]any{
			"owner":       githubv4.String(login),
			"number":      githubv4.Int(number),
			"itemsFirst":  githubv4.Int(first),
			"itemsCursor": (*githubv4.String)(nil),
		}
	}
	var allItemsQuery struct {
		User struct {
			ProjectV2 projectV2AllItemsQueryResult[FV] `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}
	items, err := paginateProjectV2Items(ctx, gql, &allItemsQuery, newVariables(), &allItemsQuery.User.ProjectV2.Items)
	if !unsupportedArchivedStates(err) {
		return items, err
	}
	var query struct {
		User struct {
			ProjectV2 projectV2ItemsQueryResult[FV] `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $owner)"`
	}
	return paginateProjectV2Items(ctx, gql, &query, newVariables(), &query.User.ProjectV2.Items)
}

// ListOrgProjectV2Items lists all items, including archived ones, for an org's ProjectV2.
func (g *GitHubClient) ListOrgProjectV2Items(ctx context.Context, org string, number int, first int) ([]ProjectV2Item, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	items, err := listOrgProjectV2Items[multiSelectItemFieldValueNode](ctx, gql, org, number, first)
	if !unsupportedMultiSelect(err) {
		return items, err
	}
	return listOrgProjectV2Items[projectV2ItemFieldValueNode](ctx, gql, org, number, first)
}

func listOrgProjectV2Items[FV itemFieldValueNode](ctx context.Context, gql *githubv4.Client, org string, number int, first int) ([]ProjectV2Item, error) {
	newVariables := func() map[string]any {
		return map[string]any{
			"owner":       githubv4.String(org),
			"number":      githubv4.Int(number),
			"itemsFirst":  githubv4.Int(first),
			"itemsCursor": (*githubv4.String)(nil),
		}
	}
	var allItemsQuery struct {
		Organization struct {
			ProjectV2 projectV2AllItemsQueryResult[FV] `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}
	items, err := paginateProjectV2Items(ctx, gql, &allItemsQuery, newVariables(), &allItemsQuery.Organization.ProjectV2.Items)
	if !unsupportedArchivedStates(err) {
		return items, err
	}
	var query struct {
		Organization struct {
			ProjectV2 projectV2ItemsQueryResult[FV] `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $owner)"`
	}
	return paginateProjectV2Items(ctx, gql, &query, newVariables(), &query.Organization.ProjectV2.Items)
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

// CreateProjectV2FieldMultiSelectOptionInput is a multi-select option for field creation.
type CreateProjectV2FieldMultiSelectOptionInput struct {
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
	MultiSelectOptions     []CreateProjectV2FieldMultiSelectOptionInput  `json:"multiSelectOptions,omitempty"`
	IterationConfiguration *ProjectV2IterationFieldConfigInput           `json:"iterationConfiguration,omitempty"`
}

// AddProjectV2DraftIssueInput is the input for adding a draft issue to a Project v2.
type AddProjectV2DraftIssueInput struct {
	ProjectID githubv4.ID      `json:"projectId"`
	Title     githubv4.String  `json:"title"`
	Body      *githubv4.String `json:"body,omitempty"`
}

// AddProjectV2ItemByIdInput is the input for linking an existing issue or PR to a Project v2.
// The JSON tags map to the GraphQL schema type AddProjectV2ItemByIdInput.
type AddProjectV2ItemByIdInput struct {
	ProjectID githubv4.ID `json:"projectId"`
	ContentID githubv4.ID `json:"contentId"`
}

// ProjectV2FieldValueInput represents the value to set on a project item field.
// Only one of Text/Number/Date/SingleSelectOptionID/MultiSelectOptionIDs/IterationID should be set.
type ProjectV2FieldValueInput struct {
	Text                 *githubv4.String   `json:"text,omitempty"`
	Number               *githubv4.Float    `json:"number,omitempty"`
	Date                 *githubv4.Date     `json:"date,omitempty"`
	SingleSelectOptionID *githubv4.String   `json:"singleSelectOptionId,omitempty"`
	MultiSelectOptionIDs *[]githubv4.String `json:"multiSelectOptionIds,omitempty"`
	IterationID          *githubv4.String   `json:"iterationId,omitempty"`
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
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"archiveProjectV2Item(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// ListUserProjectV2Views lists all views for a user's ProjectV2.
func (g *GitHubClient) ListUserProjectV2Views(ctx context.Context, login string, number int) ([]ProjectV2View, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	views, err := listUserProjectV2Views[multiSelectFieldConfigNameRef](ctx, gql, login, number)
	if !unsupportedMultiSelect(err) {
		return views, err
	}
	return listUserProjectV2Views[fieldConfigNameRef](ctx, gql, login, number)
}

func listUserProjectV2Views[R nameRef](ctx context.Context, gql *githubv4.Client, login string, number int) ([]ProjectV2View, error) {
	var query struct {
		User struct {
			ProjectV2 struct {
				Views struct {
					Nodes    []projectV2ViewNode[R]
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
		for i := range query.User.ProjectV2.Views.Nodes {
			all = append(all, query.User.ProjectV2.Views.Nodes[i].toProjectV2View())
		}
		if !query.User.ProjectV2.Views.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.User.ProjectV2.Views.PageInfo.EndCursor)
	}
	return all, nil
}

// ListOrgProjectV2Views lists all views for an org's ProjectV2.
func (g *GitHubClient) ListOrgProjectV2Views(ctx context.Context, org string, number int) ([]ProjectV2View, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	views, err := listOrgProjectV2Views[multiSelectFieldConfigNameRef](ctx, gql, org, number)
	if !unsupportedMultiSelect(err) {
		return views, err
	}
	return listOrgProjectV2Views[fieldConfigNameRef](ctx, gql, org, number)
}

func listOrgProjectV2Views[R nameRef](ctx context.Context, gql *githubv4.Client, org string, number int) ([]ProjectV2View, error) {
	var query struct {
		Organization struct {
			ProjectV2 struct {
				Views struct {
					Nodes    []projectV2ViewNode[R]
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
		for i := range query.Organization.ProjectV2.Views.Nodes {
			all = append(all, query.Organization.ProjectV2.Views.Nodes[i].toProjectV2View())
		}
		if !query.Organization.ProjectV2.Views.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.ProjectV2.Views.PageInfo.EndCursor)
	}
	return all, nil
}

// GetProjectV2ByID retrieves a ProjectV2 by its GraphQL node ID.
func (g *GitHubClient) GetProjectV2ByID(ctx context.Context, id string) (*ProjectV2, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		Node struct {
			ProjectV2 ProjectV2 `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]any{
		"id": githubv4.ID(id),
	}
	if err := gql.Query(ctx, &query, variables); err != nil {
		return nil, err
	}
	p := query.Node.ProjectV2
	return &p, nil
}

// ListRepositoryProjectsV2 lists all ProjectV2s linked to a repository.
func (g *GitHubClient) ListRepositoryProjectsV2(ctx context.Context, owner string, name string, first int) ([]ProjectV2, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var query struct {
		Repository struct {
			ProjectsV2 struct {
				Nodes    []ProjectV2
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"projectsV2(first: $first, after: $cursor)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]any{
		"owner":  githubv4.String(owner),
		"name":   githubv4.String(name),
		"first":  githubv4.Int(first),
		"cursor": (*githubv4.String)(nil),
	}
	var all []ProjectV2
	for {
		if err := gql.Query(ctx, &query, variables); err != nil {
			return nil, err
		}
		all = append(all, query.Repository.ProjectsV2.Nodes...)
		if !query.Repository.ProjectsV2.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Repository.ProjectsV2.PageInfo.EndCursor)
	}
	return all, nil
}

// CopyProjectV2Input is the input for copying a GitHub Project v2.
type CopyProjectV2Input struct {
	ProjectID          githubv4.ID       `json:"projectId"`
	OwnerID            githubv4.ID       `json:"ownerId"`
	Title              githubv4.String   `json:"title"`
	IncludeDraftIssues *githubv4.Boolean `json:"includeDraftIssues,omitempty"`
}

// MarkProjectV2AsTemplateInput is the input for marking a Project v2 as a template.
type MarkProjectV2AsTemplateInput struct {
	ProjectID githubv4.ID `json:"projectId"`
}

// UnmarkProjectV2AsTemplateInput is the input for unmarking a Project v2 as a template.
type UnmarkProjectV2AsTemplateInput struct {
	ProjectID githubv4.ID `json:"projectId"`
}

// CopyProjectV2 copies a GitHub Project v2 to the given owner.
func (g *GitHubClient) CopyProjectV2(ctx context.Context, input CopyProjectV2Input) (*ProjectV2, error) {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}
	var mutation struct {
		CopyProjectV2 struct {
			ProjectV2 ProjectV2
		} `graphql:"copyProjectV2(input: $input)"`
	}
	if err := gql.Mutate(ctx, &mutation, input, nil); err != nil {
		return nil, err
	}
	p := mutation.CopyProjectV2.ProjectV2
	return &p, nil
}

// MarkProjectV2AsTemplate marks a GitHub Project v2 as a template.
func (g *GitHubClient) MarkProjectV2AsTemplate(ctx context.Context, input MarkProjectV2AsTemplateInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		MarkProjectV2AsTemplate struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"markProjectV2AsTemplate(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}

// UnmarkProjectV2AsTemplate unmarks a GitHub Project v2 as a template.
func (g *GitHubClient) UnmarkProjectV2AsTemplate(ctx context.Context, input UnmarkProjectV2AsTemplateInput) error {
	gql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return err
	}
	var mutation struct {
		UnmarkProjectV2AsTemplate struct {
			ClientMutationID githubv4.String `graphql:"clientMutationId"`
		} `graphql:"unmarkProjectV2AsTemplate(input: $input)"`
	}
	return gql.Mutate(ctx, &mutation, input, nil)
}
