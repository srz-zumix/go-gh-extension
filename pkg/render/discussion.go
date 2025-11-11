package render

import (
	"strings"

	"github.com/fatih/color"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

type discussionFieldGetter func(discussion *client.Discussion) string
type discussionFieldGetters struct {
	Func map[string]discussionFieldGetter
}

// NewDiscussionFieldGetters returns field getter functions for client.Discussion
func NewDiscussionFieldGetters(enableColor bool) *discussionFieldGetters {
	return &discussionFieldGetters{
		Func: map[string]discussionFieldGetter{
			"NUMBER": func(discussion *client.Discussion) string {
				return ToString(discussion.Number)
			},
			"TITLE": func(discussion *client.Discussion) string {
				return ToString(discussion.Title)
			},
			"AUTHOR": func(discussion *client.Discussion) string {
				return ToString(discussion.Author.Login)
			},
			"CATEGORY": func(discussion *client.Discussion) string {
				return ToString(discussion.Category.Name)
			},
			"CATEGORY_SLUG": func(discussion *client.Discussion) string {
				return ToString(discussion.Category.Slug)
			},
			"LABELS": func(discussion *client.Discussion) string {
				var labelNames []string
				for _, label := range discussion.Labels.Nodes {
					labelName := "'" + ToString(label.Name) + "'"
					if enableColor && label.Color != "" {
						r, g, b, err := ToRGB((string)(label.Color))
						if err == nil {
							labelName = color.RGB(r, g, b).Sprint(labelName)
						}
					}
					labelNames = append(labelNames, labelName)
				}
				return strings.Join(labelNames, ", ")
			},
			"URL": func(discussion *client.Discussion) string {
				return ToString(discussion.URL)
			},
			"CREATED_AT": func(discussion *client.Discussion) string {
				return ToString(discussion.CreatedAt)
			},
			"UPDATED_AT": func(discussion *client.Discussion) string {
				return ToString(discussion.UpdatedAt)
			},
			"CLOSED_AT": func(discussion *client.Discussion) string {
				return ToString(discussion.ClosedAt)
			},
			"LOCKED": func(discussion *client.Discussion) string {
				return ToString(discussion.Locked)
			},
			"STATE_REASON": func(discussion *client.Discussion) string {
				return ToString(*discussion.StateReason)
			},
		},
	}
}

// GetField returns the string value for the given field
func (g *discussionFieldGetters) GetField(discussion *client.Discussion, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(discussion)
	}
	return ""
}

// RenderDiscussions renders a table of discussions with the specified headers
func (r *Renderer) RenderDiscussions(discussions []client.Discussion, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(discussions)
		return
	}

	getter := NewDiscussionFieldGetters(r.Color)
	table := r.newTableWriter(headers)
	table.SetAutoWrapText(false)

	for _, disc := range discussions {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(&disc, header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderDiscussionsDefault renders discussions with default columns
func (r *Renderer) RenderDiscussionsDefault(discussions []client.Discussion) {
	headers := []string{"NUMBER", "TITLE", "AUTHOR", "CATEGORY", "LABELS"}
	r.RenderDiscussions(discussions, headers)
}
