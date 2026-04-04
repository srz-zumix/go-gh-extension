package render

import (
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// ProjectV1Fields lists the built-in field names available for --field flag completion.
var ProjectV1Fields = []string{"ID", "NUMBER", "NAME", "STATE", "BODY", "URL"}

// ProjectV1ColumnFields lists the built-in field names for columns.
var ProjectV1ColumnFields = []string{"ID", "NAME", "URL"}

// ProjectV1CardFields lists the built-in field names for cards.
var ProjectV1CardFields = []string{"ID", "NOTE", "ARCHIVED", "CONTENT_URL", "URL"}

// RenderProjectsV1 renders a table of classic projects with the specified headers.
func (r *Renderer) RenderProjectsV1(projects []client.ProjectV1, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(projects)
	}

	if len(projects) == 0 {
		r.writeLine("No projects.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"NUMBER", "NAME", "STATE", "URL"}
	}

	table := r.newTableWriter(headers)
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Row.Formatting.AutoWrap = tw.WrapNone
	})

	for i := range projects {
		p := &projects[i]
		row := make([]string, len(headers))
		for j, h := range headers {
			row[j] = projectV1Field(p, h)
		}
		table.Append(row)
	}
	return table.Render()
}

// projectV1Field returns the string value for the given field name of a classic project.
func projectV1Field(p *client.ProjectV1, field string) string {
	switch strings.ToUpper(field) {
	case "ID":
		return fmt.Sprintf("%d", p.ID)
	case "NUMBER":
		return fmt.Sprintf("%d", p.Number)
	case "NAME":
		return p.Name
	case "STATE":
		return p.State
	case "BODY":
		return p.Body
	case "URL":
		return p.HTMLURL
	}
	return ""
}

// RenderProjectV1Columns renders a table of classic project columns with the specified headers.
func (r *Renderer) RenderProjectV1Columns(columns []client.ProjectV1Column, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(columns)
	}

	if len(columns) == 0 {
		r.writeLine("No columns.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"ID", "NAME", "URL"}
	}

	table := r.newTableWriter(headers)
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Row.Formatting.AutoWrap = tw.WrapNone
	})

	for i := range columns {
		col := &columns[i]
		row := make([]string, len(headers))
		for j, h := range headers {
			row[j] = projectV1ColumnField(col, h)
		}
		table.Append(row)
	}
	return table.Render()
}

// projectV1ColumnField returns the string value for the given field name of a classic project column.
func projectV1ColumnField(col *client.ProjectV1Column, field string) string {
	switch strings.ToUpper(field) {
	case "ID":
		return fmt.Sprintf("%d", col.ID)
	case "NAME":
		return col.Name
	case "URL":
		return col.URL
	}
	return ""
}

// RenderProjectV1Cards renders a table of classic project cards with the specified headers.
func (r *Renderer) RenderProjectV1Cards(cards []client.ProjectV1Card, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(cards)
	}

	if len(cards) == 0 {
		r.writeLine("No cards.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"ID", "NOTE", "CONTENT_URL", "ARCHIVED"}
	}

	table := r.newTableWriter(headers)
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Row.Formatting.AutoWrap = tw.WrapNone
	})

	for i := range cards {
		card := &cards[i]
		row := make([]string, len(headers))
		for j, h := range headers {
			row[j] = projectV1CardField(card, h)
		}
		table.Append(row)
	}
	return table.Render()
}

// projectV1CardField returns the string value for the given field name of a classic project card.
func projectV1CardField(card *client.ProjectV1Card, field string) string {
	switch strings.ToUpper(field) {
	case "ID":
		return fmt.Sprintf("%d", card.ID)
	case "NOTE":
		if card.Note != nil {
			return *card.Note
		}
		return ""
	case "ARCHIVED":
		return ToString(card.Archived)
	case "CONTENT_URL":
		if card.ContentURL != nil {
			return *card.ContentURL
		}
		return ""
	case "URL":
		return card.URL
	}
	return ""
}
