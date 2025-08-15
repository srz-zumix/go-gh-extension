package render

import (
	"github.com/google/go-github/v73/github"
)

type copilotMetricsFieldGetter func(m *github.CopilotMetrics) string
type copilotMetricsFieldGetters struct {
	Func map[string]copilotMetricsFieldGetter
}

// NewCopilotMetricsFieldGetters returns field getter functions for github.CopilotMetrics
func NewCopilotMetricsFieldGetters() *copilotMetricsFieldGetters {
	return &copilotMetricsFieldGetters{
		Func: map[string]copilotMetricsFieldGetter{
			"DATE": func(m *github.CopilotMetrics) string {
				return ToString(m.Date)
			},
			"TOTAL_ACTIVE_USERS": func(m *github.CopilotMetrics) string {
				return ToString(m.TotalActiveUsers)
			},
			"TOTAL_ENGAGED_USERS": func(m *github.CopilotMetrics) string {
				return ToString(m.TotalEngagedUsers)
			},
		},
	}
}

func (g *copilotMetricsFieldGetters) GetField(m *github.CopilotMetrics, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(m)
	}
	return ""
}

// RenderCopilotMetrics renders a table of Copilot metrics with the specified headers
func (r *Renderer) RenderCopilotMetrics(metrics []*github.CopilotMetrics, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(metrics)
		return
	}

	if len(metrics) == 0 {
		r.writeLine("no copilot metrics found")
		return
	}

	getter := NewCopilotMetricsFieldGetters()
	table := r.newTableWriter(headers)

	for _, m := range metrics {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(m, header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderCopilotMetricsDefault renders Copilot metrics with default columns
func (r *Renderer) RenderCopilotMetricsDefault(metrics []*github.CopilotMetrics) {
	headers := []string{"DATE", "TOTAL_ACTIVE_USERS", "TOTAL_ENGAGED_USERS"}
	r.RenderCopilotMetrics(metrics, headers)
}
