package render

import (
	"github.com/google/go-github/v73/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

func (r *Renderer) RenderRepository(repos []*github.Repository) {
	if r.exporter != nil {
		r.RenderExportedData(repos)
		return
	}
	headers := []string{"NAME", "PERMISSION", "VISIBILITY"}
	table := r.newTableWriter(headers)

	for _, repo := range repos {
		permission := gh.GetRepositoryPermissions(repo)
		row := []string{
			*repo.FullName,
			permission,
			*repo.Visibility,
		}
		table.Append(row)
	}
	table.Render()
}
