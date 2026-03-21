package render

import (
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// MannequinFieldList is the list of valid field names for mannequin display.
var MannequinFieldList = []string{"ID", "LOGIN", "EMAIL", "URL", "CREATED_AT", "CLAIMANT"}

type mannequinFieldGetter func(m *client.Mannequin) string
type mannequinFieldGetters struct {
	Func map[string]mannequinFieldGetter
}

// NewMannequinFieldGetters creates a new mannequinFieldGetters with default field mappings.
func NewMannequinFieldGetters() *mannequinFieldGetters {
	return &mannequinFieldGetters{
		Func: map[string]mannequinFieldGetter{
			"ID": func(m *client.Mannequin) string {
				return ToString(m.ID)
			},
			"LOGIN": func(m *client.Mannequin) string {
				return ToString(m.Login)
			},
			"EMAIL": func(m *client.Mannequin) string {
				return ToString(m.Email)
			},
			"URL": func(m *client.Mannequin) string {
				return ToString(m.URL)
			},
			"CREATED_AT": func(m *client.Mannequin) string {
				return ToString(m.CreatedAt)
			},
			"CLAIMANT": func(m *client.Mannequin) string {
				return ToString(m.Claimant.Login)
			},
		},
	}
}

func (g *mannequinFieldGetters) GetField(m *client.Mannequin, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(m)
	}
	return ""
}

// RenderMannequins renders a list of mannequins as a table with the given headers.
func (r *Renderer) RenderMannequins(mannequins []client.Mannequin, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(mannequins)
	}
	if len(mannequins) == 0 {
		r.writeLine("No mannequins.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"LOGIN", "EMAIL", "CLAIMANT"}
	}

	getter := NewMannequinFieldGetters()
	table := r.newTableWriter(headers)
	for _, m := range mannequins {
		m := m
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(&m, header)
		}
		table.Append(row)
	}
	return table.Render()
}
