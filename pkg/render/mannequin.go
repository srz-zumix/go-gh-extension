package render

import (
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

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
				if string(m.Claimant.Login) == "" {
					return ""
				}
				return ToString(m.Claimant.Login)
			},
		},
	}
}

func (g *mannequinFieldGetters) GetField(m *client.Mannequin, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(m)
	}
	return ""
}

// RenderMannequins renders a list of mannequins as a table with the given headers.
func (r *Renderer) RenderMannequins(mannequins []client.Mannequin, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(mannequins)
		return
	}
	if len(mannequins) == 0 {
		r.writeLine("No mannequins.")
		return
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
	table.Render()
}

// RenderMannequinsDefault renders mannequins with default columns.
func (r *Renderer) RenderMannequinsDefault(mannequins []client.Mannequin) {
	r.RenderMannequins(mannequins, nil)
}
