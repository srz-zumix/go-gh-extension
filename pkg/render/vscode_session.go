package render

import (
	"strconv"

	"github.com/srz-zumix/go-gh-extension/pkg/copilotext"
)

var vscodeToolTableHeader = []string{"KEY", "TOTAL", "OK", "ERROR", "TOTAL_MS", "AVG_MS"}
var vscodeModelTableHeader = []string{"KEY", "REQUESTS", "INPUT", "OUTPUT", "CACHED", "AVG_TTFT_MS", "USAGE_AIU"}
var vscodeAgentTableHeader = []string{"KEY", "TOTAL"}
var vscodeWorkspaceTableHeader = []string{"KEY", "TOOL_CALLS", "LLM_REQUESTS", "TURNS"}

func (r *Renderer) renderVSCodeToolTable(counts []copilotext.VSCodeToolCount) error {
	if len(counts) == 0 {
		return nil
	}
	r.writeLine("")
	r.writeLine("TOOL")
	table := r.newTableWriter(vscodeToolTableHeader)
	for _, c := range counts {
		table.Append([]string{
			c.Key,
			strconv.Itoa(c.Total),
			strconv.Itoa(c.OK),
			strconv.Itoa(c.Error),
			strconv.FormatFloat(c.TotalMs, 'f', 0, 64),
			strconv.FormatFloat(c.AvgMs, 'f', 1, 64),
		})
	}
	return table.Render()
}

func (r *Renderer) renderVSCodeModelTable(counts []copilotext.VSCodeModelCount) error {
	if len(counts) == 0 {
		return nil
	}
	r.writeLine("")
	r.writeLine("MODEL")
	table := r.newTableWriter(vscodeModelTableHeader)
	for _, c := range counts {
		table.Append([]string{
			c.Key,
			strconv.Itoa(c.Requests),
			strconv.FormatInt(c.InputTokens, 10),
			strconv.FormatInt(c.OutputTokens, 10),
			strconv.FormatInt(c.CachedTokens, 10),
			strconv.FormatFloat(c.AvgTTFTMs, 'f', 1, 64),
			strconv.FormatFloat(c.UsageAIU, 'f', 3, 64),
		})
	}
	return table.Render()
}

func (r *Renderer) renderVSCodeAgentTable(counts []copilotext.VSCodeAgentCount) error {
	if len(counts) == 0 {
		return nil
	}
	r.writeLine("")
	r.writeLine("AGENT")
	table := r.newTableWriter(vscodeAgentTableHeader)
	for _, c := range counts {
		table.Append([]string{c.Key, strconv.Itoa(c.Total)})
	}
	return table.Render()
}

func (r *Renderer) renderVSCodeWorkspaceTable(counts []copilotext.VSCodeWorkspaceCount) error {
	if len(counts) == 0 {
		return nil
	}
	r.writeLine("")
	r.writeLine("WORKSPACE")
	table := r.newTableWriter(vscodeWorkspaceTableHeader)
	for _, c := range counts {
		table.Append([]string{
			c.Key,
			strconv.Itoa(c.ToolCalls),
			strconv.Itoa(c.LLMRequests),
			strconv.Itoa(c.Turns),
		})
	}
	return table.Render()
}

// RenderVSCodeStats renders a copilotext.VSCodeStats summary. With an exporter configured
// (e.g. --format json) it writes the full struct as exported data; otherwise it prints a
// summary line followed by one table per non-empty axis.
func (r *Renderer) RenderVSCodeStats(stats *copilotext.VSCodeStats) error {
	if r.exporter != nil {
		return r.RenderExportedData(stats)
	}
	if stats == nil {
		return nil
	}

	r.writeLine("SUMMARY")
	r.writeLine("sessions: " + strconv.Itoa(stats.Sessions) +
		", turns: " + strconv.Itoa(stats.Turns) +
		", llm_requests: " + strconv.Itoa(stats.LLMRequests) +
		", tool_calls: " + strconv.Itoa(stats.ToolCalls))

	if err := r.renderVSCodeToolTable(stats.ByTool); err != nil {
		return err
	}
	if err := r.renderVSCodeModelTable(stats.ByModel); err != nil {
		return err
	}
	if err := r.renderVSCodeAgentTable(stats.ByAgent); err != nil {
		return err
	}
	if err := r.renderVSCodeWorkspaceTable(stats.ByWorkspace); err != nil {
		return err
	}
	return nil
}
