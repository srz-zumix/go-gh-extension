package render

import (
	"fmt"

	"github.com/google/go-github/v84/github"
)

// RenderTagProtections renders a list of tag protections in a table format.
func (r *Renderer) RenderTagProtections(tagProtections []*github.TagProtection) error {
	if r.exporter != nil {
		return r.RenderExportedData(tagProtections)
	}

	if len(tagProtections) == 0 {
		r.writeLine("No tag protections.")
		return nil
	}

	table := r.newTableWriter([]string{"ID", "PATTERN"})
	for _, tagProtection := range tagProtections {
		table.Append([]string{ToString(tagProtection.ID), ToString(tagProtection.Pattern)})
	}
	return table.Render()
}

// RenderTagProtection renders detailed tag protection settings as a key-value table.
func (r *Renderer) RenderTagProtection(tagProtection *github.TagProtection) error {
	if r.exporter != nil {
		return r.RenderExportedData(tagProtection)
	}

	if tagProtection == nil {
		r.writeLine("Tag protection not found.")
		return nil
	}

	r.writeLine(fmt.Sprintf("Pattern: %s", ToString(tagProtection.Pattern)))
	r.writeLine("")

	table := r.newTableWriter([]string{"SETTING", "VALUE"})
	table.Append([]string{"ID", ToString(tagProtection.ID)})
	table.Append([]string{"Pattern", ToString(tagProtection.Pattern)})
	return table.Render()
}
