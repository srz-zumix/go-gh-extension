package render

import (
	"github.com/google/go-github/v71/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

func getPermissions(v any) map[string]bool {
	switch v := v.(type) {
	case *github.Repository:
		return v.Permissions
	case *github.Team:
		return v.Permissions
	case *github.User:
		return v.Permissions
	case *github.RepositoryPermissionLevel:
		return v.User.Permissions
	default:
		return nil
	}
}

func (r *Renderer) RenderPermission(v any) {
	var permissions = getPermissions(v)

	if r.exporter != nil {
		r.RenderExportedData(permissions)
		return
	}

	r.WriteLine(gh.GetPermissionName(permissions))
}

type nameWithPermissions struct {
	Name        string
	Permissions map[string]bool
}

func (r *Renderer) RenderPermissions(v any) {
	switch v := v.(type) {
	case []*github.Repository:
		r.RenderRepositoryPermissions(v)
	default:
		r.RenderPermission(v)
	}
}

func (r *Renderer) RenderRepositoryPermissions(v []*github.Repository) {
	var permissionsList []nameWithPermissions
	for _, item := range v {
		var name = getName(item)
		var permissions = getPermissions(item)
		permissionsList = append(permissionsList, nameWithPermissions{
			Name:        name,
			Permissions: permissions,
		})
	}

	if r.exporter != nil {
		r.RenderExportedData(permissionsList)
		return
	}

	headers := []string{"NAME", "PERMISSION"}
	table := r.newTableWriter(headers)

	for _, v := range permissionsList {
		permission := gh.GetPermissionName(v.Permissions)
		row := []string{
			v.Name,
			permission,
		}
		table.Append(row)
	}
	table.Render()
}
