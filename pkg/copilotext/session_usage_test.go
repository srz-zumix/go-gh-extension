package copilotext

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func usageRecordEvent(t *testing.T, id, model, initiator, source string, nano int64) string {
	t.Helper()
	data := map[string]any{
		"model": model, "initiator": initiator, "cost": 3, "duration": 10,
		"inputTokens": 150, "cacheReadTokens": 100, "cacheWriteTokens": 20, "outputTokens": 10,
		"accounting":   map[string]any{"sourceSessionId": source, "usageId": id},
		"copilotUsage": map[string]any{"totalNanoAiu": nano},
	}
	payload, err := json.Marshal(map[string]any{
		"type": "session.usage_record", "timestamp": "2026-01-01T01:00:00Z",
		"data": map[string]any{"usage": data},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(payload) + "\n"
}

func collectUsageLog(t *testing.T, events string) SessionUsage {
	t.Helper()
	dir := writeSessionDir(t, t.TempDir(), "session-a", "", events)
	u, found, err := readSessionUsage(Session{ID: "session-a", Dir: dir})
	if err != nil || !found {
		t.Fatalf("readSessionUsage() found=%v error=%v", found, err)
	}
	return u
}

const completeUsageShutdown = `{"type":"session.shutdown","timestamp":"2026-01-01T02:00:00Z","data":{"totalPremiumRequests":3,"totalNanoAiu":3000000000,"totalApiDurationMs":0,"tokenDetails":{"input":{"tokenCount":60},"cache_read":{"tokenCount":200},"cache_write":{"tokenCount":40},"output":{"tokenCount":20}},"modelMetrics":{"model-a":{"requests":{"count":1,"cost":3},"totalNanoAiu":2000000000,"usage":{"inputTokens":30,"cacheReadTokens":100,"cacheWriteTokens":20,"outputTokens":10}},"model-b":{"requests":{"count":1,"cost":0},"totalNanoAiu":1000000000,"usage":{"inputTokens":30,"cacheReadTokens":100,"cacheWriteTokens":20,"outputTokens":10}}}}}
`

func TestSessionUsageRecordsAndSnapshots(t *testing.T) {
	main := usageRecordEvent(t, "usage-a", "model-a", "user", "root", 2000000000)
	child := usageRecordEvent(t, "usage-b", "model-b", "sub-agent", "child", 1000000000)
	checkpoint := `{"type":"session.usage_checkpoint","timestamp":"2026-01-01T01:30:00Z","data":{"totalPremiumRequests":3,"totalNanoAiu":3000000000}}` + "\n"
	for _, test := range []struct {
		name, events string
		unknown      bool
	}{
		{"records only", main + child, false},
		{"checkpoint", main + child + checkpoint, false},
		{"shutdown", main + child + checkpoint + completeUsageShutdown, false},
		{"repeated snapshots", main + child + checkpoint + checkpoint + completeUsageShutdown + completeUsageShutdown, false},
		{"replayed records after checkpoint", main + child + checkpoint + main + child, false},
		{"shutdown without records", completeUsageShutdown, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			u := collectUsageLog(t, test.events)
			if u.Requests != 2 || u.PremiumRequests != 3 || math.Abs(u.AIU-3) > 1e-9 ||
				u.InputTokens != 60 || u.CacheReadTokens != 200 || u.CacheWriteTokens != 40 || u.OutputTokens != 20 || u.Incomplete {
				t.Fatalf("usage = %+v, want nonduplicated complete request and token totals", u)
			}
			if len(u.Models) != 2 || u.Models["model-a"].Requests != 1 || u.Models["model-b"].Requests != 1 {
				t.Fatalf("models = %+v, want one main and one child request", u.Models)
			}
			if test.unknown {
				if len(u.Roles) != 1 || u.Roles[UsageRoleUnknown].AIU != 3 {
					t.Fatalf("roles = %+v, want only unknown without request attribution", u.Roles)
				}
			} else if len(u.Roles) != 2 || u.Roles[UsageRoleMain].AIU != 2 ||
				u.Roles[UsageRoleSubagent].AIU != 1 || u.Roles[UsageRoleSubagent].PremiumRequests != 0 {
				t.Fatalf("roles = %+v, want main AIU=2 and subagent AIU=1 without premium charge", u.Roles)
			}
		})
	}
}

func TestSessionUsageResumedRecords(t *testing.T) {
	record := usageRecordEvent(t, "new-usage", "model-b", "user", "resumed", 4000000000)
	record = strings.ReplaceAll(record, "2026-01-01T01:00:00Z", "2026-01-01T03:00:00Z")
	u := collectUsageLog(t, completeUsageShutdown+record)
	if u.Requests != 3 || u.AIU != 7 || u.PremiumRequests != 6 ||
		u.InputTokens != 90 || u.OutputTokens != 30 || u.Models["model-b"].Requests != 2 {
		t.Fatalf("resumed usage = %+v, want snapshot plus only new request", u)
	}
	if !u.At.Equal(time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC)) {
		t.Fatalf("At = %v, want latest request time", u.At)
	}
	if u.Roles[UsageRoleMain].AIU != 4 || u.Roles[UsageRoleUnknown].AIU != 3 {
		t.Fatalf("roles = %+v, want new main usage plus unknown historical snapshot", u.Roles)
	}
}

