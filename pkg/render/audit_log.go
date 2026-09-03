package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v90/github"
)

type auditEntryFieldGetter func(entry *github.AuditEntry) string
type auditEntryFieldGetters struct {
	Func map[string]auditEntryFieldGetter
}

func newAuditEntryFieldGetters() *auditEntryFieldGetters {
	return &auditEntryFieldGetters{
		Func: map[string]auditEntryFieldGetter{
			"ACTION": func(entry *github.AuditEntry) string {
				return ToString(entry.Action)
			},
			"ACTOR": func(entry *github.AuditEntry) string {
				return ToString(entry.Actor)
			},
			"BUSINESS": func(entry *github.AuditEntry) string {
				return ToString(entry.Business)
			},
			"CREATED_AT": func(entry *github.AuditEntry) string {
				return ToString(entry.CreatedAt)
			},
			"DOCUMENT_ID": func(entry *github.AuditEntry) string {
				return ToString(entry.DocumentID)
			},
			"ORG": func(entry *github.AuditEntry) string {
				return ToString(entry.Org)
			},
			"TIMESTAMP": func(entry *github.AuditEntry) string {
				return ToString(entry.Timestamp)
			},
			"TOKEN_SCOPES": func(entry *github.AuditEntry) string {
				return ToString(entry.TokenScopes)
			},
			"USER": func(entry *github.AuditEntry) string {
				return ToString(entry.User)
			},
		},
	}
}

// getField returns the value of the requested column. Headers that are not
// backed by a struct field fall back to the entry's additional fields, so
// action specific fields such as "KEY" or "ENVIRONMENT_NAME" can be rendered.
func (g *auditEntryFieldGetters) getField(entry *github.AuditEntry, field string) string {
	upper := strings.ToUpper(field)
	if getter, ok := g.Func[upper]; ok {
		return getter(entry)
	}
	return AuditEntryAdditionalField(entry, strings.ToLower(field))
}

// AuditEntryAdditionalField returns the string representation of an audit log
// entry field that is not part of the github.AuditEntry struct. It looks the
// key up in AdditionalFields first, then in Data.
func AuditEntryAdditionalField(entry *github.AuditEntry, key string) string {
	if entry == nil {
		return ""
	}
	for _, fields := range []map[string]any{entry.AdditionalFields, entry.Data} {
		value, ok := fields[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case string:
			return v
		case bool:
			return strconv.FormatBool(v)
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// RenderAuditEntries renders a table of audit log entries with the specified headers.
// Headers that do not match a github.AuditEntry field are resolved from the
// entry's additional fields using the lower cased header name.
func (r *Renderer) RenderAuditEntries(entries []*github.AuditEntry, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(entries)
	}

	if len(entries) == 0 {
		r.writeLine("No audit log entries found.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"TIMESTAMP", "ACTION", "ACTOR", "REPO"}
	}

	getter := newAuditEntryFieldGetters()
	table := r.newTableWriter(headers)

	for _, entry := range entries {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.getField(entry, header)
		}
		table.Append(row)
	}

	return table.Render()
}
