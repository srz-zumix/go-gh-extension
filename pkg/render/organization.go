package render

import (
	"github.com/google/go-github/v84/github"
)

// OrgMemberPrivilegeFieldList is the ordered list of fields used when displaying member privileges.
var OrgMemberPrivilegeFieldList = []string{
	"DEFAULT_REPO_PERMISSION",
	"MEMBERS_CAN_CREATE_REPOS",
	"MEMBERS_CAN_CREATE_PUBLIC_REPOS",
	"MEMBERS_CAN_CREATE_PRIVATE_REPOS",
	"MEMBERS_CAN_CREATE_INTERNAL_REPOS",
	"MEMBERS_CAN_FORK_PRIVATE_REPOS",
	"MEMBERS_CAN_CREATE_PAGES",
	"MEMBERS_CAN_CREATE_PUBLIC_PAGES",
	"MEMBERS_CAN_CREATE_PRIVATE_PAGES",
	"MEMBERS_CAN_CREATE_TEAMS",
	"WEB_COMMIT_SIGNOFF_REQUIRED",
}

type orgMemberPrivilegeFieldGetter func(org *github.Organization) string
type orgMemberPrivilegeFieldGetters struct {
	Func map[string]orgMemberPrivilegeFieldGetter
}

// newOrgMemberPrivilegeFieldGetters returns field getter functions for member privilege fields.
func newOrgMemberPrivilegeFieldGetters() *orgMemberPrivilegeFieldGetters {
	return &orgMemberPrivilegeFieldGetters{
		Func: map[string]orgMemberPrivilegeFieldGetter{
			"DEFAULT_REPO_PERMISSION": func(org *github.Organization) string {
				return ToString(org.DefaultRepoPermission)
			},
			"MEMBERS_CAN_CREATE_REPOS": func(org *github.Organization) string {
				return ToString(org.MembersCanCreateRepos)
			},
			"MEMBERS_CAN_CREATE_PUBLIC_REPOS": func(org *github.Organization) string {
				return ToString(org.MembersCanCreatePublicRepos)
			},
			"MEMBERS_CAN_CREATE_PRIVATE_REPOS": func(org *github.Organization) string {
				return ToString(org.MembersCanCreatePrivateRepos)
			},
			"MEMBERS_CAN_CREATE_INTERNAL_REPOS": func(org *github.Organization) string {
				return ToString(org.MembersCanCreateInternalRepos)
			},
			"MEMBERS_CAN_FORK_PRIVATE_REPOS": func(org *github.Organization) string {
				return ToString(org.MembersCanForkPrivateRepos)
			},
			"MEMBERS_CAN_CREATE_PAGES": func(org *github.Organization) string {
				return ToString(org.MembersCanCreatePages)
			},
			"MEMBERS_CAN_CREATE_PUBLIC_PAGES": func(org *github.Organization) string {
				return ToString(org.MembersCanCreatePublicPages)
			},
			"MEMBERS_CAN_CREATE_PRIVATE_PAGES": func(org *github.Organization) string {
				return ToString(org.MembersCanCreatePrivatePages)
			},
			"MEMBERS_CAN_CREATE_TEAMS": func(org *github.Organization) string {
				return ToString(org.MembersCanCreateTeams)
			},
			"WEB_COMMIT_SIGNOFF_REQUIRED": func(org *github.Organization) string {
				return ToString(org.WebCommitSignoffRequired)
			},
		},
	}
}

// RenderOrgMemberPrivileges renders the member privileges of an organization as a key-value table.
func (r *Renderer) RenderOrgMemberPrivileges(org *github.Organization, fields []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(org)
	}

	if len(fields) == 0 {
		fields = OrgMemberPrivilegeFieldList
	}

	getter := newOrgMemberPrivilegeFieldGetters()
	table := r.newTableWriter([]string{"FIELD", "VALUE"})

	for _, field := range fields {
		if fn, ok := getter.Func[field]; ok {
			table.Append([]string{field, fn(org)})
		}
	}

	return table.Render()
}
