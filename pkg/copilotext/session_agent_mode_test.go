package copilotext

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestPermissionRequestAgentModeSnapshots(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mode    any
		missing bool
		want    string
	}{
		{name: "interactive", mode: "interactive", want: "interactive"},
		{name: "plan", mode: "plan", want: "plan"},
		{name: "autopilot", mode: "autopilot", want: "autopilot"},
		{name: "future", mode: "future-mode", want: "future-mode"},
		{name: "missing", missing: true},
		{name: "null", mode: nil},
		{name: "empty", mode: ""},
		{name: "number", mode: 42},
		{name: "object", mode: map[string]string{"mode": "plan"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]any{
				"requestId": "request-a", "permissionMode": "manual",
				"permissionRequest": map[string]any{"kind": "read"},
			}
			if !tc.missing {
				data["agentMode"] = tc.mode
			}
			payload, err := json.Marshal(data)
			if err != nil {
				t.Fatal(err)
			}
			pc := newPermissionCollector(Session{})
			pc.handle(rawEvent{Type: "session.mode_changed", Data: json.RawMessage(`{"newMode":"autopilot"}`)})
			pc.handle(rawEvent{Type: "permission.requested", Data: payload})
			records := pc.records()
			if len(records) != 1 || records[0].Request.AgentMode != tc.want ||
				records[0].Request.Kind != "read" || records[0].Result != PermissionUnresolved {
				t.Fatalf("records = %+v, want one preserved request with mode %q", records, tc.want)
			}
		})
	}
}

func TestCollectPermissionStatsAgentModeIdentityAndOutcomes(t *testing.T) {
	root := t.TempDir()
	events := `{"type":"session.mode_changed","data":{"previousMode":"interactive","newMode":"plan"}}
{"type":"permission.completed","data":{"requestId":"same","result":{"kind":"approved"}}}
{"type":"permission.requested","data":{"requestId":"same","agentMode":"interactive","permissionRequest":{"kind":"read"}}}
{"type":"permission.requested","data":{"requestId":"same","agentMode":"autopilot","permissionRequest":{"kind":"read"}}}
{"type":"permission.completed","data":{"requestId":"same","result":{"kind":"denied"}}}
{"type":"permission.requested","agentId":"child","data":{"requestId":"same","agentMode":"plan","permissionRequest":{"kind":"write"}}}
{"type":"permission.completed","agentId":"child","data":{"requestId":"same","agentMode":"interactive","result":{"kind":"denied-interactively-by-user"}}}
{"type":"permission.requested","data":{"requestId":"future","agentMode":"future-mode","permissionRequest":{"kind":"read"}}}
{"type":"permission.completed","data":{"requestId":"future","result":{"kind":"approved-for-location"}}}
{"type":"permission.requested","data":{"requestId":"autonomous","agentMode":"autopilot","permissionRequest":{"kind":"shell"}}}
{"type":"permission.completed","data":{"requestId":"autonomous","result":{"kind":"denied-no-rule"}}}
{"type":"permission.requested","data":{"requestId":"empty","agentMode":"","permissionRequest":{"kind":"read"}}}
{"type":"permission.requested","data":{"requestId":"null","agentMode":null,"permissionRequest":{"kind":"read"}}}
{"type":"permission.completed","data":{"requestId":"null","result":{"kind":"approved"}}}
{"type":"permission.requested","data":{"requestId":"missing","permissionRequest":{"kind":"read"}}}
{"type":"permission.requested","data":{"requestId":"number","agentMode":42,"permissionRequest":{"kind":"read"}}}
{"type":"permission.requested","data":{"requestId":"object","agentMode":{"mode":"plan"},"permissionRequest":{"kind":"read"}}}
{"type":"permission.completed","data":{"requestId":"object","result":{"kind":"approved-for-location"}}}
{"type":"permission.requested","data":{"requestId":"approved-variant","agentMode":"interactive","permissionRequest":{"kind":"read"}}}
{"type":"permission.completed","data":{"requestId":"approved-variant","result":{"kind":"approved-by-rule"}}}
{"type":"session.mode_changed","data":{"previousMode":"plan","newMode":"autopilot"}}
`
	writeSessionDir(t, root, "session-a", "id: session-a\ncwd: /repo/a\n", events)
	writeSessionDir(t, root, "session-b", "id: session-b\ncwd: /repo/b\n",
		`{"type":"permission.requested","data":{"requestId":"same","agentMode":"interactive","permissionRequest":{"kind":"read"}}}
{"type":"permission.completed","data":{"requestId":"same","result":{"kind":"denied"}}}
`)
	stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	want := []Count{
		{Key: "unknown", Total: 5, Approved: 1, ApprovedForLocation: 1, Unresolved: 3},
		{Key: "interactive", Total: 3, Approved: 2, Denied: 1},
		{Key: "autopilot", Total: 1, Denied: 1},
		{Key: "future-mode", Total: 1, ApprovedForLocation: 1},
		{Key: "plan", Total: 1, Denied: 1},
	}
	if stats.Sessions != 2 || stats.Requests != 11 || !reflect.DeepEqual(stats.ByAgentMode, want) {
		t.Fatalf("sessions=%d requests=%d modes=%+v, want %+v", stats.Sessions, stats.Requests, stats.ByAgentMode, want)
	}
	filtered, err := CollectPermissionStats(PermissionStatsOptions{Root: root, AgentMode: []string{"unknown"}})
	if err != nil {
		t.Fatal(err)
	}
	if filtered.Sessions != 1 || filtered.Requests != 5 || !reflect.DeepEqual(filtered.ByAgentMode, want[:1]) {
		t.Fatalf("unknown filter = %+v, want %+v", filtered, want[:1])
	}
}

