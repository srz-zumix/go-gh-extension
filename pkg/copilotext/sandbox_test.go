package copilotext

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func sandboxCounts(counts []Count) map[string]int {
	m := make(map[string]int, len(counts))
	for _, c := range counts {
		m[c.Key] = c.Total
	}
	return m
}

const sandboxEvents = `{"type":"tool.execution_start","timestamp":"2026-01-01T00:00:00Z","data":{"toolCallId":"tc-1","toolName":"bash"}}
{"type":"permission.requested","timestamp":"2026-01-01T00:00:01Z","data":{"requestId":"req-1","permissionRequest":{"kind":"shell","toolCallId":"tc-1","commands":[{"identifier":"make","readOnly":false}]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:02Z","data":{"requestId":"req-1","result":{"kind":"approved"}}}
{"type":"tool.execution_complete","timestamp":"2026-01-01T00:00:03Z","data":{"toolCallId":"tc-1","toolTelemetry":{"properties":{"sandboxApplied":"true","sandboxOptOutRequested":"false"}}}}
{"type":"tool.execution_start","timestamp":"2026-01-02T00:00:00Z","data":{"toolCallId":"tc-2","toolName":"bash"}}
{"type":"permission.requested","timestamp":"2026-01-02T00:00:01Z","data":{"requestId":"req-2","permissionRequest":{"kind":"shell","toolCallId":"tc-2","commands":[{"identifier":"curl","readOnly":false}]}}}
{"type":"permission.completed","timestamp":"2026-01-02T00:00:02Z","data":{"requestId":"req-2","result":{"kind":"approved"}}}
{"type":"tool.execution_complete","timestamp":"2026-01-02T00:00:03Z","data":{"toolCallId":"tc-2","toolTelemetry":{"properties":{"sandboxApplied":"false","sandboxOptOutRequested":"true"}}}}
{"type":"tool.execution_start","agentId":"sub-1","timestamp":"2026-01-03T00:00:00Z","data":{"toolCallId":"tc-3","toolName":"rg"}}
{"type":"permission.requested","timestamp":"2026-01-03T00:00:01Z","data":{"requestId":"req-3","permissionRequest":{"kind":"read","toolCallId":"tc-3","possiblePaths":["/repo/a"]}}}
{"type":"permission.completed","timestamp":"2026-01-03T00:00:02Z","data":{"requestId":"req-3","result":{"kind":"approved"}}}
{"type":"tool.execution_complete","agentId":"sub-1","timestamp":"2026-01-03T00:00:03Z","data":{"toolCallId":"tc-3","toolTelemetry":{"properties":{"sandboxApplied":"true"}}}}
{"type":"tool.execution_start","timestamp":"2026-01-04T00:00:00Z","data":{"toolCallId":"tc-4","toolName":"bash"}}
{"type":"tool.execution_complete","timestamp":"2026-01-04T00:00:03Z","data":{"toolCallId":"tc-4","toolTelemetry":{"properties":{}}}}
{"type":"tool.execution_start","timestamp":"2026-01-04T00:00:04Z","data":{"toolCallId":"tc-5","toolName":"view"}}
{"type":"tool.execution_complete","timestamp":"2026-01-04T00:00:05Z","data":{"toolCallId":"tc-5"}}
{"type":"permission.requested","timestamp":"2026-01-05T00:00:01Z","data":{"requestId":"req-6","permissionRequest":{"kind":"url","toolCallId":"tc-6","possibleUrls":["https://example.com"]}}}
`

