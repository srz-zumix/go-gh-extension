package render

import (
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

// IDPGroupFieldList is the list of valid field names for IDP group display.
var IDPGroupFieldList = []string{"ID", "NAME", "DESCRIPTION"}

type idpGroupFieldGetter func(g *github.IDPGroup) string
type idpGroupFieldGetters struct {
	Func map[string]idpGroupFieldGetter
}

// NewIDPGroupFieldGetters creates a new idpGroupFieldGetters with default field mappings.
func NewIDPGroupFieldGetters() *idpGroupFieldGetters {
	return &idpGroupFieldGetters{
		Func: map[string]idpGroupFieldGetter{
			"ID": func(g *github.IDPGroup) string {
				return ToString(g.GroupID)
			},
			"NAME": func(g *github.IDPGroup) string {
				return ToString(g.GroupName)
			},
			"DESCRIPTION": func(g *github.IDPGroup) string {
				return ToString(g.GroupDescription)
			},
		},
	}
}

func (u *idpGroupFieldGetters) GetField(g *github.IDPGroup, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := u.Func[field]; ok {
		return getter(g)
	}
	return ""
}

// RenderIDPGroups renders a list of IDP groups as a table with the given headers.
func (r *Renderer) RenderIDPGroups(groups []*github.IDPGroup, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(groups)
		return
	}

	if len(groups) == 0 {
		r.writeLine("No IDP groups.")
		return
	}

	if len(headers) == 0 {
		headers = []string{"ID", "NAME", "DESCRIPTION"}
	}

	getter := NewIDPGroupFieldGetters()
	table := r.newTableWriter(headers)

	for _, g := range groups {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(g, header)
		}
		table.Append(row)
	}

	table.Render()
}

// RenderIDPGroupsDefault renders IDP groups with default columns.
func (r *Renderer) RenderIDPGroupsDefault(groups []*github.IDPGroup) {
	r.RenderIDPGroups(groups, nil)
}

// ExternalGroupFieldList is the list of valid field names for external group (EMU) display.
// NOTE: MEMBER_COUNT and MEMBERS are omitted because the list API (ListExternalGroupsInOrganization /
// ListExternalGroupsForTeam) does not populate the Members field. GetExternalGroup also does not
// support pagination for members, so these fields cannot be reliably displayed.
var ExternalGroupFieldList = []string{"ID", "NAME", "UPDATED_AT", "TEAM_COUNT", "TEAMS"}

type externalGroupFieldGetter func(g *github.ExternalGroup) string
type externalGroupFieldGetters struct {
	Func map[string]externalGroupFieldGetter
}

// NewExternalGroupFieldGetters creates a new externalGroupFieldGetters with default field mappings.
func NewExternalGroupFieldGetters() *externalGroupFieldGetters {
	return &externalGroupFieldGetters{
		Func: map[string]externalGroupFieldGetter{
			"ID": func(g *github.ExternalGroup) string {
				return ToString(g.GroupID)
			},
			"NAME": func(g *github.ExternalGroup) string {
				return ToString(g.GroupName)
			},
			"UPDATED_AT": func(g *github.ExternalGroup) string {
				return ToString(g.UpdatedAt)
			},
			"TEAM_COUNT": func(g *github.ExternalGroup) string {
				count := len(g.Teams)
				return ToString(&count)
			},
			// NOTE: MEMBER_COUNT is not supported because GetExternalGroup does not support
			// pagination for members, so the Members slice may be incomplete.
			// "MEMBER_COUNT": func(g *github.ExternalGroup) string {
			// 	count := len(g.Members)
			// 	return ToString(&count)
			// },
			"TEAMS": func(g *github.ExternalGroup) string {
				teamNames := gh.GetObjectNames(g.Teams)
				return strings.Join(teamNames, ", ")
			},
			// NOTE: MEMBER_COUNT and MEMBERS are not supported because GetExternalGroup does not
			// support pagination for members, so the Members slice may be incomplete.
			// "MEMBER_COUNT": func(g *github.ExternalGroup) string {
			// 	count := len(g.Members)
			// 	return ToString(&count)
			// },
			// "MEMBERS": func(g *github.ExternalGroup) string {
			// 	memberNames := gh.GetObjectNames(g.Members)
			// 	return strings.Join(memberNames, ", ")
			// },
		},
	}
}

