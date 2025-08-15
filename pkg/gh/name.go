package gh

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v73/github"
)

func GetObjectName(item any) string {
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
	case *RepositoryPermissionLevel:
		return fmt.Sprintf("%s/%s", v.Repository.Owner, v.Repository.Name)
	case *github.Label:
		return *v.Name
	case *github.RepoDependencies:
		s := strings.Split(*v.Name, ":")
		return s[1]
	case *github.SBOM:
		return *v.SBOM.Name
	case *github.SBOMInfo:
		return *v.Name
	default:
		return ""
	}
}

func GetObjectNames(items any) []string {
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
	case []*github.RepoDependencies:
		names := make([]string, len(v))
		for i, item := range v {
			s := strings.Split(*item.Name, ":")
			names[i] = s[1]
		}
		return names
	default:
		return nil
	}
}