func TestSessionUsageCheckpointPartialCoverage(t *testing.T) {
	checkpoint := `{"type":"session.usage_checkpoint","timestamp":"2026-01-01T02:00:00Z","data":{"totalPremiumRequests":8,"totalNanoAiu":7000000000}}` + "\n"
	for _, prefix := range []string{"", usageRecordEvent(t, "usage-a", "model-a", "user", "root", 2000000000)} {
		u := collectUsageLog(t, prefix+checkpoint)
		if u.AIU != 7 || u.PremiumRequests != 8 || !u.Incomplete {
			t.Fatalf("usage = %+v, want checkpoint totals with incomplete breakdown", u)
		}
		known := 0.0
		for _, role := range u.Roles {
			known += role.AIU
		}
		if math.Abs(known-u.AIU) > 1e-9 {
			t.Fatalf("role AIU sum=%v, want aggregate=%v", known, u.AIU)
		}
		if prefix == "" && (u.Requests != 0 || len(u.Models) != 0 || u.Roles[UsageRoleUnknown].AIU != 7) {
			t.Fatalf("checkpoint-only usage = %+v, must not invent model/request/token counts", u)
		}
		if prefix != "" && (u.Requests != 1 || u.Roles[UsageRoleUnknown].AIU != 5 || u.InputTokens != 30) {
			t.Fatalf("partial usage = %+v, want known record plus unknown cost remainder", u)
		}
	}
	record := strings.ReplaceAll(usageRecordEvent(t, "new", "model-a", "user", "root", 2000000000),
		"2026-01-01T01:00:00Z", "2026-01-01T03:00:00Z")
	u := collectUsageLog(t, checkpoint+record)
	if u.AIU != 9 || u.PremiumRequests != 11 || u.Requests != 1 {
		t.Fatalf("checkpoint then record = %+v, want cumulative checkpoint plus new delta", u)
	}
}

func TestSessionUsageLateRecordsAndSnapshots(t *testing.T) {
	main := usageRecordEvent(t, "main", "model-a", "user", "root", 2000000000)
	child := usageRecordEvent(t, "child", "model-b", "sub-agent", "child", 1000000000)
	for _, events := range []string{completeUsageShutdown + main + child, main + completeUsageShutdown + child} {
		u := collectUsageLog(t, events)
		if u.AIU != 3 || u.Requests != 2 || u.PremiumRequests != 3 || u.Incomplete ||
			u.Roles[UsageRoleMain].AIU != 2 || u.Roles[UsageRoleSubagent].AIU != 1 ||
			!u.At.Equal(time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)) {
			t.Fatalf("late records were counted twice or changed latest usage time: %+v", u)
		}
	}
	newer := `{"type":"session.usage_checkpoint","timestamp":"2026-01-01T03:00:00Z","data":{"totalNanoAiu":4000000000}}` + "\n"
	if u := collectUsageLog(t, newer+completeUsageShutdown); u.AIU != 4 || u.At.Hour() != 3 || !u.Incomplete {
		t.Fatalf("older snapshot replaced newer totals: %+v", u)
	}
}

func TestSessionUsageChildSnapshotAndFallbackIdentity(t *testing.T) {
	main := usageRecordEvent(t, "shared-id", "model-a", "user", "", 2000000000)
	child := usageRecordEvent(t, "shared-id", "model-b", "sub-agent", "", 1000000000)
	child = strings.ReplaceAll(child, `"type":"session.usage_record"`, `"type":"session.usage_record","agentId":"child-a"`)
	childSnapshot := `{"type":"session.shutdown","agentId":"child-a","timestamp":"2026-01-01T04:00:00Z","data":{"totalNanoAiu":1000000000}}` + "\n"
	u := collectUsageLog(t, main+child+child+completeUsageShutdown+childSnapshot)
	if u.AIU != 3 || u.Requests != 2 || len(u.Roles) != 2 || u.At.Hour() != 2 {
		t.Fatalf("child snapshot or identity collided with parent usage: %+v", u)
	}
}

