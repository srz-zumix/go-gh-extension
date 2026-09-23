package render

import (
	"strconv"

	"github.com/srz-zumix/go-gh-extension/pkg/copilotext"
)

// countTableHeader lists the columns used for every axis table in
// RenderCopilotPermissionStats.
var countTableHeader = []string{"KEY", "TOTAL", "APPROVED", "APPROVED_FOR_LOCATION", "DENIED", "UNRESOLVED"}

// renderCountTable renders one axis (e.g. ByCommand) as a table titled by section.
func (r *Renderer) renderCountTable(section string, counts []copilotext.Count) error {
	if len(counts) == 0 {
		return nil
	}
	r.writeLine("")
	r.writeLine(section)
	table := r.newTableWriter(countTableHeader)
	for _, c := range counts {
		table.Append([]string{
			c.Key,
			strconv.Itoa(c.Total),
			strconv.Itoa(c.Approved),
			strconv.Itoa(c.ApprovedForLocation),
			strconv.Itoa(c.Denied),
			strconv.Itoa(c.Unresolved),
		})
	}
	return table.Render()
}

// RenderCopilotPermissionStats renders a copilotext.PermissionStats summary. With an
// exporter configured (e.g. --format json) it writes the full struct as exported data;
// otherwise it prints a summary line followed by one table per non-empty axis.
func (r *Renderer) RenderCopilotPermissionStats(stats *copilotext.PermissionStats) error {
	if r.exporter != nil {
		return r.RenderExportedData(stats)
	}

	r.writeLine("SUMMARY")
	r.writeLine("sessions: " + strconv.Itoa(stats.Sessions) + ", requests: " + strconv.Itoa(stats.Requests))

	sections := []struct {
		name   string
		counts []copilotext.Count
	}{
		{"RESULT", stats.ByResult},
		{"KIND", stats.ByKind},
		{"READONLY", stats.ByReadOnly},
		{"COMMAND", stats.ByCommand},
		{"PATH", stats.ByPath},
		{"URL", stats.ByURL},
		{"CWD", stats.ByCWD},
	}
	for _, section := range sections {
		if err := r.renderCountTable(section.name, section.counts); err != nil {
			return err
		}
	}
	return nil
}
