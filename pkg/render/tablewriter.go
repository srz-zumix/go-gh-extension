package render

import (
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// TableWriter wraps *tablewriter.Table so that Append errors are logged
// rather than returned, eliminating per-call error checks at the call site.
// Render returns the first error from the underlying table.Render call; Append
// errors are only logged and are not returned by Render.
type TableWriter struct {
	table    *tablewriter.Table
	renderer *Renderer
}

// Append adds a row to the table. If an error occurs it is logged and discarded.
func (s *TableWriter) Append(row []string) {
	if err := s.table.Append(row); err != nil {
		s.renderer.WriteError(err)
	}
}

// Configure allows callers to customize the underlying tablewriter.Table.
func (s *TableWriter) Configure(f func(*tablewriter.Config)) {
	s.table.Configure(f)
}

// Render flushes the table to the output stream.
func (s *TableWriter) Render() error {
	return s.table.Render()
}

// newTableWriter creates a TableWriter with the given column headers.
func (r *Renderer) newTableWriter(header []string) *TableWriter {
	table := tablewriter.NewTable(r.IO.Out)
	table.Configure(func(config *tablewriter.Config) {
		config.Row.Alignment.Global = tw.AlignLeft
	})

	anyHeader := make([]any, len(header))
	for i, h := range header {
		anyHeader[i] = h
	}
	table.Header(anyHeader...)

	return &TableWriter{table: table, renderer: r}
}