func TestSessionUsageMissingSnapshotTokenBreakdown(t *testing.T) {
	snapshot := strings.ReplaceAll(completeUsageShutdown,
		`"tokenDetails":{"input":{"tokenCount":60},"cache_read":{"tokenCount":200},"cache_write":{"tokenCount":40},"output":{"tokenCount":20}},`, "")
	u := collectUsageLog(t, snapshot)
	if !u.Incomplete || u.AIU != 3 || u.Requests != 2 || u.InputTokens != 0 {
		t.Fatalf("snapshot with missing total tokens should have incomplete coverage: %+v", u)
	}
}

func TestSessionUsageAccountingIdentityAndPremiumCharges(t *testing.T) {
	main := usageRecordEvent(t, "same-id", "model-a", "user", "root", 2000000000)
	agent := usageRecordEvent(t, "next-id", "model-a", "agent", "root", 1000000000)
	child := usageRecordEvent(t, "same-id", "model-b", "sub-agent", "child", 1000000000)
	u := collectUsageLog(t, main+main+agent+child+child)
	if u.Requests != 3 || u.AIU != 4 || u.PremiumRequests != 3 || u.Roles[UsageRoleMain].Requests != 2 {
		t.Fatalf("usage = %+v, want source-scoped deduplication and only user premium charge", u)
	}
}

func TestSessionUsageBillingTokensAndRoleMarkers(t *testing.T) {
	record := `{"type":"session.usage_record","agentId":"child-agent","timestamp":"2026-01-01T01:00:00Z","data":{"usage":{"model":"model-a","initiator":"agent","inputTokens":999,"cacheReadTokens":900,"outputTokens":999,"copilotUsage":{"totalNanoAiu":1000000000,"tokenDetails":[{"tokenType":"input","tokenCount":10},{"tokenType":"cache_read","tokenCount":20},{"tokenType":"cache_write","tokenCount":30},{"tokenType":"output","tokenCount":40}]}}}}` + "\n"
	u := collectUsageLog(t, record)
	if u.InputTokens != 10 || u.CacheReadTokens != 20 || u.CacheWriteTokens != 30 || u.OutputTokens != 40 ||
		u.Roles[UsageRoleSubagent].Requests != 1 || u.PremiumRequests != 0 {
		t.Fatalf("usage = %+v, want billing token details and envelope subagent attribution", u)
	}
	for _, marker := range []string{`"parentToolCallId":"child-call"`, `"interactionType":"conversation-subagent"`} {
		record := fmt.Sprintf(`{"type":"session.usage_record","timestamp":"2026-01-01T01:00:00Z","data":{"usage":{"model":"model-a","initiator":"agent",%s,"copilotUsage":{"totalNanoAiu":0}}}}`+"\n", marker)
		if u := collectUsageLog(t, record); u.Roles[UsageRoleSubagent].Requests != 1 {
			t.Fatalf("usage = %+v, want subagent marker %s", u, marker)
		}
	}
}

func TestSessionUsageInvalidRecordsAndUnknownAttribution(t *testing.T) {
	invalid := `{"type":"session.usage_record","timestamp":"2026-01-01T01:00:00Z","data":{"usage":{"model":"model-a","inputTokens":1,"cacheReadTokens":10}}}
{"type":"session.usage_record","timestamp":"2026-01-01T01:00:00Z","data":{"usage":null}}
`
	u := collectUsageLog(t, invalid+completeUsageShutdown)
	if u.Requests != 2 || u.AIU != 3 || len(u.Roles) != 1 {
		t.Fatalf("invalid records corrupted snapshot: %+v", u)
	}
	unknown := usageRecordEvent(t, "sample", "model-a", "mcp-sampling", "root", 1000000000)
	u = collectUsageLog(t, unknown)
	if u.Roles[UsageRoleUnknown].Requests != 1 || u.PremiumRequests != 0 {
		t.Fatalf("usage = %+v, want unknown rather than main for unrecognized initiator", u)
	}
	mismatch := usageRecordEvent(t, "sample", "model-a", "user", "root", 9000000000)
	u = collectUsageLog(t, mismatch+completeUsageShutdown)
	if u.AIU != 3 || !u.Incomplete || len(u.Roles) != 1 || u.Roles[UsageRoleUnknown].AIU != 3 {
		t.Fatalf("usage = %+v, want authoritative snapshot and unknown conflicting attribution", u)
	}
}

