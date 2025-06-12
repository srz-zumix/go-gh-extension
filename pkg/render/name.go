package render

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v71/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

func getName(item any) string {
	switch v := item.(type) {
	case *github.Repository:
		return *v.FullName
	case *github.Team:
		return *v.Slug
	case *github.User:
		return *v.Login
	case *github.CustomOrgRoles:
		return *v.Name
	case *github.RepositoryPermissionLevel:
		return *v.User.Login
	case *gh.RepositoryPermissionLevel:
		return fmt.Sprintf("%s/%s", v.Repository.Owner, v.Repository.Name)
	case *github.Label:
		return *v.Name
	default:
		return ""
	}
}

func getNames(items any) []string {
	switch v := items.(type) {
	case []*github.Repository:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = *item.FullName
		}
		return names
	case []*github.Team:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = *item.Slug
		}
		return names
	case []*github.User:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = *item.Login
		}
		return names
	case []*github.CustomOrgRoles:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = *item.Name
		}
		return names
	case []*github.Label:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = *item.Name
		}
		return names
	default:
		return nil
	}
}

func (r *Renderer) RenderNames(items any) {
	names := getNames(items)
	if r.exporter != nil {
		r.RenderExportedData(names)
		return
	}

	if names == nil {
		return
	}

	r.writeLine(strings.Join(names, "\n"))
}

// RenderNamesWithSeparator renders the names joined by the specified separator
func (r *Renderer) RenderNamesWithSeparator(items any, sep string) {
	names := getNames(items)
	if r.exporter != nil {
		r.RenderExportedData(names)
		return
	}

	if names == nil {
		return
	}

	r.writeLine(strings.Join(names, sep))
}
