package render

import (
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// RenderProjectV2Fields renders a table of project v2 field definitions.
// One row per field, with select options and iterations summarized in the OPTIONS column.
func (r *Renderer) RenderProjectV2Fields(fields []client.ProjectV2Field) error {
	if r.exporter != nil {
		return r.RenderExportedData(fields)
	}

	if len(fields) == 0 {
		r.writeLine("No fields.")
		return nil
	}

	table := r.newTableWriter([]string{"NAME", "DATA TYPE", "OPTIONS"})
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Row.Formatting.AutoWrap = tw.WrapNone
	})

	for i := range fields {
		f := &fields[i]
		table.Append([]string{f.Name, f.DataType, projectV2FieldOptionsSummary(f)})
	}
	return table.Render()
}

// RenderProjectV2FieldOptions renders one row per select option or iteration,
// skipping fields that have neither.
func (r *Renderer) RenderProjectV2FieldOptions(fields []client.ProjectV2Field) error {
	if r.exporter != nil {
		return r.RenderExportedData(fields)
	}

	table := r.newTableWriter([]string{"FIELD", "DATA TYPE", "OPTION", "DETAIL"})
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Row.Formatting.AutoWrap = tw.WrapNone
	})

	rows := 0
	for i := range fields {
		f := &fields[i]
		for _, o := range f.Options {
			table.Append([]string{f.Name, f.DataType, o.Name, projectV2OptionDetail(o)})
			rows++
		}
		for _, it := range f.CompletedIterations {
			table.Append([]string{f.Name, f.DataType, it.Title, projectV2IterationDetail(it, true)})
			rows++
		}
		for _, it := range f.Iterations {
			table.Append([]string{f.Name, f.DataType, it.Title, projectV2IterationDetail(it, false)})
			rows++
		}
	}
	if rows == 0 {
		r.writeLine("No field options.")
		return nil
	}
	return table.Render()
}

// projectV2FieldOptionsSummary summarizes the options or iterations of a field for the table view.
func projectV2FieldOptionsSummary(f *client.ProjectV2Field) string {
	if len(f.Options) > 0 {
		names := make([]string, len(f.Options))
		for i, o := range f.Options {
			names[i] = o.Name
		}
		return strings.Join(names, ", ")
	}
	total := len(f.Iterations) + len(f.CompletedIterations)
	if total > 0 {
		return fmt.Sprintf("%d iterations (%d completed)", total, len(f.CompletedIterations))
	}
	return ""
}

// projectV2OptionDetail formats the color and description of a select option.
func projectV2OptionDetail(o client.ProjectV2SelectOption) string {
	parts := []string{}
	if o.Color != "" {
		parts = append(parts, strings.ToLower(o.Color))
	}
	if o.Description != "" {
		parts = append(parts, o.Description)
	}
	return strings.Join(parts, " - ")
}

// projectV2IterationDetail formats the schedule of an iteration.
func projectV2IterationDetail(it client.ProjectV2IterationOption, completed bool) string {
	detail := fmt.Sprintf("%s +%dd", it.StartDate, it.Duration)
	if completed {
		detail += " (completed)"
	}
	return detail
}
