package render

import (
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// ProjectV2ViewFields lists the built-in view list output columns available for --field flag completion.
var ProjectV2ViewFields = []string{"ID", "NUMBER", "NAME", "LAYOUT", "FILTER", "GROUPBY", "VERTICALGROUPBY", "SORTBY", "VISIBLEFIELDS"}

// RenderProjectV2Views renders a table of project v2 views with the specified headers.
func (r *Renderer) RenderProjectV2Views(views []client.ProjectV2View, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(views)
	}

	if len(views) == 0 {
		r.writeLine("No views.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"NUMBER", "NAME", "LAYOUT", "FILTER", "GROUPBY", "SORTBY"}
	}

	table := r.newTableWriter(headers)
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Row.Formatting.AutoWrap = tw.WrapNone
	})

	for i := range views {
		v := &views[i]
		row := make([]string, len(headers))
		for j, h := range headers {
			row[j] = projectV2ViewField(v, h)
		}
		table.Append(row)
	}
	return table.Render()
}

// projectV2ViewField returns the string value for the given field name of a project view.
func projectV2ViewField(v *client.ProjectV2View, field string) string {
	switch strings.ToUpper(field) {
	case "ID":
		return v.ID
	case "NUMBER":
		return fmt.Sprintf("%d", v.Number)
	case "NAME":
		return v.Name
	case "LAYOUT":
		return strings.TrimSuffix(v.Layout, "_LAYOUT")
	case "FILTER":
		return v.Filter
	case "GROUPBY":
		return strings.Join(v.GroupByFields, ", ")
	case "VERTICALGROUPBY":
		return strings.Join(v.VerticalGroupByFields, ", ")
	case "SORTBY":
		return projectV2ViewSortByString(v.SortBy)
	case "VISIBLEFIELDS":
		return strings.Join(v.VisibleFields, ", ")
	}
	return ""
}

// projectV2ViewSortByString formats sort criteria as "field (ASC), field (DESC)".
func projectV2ViewSortByString(sortBy []client.ProjectV2ViewSortBy) string {
	parts := make([]string, len(sortBy))
	for i, s := range sortBy {
		parts[i] = fmt.Sprintf("%s (%s)", s.FieldName, s.Direction)
	}
	return strings.Join(parts, ", ")
}
