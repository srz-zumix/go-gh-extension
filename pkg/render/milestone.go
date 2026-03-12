package render

import (
	"strings"

	"github.com/google/go-github/v79/github"
)

// MilestoneFieldGetter defines a function to get a field value from a github.Milestone.
type MilestoneFieldGetter func(m *github.Milestone) string

// MilestoneFieldGetters holds field getters for Milestone table rendering.
type MilestoneFieldGetters struct {
	Func map[string]MilestoneFieldGetter
}

// NewMilestoneFieldGetters creates field getters for Milestone table rendering.
func NewMilestoneFieldGetters() *MilestoneFieldGetters {
	return &MilestoneFieldGetters{
		Func: map[string]MilestoneFieldGetter{
			"NUMBER": func(m *github.Milestone) string {
				return ToString(m.GetNumber())
			},
			"TITLE": func(m *github.Milestone) string {
				return truncateString(m.GetTitle(), 50)
			},
			"STATE": func(m *github.Milestone) string {
				return m.GetState()
			},
			"OPEN_ISSUES": func(m *github.Milestone) string {
				return ToString(m.GetOpenIssues())
			},
			"CLOSED_ISSUES": func(m *github.Milestone) string {
				return ToString(m.GetClosedIssues())
			},
			"DESCRIPTION": func(m *github.Milestone) string {
				return m.GetDescription()
			},
			"DUE_ON": func(m *github.Milestone) string {
				return ToString(m.GetDueOn())
			},
			"CREATED_AT": func(m *github.Milestone) string {
				return ToString(m.GetCreatedAt())
			},
			"UPDATED_AT": func(m *github.Milestone) string {
				return ToString(m.GetUpdatedAt())
			},
			"URL": func(m *github.Milestone) string {
				return m.GetHTMLURL()
			},
			"CREATOR": func(m *github.Milestone) string {
				if m.Creator == nil {
					return ""
				}
				return m.Creator.GetLogin()
			},
		},
	}
}

// GetField returns the value of the specified field for the given milestone.
func (g *MilestoneFieldGetters) GetField(m *github.Milestone, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(m)
	}
	return ""
}

// RenderMilestones renders a table of milestones with the specified headers.
func (r *Renderer) RenderMilestones(milestones []*github.Milestone, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(milestones)
		return
	}

	if len(milestones) == 0 {
		r.WriteLine("No milestones found")
		return
	}

	if len(headers) == 0 {
		headers = []string{"NUMBER", "TITLE", "STATE", "OPEN_ISSUES", "CLOSED_ISSUES"}
	}

	getter := NewMilestoneFieldGetters()
	table := r.newTableWriter(headers)

	for _, milestone := range milestones {
		row := make([]string, len(headers))
		for j, header := range headers {
			row[j] = getter.GetField(milestone, header)
		}
		table.Append(row)
	}

	table.Render()
}