func TestCollectPermissionStatsSandbox(t *testing.T) {
	root := t.TempDir()
	writeSessionDir(t, root, "session-a",
		"id: session-a\ncwd: /repo/a\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n",
		sandboxEvents)

	stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root})
	if err != nil {
		t.Fatalf("CollectPermissionStats() error = %v", err)
	}
	// req-3 was logged without the subagent identity; it resolves through the unique
	// session-local toolCallId. req-6 never executed.
	got := sandboxCounts(stats.BySandbox)
	if got[SandboxApplied] != 2 || got[SandboxNotApplied] != 1 || got[SandboxUnknown] != 1 {
		t.Fatalf("BySandbox = %+v, want applied=2 not-applied=1 unknown=1", stats.BySandbox)
	}
	sb := stats.Sandbox
	// tc-5 (view, no telemetry) is excluded; tc-4 (bash, no telemetry) is unknown.
	if sb.Sessions != 1 || sb.Executions != 4 || sb.Applied != 2 || sb.NotApplied != 1 || sb.Unknown != 1 || sb.OptOutRequested != 1 {
		t.Fatalf("Sandbox = %+v, want 4 executions (2 applied, 1 not-applied, 1 unknown, 1 opt-out)", sb)
	}
	byTool := map[string]SandboxCount{}
	for _, c := range sb.ByTool {
		byTool[c.Key] = c
	}
	if byTool["bash"].Total != 3 || byTool["bash"].Applied != 1 || byTool["rg"].Applied != 1 {
		t.Fatalf("Sandbox.ByTool = %+v, want bash=3 (1 applied) rg=1 applied", sb.ByTool)
	}
}

func TestCollectPermissionStatsSandboxFilter(t *testing.T) {
	root := t.TempDir()
	writeSessionDir(t, root, "session-a",
		"id: session-a\ncwd: /repo/a\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n",
		sandboxEvents)

	stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Sandbox: []string{SandboxNotApplied}})
	if err != nil {
		t.Fatalf("CollectPermissionStats() error = %v", err)
	}
	if stats.Requests != 1 || len(stats.ByCommand) != 1 || stats.ByCommand[0].Key != "curl" {
		t.Fatalf("Requests = %d ByCommand = %+v, want only curl", stats.Requests, stats.ByCommand)
	}
	if stats.Sandbox.Executions != 1 || stats.Sandbox.NotApplied != 1 {
		t.Fatalf("Sandbox = %+v, want one not-applied execution", stats.Sandbox)
	}

	// Time bounds use completion time for executions.
	stats, err = CollectPermissionStats(PermissionStatsOptions{Root: root,
		Since: time.Date(2026, 1, 2, 0, 0, 3, 0, time.UTC), Until: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("CollectPermissionStats() error = %v", err)
	}
	if stats.Sandbox.Executions != 1 || stats.Sandbox.NotApplied != 1 {
		t.Fatalf("windowed Sandbox = %+v, want only tc-2", stats.Sandbox)
	}

	if _, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Sandbox: []string{"maybe"}}); err == nil {
		t.Fatal("CollectPermissionStats() with an invalid sandbox state should fail")
	}
}

func TestToolExecutionCollectorAmbiguousFallback(t *testing.T) {
	ec := newToolExecutionCollector(Session{})
	for _, agent := range []string{"sub-1", "sub-2"} {
		ec.handle(rawEvent{AgentID: agent, Type: "tool.execution_complete",
			Data: []byte(`{"toolCallId":"tc-1","toolTelemetry":{"properties":{"sandboxApplied":"true"}}}`)})
	}
	if got := ec.sandboxFor("", "tc-1"); got != SandboxUnknown {
		t.Fatalf("sandboxFor() = %q, want unknown for an ambiguous toolCallId", got)
	}
	if got := ec.sandboxFor("sub-2", "tc-1"); got != SandboxApplied {
		t.Fatalf("sandboxFor() exact = %q, want applied", got)
	}
}

