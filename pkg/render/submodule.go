package render

import (
	"maps"
	"slices"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// SubmoduleFieldList is the list of default field names for submodule table rendering.
var SubmoduleFieldList = []string{"NAME", "PATH", "REPOSITORY", "BRANCH", "GIT_URL"}

// SubmoduleFieldGetter defines a function to get a field value from a client.RepositorySubmodule.
type SubmoduleFieldGetter func(sub *client.RepositorySubmodule) string

// SubmoduleFieldGetters holds field getters for RepositorySubmodule table rendering.
type SubmoduleFieldGetters struct {
	Func map[string]SubmoduleFieldGetter
}

// NewSubmoduleFieldGetters creates field getters for RepositorySubmodule table rendering.
func NewSubmoduleFieldGetters() *SubmoduleFieldGetters {
	return &SubmoduleFieldGetters{
		Func: map[string]SubmoduleFieldGetter{
			"NAME": func(sub *client.RepositorySubmodule) string {
				return sub.Name
			},
			"PATH": func(sub *client.RepositorySubmodule) string {
				return sub.Path
			},
			"REPOSITORY": func(sub *client.RepositorySubmodule) string {
				return parser.GetRepositoryFullName(sub.Repository)
			},
			"BRANCH": func(sub *client.RepositorySubmodule) string {
				return sub.Branch
			},
			"GIT_URL": func(sub *client.RepositorySubmodule) string {
				return sub.GitUrl
			},
		},
	}
}

// GetField returns the value of the specified field for the given submodule.
func (g *SubmoduleFieldGetters) GetField(sub *client.RepositorySubmodule, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(sub)
	}
	return ""
}

// Fields returns a sorted list of available field names.
func (g *SubmoduleFieldGetters) Fields() []string {
	return slices.Sorted(maps.Keys(g.Func))
}

// RenderSubmodules renders repository submodules as a table.
// headers specifies which fields to display; defaults to SubmoduleFieldList when empty.
func (r *Renderer) RenderSubmodules(submodules []client.RepositorySubmodule, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(submodules)
	}

	if len(submodules) == 0 {
		r.writeLine("No submodules.")
		return nil
	}

	if len(headers) == 0 {
		headers = SubmoduleFieldList
	}
	getter := NewSubmoduleFieldGetters()
	table := r.newTableWriter(headers)
	for i := range submodules {
		row := make([]string, len(headers))
		for j, header := range headers {
			row[j] = getter.GetField(&submodules[i], header)
		}
		table.Append(row)
	}
	return table.Render()
}
