package render

import (
	"github.com/google/go-github/v71/github"
)

func (r *Renderer) RenderCustomOrgRoles(roles []*github.CustomOrgRoles) {
	if r.exporter != nil {
		r.RenderExportedData(roles)
		return
	}

	headers := []string{"ID", "NAME", "DESCRIPTION"}
	table := r.newTableWriter(headers)

	for _, role := range roles {
		data := []string{
			ToString(role.ID),
			*role.Name,
			*role.Description,
		}
		table.Append(data)
	}

	table.Render()
}
