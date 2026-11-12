package render

import (
	"strconv"

	"github.com/srz-zumix/go-gh-extension/pkg/copilotext"
)

var sandboxToolTableHeader = []string{"TOOL", "TOTAL", "APPLIED", "NOT_APPLIED", "UNKNOWN"}

// renderSandboxStats renders a SANDBOX summary line and a SANDBOX_TOOL table, or nothing
// when no execution was observed.
func (r *Renderer) renderSandboxStats(stats *copilotext.SandboxStats, withOptOut bool) error {
	if stats.IsEmpty() {
		return nil
	}
	r.writeLine("")
	r.writeLine("SANDBOX")
	line := "sessions: " + strconv.Itoa(stats.Sessions) +
		", executions: " + strconv.Itoa(stats.Executions) +
		", applied: " + strconv.Itoa(stats.Applied) +
		", not_applied: " + strconv.Itoa(stats.NotApplied) +
		", unknown: " + strconv.Itoa(stats.Unknown)
	if withOptOut {
		line += ", opt_out_requested: " + strconv.Itoa(stats.OptOutRequested)
	}
	r.writeLine(line)
	if len(stats.ByTool) == 0 {
		return nil
	}
	r.writeLine("")
	r.writeLine("SANDBOX_TOOL")
	table := r.newTableWriter(sandboxToolTableHeader)
	for _, c := range stats.ByTool {
		table.Append([]string{
			c.Key,
			strconv.Itoa(c.Total),
			strconv.Itoa(c.Applied),
			strconv.Itoa(c.NotApplied),
			strconv.Itoa(c.Unknown),
		})
	}
	return table.Render()
}
