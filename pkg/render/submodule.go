package render

import (
	"fmt"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
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
			fmt.Sprintf("%s/%s", sub.Repository.Owner, sub.Repository.Name),
			sub.Branch,
		})
	}
	table.Render()
}

// RenderSubmoduleNames renders only the names of submodules
func (r *Renderer) RenderSubmoduleNames(submodules []client.RepositorySubmodule) {
	names := make([]string, len(submodules))
	for i, sub := range submodules {
		names[i] = sub.Name
	}
	if r.exporter != nil {
		r.RenderExportedData(names)
		return
	}
	if len(names) == 0 {
		return
	}
	r.writeLine(strings.Join(names, "\n"))
}
