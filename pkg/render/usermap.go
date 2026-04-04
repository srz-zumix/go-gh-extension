package render

import (
	"fmt"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/settings"
)

// UserMappingFieldList is the list of valid field names for user mapping display.
var UserMappingFieldList = []string{"SRC", "DST", "EMAIL"}

type userMappingFieldGetter func(m settings.UserMapping) string
type userMappingFieldGetters struct {
	Func map[string]userMappingFieldGetter
}

// NewUserMappingFieldGetters creates a new userMappingFieldGetters with default field mappings.
func NewUserMappingFieldGetters() *userMappingFieldGetters {
	return &userMappingFieldGetters{
		Func: map[string]userMappingFieldGetter{
			"SRC": func(m settings.UserMapping) string {
				return m.Src
			},
			"DST": func(m settings.UserMapping) string {
				return m.Dst
			},
			"EMAIL": func(m settings.UserMapping) string {
				return m.Email
			},
		},
	}
}

func (g *userMappingFieldGetters) GetField(m settings.UserMapping, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(m)
	}
	return ""
}

// RenderUserMappings renders user mappings as a table or via exporter (JSON/template).
func (r *Renderer) RenderUserMappings(mappings []settings.UserMapping, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(mappings)
	}

	if len(mappings) == 0 {
		r.writeLine("No user mappings found.")
		return nil
	}

	if len(headers) == 0 {
		headers = UserMappingFieldList
	}

	getter := NewUserMappingFieldGetters()
	table := r.newTableWriter(headers)
	for _, m := range mappings {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(m, header)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderUserMappingsYAML renders user mappings as YAML to the renderer's output stream.
func (r *Renderer) RenderUserMappingsYAML(mappings []settings.UserMapping) error {
	data, err := settings.Marshal(mappings)
	if err != nil {
		return fmt.Errorf("failed to marshal user mappings: %w", err)
	}
	r.writeLine(strings.TrimSuffix(string(data), "\n"))
	return nil
}
