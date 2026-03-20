package render

import (
	"strings"

	"github.com/google/go-github/v79/github"
)

type deployKeyFieldGetter func(key *github.Key) string
type deployKeyFieldGetters struct {
	Func map[string]deployKeyFieldGetter
}

func newDeployKeyFieldGetters() *deployKeyFieldGetters {
	return &deployKeyFieldGetters{
		Func: map[string]deployKeyFieldGetter{
			"ID": func(key *github.Key) string {
				return ToString(key.ID)
			},
			"TITLE": func(key *github.Key) string {
				return ToString(key.Title)
			},
			"CREATED_AT": func(key *github.Key) string {
				return ToString(key.CreatedAt)
			},
			"READ_ONLY": func(key *github.Key) string {
				return ToString(key.ReadOnly)
			},
			"VERIFIED": func(key *github.Key) string {
				return ToString(key.Verified)
			},
			"ADDED_BY": func(key *github.Key) string {
				return ToString(key.AddedBy)
			},
			"LAST_USED": func(key *github.Key) string {
				return ToString(key.LastUsed)
			},
		},
	}
}

func (g *deployKeyFieldGetters) getField(key *github.Key, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(key)
	}
	return ""
}

// RenderDeployKeys renders a table of deploy keys with the specified headers.
func (r *Renderer) RenderDeployKeys(keys []*github.Key, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(keys)
		return
	}

	if len(keys) == 0 {
		r.writeLine("No deploy keys found.")
		return
	}

	if len(headers) == 0 {
		headers = []string{"ID", "TITLE", "READ_ONLY", "CREATED_AT"}
	}

	getter := newDeployKeyFieldGetters()
	table := r.newTableWriter(headers)

	for _, key := range keys {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.getField(key, header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderDeployKey prints the details of a deploy key in a two-column FIELD/VALUE table.
// fields specifies which fields to display; if empty, all fields are shown.
func (r *Renderer) RenderDeployKey(key *github.Key, fields []string) {
	if r.exporter != nil {
		r.RenderExportedData(key)
		return
	}

	if len(fields) == 0 {
		fields = []string{"ID", "TITLE", "READ_ONLY", "VERIFIED", "ADDED_BY", "CREATED_AT", "LAST_USED", "URL", "KEY"}
	}

	getter := newDeployKeyFieldGetters()
	table := r.newTableWriter([]string{"FIELD", "VALUE"})
	for _, field := range fields {
		table.Append([]string{field, getter.getField(key, field)})
	}
	table.Render()
}
