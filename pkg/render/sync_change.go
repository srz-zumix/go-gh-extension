package render

import (
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

// SyncChangeFieldGetters holds field getters for planned synchronization changes.
type SyncChangeFieldGetters struct {
	Func map[string]func(gh.SyncChange) string
}

// NewSyncChangeFieldGetters creates field getters for planned synchronization changes.
func NewSyncChangeFieldGetters() *SyncChangeFieldGetters {
	return &SyncChangeFieldGetters{
		Func: map[string]func(gh.SyncChange) string{
			"DESTINATION": func(change gh.SyncChange) string { return change.Destination },
			"ACTION":      func(change gh.SyncChange) string { return string(change.Action) },
			"TARGET":      func(change gh.SyncChange) string { return change.Target },
			"PERMISSION":  func(change gh.SyncChange) string { return change.Permission },
			"CURRENT":     func(change gh.SyncChange) string { return change.Current },
		},
	}
}

// GetField returns the value of the specified field for a planned change.
func (g *SyncChangeFieldGetters) GetField(change gh.SyncChange, field string) string {
	if getter, ok := g.Func[strings.ToUpper(field)]; ok {
		return getter(change)
	}
	return ""
}

// RenderSyncChanges renders planned synchronization changes as a table or exported data.
func (r *Renderer) RenderSyncChanges(changes []gh.SyncChange, headers []string) error {
	if changes == nil {
		changes = []gh.SyncChange{}
	}
	if r.exporter != nil {
		return r.RenderExportedData(changes)
	}
	if len(changes) == 0 {
		r.writeLine("No changes.")
		return nil
	}
	if len(headers) == 0 {
		headers = []string{"DESTINATION", "ACTION", "TARGET", "PERMISSION", "CURRENT"}
	}
	getters := NewSyncChangeFieldGetters()
	table := r.newTableWriter(headers)
	for _, change := range changes {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getters.GetField(change, header)
		}
		table.Append(row)
	}
	return table.Render()
}
