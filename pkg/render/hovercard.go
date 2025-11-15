package render

import (
	"github.com/google/go-github/v79/github"
)

func (r *Renderer) RenderHovercard(hovercard *github.Hovercard) {
	if r.exporter != nil {
		r.RenderExportedData(hovercard)
		return
	}

	headers := []string{"MESSAGE", "OCTION"}
	table := r.newTableWriter(headers)

	for _, context := range hovercard.Contexts {
		row := make([]string, len(headers))
		row[0] = ToString(context.Message)
		row[1] = ToString(context.Octicon)
		table.Append(row)
	}
	table.Render()
}