func TestCollectPermissionStatsAgentModeFilters(t *testing.T) {
	root := t.TempDir()
	events := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"a","agentMode":"interactive","permissionRequest":{"kind":"shell","commands":[{"identifier":"list","readOnly":true}],"possiblePaths":["/repo/a/first"],"possibleUrls":["https://example.test/a"]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:01:10Z","data":{"requestId":"a","result":{"kind":"approved"},"decisionSource":"human_response"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"b","agentMode":"plan","permissionRequest":{"kind":"shell","commands":[{"identifier":"mutate","readOnly":false}],"possiblePaths":["/repo/a/other"]}}}
{"type":"permission.completed","data":{"requestId":"b","result":{"kind":"denied-no-rule"},"decisionSource":"unattended_fallback"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:02:00Z","data":{"requestId":"c","agentMode":"autopilot","permissionRequest":{"kind":"read","possiblePaths":["/repo/a/other"]}}}
{"type":"permission.completed","data":{"requestId":"c","result":{"kind":"approved-for-location"}}}
{"type":"permission.requested","timestamp":"2026-01-01T00:03:00Z","data":{"requestId":"d","permissionRequest":{"kind":"write","possiblePaths":["/repo/a/other"]}}}
{"type":"permission.requested","timestamp":"2026-01-01T00:04:00Z","data":{"requestId":"e","agentMode":"future-mode","permissionRequest":{"kind":"read","possiblePaths":["/repo/a/other"]}}}
{"type":"permission.completed","data":{"requestId":"e","result":{"kind":"approved"},"decisionSource":"host_policy"}}
`
	writeSessionDir(t, root, "session-a", "id: session-a\ncwd: /repo/a\n", events)
	writeSessionDir(t, root, "session-b", "id: session-b\ncwd: /repo/b\n",
		`{"type":"permission.requested","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"a","agentMode":"interactive","permissionRequest":{"kind":"shell","commands":[{"identifier":"list","readOnly":true}],"possibleUrls":["https://example.test/b"]}}}
{"type":"permission.completed","data":{"requestId":"a","result":{"kind":"approved"},"decisionSource":"human_response"}}
`)
	interactive := Count{Key: "interactive", Total: 1, Approved: 1}
	interactiveTwo := Count{Key: "interactive", Total: 2, Approved: 2}
	plan := Count{Key: "plan", Total: 1, Denied: 1}
	autopilot := Count{Key: "autopilot", Total: 1, ApprovedForLocation: 1}
	future := Count{Key: "future-mode", Total: 1, Approved: 1}
	unknown := Count{Key: "unknown", Total: 1, Unresolved: 1}
	for _, tc := range []struct {
		name     string
		opts     PermissionStatsOptions
		requests int
		want     []Count
	}{
		{"all", PermissionStatsOptions{}, 6, []Count{interactiveTwo, autopilot, future, plan, unknown}},
		{"top", PermissionStatsOptions{Top: 1}, 6, []Count{interactiveTwo}},
		{"top-tie", PermissionStatsOptions{Kind: []string{"read"}, Top: 1}, 2, []Count{autopilot}},
		{"mode", PermissionStatsOptions{AgentMode: []string{"plan"}}, 1, []Count{plan}},
		{"mode-or", PermissionStatsOptions{AgentMode: []string{"plan", "unknown", "plan"}}, 2, []Count{plan, unknown}},
		{"mode-and-source", PermissionStatsOptions{AgentMode: []string{"interactive", "plan"}, Kind: []string{"shell"}, DecisionSource: []string{DecisionSourceHumanResponse}}, 2, []Count{interactiveTwo}},
		{"mode-and-operation", PermissionStatsOptions{AgentMode: []string{"interactive", "plan", "unknown"}, Operations: []string{"write"}}, 2, []Count{plan, unknown}},
		{"mode-before-top", PermissionStatsOptions{AgentMode: []string{"interactive", "plan", "unknown"}, Top: 1}, 4, []Count{interactiveTwo}},
		{"mode-no-match", PermissionStatsOptions{AgentMode: []string{"absent"}}, 0, []Count{}},
		{"cwd", PermissionStatsOptions{Scope: SessionScope{ExactCWD: "/repo/a"}}, 5, []Count{autopilot, future, interactive, plan, unknown}},
		{"worktree", PermissionStatsOptions{Scope: SessionScope{UnderDir: "/repo/a"}}, 5, []Count{autopilot, future, interactive, plan, unknown}},
		{"session", PermissionStatsOptions{SessionID: "session-b"}, 1, []Count{interactive}},
		{"absent", PermissionStatsOptions{SessionID: "absent"}, 0, []Count{}},
		{"operation", PermissionStatsOptions{Operations: []string{"write"}}, 2, []Count{plan, unknown}},
		{"kind", PermissionStatsOptions{Kind: []string{"read"}}, 2, []Count{autopilot, future}},
		{"command", PermissionStatsOptions{Command: []string{"list"}}, 2, []Count{interactiveTwo}},
		{"path", PermissionStatsOptions{Path: []string{"^/repo/a/first$"}}, 1, []Count{interactive}},
		{"url", PermissionStatsOptions{URL: []string{"example\\.test/b"}}, 1, []Count{interactive}},
		{"human", PermissionStatsOptions{DecisionSource: []string{DecisionSourceHumanResponse}}, 2, []Count{interactiveTwo}},
		{"unknown-source", PermissionStatsOptions{DecisionSource: []string{DecisionSourceUnknown}}, 2, []Count{autopilot, unknown}},
		{"until", PermissionStatsOptions{Until: time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)}, 1, []Count{interactive}},
		{"window", PermissionStatsOptions{Since: time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC), Until: time.Date(2026, 1, 1, 0, 2, 0, 0, time.UTC)}, 2, []Count{interactive, plan}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts.Root = root
			stats, err := CollectPermissionStats(tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if stats.Requests != tc.requests || !reflect.DeepEqual(stats.ByAgentMode, tc.want) {
				t.Fatalf("requests=%d modes=%+v, want requests=%d modes=%+v", stats.Requests, stats.ByAgentMode, tc.requests, tc.want)
			}
		})
	}
}

func TestCollectPermissionStatsAgentModeExactValues(t *testing.T) {
	root := t.TempDir()
	events := `{"type":"permission.requested","data":{"requestId":"a","agentMode":" plan ","permissionRequest":{"kind":"read"}}}
{"type":"permission.requested","data":{"requestId":"b","agentMode":"plan","permissionRequest":{"kind":"read"}}}
{"type":"permission.requested","data":{"requestId":"c","agentMode":"Plan","permissionRequest":{"kind":"read"}}}
{"type":"permission.requested","data":{"requestId":"d","agentMode":"future,a","permissionRequest":{"kind":"read"}}}
{"type":"permission.requested","data":{"requestId":"e","agentMode":" ","permissionRequest":{"kind":"read"}}}
`
	writeSessionDir(t, root, "session-a", "id: session-a\ncwd: /repo/a\n", events)
	for _, mode := range []string{" plan ", "plan", "Plan", "future,a", " "} {
		t.Run(mode, func(t *testing.T) {
			stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, AgentMode: []string{mode}})
			if err != nil {
				t.Fatal(err)
			}
			want := []Count{{Key: mode, Total: 1, Unresolved: 1}}
			if stats.Requests != 1 || !reflect.DeepEqual(stats.ByAgentMode, want) {
				t.Fatalf("mode %q: requests=%d modes=%+v, want %+v", mode, stats.Requests, stats.ByAgentMode, want)
			}
		})
	}
	stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, AgentMode: []string{"PLAN"}})
	if err != nil || stats.Requests != 0 {
		t.Fatalf("mode filter must be case-sensitive: %+v, %v", stats, err)
	}
	if _, err := CollectPermissionStats(PermissionStatsOptions{Root: root, AgentMode: []string{""}}); err == nil {
		t.Fatal("empty agent mode must be rejected")
	}
}

func TestCollectPermissionWaitFollowsAgentModeFilter(t *testing.T) {
	root := t.TempDir()
	events := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"a","agentMode":"interactive","permissionRequest":{"kind":"read"}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:10Z","data":{"requestId":"a","result":{"kind":"approved"},"decisionSource":"human_response"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"b","agentMode":"plan","permissionRequest":{"kind":"write"}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:01:30Z","data":{"requestId":"b","result":{"kind":"denied"},"decisionSource":"human_response"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:02:00Z","data":{"requestId":"c","permissionRequest":{"kind":"read"}}}
`
	writeSessionDir(t, root, "session-a", "id: session-a\ncwd: /repo/a\n", events)
	stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, AgentMode: []string{"plan", "unknown"}, Top: 1})
	if err != nil {
		t.Fatal(err)
	}
	wait := stats.PermissionWait
	if stats.Requests != 2 || wait.HumanResponses.Completed != 1 || wait.HumanResponses.TotalWaitMs != 30000 ||
		wait.Unresolved != 1 || len(stats.ByAgentMode) != 1 || stats.ByAgentMode[0].Key != "plan" {
		t.Fatalf("mode-filtered requests=%d modes=%+v wait=%+v", stats.Requests, stats.ByAgentMode, wait)
	}
}

func TestCollectPermissionStatsAgentModeEmpty(t *testing.T) {
	stats, err := CollectPermissionStats(PermissionStatsOptions{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Requests != 0 || stats.ByAgentMode == nil || len(stats.ByAgentMode) != 0 {
		t.Fatalf("empty requests invented modes or omitted the supported axis: %+v", stats)
	}
}
