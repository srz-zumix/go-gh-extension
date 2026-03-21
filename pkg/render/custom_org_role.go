package render

import (
	"strings"

	"github.com/google/go-github/v84/github"
)

// CustomOrgRoleFieldList is the list of valid field names for custom organization role display.
var CustomOrgRoleFieldList = []string{"ID", "NAME", "DESCRIPTION", "BASE_ROLE", "SOURCE", "PERMISSIONS", "CREATED_AT", "UPDATED_AT"}

type customOrgRoleFieldGetter func(r *github.CustomOrgRoles) string
type customOrgRoleFieldGetters struct {
	Func map[string]customOrgRoleFieldGetter
}

// NewCustomOrgRoleFieldGetters creates a new customOrgRoleFieldGetters with default field mappings.
func NewCustomOrgRoleFieldGetters() *customOrgRoleFieldGetters {
	return &customOrgRoleFieldGetters{
		Func: map[string]customOrgRoleFieldGetter{
			"ID": func(r *github.CustomOrgRoles) string {
				return ToString(r.ID)
			},
			"NAME": func(r *github.CustomOrgRoles) string {
				return ToString(r.Name)
			},
			"DESCRIPTION": func(r *github.CustomOrgRoles) string {
				return ToString(r.Description)
			},
			"BASE_ROLE": func(r *github.CustomOrgRoles) string {
				return ToString(r.BaseRole)
			},
			"SOURCE": func(r *github.CustomOrgRoles) string {
				return ToString(r.Source)
			},
			"PERMISSIONS": func(r *github.CustomOrgRoles) string {
				return strings.Join(r.Permissions, ",")
			},
			"CREATED_AT": func(r *github.CustomOrgRoles) string {
				return ToString(r.CreatedAt)
			},
			"UPDATED_AT": func(r *github.CustomOrgRoles) string {
				return ToString(r.UpdatedAt)
			},
		},
	}
}

func (g *customOrgRoleFieldGetters) GetField(r *github.CustomOrgRoles, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(r)
	}
	return ""
}

// RenderCustomOrgRoles renders a list of custom organization roles as a table with the given headers.
func (r *Renderer) RenderCustomOrgRoles(roles []*github.CustomOrgRoles, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(roles)
	}

	if len(roles) == 0 {
		r.writeLine("No custom organization roles.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"ID", "NAME", "DESCRIPTION", "BASE_ROLE", "SOURCE"}
	}

	getter := NewCustomOrgRoleFieldGetters()
	table := r.newTableWriter(headers)

	for _, role := range roles {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(role, header)
		}
		table.Append(row)
	}

	return table.Render()
}
