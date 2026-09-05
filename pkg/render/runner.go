package render

import (
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"
)

// RunnerFieldGetter renders a single field of a runner
type RunnerFieldGetter func(runner *github.Runner) string

type RunnerFieldGetters struct {
	Func map[string]RunnerFieldGetter
}

// NewRunnerFieldGetters returns field getter functions for github.Runner
func NewRunnerFieldGetters() *RunnerFieldGetters {
	return &RunnerFieldGetters{
		Func: map[string]RunnerFieldGetter{
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
			"GROUP": func(r *github.Runner) string {
				return ToString(r.RunnerGroupID)
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

func (g *RunnerFieldGetters) GetField(runner *github.Runner, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(runner)
	}
	return ""
}

// Fields returns the sorted field names handled by g
func (g *RunnerFieldGetters) Fields() []string {
	fields := make([]string, 0, len(g.Func))
	for field := range g.Func {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields
}

// RunnerFields returns the sorted field names accepted as RenderRunners headers
func RunnerFields() []string {
	return NewRunnerFieldGetters().Fields()
}

// RenderRunners renders a table of runners with the specified headers
func (r *Renderer) RenderRunners(runners []*github.Runner, headers []string) error {
	return r.RenderRunnersWithFieldGetters(runners, headers, NewRunnerFieldGetters())
}

// RenderRunnersWithFieldGetters renders a table of runners with the specified headers,
// resolving each column through getter so callers can add their own fields
func (r *Renderer) RenderRunnersWithFieldGetters(runners []*github.Runner, headers []string, getter *RunnerFieldGetters) error {
	if r.exporter != nil {
		return r.RenderExportedData(runners)
	}

	if len(runners) == 0 {
		r.writeLine("No runners.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"ID", "NAME", "OS", "STATUS", "LABELS"}
	}

	if getter == nil {
		getter = NewRunnerFieldGetters()
	}

	table := r.newTableWriter(headers)

	for _, runner := range runners {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(runner, header)
		}
		table.Append(row)
	}
	return table.Render()
}