func (u *externalGroupFieldGetters) GetField(g *github.ExternalGroup, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := u.Func[field]; ok {
		return getter(g)
	}
	return ""
}

// RenderExternalGroups renders a list of external groups (EMU) as a table with the given headers.
func (r *Renderer) RenderExternalGroups(groups []*github.ExternalGroup, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(groups)
	}

	if len(groups) == 0 {
		r.writeLine("No external groups.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"ID", "NAME", "TEAM_COUNT"}
	}

	getter := NewExternalGroupFieldGetters()
	table := r.newTableWriter(headers)

	for _, g := range groups {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(g, header)
		}
		table.Append(row)
	}

	return table.Render()
}

// RenderExternalGroupsDetails renders detailed information for external groups using the provided headers.
func (r *Renderer) RenderExternalGroupsDetails(groups []*github.ExternalGroup, headers []string) error {
	if len(headers) == 0 {
		headers = ExternalGroupFieldList
	}
	return r.RenderExternalGroups(groups, headers)
}

// ExternalGroupTeamFieldList is the list of valid field names for external group team display.
var ExternalGroupTeamFieldList = []string{"TEAM_ID", "TEAM_NAME"}

type externalGroupTeamFieldGetter func(t *github.ExternalGroupTeam) string
type externalGroupTeamFieldGetters struct {
	Func map[string]externalGroupTeamFieldGetter
}

// NewExternalGroupTeamFieldGetters creates a new externalGroupTeamFieldGetters with default field mappings.
func NewExternalGroupTeamFieldGetters() *externalGroupTeamFieldGetters {
	return &externalGroupTeamFieldGetters{
		Func: map[string]externalGroupTeamFieldGetter{
			"TEAM_ID": func(t *github.ExternalGroupTeam) string {
				return ToString(t.TeamID)
			},
			"TEAM_NAME": func(t *github.ExternalGroupTeam) string {
				return ToString(t.TeamName)
			},
		},
	}
}

func (u *externalGroupTeamFieldGetters) GetField(t *github.ExternalGroupTeam, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := u.Func[field]; ok {
		return getter(t)
	}
	return ""
}

// RenderExternalGroupTeams renders the teams connected to an external group as a table.
func (r *Renderer) RenderExternalGroupTeams(teams []*github.ExternalGroupTeam, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(teams)
		return
	}

	if len(teams) == 0 {
		r.writeLine("No teams connected to this external group.")
		return
	}

	if len(headers) == 0 {
		headers = []string{"TEAM_ID", "TEAM_NAME"}
	}

	getter := NewExternalGroupTeamFieldGetters()
	table := r.newTableWriter(headers)

	for _, t := range teams {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(t, header)
		}
		table.Append(row)
	}

	table.Render()
}

// ExternalGroupDetailFieldList is the list of valid field names for a single external group detail.
// NOTE: MEMBER_COUNT is omitted because GetExternalGroup does not support pagination for members.
var ExternalGroupDetailFieldList = []string{"ID", "NAME", "UPDATED_AT", "TEAM_COUNT"}

type externalGroupDetailFieldGetter func(g *github.ExternalGroup) string
type externalGroupDetailFieldGetters struct {
	Func map[string]externalGroupDetailFieldGetter
}

