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
var usageCountTableHeader = []string{"KEY", "SESSIONS", "REQUESTS", "PREMIUM_REQUESTS", "AIU", "INPUT", "CACHE_READ", "CACHE_WRITE", "OUTPUT", "API_DURATION_MS"}

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
			strconv.Itoa(c.Requests),
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
	if stats.UsageIncompleteSessions > 0 {
		r.writeLine("incomplete_usage_sessions: " + strconv.Itoa(stats.UsageIncompleteSessions) +
			" (model, request, or token breakdown may be incomplete)")
	}

	if !stats.AskUser.IsEmpty() {
		q := stats.AskUser
		r.writeLine("")
		r.writeLine("ASK_USER")
		r.writeLine("sessions: " + strconv.Itoa(q.Sessions) + ", questions: " + strconv.Itoa(q.Questions) +
			", answered: " + strconv.Itoa(q.Answered) + ", cancelled: " + strconv.Itoa(q.Cancelled) +
			", failed: " + strconv.Itoa(q.Failed) + ", unresolved: " + strconv.Itoa(q.Unresolved) +
			", unknown: " + strconv.Itoa(q.Unknown))
		r.renderQuestionWait(q.TimedAnswers, q.UntimedAnswers, q.TotalWaitMs, q.AvgWaitMs, q.MaxWaitMs)
	}
	if w := stats.PermissionWait; w != nil && stats.Requests > 0 {
		r.writeLine("")
		r.writeLine("PERMISSION_WAIT")
		r.writeLine("unresolved: " + strconv.Itoa(w.Unresolved))
		for _, entry := range []struct {
			name string
			wait copilotext.PermissionDecisionWait
		}{
			{"human_response", w.HumanResponses},
			{"non_human", w.NonHumanDecisions},
			{"unknown_source", w.UnknownSourceDecisions},
		} {
			t := entry.wait
			r.writeLine(entry.name + ": completed: " + strconv.Itoa(t.Completed) +
				", timed: " + strconv.Itoa(t.Timed) + ", untimed: " + strconv.Itoa(t.Untimed) +
				", total_wait_ms: " + strconv.FormatFloat(t.TotalWaitMs, 'f', 3, 64) +
				", avg_wait_ms: " + optionalMilliseconds(t.AvgWaitMs) +
				", max_wait_ms: " + optionalMilliseconds(t.MaxWaitMs))
		}
		r.writeLine("request-to-decision elapsed time includes UI/transport overhead; overlapping intervals are summed, not unique idle time")
		r.writeLine("only human_response is confirmed user wait; other and unknown-source decisions remain separate; unresolved requests have no duration")
	}

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
	if err := r.renderUsageCountTable("CWD_USAGE", stats.ByCWDUsage); err != nil {
		return err
	}
	if err := r.renderUsageCountTable("MODEL_USAGE", stats.ByModelUsage); err != nil {
		return err
	}
	if len(stats.ByRoleUsage) == 0 {
		return nil
	}
	r.writeLine("")
	r.writeLine("ROLE_USAGE")
	table := r.newTableWriter([]string{"ROLE", "SESSIONS", "REQUESTS", "PREMIUM_REQUESTS", "AIU", "INPUT", "CACHE_READ", "CACHE_WRITE", "OUTPUT"})
	for _, c := range stats.ByRoleUsage {
		table.Append([]string{
			c.Key, strconv.Itoa(c.Sessions), strconv.Itoa(c.Requests),
			strconv.FormatFloat(c.PremiumRequests, 'f', -1, 64), strconv.FormatFloat(c.AIU, 'f', 3, 64),
			strconv.FormatInt(c.InputTokens, 10), strconv.FormatInt(c.CacheReadTokens, 10),
			strconv.FormatInt(c.CacheWriteTokens, 10), strconv.FormatInt(c.OutputTokens, 10),
		})
	}
	return table.Render()
}
