package render

import (
	"slices"
	"strings"

	"github.com/google/go-github/v73/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

type teamFiledGetter func(team *github.Team) string
type teamFiledGetters struct {
	Func map[string]teamFiledGetter
}

func NewTeamFieldGetters() *teamFiledGetters {
	return &teamFiledGetters{
		Func: map[string]teamFiledGetter{
			"NAME": func(team *github.Team) string {
				return ToString(team.Name)
			},
			"DESCRIPTION": func(team *github.Team) string {
				return ToString(team.Description)
			},
			"MEMBER_COUNT": func(team *github.Team) string {
				return ToString(team.MembersCount)
			},
			"REPOS_COUNT": func(team *github.Team) string {
				return ToString(team.ReposCount)
			},
			"PERMISSION": func(team *github.Team) string {
				return ToString(team.Permission)
			},
			"PRIVACY": func(team *github.Team) string {
				return ToString(team.Privacy)
			},
			"URL": func(team *github.Team) string {
				return ToString(team.HTMLURL)
			},
			"PARENT_SLUG": func(team *github.Team) string {
				if team.Parent == nil {
					return ""
				}
				return ToString(team.Parent.Slug)
			},
			"ORGANIZATION": func(team *github.Team) string {
				return ToString(team.Organization.Login)
			},
		},
	}
}

func (u *teamFiledGetters) GetField(team *github.Team, field string) string {
	if getter, ok := u.Func[field]; ok {
		return getter(team)
	}
	return ""
}

func (r *Renderer) RenderTeams(teams []*github.Team, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(teams)
		return
	}

	headers = slices.DeleteFunc(headers, func(s string) bool {
		switch s {
		case "MEMBER_COUNT":
			return len(teams) == 0 || teams[0].MembersCount == nil
		case "REPOS_COUNT":
			return len(teams) == 0 || teams[0].ReposCount == nil
		}
		return false
	})

	getter := NewTeamFieldGetters()
	table := r.newTableWriter(headers)

	for _, team := range teams {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(team, header)
		}
		table.Append(row)
	}

	table.Render()
}

func (r *Renderer) RenderTeamsDefault(teams []*github.Team) {
	headers := []string{"NAME", "DESCRIPTION", "MEMBER_COUNT", "REPOS_COUNT", "PRIVACY", "PARENT_SLUG"}
	r.RenderTeams(teams, headers)
}

func (r *Renderer) RenderTeamsWithPermission(teams []*github.Team) {
	headers := []string{"NAME", "DESCRIPTION", "MEMBER_COUNT", "REPOS_COUNT", "PRIVACY", "PERMISSION", "PARENT_SLUG"}
	r.RenderTeams(teams, headers)
}

type teamCodeReviewFiledGetter func(s *gh.TeamCodeReviewSettings) string
type teamCodeReviewFiledGetters struct {
	Func map[string]teamCodeReviewFiledGetter
}

func NewTeamCodeReviewFieldGetters() *teamCodeReviewFiledGetters {
	return &teamCodeReviewFiledGetters{
		Func: map[string]teamCodeReviewFiledGetter{
			"NAME": func(s *gh.TeamCodeReviewSettings) string {
				return s.TeamSlug
			},
			"TEAM_SLUG": func(s *gh.TeamCodeReviewSettings) string {
				return s.TeamSlug
			},
			"AUTO_ASSIGNMENT": func(s *gh.TeamCodeReviewSettings) string {
				return ToString(s.Enabled)
			},
			"ALGORITHM": func(s *gh.TeamCodeReviewSettings) string {
				return s.Algorithm
			},
			"MEMBER_COUNT": func(s *gh.TeamCodeReviewSettings) string {
				return ToString(s.TeamMemberCount)
			},
			"NOTIFY_TEAM": func(s *gh.TeamCodeReviewSettings) string {
				return ToString(s.NotifyTeam)
			},
			"EXCLUDED_TEAM_MEMBERS": func(s *gh.TeamCodeReviewSettings) string {
				return strings.Join(s.ExcludedTeamMembers, ", ")
			},
			"INCLUDE_CHILD_TEAM_MEMBERS": func(s *gh.TeamCodeReviewSettings) string {
				return ToString(s.IncludeChildTeamMembers)
			},
			"COUNT_MEMBERS_ALREADY_REQUESTED": func(s *gh.TeamCodeReviewSettings) string {
				return ToString(s.CountMembersAlreadyRequested)
			},
			"REMOVE_TEAM_REQUEST": func(s *gh.TeamCodeReviewSettings) string {
				return ToString(s.RemoveTeamRequest)
			},
		},
	}
}

func (u *teamCodeReviewFiledGetters) GetField(s *gh.TeamCodeReviewSettings, field string) string {
	if getter, ok := u.Func[field]; ok {
		return getter(s)
	}
	return ""
}

func (r *Renderer) RenderTeamCodeReviewSettings(codeReviewSettings *gh.TeamCodeReviewSettings, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(codeReviewSettings)
		return
	}

	getter := NewTeamCodeReviewFieldGetters()
	table := r.newTableWriter([]string{"FIELD", "VALUE"})

	for _, header := range headers {
		value := getter.GetField(codeReviewSettings, header)
		table.Append([]string{header, value})
	}

	table.Render()
}

func (r *Renderer) RenderTeamCodeReviewSettingsDefault(codeReviewSettings *gh.TeamCodeReviewSettings) {
	headers := []string{"TEAM_SLUG", "AUTO_ASSIGNMENT", "MEMBER_COUNT", "ALGORITHM", "NOTIFY_TEAM"}
	r.RenderTeamCodeReviewSettings(codeReviewSettings, headers)
}
