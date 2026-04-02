package render

import (
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// ProjectV2ItemFields lists the built-in field names available for --field flag completion.
var ProjectV2ItemFields = []string{"ID", "TYPE", "NUMBER", "TITLE", "AUTHOR", "URL", "ARCHIVED"}

type projectV2ItemFieldGetter func(item *client.ProjectV2Item) string
type projectV2ItemFieldGetters struct {
	Func map[string]projectV2ItemFieldGetter
}

// NewProjectV2ItemFieldGetters returns field getter functions for client.ProjectV2Item.
func NewProjectV2ItemFieldGetters() *projectV2ItemFieldGetters {
	return &projectV2ItemFieldGetters{
		Func: map[string]projectV2ItemFieldGetter{
			"ID": func(item *client.ProjectV2Item) string {
				return item.ID
			},
			"TYPE": func(item *client.ProjectV2Item) string {
				return string(item.Content.Type)
			},
			"NUMBER": func(item *client.ProjectV2Item) string {
				if item.Content.Number == 0 {
					return ""
				}
				return fmt.Sprintf("%d", item.Content.Number)
			},
			"TITLE": func(item *client.ProjectV2Item) string {
				return item.Content.Title
			},
			"AUTHOR": func(item *client.ProjectV2Item) string {
				return item.Content.Author
			},
			"URL": func(item *client.ProjectV2Item) string {
				return item.Content.URL
			},
			"ARCHIVED": func(item *client.ProjectV2Item) string {
				return ToString(item.IsArchived)
			},
		},
	}
}

// GetField returns the string value for the given field name.
// Custom project fields (not built-in) are looked up in FieldValues.
func (g *projectV2ItemFieldGetters) GetField(item *client.ProjectV2Item, field string) string {
	upper := strings.ToUpper(field)
	if getter, ok := g.Func[upper]; ok {
		return getter(item)
	}
	// Fall back to custom field values.
	for _, fv := range item.FieldValues {
		if strings.EqualFold(fv.FieldName, field) {
			return projectV2FieldValueString(fv)
		}
	}
	return ""
}

// projectV2FieldValueString formats a field value as a string.
func projectV2FieldValueString(fv client.ProjectV2FieldValue) string {
	switch fv.ValueType {
	case "TEXT":
		return fv.Text
	case "NUMBER":
		if fv.Number != nil {
			return fmt.Sprintf("%g", *fv.Number)
		}
		return ""
	case "DATE":
		return fv.Date
	case "SINGLE_SELECT":
		return fv.SelectName
	case "ITERATION":
		return fv.IterationTitle
	}
	return ""
}

// ProjectV2Fields lists the built-in field names available for projects list --field flag completion.
var ProjectV2Fields = []string{"ID", "NUMBER", "TITLE", "STATE", "PUBLIC", "URL"}

// RenderProjectsV2 renders a table of GitHub Project v2 projects with the specified headers.
func (r *Renderer) RenderProjectsV2(projects []client.ProjectV2, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(projects)
	}

	if len(projects) == 0 {
		r.writeLine("No projects.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"NUMBER", "TITLE", "STATE", "URL"}
	}

	table := r.newTableWriter(headers)
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Row.Formatting.AutoWrap = tw.WrapNone
	})

	for i := range projects {
		p := &projects[i]
		row := make([]string, len(headers))
		for j, h := range headers {
			row[j] = projectV2Field(p, h)
		}
		table.Append(row)
	}
	return table.Render()
}

// projectV2Field returns the string value for the given field name of a GitHub Project v2.
func projectV2Field(p *client.ProjectV2, field string) string {
	switch strings.ToUpper(field) {
	case "ID":
		return p.ID
	case "NUMBER":
		return fmt.Sprintf("%d", p.Number)
	case "TITLE":
		return p.Title
	case "STATE":
		if p.Closed {
			return "closed"
		}
		return "open"
	case "PUBLIC":
		return ToString(p.Public)
	case "URL":
		return p.URL
	}
	return ""
}

// RenderProjectV2Items renders a table of project v2 items with the specified headers.
func (r *Renderer) RenderProjectV2Items(items []client.ProjectV2Item, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(items)
	}

	if len(items) == 0 {
		r.writeLine("No items.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"TYPE", "NUMBER", "TITLE", "URL"}
	}

	getter := NewProjectV2ItemFieldGetters()
	table := r.newTableWriter(headers)
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Row.Formatting.AutoWrap = tw.WrapNone
	})

	for i := range items {
		row := make([]string, len(headers))
		for j, header := range headers {
			row[j] = getter.GetField(&items[i], header)
		}
		table.Append(row)
	}
	return table.Render()
}
