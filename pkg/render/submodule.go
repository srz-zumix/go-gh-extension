package render

import (
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// RenderSubmodules renders repository submodules as a table
func (r *Renderer) RenderSubmodules(submodules []client.RepositorySubmodule) {
	if r.exporter != nil {
		r.RenderExportedData(submodules)
		return
	}
	headers := []string{"Name", "Path", "Repository", "Branch"}
	table := r.newTableWriter(headers)
	for _, sub := range submodules {
		table.Append([]string{
			sub.Name,
			sub.Path,
			parser.GetRepositoryFullName(sub.Repository),
			sub.Branch,
		})
	}
	table.Render()
}
