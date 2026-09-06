package render

import (
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"
)

// RunnerGroupFieldGetter renders a single field of a runner group
type RunnerGroupFieldGetter func(group *github.RunnerGroup) string

type RunnerGroupFieldGetters struct {
	Func map[string]RunnerGroupFieldGetter
}

// NewRunnerGroupFieldGetters returns field getter functions for github.RunnerGroup
func NewRunnerGroupFieldGetters() *RunnerGroupFieldGetters {
	return &RunnerGroupFieldGetters{
		Func: map[string]RunnerGroupFieldGetter{
			"ID": func(g *github.RunnerGroup) string {
				return ToString(g.ID)
			},
			"NAME": func(g *github.RunnerGroup) string {
				return ToString(g.Name)
			},
			"VISIBILITY": func(g *github.RunnerGroup) string {
				return ToString(g.Visibility)
			},
			"DEFAULT": func(g *github.RunnerGroup) string {
				return ToString(g.Default)
			},
			"INHERITED": func(g *github.RunnerGroup) string {
				return ToString(g.Inherited)
			},
			"PUBLIC_REPOSITORIES": func(g *github.RunnerGroup) string {
				return ToString(g.AllowsPublicRepositories)
			},
			"RESTRICTED_TO_WORKFLOWS": func(g *github.RunnerGroup) string {
				return ToString(g.RestrictedToWorkflows)
			},
		},
	}
}

func (g *RunnerGroupFieldGetters) GetField(group *github.RunnerGroup, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(group)
	}
	return ""
}

// Fields returns the sorted field names handled by g
func (g *RunnerGroupFieldGetters) Fields() []string {
	fields := make([]string, 0, len(g.Func))
	for field := range g.Func {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields
}

// RunnerGroupFields returns the sorted field names accepted as RenderRunnerGroups headers
func RunnerGroupFields() []string {
	return NewRunnerGroupFieldGetters().Fields()
}

// RenderRunnerGroups renders a table of runner groups with the specified headers
func (r *Renderer) RenderRunnerGroups(groups []*github.RunnerGroup, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(groups)
	}

	if len(groups) == 0 {
		r.writeLine("No runner groups.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"ID", "NAME", "VISIBILITY", "DEFAULT", "INHERITED"}
	}

	getter := NewRunnerGroupFieldGetters()
	table := r.newTableWriter(headers)

	for _, group := range groups {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(group, header)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderRunnerGroup renders a single runner group as a field/value table.
// fields specifies which fields to display; if empty, all fields are shown.
func (r *Renderer) RenderRunnerGroup(group *github.RunnerGroup, fields []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(group)
	}

	if group == nil {
		return nil
	}

	if len(fields) == 0 {
		fields = RunnerGroupFields()
	}

	getter := NewRunnerGroupFieldGetters()
	table := r.newTableWriter([]string{"FIELD", "VALUE"})
	for _, field := range fields {
		table.Append([]string{strings.ToUpper(field), getter.GetField(group, field)})
	}
	return table.Render()
}