// NewExternalGroupDetailFieldGetters creates field getters for a single external group, including counts.
func NewExternalGroupDetailFieldGetters() *externalGroupDetailFieldGetters {
	return &externalGroupDetailFieldGetters{
		Func: map[string]externalGroupDetailFieldGetter{
			"ID": func(g *github.ExternalGroup) string {
				return ToString(g.GroupID)
			},
			"NAME": func(g *github.ExternalGroup) string {
				return ToString(g.GroupName)
			},
			"UPDATED_AT": func(g *github.ExternalGroup) string {
				return ToString(g.UpdatedAt)
			},
			"TEAM_COUNT": func(g *github.ExternalGroup) string {
				count := len(g.Teams)
				return ToString(&count)
			},
			// NOTE: MEMBER_COUNT is not supported because GetExternalGroup does not support
			// pagination for members, so the Members slice may be incomplete.
			// "MEMBER_COUNT": func(g *github.ExternalGroup) string {
			// 	count := len(g.Members)
			// 	return ToString(&count)
			// },
		},
	}
}

func (u *externalGroupDetailFieldGetters) GetField(g *github.ExternalGroup, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := u.Func[field]; ok {
		return getter(g)
	}
	return ""
}

// RenderExternalGroup renders a single external group as a table row.
func (r *Renderer) RenderExternalGroup(group *github.ExternalGroup, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(group)
	}

	if group == nil {
		r.writeLine("External group not found.")
		return nil
	}

	if len(headers) == 0 {
		headers = ExternalGroupDetailFieldList
	}

	getter := NewExternalGroupDetailFieldGetters()
	table := r.newTableWriter(headers)

	row := make([]string, len(headers))
	for i, header := range headers {
		row[i] = getter.GetField(group, header)
	}
	table.Append(row)
	return table.Render()
}

// ExternalGroupTeamDetailFieldList is the list of valid field names for external group team details.
var ExternalGroupTeamDetailFieldList = []string{"TEAM_ID", "TEAM_NAME", "SLUG", "DESCRIPTION", "PRIVACY", "HTML_URL"}

type externalGroupTeamDetailFieldGetter func(t *gh.ExternalGroupTeamDetail) string
type externalGroupTeamDetailFieldGetters struct {
	Func map[string]externalGroupTeamDetailFieldGetter
}

// NewExternalGroupTeamDetailFieldGetters creates field getters for ExternalGroupTeamDetail.
func NewExternalGroupTeamDetailFieldGetters() *externalGroupTeamDetailFieldGetters {
	return &externalGroupTeamDetailFieldGetters{
		Func: map[string]externalGroupTeamDetailFieldGetter{
			"TEAM_ID": func(t *gh.ExternalGroupTeamDetail) string {
				return ToString(t.Team.ID)
			},
			"TEAM_NAME": func(t *gh.ExternalGroupTeamDetail) string {
				return ToString(t.Team.Name)
			},
			"SLUG": func(t *gh.ExternalGroupTeamDetail) string {
				return ToString(t.Team.Slug)
			},
			"DESCRIPTION": func(t *gh.ExternalGroupTeamDetail) string {
				return ToString(t.Team.Description)
			},
			"PRIVACY": func(t *gh.ExternalGroupTeamDetail) string {
				return ToString(t.Team.Privacy)
			},
			"HTML_URL": func(t *gh.ExternalGroupTeamDetail) string {
				return ToString(t.Team.HTMLURL)
			},
		},
	}
}

func (u *externalGroupTeamDetailFieldGetters) GetField(t *gh.ExternalGroupTeamDetail, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := u.Func[field]; ok {
		return getter(t)
	}
	return ""
}

// RenderExternalGroupTeamDetails renders detailed team info from an external group.
func (r *Renderer) RenderExternalGroupTeamDetails(teams []*gh.ExternalGroupTeamDetail, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(teams)
	}

	if len(teams) == 0 {
		r.writeLine("No teams connected to this external group.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"TEAM_ID", "TEAM_NAME", "SLUG", "DESCRIPTION", "PRIVACY"}
	}

	getter := NewExternalGroupTeamDetailFieldGetters()
	table := r.newTableWriter(headers)

	for _, t := range teams {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(t, header)
		}
		table.Append(row)
	}
	return table.Render()
}
