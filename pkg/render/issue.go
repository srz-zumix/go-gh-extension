package render

import (
	"strings"

	"github.com/fatih/color"
	"github.com/google/go-github/v79/github"
)

type issueFieldGetter func(issue *github.Issue) string
type issueFieldGetters struct {
	Func map[string]issueFieldGetter
}

// NewIssueFieldGetters returns field getter functions for github.Issue
func NewIssueFieldGetters(enableColor bool) *issueFieldGetters {
	return &issueFieldGetters{
		Func: map[string]issueFieldGetter{
			"NUMBER": func(issue *github.Issue) string {
				return ToString(issue.Number)
			},
			"TITLE": func(issue *github.Issue) string {
				return ToString(issue.Title)
			},
			"AUTHOR": func(issue *github.Issue) string {
				if issue.User != nil {
					return ToString(issue.User.Login)
				}
				return ""
			},
			"STATE": func(issue *github.Issue) string {
				return ToString(issue.State)
			},
			"LABELS": func(issue *github.Issue) string {
				var labelNames []string
				for _, label := range issue.Labels {
					labelName := "'" + ToString(label.Name) + "'"
					if enableColor && label.Color != nil && *label.Color != "" {
						r, g, b, err := ToRGB(*label.Color)
						if err == nil {
							labelName = color.RGB(r, g, b).Sprint(labelName)
						}
					}
					labelNames = append(labelNames, labelName)
				}
				return strings.Join(labelNames, ", ")
			},
			"CREATED_AT": func(issue *github.Issue) string {
				return ToString(issue.CreatedAt)
			},
			"UPDATED_AT": func(issue *github.Issue) string {
				return ToString(issue.UpdatedAt)
			},
			"CLOSED_AT": func(issue *github.Issue) string {
				return ToString(issue.ClosedAt)
			},
			"URL": func(issue *github.Issue) string {
				return ToString(issue.HTMLURL)
			},
			"ASSIGNEES": func(issue *github.Issue) string {
				var assigneeLogins []string
				for _, assignee := range issue.Assignees {
					if assignee.Login != nil {
						assigneeLogins = append(assigneeLogins, *assignee.Login)
					}
				}
				return strings.Join(assigneeLogins, ", ")
			},
			"MILESTONE": func(issue *github.Issue) string {
				if issue.Milestone != nil {
					return ToString(issue.Milestone.Title)
				}
				return ""
			},
		},
	}
}

// GetField returns the string value for the given field
func (g *issueFieldGetters) GetField(issue *github.Issue, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(issue)
	}
	return ""
}

// RenderIssues renders a table of issues with the specified headers
func (r *Renderer) RenderIssues(issues []*github.Issue, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(issues)
		return
	}

	if len(issues) == 0 {
		r.writeLine("No issues.")
		return
	}

	getter := NewIssueFieldGetters(r.Color)
	table := r.newTableWriter(headers)
	table.SetAutoWrapText(false)

	for _, issue := range issues {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(issue, header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderIssuesDefault renders issues with default columns
func (r *Renderer) RenderIssuesDefault(issues []*github.Issue) {
	headers := []string{"NUMBER", "TITLE", "AUTHOR", "STATE", "LABELS"}
	r.RenderIssues(issues, headers)
}
