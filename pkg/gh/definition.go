package gh

import (
	"slices"

	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// defaultHost is an alias for client.DefaultHost for convenience within this package.
const defaultHost = client.DefaultHost

// defaultV3Endpoint is an alias for client.DefaultV3Endpoint for convenience within this package.
const defaultV3Endpoint = client.DefaultV3Endpoint

var PermissionsList = []string{
	"admin",
	"maintain",
	"push",
	"triage",
	"pull",
}

// TeamPrivacyList is the list of valid values for the privacy setting of a team.
var TeamPrivacyList = []string{
	"secret",
	"closed",
}

// TeamNotificationSettingList is the list of valid values for the notification setting of a team.
var TeamNotificationSettingList = []string{
	"notifications_enabled",
	"notifications_disabled",
}

var TeamMembershipList = []string{
	TeamMembershipRoleMember,
	TeamMembershipRoleMaintainer,
}

var OrgMembershipList = []string{
	"member",
	"admin",
}

// OrgDefaultRepoPermissionList is the list of valid values for the default repository permission in an organization.
var OrgDefaultRepoPermissionList = []string{
	"read",
	"write",
	"admin",
	"none",
}

// OrgCustomRoleSourceList is the list of valid source values for custom organization roles.
var OrgCustomRoleSourceList = []string{
	"Organization",
	"Predefined",
	"System",
}

var RepoSearchTypeList = []string{
	"public",
	"internal",
	"private",
	"forks",
	"sources",
	"member",
}

var RepoVisibilityList = []string{
	"public",
	"private",
	"internal",
}

var PackageVisibilityList = []string{
	"public",
	"private",
	"internal",
}

var CollaboratorAffiliationList = []string{
	"outside",
	"direct",
}

type RespositorySearchOptions struct {
	Visibility []string
	Fork       *bool
	Archived   *bool
	Mirror     *bool
	Template   *bool
}

func (opt *RespositorySearchOptions) SetFork(fork bool) {
	opt.Fork = new(bool)
	*opt.Fork = fork
}
func (opt *RespositorySearchOptions) SetArchived(archived bool) {
	opt.Archived = new(bool)
	*opt.Archived = archived
}
func (opt *RespositorySearchOptions) SetMirror(mirror bool) {
	opt.Mirror = new(bool)
	*opt.Mirror = mirror
}
func (opt *RespositorySearchOptions) SetTemplate(template bool) {
	opt.Template = new(bool)
	*opt.Template = template
}
func (opt *RespositorySearchOptions) SetSources() {
	opt.SetFork(false)
	opt.SetArchived(false)
	opt.SetMirror(false)
}

func (opt *RespositorySearchOptions) Sources() bool {
	if opt.Fork == nil || *opt.Fork {
		return false
	}
	if opt.Archived == nil || *opt.Archived {
		return false
	}
	if opt.Mirror == nil || *opt.Mirror {
		return false
	}
	return true
}

func (opt *RespositorySearchOptions) GetFilterString() string {
	if opt != nil {
		matched := 0
		for _, role := range opt.Visibility {
			if slices.Contains(RepoVisibilityList, role) {
				matched++
			}
		}
		if matched == 1 && len(opt.Visibility) == 1 {
			return opt.Visibility[0]
		}

		if opt.Sources() {
			return "sources"
		}
		if opt.Fork != nil && *opt.Fork {
			return "forks"
		}
	}
	return "all"
}

func (opt *RespositorySearchOptions) Filter(repos []*github.Repository) []*github.Repository {
	if opt == nil {
		return repos
	}
	var filteredRepos []*github.Repository
	for _, repo := range repos {
		if opt.Fork != nil {
			if *opt.Fork != *repo.Fork {
				continue
			}
		}
		if opt.Archived != nil {
			if *opt.Archived != *repo.Archived {
				continue
			}
		}
		if opt.Mirror != nil {
			hasMirror := repo.MirrorURL != nil && *repo.MirrorURL != ""
			if *opt.Mirror != hasMirror {
				continue
			}
		}
		if opt.Template != nil {
			if *opt.Template != *repo.IsTemplate {
				continue
			}
		}
		if opt.Visibility != nil {
			if !slices.Contains(opt.Visibility, *repo.Visibility) {
				continue
			}
		}
		filteredRepos = append(filteredRepos, repo)
	}
	return filteredRepos
}

func GetRepositoryPermissions(repo *github.Repository) string {
	if repo != nil {
		if repo.Permissions != nil {
			return GetPermissionName(repo.Permissions)
		}
	}
	return "none"
}

func GetPermissionName(permissions *github.RepositoryPermissions) string {
	if permissions == nil {
		return "none"
	}
	if permissions.GetAdmin() {
		return "admin"
	}
	if permissions.GetMaintain() {
		return "maintain"
	}
	if permissions.GetPush() {
		return "push"
	}
	if permissions.GetTriage() {
		return "triage"
	}
	if permissions.GetPull() {
		return "pull"
	}
	return "none"
}

// GetPermissionNameFromMap returns the highest permission name from a map of permissions.
// Used for types that still use map[string]bool (e.g. github.Team).
func GetPermissionNameFromMap(permissions map[string]bool) string {
	for _, permission := range PermissionsList {
		if permissions[permission] {
			return permission
		}
	}
	return "none"
}

type PermissionInterface interface {
	GetPermissions() *github.RepositoryPermissions
}

func HasPermission(obj PermissionInterface, roles []string) bool {
	if obj != nil {
		perms := obj.GetPermissions()
		if perms == nil {
			return false
		}
		for _, role := range roles {
			switch role {
			case "admin":
				if perms.GetAdmin() {
					return true
				}
			case "maintain":
				if perms.GetMaintain() {
					return true
				}
			case "push":
				if perms.GetPush() {
					return true
				}
			case "triage":
				if perms.GetTriage() {
					return true
				}
			case "pull":
				if perms.GetPull() {
					return true
				}
			}
		}
	}
	return false
}

// CreateRepositoryPermissions creates a *github.RepositoryPermissions from a list of permission names.
func CreateRepositoryPermissions(permissions []string) *github.RepositoryPermissions {
	admin := slices.Contains(permissions, "admin")
	maintain := slices.Contains(permissions, "maintain")
	push := slices.Contains(permissions, "push")
	triage := slices.Contains(permissions, "triage")
	pull := slices.Contains(permissions, "pull")
	return &github.RepositoryPermissions{
		Admin:    &admin,
		Maintain: &maintain,
		Push:     &push,
		Triage:   &triage,
		Pull:     &pull,
	}
}

func FilterByRepositoryPermissions(repos []*github.Repository, permissions []string) []*github.Repository {
	if len(permissions) > 0 {
		var filtered []*github.Repository
		for _, r := range repos {
			if HasPermission(r, permissions) {
				filtered = append(filtered, r)
			}
		}
		return filtered
	}
	return repos
}

func FilterByUserPermissions(repos []*github.User, permissions []string) []*github.User {
	if len(permissions) > 0 {
		var filtered []*github.User
		for _, r := range repos {
			if HasPermission(r, permissions) {
				filtered = append(filtered, r)
			}
		}
		return filtered
	}
	return repos
}

func GetTeamMembershipFilter(roles []string) string {
	if roles != nil {
		matched := 0
		for _, role := range roles {
			if slices.Contains(TeamMembershipList, role) {
				matched++
			}
		}
		if matched == 1 && len(roles) == 1 {
			return roles[0]
		}
	}
	return "all"
}

func GetOrgMembershipFilter(roles []string) string {
	if roles != nil {
		matched := 0
		for _, role := range roles {
			if slices.Contains(OrgMembershipList, role) {
				matched++
			}
		}
		if matched == 1 && len(roles) == 1 {
			return roles[0]
		}
	}
	return "all"
}

func GetCollaboratorAffiliationsFilter(affiliations []string) string {
	if affiliations != nil {
		matched := 0
		for _, role := range affiliations {
			if slices.Contains(CollaboratorAffiliationList, role) {
				matched++
			}
		}
		if matched == 1 && len(affiliations) == 1 {
			return affiliations[0]
		}
	}
	return "all"
}
