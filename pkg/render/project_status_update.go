package render

import (
	"sort"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// RenderProjectV2StatusUpdates renders a table of project v2 status updates,
// ordered by CreatedAt descending (newest first).
func (r *Renderer) RenderProjectV2StatusUpdates(updates []client.ProjectV2StatusUpdate) error {
	sorted := make([]client.ProjectV2StatusUpdate, len(updates))
	copy(sorted, updates)
	sort.SliceStable(sorted, func(i, j int) bool {
		return statusUpdateCreatedAt(sorted[i].CreatedAt).After(statusUpdateCreatedAt(sorted[j].CreatedAt))
	})

	if r.exporter != nil {
		return r.RenderExportedData(sorted)
	}

	if len(sorted) == 0 {
		r.writeLine("No status updates.")
		return nil
	}

	table := r.newTableWriter([]string{"CREATED", "STATUS", "START", "TARGET", "CREATOR", "BODY"})
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Row.Formatting.AutoWrap = tw.WrapNone
	})

	for i := range sorted {
		u := &sorted[i]
		table.Append([]string{
			formatRFC3339(u.CreatedAt),
			string(u.Status),
			u.StartDate,
			u.TargetDate,
			u.Creator,
			truncateString(firstLine(u.Body), 60),
		})
	}
	return table.Render()
}

// statusUpdateCreatedAt parses an RFC3339 CreatedAt value for ordering.
// Unparsable or empty values sort last, represented by the zero time.
func statusUpdateCreatedAt(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// formatRFC3339 renders an RFC3339 timestamp using TimeFormat, falling back to the raw value.
func formatRFC3339(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format(TimeFormat)
}

// firstLine returns the first non-empty line of s with surrounding whitespace removed.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
