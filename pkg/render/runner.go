package render

import (
	"strings"

	"github.com/google/go-github/v71/github"
)

type runnerFieldGetter func(runner *github.Runner) string
type runnerFieldGetters struct {
	Func map[string]runnerFieldGetter
}

// NewRunnerFieldGetters returns field getter functions for github.Runner
func NewRunnerFieldGetters() *runnerFieldGetters {
	return &runnerFieldGetters{
		Func: map[string]runnerFieldGetter{
			"ID": func(r *github.Runner) string {
				return ToString(r.ID)
			},
			"NAME": func(r *github.Runner) string {
				return ToString(r.Name)
			},
			"OS": func(r *github.Runner) string {
				return ToString(r.OS)
			},
			"STATUS": func(r *github.Runner) string {
				return ToString(r.Status)
			},
			"BUSY": func(r *github.Runner) string {
				return ToString(r.Busy)
			},
			"LABELS": func(r *github.Runner) string {
				if r.Labels == nil {
					return ""
				}
				labels := make([]string, len(r.Labels))
				for i, l := range r.Labels {
					labels[i] = ToString(l.Name)
				}
				return strings.Join(labels, ", ")
			},
		},
	}
}

func (g *runnerFieldGetters) GetField(runner *github.Runner, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(runner)
	}
	return ""
}

// RenderRunners renders a table of runners with the specified headers
func (r *Renderer) RenderRunners(runners []*github.Runner, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(runners)
		return
	}

	if len(runners) == 0 {
		r.writeLine("no runners found")
		return
	}

	getter := NewRunnerFieldGetters()
	table := r.newTableWriter(headers)

	for _, runner := range runners {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(runner, header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderRunnersDefault renders runners with default columns
func (r *Renderer) RenderRunnersDefault(runners []*github.Runner) {
	headers := []string{"ID", "NAME", "OS", "STATUS", "LABELS"}
	r.RenderRunners(runners, headers)
}
