package copilotext

import (
	"testing"
)

func TestCollectVSCodeStatsToolsAndModels(t *testing.T) {
	root := t.TempDir()
	hashA := writeVSCodeWorkspace(t, root, "hash-a", `{"folder":"file:///repo/a"}`)

	log := `{"type":"session_start","ts":1000}
{"type":"turn_start","ts":1000,"attrs":{"turnId":"0"}}
{"type":"tool_call","ts":1001,"name":"read_file","status":"ok","dur":10}
{"type":"tool_call","ts":1002,"name":"read_file","status":"ok","dur":20}
{"type":"tool_call","ts":1003,"name":"run_command","status":"error","dur":5}
{"type":"llm_request","ts":1004,"attrs":{"model":"gpt-5.5","inputTokens":100,"outputTokens":20,"cachedTokens":5,"ttft":50,"copilotUsageNanoAiu":2000000000}}
{"type":"llm_request","ts":1005,"attrs":{"model":"gpt-5.5","inputTokens":200,"outputTokens":40,"cachedTokens":10,"ttft":150,"copilotUsageNanoAiu":1000000000}}
{"type":"child_session_ref","ts":1006,"attrs":{"label":"runSubagent-Explore"}}
{"type":"child_session_ref","ts":1007,"attrs":{"label":"title"}}
{"type":"turn_end","ts":1008,"attrs":{"turnId":"0"}}
`
	writeVSCodeSessionLog(t, hashA, "session-a", log)

	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root})
	if err != nil {
		t.Fatalf("CollectVSCodeStats() error = %v", err)
	}

	if stats.Sessions != 1 {
		t.Fatalf("Sessions = %d, want 1", stats.Sessions)
	}
	if stats.Turns != 1 {
		t.Fatalf("Turns = %d, want 1", stats.Turns)
	}
	if stats.ToolCalls != 3 {
		t.Fatalf("ToolCalls = %d, want 3", stats.ToolCalls)
	}
	if stats.LLMRequests != 2 {
		t.Fatalf("LLMRequests = %d, want 2", stats.LLMRequests)
	}

	if len(stats.ByTool) != 2 {
		t.Fatalf("ByTool len = %d, want 2: %+v", len(stats.ByTool), stats.ByTool)
	}
	readFile := stats.ByTool[0]
	if readFile.Key != "read_file" || readFile.Total != 2 || readFile.OK != 2 || readFile.Error != 0 {
		t.Fatalf("ByTool[read_file] = %+v, want Total=2 OK=2 Error=0", readFile)
	}
	if readFile.AvgMs != 15 {
		t.Fatalf("ByTool[read_file].AvgMs = %v, want 15", readFile.AvgMs)
	}
	runCommand := stats.ByTool[1]
	if runCommand.Key != "run_command" || runCommand.Total != 1 || runCommand.Error != 1 {
		t.Fatalf("ByTool[run_command] = %+v, want Total=1 Error=1", runCommand)
	}

	if len(stats.ByModel) != 1 {
		t.Fatalf("ByModel len = %d, want 1: %+v", len(stats.ByModel), stats.ByModel)
	}
	model := stats.ByModel[0]
	if model.Key != "gpt-5.5" || model.Requests != 2 {
		t.Fatalf("ByModel[0] = %+v, want Key=gpt-5.5 Requests=2", model)
	}
	if model.InputTokens != 300 || model.OutputTokens != 60 || model.CachedTokens != 15 {
		t.Fatalf("ByModel[0] token sums = %+v, want Input=300 Output=60 Cached=15", model)
	}
	if model.AvgTTFTMs != 100 {
		t.Fatalf("ByModel[0].AvgTTFTMs = %v, want 100", model.AvgTTFTMs)
	}
	if model.UsageAIU != 3 {
		t.Fatalf("ByModel[0].UsageAIU = %v, want 3 (2e9+1e9 nano-AIU)", model.UsageAIU)
	}

	// The "title" child session must be excluded from the subagent axis, leaving only
	// the runSubagent-prefixed invocation.
	if len(stats.ByAgent) != 1 || stats.ByAgent[0].Key != "Explore" || stats.ByAgent[0].Total != 1 {
		t.Fatalf("ByAgent = %+v, want [{Explore 1}]", stats.ByAgent)
	}

	if len(stats.ByWorkspace) != 1 || stats.ByWorkspace[0].Key != "/repo/a" {
		t.Fatalf("ByWorkspace = %+v, want one entry for /repo/a", stats.ByWorkspace)
	}
	ws := stats.ByWorkspace[0]
	if ws.ToolCalls != 3 || ws.LLMRequests != 2 || ws.Turns != 1 {
		t.Fatalf("ByWorkspace[0] = %+v, want ToolCalls=3 LLMRequests=2 Turns=1", ws)
	}
}

func TestCollectVSCodeStatsScope(t *testing.T) {
	root := t.TempDir()
	hashA := writeVSCodeWorkspace(t, root, "hash-a", `{"folder":"file:///repo/a"}`)
	writeVSCodeSessionLog(t, hashA, "session-a", `{"type":"tool_call","ts":1,"name":"read_file","status":"ok"}`+"\n")

	hashB := writeVSCodeWorkspace(t, root, "hash-b", `{"folder":"file:///repo/b"}`)
	writeVSCodeSessionLog(t, hashB, "session-b", `{"type":"tool_call","ts":1,"name":"write_file","status":"ok"}`+"\n")

	stats, err := CollectVSCodeStats(VSCodeStatsOptions{
		Root:  root,
		Scope: SessionScope{ExactCWD: "/repo/a"},
	})
	if err != nil {
		t.Fatalf("CollectVSCodeStats() error = %v", err)
	}
	if stats.Sessions != 1 || stats.ToolCalls != 1 {
		t.Fatalf("scoped stats = %+v, want Sessions=1 ToolCalls=1", stats)
	}
	if len(stats.ByTool) != 1 || stats.ByTool[0].Key != "read_file" {
		t.Fatalf("scoped ByTool = %+v, want only read_file", stats.ByTool)
	}
}

func TestCollectVSCodeStatsToolFilter(t *testing.T) {
	root := t.TempDir()
	hashA := writeVSCodeWorkspace(t, root, "hash-a", `{"folder":"file:///repo/a"}`)
	log := `{"type":"tool_call","ts":1,"name":"read_file","status":"ok"}
{"type":"tool_call","ts":2,"name":"write_file","status":"ok"}
`
	writeVSCodeSessionLog(t, hashA, "session-a", log)

	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root, Tool: []string{"^read_"}})
	if err != nil {
		t.Fatalf("CollectVSCodeStats() error = %v", err)
	}
	if len(stats.ByTool) != 1 || stats.ByTool[0].Key != "read_file" {
		t.Fatalf("ByTool = %+v, want only read_file", stats.ByTool)
	}
	if stats.ToolCalls != 1 {
		t.Fatalf("ToolCalls = %d, want 1", stats.ToolCalls)
	}
}

func TestCollectVSCodeStatsMissingRoot(t *testing.T) {
	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: "does-not-exist"})
	if err != nil {
		t.Fatalf("CollectVSCodeStats() error = %v", err)
	}
	if stats.Sessions != 0 {
		t.Fatalf("Sessions = %d, want 0", stats.Sessions)
	}
}
