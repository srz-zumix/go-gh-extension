package gh

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
	"github.com/srz-zumix/go-gh-extension/pkg/unity"
)

func GetObjectName(item any) string {
	switch v := item.(type) {
	case *parser.ActionReference:
		if v == nil {
			return ""
		}
		return v.Name()
	case parser.ActionReference:
		return v.Name()
	case *github.App:
		if v == nil || v.Slug == nil {
			return ""
		}
		return *v.Slug
	case *github.CustomOrgRoles:
		if v == nil {
			return ""
		}
		return v.GetName()
	case *github.Label:
		if v == nil {
			return ""
		}
		return v.GetName()
	case *client.Mannequin:
		if v == nil {
			return ""
		}
		return string(v.Login)
	case client.Mannequin:
		return string(v.Login)
	case *github.RepoDependencies:
		if v == nil {
			return ""
		}
		name := v.GetName()
		if name == "" {
			return ""
		}
		s := strings.Split(name, ":")
		if len(s) < 2 {
			return name
		}
		return s[1]
	case *github.Repository:
		if v == nil {
			return ""
		}
		return v.GetFullName()
	case repository.Repository:
		return parser.GetRepositoryFullName(v)
	case *github.RepositoryPermissionLevel:
		if v == nil {
			return ""
		}
		return GetObjectName(v.User)
	case *RepositoryPermissionLevel:
		if v == nil {
			return ""
		}
		return parser.GetRepositoryFullName(v.Repository)
	case RepositorySubmodule:
		return parser.GetRepositoryFullName(v.Repository)
	case *github.SBOM:
		if v == nil {
			return ""
		}
		return *v.SBOM.Name
	case *github.SBOMInfo:
		if v == nil {
			return ""
		}
		return *v.Name
	case *github.Secret:
		if v == nil {
			return ""
		}
		return v.Name
	case *github.Team:
		if v == nil {
			return ""
		}
		return *v.Slug
	case unity.UnityPackage:
		return v.Name
	case *github.User:
		if v == nil {
			return ""
		}
		return *v.Login
	default:
		return ""
	}
}

func GetObjectNames(items any) []string {
	switch v := items.(type) {
	case []parser.ActionReference:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.App:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.CustomOrgRoles:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.Label:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []client.Mannequin:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.RepoDependencies:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.Repository:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []repository.Repository:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.RepositoryPermissionLevel:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*RepositoryPermissionLevel:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []RepositorySubmodule:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.SBOM:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.SBOMInfo:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.Secret:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.Team:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []unity.UnityPackage:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	case []*github.User:
		names := make([]string, len(v))
		for i, item := range v {
			names[i] = GetObjectName(item)
		}
		return names
	default:
		return nil
	}
}

// GetObjectHTMLURL constructs a browsable URL for the given object.
// defaultHost is used when the object does not carry its own host information
// (e.g. "github.com" or a GHES hostname).
func GetObjectHTMLURL(item any, defaultHost string) string {
	host := defaultHost
	if host == "" {
		host = "github.com"
	}
	switch v := item.(type) {
	case *github.Repository:
		if v.HTMLURL != nil {
			return *v.HTMLURL
		}
		return "https://" + host + "/" + *v.FullName
	case repository.Repository:
		h := v.Host
		if h == "" {
			h = host
		}
		return "https://" + h + "/" + v.Owner + "/" + v.Name
	case *github.RepoDependencies:
		s := strings.Split(*v.Name, ":")
		if len(s) < 2 {
			return ""
		}
		return "https://" + host + "/" + s[1]
	default:
		return ""
	}
}
