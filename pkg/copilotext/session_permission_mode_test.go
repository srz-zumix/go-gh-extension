package copilotext

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestPermissionRequestPermissionModeSnapshots(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mode    any
		missing bool
		want    string
	}{
		{name: "assisted", mode: "assisted", want: "assisted"},
		{name: "manual", mode: "manual", want: "manual"},
		{name: "future", mode: " future,mode ", want: " future,mode "},
		{name: "missing", missing: true},
		{name: "null", mode: nil},
		{name: "empty", mode: ""},
		{name: "number", mode: 42},
		{name: "object", mode: map[string]string{"mode": "assisted"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]any{
				"requestId": "request-a", "agentMode": "plan",
				"permissionRequest": map[string]any{"kind": "shell"},
			}
			if !tc.missing {
				data["permissionMode"] = tc.mode
			}
			payload, err := json.Marshal(data)
			if err != nil {
				t.Fatal(err)
			}
			pc := newPermissionCollector(Session{})
			pc.handle(rawEvent{Type: "permission.requested", Data: payload})
			records := pc.records()
			if len(records) != 1 || records[0].Request.PermissionMode != tc.want ||
				records[0].Request.AgentMode != "plan" || records[0].Request.Kind != "shell" {
				t.Fatalf("records = %+v, want preserved permission mode %q", records, tc.want)
			}
		})
	}
}

func TestCollectPermissionStatsPermissionModeAndHumanResponse(t *testing.T) {
	root := t.TempDir()
	events := `{"type":"permission.requested","agentId":"agent-a","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"same","permissionMode":"assisted","agentMode":"interactive","permissionRequest":{"kind":"shell","commands":[{"identifier":"compile","readOnly":true},{"identifier":"verify","readOnly":true},{"identifier":"compile","readOnly":true}]}}}
{"type":"permission.requested","agentId":"agent-a","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"same","permissionMode":"manual","permissionRequest":{"kind":"shell"}}}
{"type":"permission.completed","agentId":"agent-a","timestamp":"2026-01-01T00:10:00Z","data":{"requestId":"same","result":{"kind":"approved"},"decisionSource":"human_response"}}
{"type":"permission.requested","agentId":"agent-b","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"same","permissionMode":"manual","permissionRequest":{"kind":"shell"}}}
{"type":"permission.completed","agentId":"agent-b","timestamp":"2026-01-01T00:02:00Z","data":{"requestId":"same","result":{"kind":"approved"},"decisionSource":"human_response"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"denied","permissionMode":"assisted","permissionRequest":{"kind":"shell"}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:02:00Z","data":{"requestId":"denied","result":{"kind":"denied"},"decisionSource":"human_response"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"auto","permissionMode":"assisted","permissionRequest":{"kind":"shell"}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:02:00Z","data":{"requestId":"auto","result":{"kind":"approved"},"decisionSource":"assisted_approval"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"legacy","permissionRequest":{"kind":"shell"}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:02:00Z","data":{"requestId":"legacy","result":{"kind":"approved"},"decisionSource":"human_response"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"invalid","permissionMode":42,"permissionRequest":{"kind":"shell"}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:02:00Z","data":{"requestId":"invalid","result":{"kind":"approved"},"decisionSource":"human_response"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"pending","permissionMode":"assisted","permissionRequest":{"kind":"shell"}}}
{"type":"permission.requested","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"future","permissionMode":" future,mode ","permissionRequest":{"kind":"shell"}}}
`
	writeSessionDir(t, root, "session-a", "id: session-a\ncwd: /repo/a\n", events)
	writeSessionDir(t, root, "session-b", "id: session-b\ncwd: /repo/b\n",
		`{"type":"permission.requested","agentId":"agent-a","timestamp":"2026-01-01T00:01:00Z","data":{"requestId":"same","permissionMode":"assisted","permissionRequest":{"kind":"shell","commands":[{"identifier":"verify","readOnly":true}]}}}
{"type":"permission.completed","agentId":"agent-a","timestamp":"2026-01-01T00:02:00Z","data":{"requestId":"same","result":{"kind":"approved"},"decisionSource":"human_response"}}
`)

	opts := PermissionStatsOptions{
		Root: root, Until: time.Date(2026, 1, 1, 0, 5, 0, 0, time.UTC),
		PermissionMode: []string{"assisted"}, DecisionSource: []string{DecisionSourceHumanResponse},
		Results: []string{"approved"}, Kind: []string{"shell"},
	}
	stats, err := CollectPermissionStats(opts)
	if err != nil {
		t.Fatal(err)
	}
	wantModes := []Count{{Key: "assisted", Total: 2, Approved: 2}}
	wantCommands := []Count{{Key: "verify", Total: 2, Approved: 2}, {Key: "compile", Total: 1, Approved: 1}}
	if stats.Requests != 2 || stats.Sessions != 2 || !reflect.DeepEqual(stats.ByPermissionMode, wantModes) ||
		!reflect.DeepEqual(stats.ByCommand, wantCommands) || stats.PermissionWait.HumanResponses.Completed != 2 {
		t.Fatalf("filtered stats = %+v, want modes %+v and commands %+v", stats, wantModes, wantCommands)
	}

	opts.Results = nil
	stats, err = CollectPermissionStats(opts)
	if err != nil || stats.Requests != 3 ||
		!reflect.DeepEqual(stats.ByPermissionMode, []Count{{Key: "assisted", Total: 3, Approved: 2, Denied: 1}}) {
		t.Fatalf("assisted human decisions must include denials: %+v, error %v", stats, err)
	}
	opts.PermissionMode = []string{"Assisted"}
	stats, err = CollectPermissionStats(opts)
	if err != nil || stats.Requests != 0 {
		t.Fatalf("permission mode matching must be case-sensitive: %+v, error %v", stats, err)
	}

	opts.PermissionMode = []string{"unknown"}
	opts.Results = []string{"approved"}
	stats, err = CollectPermissionStats(opts)
	if err != nil || stats.Requests != 2 ||
		!reflect.DeepEqual(stats.ByPermissionMode, []Count{{Key: "unknown", Total: 2, Approved: 2}}) {
		t.Fatalf("unknown permission modes = %+v, error %v", stats, err)
	}
	opts.PermissionMode = []string{"manual", " future,mode "}
	opts.DecisionSource, opts.Results = nil, nil
	stats, err = CollectPermissionStats(opts)
	if err != nil || stats.Requests != 2 {
		t.Fatalf("exact alternative permission modes = %+v, error %v", stats, err)
	}
	opts.PermissionMode = []string{""}
	if _, err := CollectPermissionStats(opts); err == nil {
		t.Fatal("empty permission mode must be rejected")
	}
}
