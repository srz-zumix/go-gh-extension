package render

import (
	"strconv"

	"github.com/srz-zumix/go-gh-extension/pkg/copilotext"
)

// countTableHeader lists the columns used for every axis table in
// RenderCopilotPermissionStats.
var countTableHeader = []string{"KEY", "TOTAL", "APPROVED", "APPROVED_FOR_LOCATION", "DENIED", "UNRESOLVED"}

// usageCountTableHeader lists the columns used for the usage-by-CWD table in
// RenderCopilotPermissionStats.
var usageCountTableHeader = []string{"KEY", "SESSIONS", "PREMIUM_REQUESTS", "AIU", "INPUT", "CACHE_READ", "CACHE_WRITE", "OUTPUT", "API_DURATION_MS"}

// renderCountTable renders one axis (e.g. ByCommand) as a table titled by section.
func (r *Renderer) renderCountTable(section string, counts []copilotext.Count) error {
	if len(counts) == 0 {
		return nil
	}
	r.writeLine("")
	r.writeLine(section)
	table := r.newTableWriter(countTableHeader)
	for _, c := range counts {
		table.Append([]string{
			c.Key,
			strconv.Itoa(c.Total),
			strconv.Itoa(c.Approved),
			strconv.Itoa(c.ApprovedForLocation),
			strconv.Itoa(c.Denied),
			strconv.Itoa(c.Unresolved),
		})
	}
	return table.Render()
}

// renderUsageCountTable renders stats.ByCWDUsage as a table titled by section.
func (r *Renderer) renderUsageCountTable(section string, counts []copilotext.UsageCount) error {
	if len(counts) == 0 {
		return nil
	}
	r.writeLine("")
	r.writeLine(section)
	table := r.newTableWriter(usageCountTableHeader)
	for _, c := range counts {
		table.Append([]string{
			c.Key,
			strconv.Itoa(c.Sessions),
			strconv.FormatFloat(c.PremiumRequests, 'f', -1, 64),
			strconv.FormatFloat(c.AIU, 'f', 3, 64),
			strconv.FormatInt(c.InputTokens, 10),
			strconv.FormatInt(c.CacheReadTokens, 10),
			strconv.FormatInt(c.CacheWriteTokens, 10),
			strconv.FormatInt(c.OutputTokens, 10),
			strconv.FormatFloat(c.APIDurationMs, 'f', 0, 64),
		})
	}
	return table.Render()
}

// RenderCopilotPermissionStats renders a copilotext.PermissionStats summary. With an
// exporter configured (e.g. --format json) it writes the full struct as exported data;
// otherwise it prints a summary line followed by one table per non-empty axis.
func (r *Renderer) RenderCopilotPermissionStats(stats *copilotext.PermissionStats) error {
	if r.exporter != nil {
		return r.RenderExportedData(stats)
	}
	if stats == nil {
		return nil
	}

	r.writeLine("SUMMARY")
	r.writeLine("sessions: " + strconv.Itoa(stats.Sessions) + ", requests: " + strconv.Itoa(stats.Requests))
	r.writeLine("usage_sessions: " + strconv.Itoa(stats.UsageSessions) +
		", premium_requests: " + strconv.FormatFloat(stats.UsagePremiumRequests, 'f', -1, 64) +
		", aiu: " + strconv.FormatFloat(stats.UsageAIU, 'f', 3, 64))

	sections := []struct {
		name   string
		counts []copilotext.Count
	}{
		{"RESULT", stats.ByResult},
		{"DECISION_SOURCE", stats.ByDecisionSource},
		{"KIND", stats.ByKind},
		{"READONLY", stats.ByReadOnly},
		{"COMMAND", stats.ByCommand},
		{"PATH", stats.ByPath},
		{"URL", stats.ByURL},
		{"CWD", stats.ByCWD},
	}
	for _, section := range sections {
		if err := r.renderCountTable(section.name, section.counts); err != nil {
			return err
		}
	}
	return r.renderUsageCountTable("CWD_USAGE", stats.ByCWDUsage)
}

