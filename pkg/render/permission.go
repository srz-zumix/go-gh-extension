package render

import (
	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

func getPermissions(v any) *github.RepositoryPermissions {
	switch v := v.(type) {
	case *github.Repository:
		return v.Permissions
	case *github.Team:
		if v.Permissions == nil {
			return nil
		}
		admin := v.Permissions["admin"]
		maintain := v.Permissions["maintain"]
		push := v.Permissions["push"]
		triage := v.Permissions["triage"]
		pull := v.Permissions["pull"]
		return &github.RepositoryPermissions{
			Admin:    &admin,
			Maintain: &maintain,
			Push:     &push,
			Triage:   &triage,
			Pull:     &pull,
		}
	case *github.User:
		return v.Permissions
	case *github.RepositoryPermissionLevel:
		return v.User.Permissions
	case *gh.RepositoryPermissionLevel:
		if v.PermissionLevel == nil {
			return nil
		}
		return v.PermissionLevel.User.Permissions
	default:
		return nil
	}
}

func (r *Renderer) RenderPermission(v any) error {
	var permissions = getPermissions(v)

	if r.exporter != nil {
		return r.RenderExportedData(permissions)
	}

	r.writeLine(gh.GetPermissionName(permissions))
	return nil
}

type nameWithPermissions struct {
	Name        string
	Permissions *github.RepositoryPermissions
}

func (r *Renderer) RenderPermissions(v any) error {
	switch v := v.(type) {
	case []*github.Repository:
		return r.RenderRepositoryPermissions(v)
	case []*gh.RepositoryPermissionLevel:
		return r.RenderRepositoryPermissionLevels(v)
	default:
		return r.RenderPermission(v)
	}
}

func (r *Renderer) RenderRepositoryPermissions(v []*github.Repository) error {
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
		return r.RenderExportedData(permissionsList)
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
	return table.Render()
}

func (r *Renderer) RenderRepositoryPermissionLevels(v []*gh.RepositoryPermissionLevel) error {
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
		return r.RenderExportedData(permissionsList)
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
	return table.Render()
}
