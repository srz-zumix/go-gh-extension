package render

import (
	"github.com/fatih/color"
	"github.com/google/go-github/v73/github"
)

type labelFieldGetter func(label *github.Label) string
type labelFieldGetters struct {
	Func map[string]labelFieldGetter
}

// NewLabelFieldGetters returns field getter functions for github.Label
func NewLabelFieldGetters() *labelFieldGetters {
	return &labelFieldGetters{
		Func: map[string]labelFieldGetter{
			"NAME": func(label *github.Label) string {
				if label.Name == nil {
					return ""
				}
				return *label.Name
			},
			"COLOR": func(label *github.Label) string {
				if label.Color == nil {
					return ""
				}
				return *label.Color
			},
			"DESCRIPTION": func(label *github.Label) string {
				if label.Description == nil {
					return ""
				}
				return *label.Description
			},
			"DEFAULT": func(label *github.Label) string {
				if label.Default == nil {
					return ""
				}
				if *label.Default {
					return "YES"
				}
				return "NO"
			},
			"URL": func(label *github.Label) string {
				if label.URL == nil {
					return ""
				}
				return *label.URL
			},
		},
	}
}

// GetField returns the string value for the given field
func (g *labelFieldGetters) GetField(label *github.Label, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(label)
	}
	return ""
}

// RenderLabels renders a table of labels with the specified headers
func (r *Renderer) RenderLabels(labels []*github.Label, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(labels)
		return
	}

	if len(labels) == 0 {
		r.writeLine("No labels.")
		return
	}

	getter := NewLabelFieldGetters()
	table := r.newTableWriter(headers)

	for _, label := range labels {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(label, header)
			if r.Color && header == "COLOR" && label.Color != nil {
				c := *label.Color
				r, g, b, err := ToRGB(c)
				if err != nil {
					break
				}
				row[i] = color.RGB(r, g, b).Sprint(row[i])
			}
		}
		table.Append(row)
	}
	table.Render()
}

// RenderLabelsDefault renders labels with default columns
func (r *Renderer) RenderLabelsDefault(labels []*github.Label) {
	headers := []string{"NAME", "COLOR", "DESCRIPTION", "DEFAULT"}
	r.RenderLabels(labels, headers)
}