func TestCollectPermissionStatsLiveUsageScopeAndTime(t *testing.T) {
	root := t.TempDir()
	events := usageRecordEvent(t, "main", "model-a", "user", "root", 2000000000) +
		usageRecordEvent(t, "child", "model-b", "sub-agent", "child", 1000000000)
	writeSessionDir(t, root, "session-a", "id: session-a\ncwd: /repo/a\n", events)
	writeSessionDir(t, root, "session-b", "id: session-b\ncwd: /repo/b\n", events)
	for _, test := range []struct {
		name string
		opts PermissionStatsOptions
		want int
	}{
		{"all", PermissionStatsOptions{}, 2},
		{"session", PermissionStatsOptions{SessionID: "session-a"}, 1},
		{"scope", PermissionStatsOptions{Scope: SessionScope{ExactCWD: "/repo/a"}}, 1},
		{"top does not truncate roles", PermissionStatsOptions{Top: 1}, 2},
		{"since includes boundary", PermissionStatsOptions{Since: time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, 2},
		{"since excludes", PermissionStatsOptions{Since: time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)}, 0},
		{"until excludes boundary", PermissionStatsOptions{Until: time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.opts.Root = root
			test.opts.Kind = []string{"write"}
			stats, err := CollectPermissionStats(test.opts)
			if err != nil {
				t.Fatal(err)
			}
			if stats.UsageSessions != test.want || stats.UsageRequests != test.want*2 ||
				stats.UsageAIU != float64(test.want)*3 || stats.Requests != 0 || stats.UsageIncompleteSessions != 0 {
				t.Fatalf("stats = %+v, want %d complete usage sessions independent of permission filter", stats, test.want)
			}
			if test.want > 0 && len(stats.ByRoleUsage) != 2 {
				t.Fatalf("roles = %+v, want both roles regardless of top limit", stats.ByRoleUsage)
			}
		})
	}
}

func TestSessionUsagePartialPremiumCheckpointIsIncomplete(t *testing.T) {
	// A later checkpoint replaces only the premium-request total, so the per-model
	// breakdown kept from the shutdown snapshot no longer sums to it.
	checkpoint := `{"type":"session.usage_checkpoint","timestamp":"2026-01-01T03:00:00Z","data":{"totalPremiumRequests":5}}` + "\n"
	u := collectUsageLog(t, completeUsageShutdown+checkpoint)
	if u.PremiumRequests != 5 {
		t.Fatalf("usage = %+v, want checkpoint premium total", u)
	}
	var models SessionUsage
	for _, model := range u.Models {
		models.add(model)
	}
	if models.PremiumRequests == u.PremiumRequests {
		t.Fatalf("model premium sum=%v unexpectedly matches total", models.PremiumRequests)
	}
	if !u.Incomplete {
		t.Fatalf("usage = %+v, want incomplete when model premium sum disagrees with total", u)
	}
}

func TestSessionUsageWithinInterval(t *testing.T) {
	early := usageRecordEvent(t, "usage-a", "model-a", "user", "root", 2000000000)
	late := strings.Replace(usageRecordEvent(t, "usage-b", "model-b", "user", "root", 1000000000),
		"2026-01-01T01:00:00Z", "2026-01-01T03:00:00Z", 1)
	dir := writeSessionDir(t, t.TempDir(), "session-a", "", early+completeUsageShutdown+late)
	since := time.Date(2026, 1, 1, 2, 30, 0, 0, time.UTC)
	u, err := readSessionActivity(Session{ID: "session-a", Dir: dir}, since, time.Time{})
	if err != nil || !u.usageFound {
		t.Fatalf("readSessionActivity() found=%v error=%v", u.usageFound, err)
	}
	if u.usage.Requests != 1 || math.Abs(u.usage.AIU-1) > 1e-9 || u.usage.Models["model-b"].Requests != 1 ||
		len(u.usage.Models) != 1 || u.usage.Roles[UsageRoleMain].Requests != 1 {
		t.Fatalf("usage since %v = %+v, want only the late record", since, u.usage)
	}
	until := time.Date(2026, 1, 1, 0, 30, 0, 0, time.UTC)
	if u, err := readSessionActivity(Session{ID: "session-a", Dir: dir}, time.Time{}, until); err != nil || u.usageFound {
		t.Fatalf("readSessionActivity() until %v found=%v error=%v, want not found", until, u.usageFound, err)
	}
}