func writeChatSession(t *testing.T, hashDir, sessionID, content string) {
	t.Helper()
	dir := filepath.Join(hashDir, "chatSessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, sessionID+".jsonl"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// chatPatches exercises snapshots, request pushes, indexed truncation, response
// replacement, and a later serialization that omits isSandboxWrapped.
const chatPatches = `{"kind":0,"v":{"requests":[{"timestamp":1000,"response":[{"kind":"toolInvocationSerialized","toolId":"run_in_terminal","toolCallId":"call-1"}]}]}}
{"kind":2,"k":["requests",0,"response"],"i":0,"v":[{"kind":"toolInvocationSerialized","toolId":"run_in_terminal","toolCallId":"call-1","toolSpecificData":{"commandLine":{"isSandboxWrapped":true}}}]}
{"kind":2,"k":["requests"],"v":[{"timestamp":2000,"response":[]},{"timestamp":3000,"response":[]}]}
{"kind":2,"k":["requests"],"i":2,"v":[{"timestamp":4000,"response":[]}]}
{"kind":2,"k":["requests",1,"response"],"v":[{"kind":"toolInvocationSerialized","toolId":"run_in_terminal","toolCallId":"call-2","toolSpecificData":{"commandLine":{"isSandboxWrapped":false}}}]}
{"kind":2,"k":["requests",2,"response"],"v":[{"kind":"toolInvocationSerialized","toolId":"run_in_terminal","toolCallId":"call-3"},{"kind":"toolInvocationSerialized","toolId":"read_file","toolCallId":"call-4"}]}
{"kind":2,"k":["requests",0,"response"],"i":0,"v":[{"kind":"toolInvocationSerialized","toolId":"run_in_terminal","toolCallId":"call-1"}]}
`

func TestReadVSCodeTerminalExecutions(t *testing.T) {
	hashDir := t.TempDir()
	writeChatSession(t, hashDir, "s", chatPatches)
	got, err := readVSCodeTerminalExecutions(filepath.Join(hashDir, "chatSessions", "s.jsonl"))
	if err != nil {
		t.Fatalf("readVSCodeTerminalExecutions() error = %v", err)
	}
	want := []vscodeTerminalExecution{
		{tool: "run_in_terminal", state: SandboxApplied, requestedAt: time.UnixMilli(1000)},
		{tool: "run_in_terminal", state: SandboxNotApplied, requestedAt: time.UnixMilli(2000)},
		{tool: "run_in_terminal", state: SandboxUnknown, requestedAt: time.UnixMilli(4000)},
	}
	if len(got) != len(want) {
		t.Fatalf("executions = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i].tool != want[i].tool || got[i].state != want[i].state || !got[i].requestedAt.Equal(want[i].requestedAt) {
			t.Fatalf("executions[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}

	missing, err := readVSCodeTerminalExecutions(filepath.Join(hashDir, "chatSessions", "missing.jsonl"))
	if err != nil || missing != nil {
		t.Fatalf("missing file = %+v, %v; want nil, nil", missing, err)
	}
}

func TestCollectVSCodeStatsSandbox(t *testing.T) {
	root := t.TempDir()
	hashA := writeVSCodeWorkspace(t, root, "hash-a", `{"folder":"file:///repo/a"}`)
	writeVSCodeSessionLog(t, hashA, "session-a", `{"type":"tool_call","ts":1001,"name":"run_in_terminal","status":"ok","dur":10}
`)
	writeChatSession(t, hashA, "session-a", chatPatches)

	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root, Tool: []string{"^nothing$"}})
	if err != nil {
		t.Fatalf("CollectVSCodeStats() error = %v", err)
	}
	sb := stats.Sandbox
	if sb.Sessions != 1 || sb.Executions != 3 || sb.Applied != 1 || sb.NotApplied != 1 || sb.Unknown != 1 {
		t.Fatalf("Sandbox = %+v, want 3 executions unaffected by the tool filter", sb)
	}
	if stats.ToolCalls != 0 {
		t.Fatalf("ToolCalls = %d, want 0 under the tool filter", stats.ToolCalls)
	}

	stats, err = CollectVSCodeStats(VSCodeStatsOptions{Root: root, Sandbox: []string{SandboxApplied}, Since: time.UnixMilli(500)})
	if err != nil {
		t.Fatalf("CollectVSCodeStats() error = %v", err)
	}
	if stats.Sandbox.Executions != 1 || stats.Sandbox.Applied != 1 || stats.ToolCalls != 1 {
		t.Fatalf("filtered Sandbox = %+v ToolCalls = %d, want 1 applied and unfiltered tools", stats.Sandbox, stats.ToolCalls)
	}

	stats, err = CollectVSCodeStats(VSCodeStatsOptions{Root: root, Since: time.UnixMilli(1500), Until: time.UnixMilli(3000)})
	if err != nil {
		t.Fatalf("CollectVSCodeStats() error = %v", err)
	}
	if stats.Sandbox.Executions != 1 || stats.Sandbox.NotApplied != 1 {
		t.Fatalf("windowed Sandbox = %+v, want only call-2", stats.Sandbox)
	}
}
