package render

import (
	"github.com/google/go-github/v79/github"
)

func (r *Renderer) RenderHovercard(hovercard *github.Hovercard) error {
	if r.exporter != nil {
		return r.RenderExportedData(hovercard)
	}

	headers := []string{"MESSAGE", "OCTICON"}
	table := r.newTableWriter(headers)

	for _, context := range hovercard.Contexts {
		row := make([]string, len(headers))
		row[0] = ToString(context.Message)
		row[1] = ToString(context.Octicon)
		table.Append(row)
	}
	return table.Render()
}
