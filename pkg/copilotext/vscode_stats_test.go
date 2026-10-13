package copilotext

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCollectVSCodeStatsLogDetails(t *testing.T) {
	root := t.TempDir()
	hash := writeVSCodeWorkspace(t, root, "hash", `{"folder":"file:///repo/a"}`)
	main := writeVSCodeSessionLog(t, hash, "parent", `{"type":"turn_start","ts":1000,"attrs":{"turnId":"main"}}
{"type":"llm_request","ts":1001,"attrs":{"model":"main","ttft":100}}
{"type":"child_session_ref","ts":1002,"attrs":{"label":"Other","childLogFile":"shared.jsonl"}}
{"type":"child_session_ref","ts":1003,"attrs":{"label":"Explore","childLogFile":"shared.jsonl"}}
`)
	writeVSCodeChildLog(t, main, "shared.jsonl", `{"type":"turn_start","ts":2000,"attrs":{"turnId":"child"}}
{"type":"llm_request","ts":2001,"attrs":{"model":"a","ttft":20,"inputTokens":100}}
{"type":"llm_request","ts":2002,"attrs":{"model":"b","ttft":0,"inputTokens":200}}
{"type":"tool_call","ts":2003,"name":"read","status":"ok","dur":10}
{"type":"child_session_ref","ts":2004,"attrs":{"label":"Explore"}}
`)
	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root, IncludeLogStats: true, Top: 1})
	if err != nil {
		t.Fatal(err)
	}
	if stats.LogDetails == nil || stats.LogDetails.Version != 1 || len(stats.LogDetails.Sessions) != 1 || len(stats.ByModel) != 1 {
		t.Fatalf("dataset = %+v, want versioned details plus normal truncated axes", stats)
	}
	session := stats.LogDetails.Sessions[0]
	if session.ID != "parent" || session.Workspace != "/repo/a" || len(session.Logs) != 2 || len(session.Invocations) != 3 {
		t.Fatalf("session = %+v, want parent provenance, unique logs, and raw invocation identities", session)
	}
	if !session.Logs[0].Main || session.Logs[1].Main {
		t.Fatal("main log classification is incorrect")
	}
	child := session.Logs[1]
	if len(child.Agents) != 2 || child.Agents[0] != "Other" || child.Agents[1] != "Explore" ||
		child.LLMRequests != 2 || child.ToolCalls != 1 || child.Turns != 1 || len(child.ByModel) != 2 {
		t.Fatalf("child = %+v, want all referencing names and untruncated own metrics", child)
	}
	if child.ByModel[0].TTFTSamples != 1 || child.ByModel[0].TotalTTFTMs != 20 ||
		child.ByModel[1].TTFTSamples != 0 || child.ByModel[1].TotalTTFTMs != 0 {
		t.Fatalf("models = %+v, want only positive TTFT samples", child.ByModel)
	}
	if session.Invocations[0].ID != "" || session.Invocations[1].Agent != "Other" || session.Invocations[2].Agent != "Explore" {
		t.Fatalf("invocations = %+v, want reference order before aggregate deduplication", session.Invocations)
	}
	plain, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(plain)
	if err != nil {
		t.Fatal(err)
	}
	if plain.LogDetails != nil || strings.Contains(string(payload), "LogDetails") {
		t.Fatal("default output must not include log details")
	}
	if _, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root, IncludeLogStats: true, Agent: []string{"Explore"}}); err == nil {
		t.Fatal("log details and server-side agent filtering must be incompatible")
	}
}

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
	if ws.Sessions != 1 || ws.ToolCalls != 3 || ws.LLMRequests != 2 || ws.Turns != 1 {
		t.Fatalf("ByWorkspace[0] = %+v, want Sessions=1 ToolCalls=3 LLMRequests=2 Turns=1", ws)
	}
	if ws.InputTokens != 300 || ws.OutputTokens != 60 || ws.CachedTokens != 15 || ws.UsageAIU != 3 {
		t.Fatalf("ByWorkspace[0] usage = %+v, want Input=300 Output=60 Cached=15 UsageAIU=3", ws)
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

func TestCollectVSCodeStatsChildLogs(t *testing.T) {
	root := t.TempDir()
	hash := writeVSCodeWorkspace(t, root, "hash", `{"folder":"file:///repo/a"}`)
	mainPath := writeVSCodeSessionLog(t, hash, "parent", `{"type":"turn_start","ts":1000,"attrs":{"turnId":"0"}}
	{"type":"llm_request","ts":1001,"attrs":{"model":"model-main","inputTokens":100,"outputTokens":10,"copilotUsageNanoAiu":1000000000}}
	{"type":"child_session_ref","ts":1002,"attrs":{"label":"runSubagent-Explore","childSessionId":"child","childLogFile":"child.jsonl"}}
	{"type":"child_session_ref","ts":1003,"attrs":{"label":"runSubagent-Explore","childSessionId":"child","childLogFile":"child.jsonl"}}
	{"type":"child_session_ref","ts":1004,"attrs":{"label":"title","childLogFile":"title.jsonl"}}
	`)
	writeVSCodeChildLog(t, mainPath, "child.jsonl", `{"type":"turn_start","ts":1005,"attrs":{"turnId":"0"}}
	{"type":"turn_start","ts":1006,"attrs":{"turnId":"0"}}
	{"type":"tool_call","ts":1007,"name":"read_file","status":"ok","dur":10}
	{"type":"llm_request","ts":1008,"attrs":{"model":"model-child","inputTokens":200,"outputTokens":20,"cachedTokens":50,"ttft":30,"copilotUsageNanoAiu":500000000}}
	{"type":"child_session_ref","ts":1009,"attrs":{"label":"executionSubagent","childSessionId":"nested","childLogFile":"nested.jsonl"}}
	`)
	writeVSCodeChildLog(t, mainPath, "nested.jsonl", `{"type":"turn_start","ts":1010,"attrs":{"turnId":"0"}}
	{"type":"tool_call","ts":1011,"name":"read_file","status":"error","dur":20}
	{"type":"llm_request","ts":1012,"attrs":{"model":"model-child","inputTokens":300,"outputTokens":30,"cachedTokens":100,"ttft":50,"copilotUsageNanoAiu":1500000000}}
	{"type":"child_session_ref","ts":1013,"attrs":{"label":"runSubagent-Explore","childLogFile":"child.jsonl"}}
	`)
	writeVSCodeChildLog(t, mainPath, "title.jsonl", `{"type":"llm_request","ts":1014,"attrs":{"model":"model-title","inputTokens":999}}`+"\n")
	writeVSCodeChildLog(t, mainPath, "orphan.jsonl", `{"type":"llm_request","ts":1015,"attrs":{"model":"model-orphan","inputTokens":999}}`+"\n")

	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Sessions != 1 || stats.Turns != 3 || stats.ToolCalls != 2 || stats.LLMRequests != 3 {
		t.Fatalf("stats = %+v, want Sessions=1 Turns=3 ToolCalls=2 LLMRequests=3", stats)
	}
	if len(stats.ByModel) != 2 {
		t.Fatalf("ByModel = %+v, want only main and child models", stats.ByModel)
	}
	child := stats.ByModel[0]
	if child.Key != "model-child" || child.Requests != 2 || child.InputTokens != 500 ||
		child.OutputTokens != 50 || child.CachedTokens != 150 || child.UsageAIU != 2 || child.AvgTTFTMs != 40 {
		t.Fatalf("child model = %+v, want Requests=2 Input=500 Output=50 Cached=150 AIU=2 TTFT=40", child)
	}
	if len(stats.ByTool) != 1 || stats.ByTool[0].Total != 2 || stats.ByTool[0].OK != 1 || stats.ByTool[0].Error != 1 {
		t.Fatalf("ByTool = %+v, want read_file Total=2 OK=1 Error=1", stats.ByTool)
	}
	if len(stats.ByAgent) != 2 || stats.ByAgent[0].Total != 1 || stats.ByAgent[1].Total != 1 {
		t.Fatalf("ByAgent = %+v, want one Explore and one executionSubagent invocation", stats.ByAgent)
	}
	if len(stats.ByWorkspace) != 1 {
		t.Fatalf("ByWorkspace = %+v, want one parent workspace", stats.ByWorkspace)
	}
	ws := stats.ByWorkspace[0]
	if ws.Key != "/repo/a" || ws.Sessions != 1 || ws.Turns != 3 || ws.ToolCalls != 2 ||
		ws.LLMRequests != 3 || ws.InputTokens != 600 || ws.OutputTokens != 60 || ws.CachedTokens != 150 || ws.UsageAIU != 3 {
		t.Fatalf("workspace = %+v, want combined parent and child activity in one session", ws)
	}
	stats, err = CollectVSCodeStats(VSCodeStatsOptions{Root: root, Top: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(stats.ByModel) != 1 || stats.ByModel[0].Key != "model-child" || stats.LLMRequests != 3 {
		t.Fatalf("top model stats = %+v, want child model without truncating overall requests", stats)
	}
	for _, test := range []struct {
		name     string
		agents   []string
		requests int
		input    int64
		output   int64
		cached   int64
		aiu      float64
	}{
		{"selected agent excludes differently named descendants", []string{"^Explore$"}, 1, 200, 20, 50, 0.5},
		{"matching descendant through unselected parent", []string{"^executionSubagent$"}, 1, 300, 30, 100, 1.5},
		{"multiple agent patterns", []string{"^Explore$", "^executionSubagent$"}, 2, 500, 50, 150, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root, Agent: test.agents})
			if err != nil {
				t.Fatal(err)
			}
			if stats.Sessions != 1 || stats.LLMRequests != test.requests || stats.Turns != test.requests ||
				stats.ToolCalls != test.requests || len(stats.ByModel) != 1 || len(stats.ByWorkspace) != 1 {
				t.Fatalf("selected activity = %+v, want only matching child logs counted once", stats)
			}
			model := stats.ByModel[0]
			if model.Key != "model-child" || model.Requests != test.requests || model.InputTokens != test.input ||
				model.OutputTokens != test.output || model.CachedTokens != test.cached || model.UsageAIU != test.aiu {
				t.Fatalf("selected model = %+v, want matching child usage only", model)
			}
			ws := stats.ByWorkspace[0]
			if ws.Key != "/repo/a" || ws.Sessions != 1 || ws.LLMRequests != test.requests ||
				ws.ToolCalls != test.requests || ws.Turns != test.requests || ws.InputTokens != test.input ||
				ws.OutputTokens != test.output || ws.CachedTokens != test.cached || ws.UsageAIU != test.aiu {
				t.Fatalf("selected workspace = %+v, want parent attribution without parent usage", ws)
			}
			if len(stats.ByAgent) != len(test.agents) {
				t.Fatalf("agents = %+v, want only selected names", stats.ByAgent)
			}
			for _, agent := range stats.ByAgent {
				if agent.Total != 1 {
					t.Fatalf("agent = %+v, want one invocation despite cyclic references", agent)
				}
			}
		})
	}
}

func TestCollectVSCodeStatsChildLogFilters(t *testing.T) {
	root := t.TempDir()
	hash := writeVSCodeWorkspace(t, root, "hash", `{"folder":"file:///repo/a"}`)
	mainPath := writeVSCodeSessionLog(t, hash, "parent", `{"type":"llm_request","ts":1000,"attrs":{"model":"model-main"}}
	{"type":"child_session_ref","ts":1001,"attrs":{"label":"runSubagent-Explore","childLogFile":"child.jsonl"}}
	`)
	writeVSCodeChildLog(t, mainPath, "child.jsonl", `{"type":"turn_start","ts":2000,"attrs":{"turnId":"0"}}
	{"type":"tool_call","ts":2000,"name":"read_file","status":"ok"}
	{"type":"tool_call","ts":2001,"name":"write_file","status":"ok"}
	{"type":"llm_request","ts":2000,"attrs":{"model":"model-child","inputTokens":200,"copilotUsageNanoAiu":500000000}}
	{"type":"llm_request","ts":3000,"attrs":{"model":"model-child","inputTokens":300,"copilotUsageNanoAiu":1000000000}}
	`)
	old := time.UnixMilli(1000)
	if err := os.Chtimes(mainPath, old, old); err != nil {
		t.Fatal(err)
	}
	writeVSCodeSessionLog(t, hash, "other", `{"type":"llm_request","ts":2000,"attrs":{"model":"model-other"}}`+"\n")

	tests := []struct {
		name        string
		opts        VSCodeStatsOptions
		llm, tools  int
		agents      int
		inputTokens int64
	}{
		{"time window", VSCodeStatsOptions{Since: time.UnixMilli(2000), Until: time.UnixMilli(3000)}, 1, 2, 0, 200},
		{"since child-only usage", VSCodeStatsOptions{Since: time.UnixMilli(3000)}, 1, 0, 0, 300},
		{"until", VSCodeStatsOptions{Until: time.UnixMilli(3000)}, 2, 2, 1, 200},
		{"model", VSCodeStatsOptions{Model: []string{"^model-child$"}}, 2, 2, 1, 500},
		{"parent model", VSCodeStatsOptions{Model: []string{"^model-main$"}}, 1, 2, 1, 0},
		{"nonmatching agent", VSCodeStatsOptions{Tool: []string{"^read_file$"}, Agent: []string{"^Other$"}}, 0, 0, 0, 0},
		{"matching agent", VSCodeStatsOptions{Agent: []string{"^Explore$"}}, 2, 2, 1, 500},
		{"tool within agent scope", VSCodeStatsOptions{Tool: []string{"^read_file$"}, Agent: []string{"^Explore$"}}, 2, 1, 1, 500},
		{"model within agent scope", VSCodeStatsOptions{Model: []string{"^model-main$"}, Agent: []string{"^Explore$"}}, 0, 2, 1, 0},
		{"agent usage after older invocation", VSCodeStatsOptions{Since: time.UnixMilli(3000), Agent: []string{"^Explore$"}}, 1, 0, 0, 300},
		{"scope excludes parent", VSCodeStatsOptions{Scope: SessionScope{ExactCWD: "/repo/b"}}, 0, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Root = root
			tt.opts.SessionID = "parent"
			stats, err := CollectVSCodeStats(tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			if stats.LLMRequests != tt.llm || stats.ToolCalls != tt.tools || len(stats.ByAgent) != tt.agents {
				t.Fatalf("stats = %+v, want LLMRequests=%d ToolCalls=%d agent entries=%d", stats, tt.llm, tt.tools, tt.agents)
			}
			if tt.llm == 0 && tt.tools == 0 && tt.agents == 0 {
				if stats.Sessions != 0 || len(stats.ByWorkspace) != 0 {
					t.Fatalf("excluded parent stats = %+v, want no sessions or workspaces", stats)
				}
			} else if stats.Sessions != 1 || len(stats.ByWorkspace) != 1 || stats.ByWorkspace[0].InputTokens != tt.inputTokens {
				t.Fatalf("workspace stats = %+v, want one session with input tokens=%d", stats.ByWorkspace, tt.inputTokens)
			}
		})
	}
}

func TestCollectVSCodeStatsAgentScopeSharedLogs(t *testing.T) {
	root := t.TempDir()
	hash := writeVSCodeWorkspace(t, root, "hash", `{"folder":"file:///repo/a"}`)
	mainPath := writeVSCodeSessionLog(t, hash, "parent", `{"type":"llm_request","ts":1000,"attrs":{"model":"model-main"}}
	{"type":"child_session_ref","ts":1001,"attrs":{"label":"Other","childLogFile":"shared.jsonl"}}
	{"type":"child_session_ref","ts":1002,"attrs":{"label":"Explore","childLogFile":"shared.jsonl"}}
	`)
	writeVSCodeChildLog(t, mainPath, "shared.jsonl", `{"type":"turn_start","ts":2000,"attrs":{"turnId":"0"}}
	{"type":"tool_call","ts":2001,"name":"read_file","status":"ok"}
	{"type":"llm_request","ts":2002,"attrs":{"model":"model-child","inputTokens":100,"outputTokens":10,"cachedTokens":20,"copilotUsageNanoAiu":500000000}}
	{"type":"child_session_ref","ts":2003,"attrs":{"label":"Explore"}}
	{"type":"child_session_ref","ts":2004,"attrs":{"label":"Explore","childLogFile":"main.jsonl"}}
	`)
	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root, Agent: []string{"^Explore$"}})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Sessions != 1 || stats.Turns != 1 || stats.ToolCalls != 1 || stats.LLMRequests != 1 ||
		len(stats.ByModel) != 1 || stats.ByModel[0].Key != "model-child" || stats.ByModel[0].UsageAIU != 0.5 {
		t.Fatalf("stats = %+v, want shared child counted once without cyclic main-log usage", stats)
	}
	if len(stats.ByAgent) != 1 || stats.ByAgent[0].Key != "Explore" || stats.ByAgent[0].Total != 3 {
		t.Fatalf("agents = %+v, want references counted once even when activity is revisited", stats.ByAgent)
	}
}

func TestCollectVSCodeStatsUnavailableChildLogs(t *testing.T) {
	root := t.TempDir()
	hash := writeVSCodeWorkspace(t, root, "hash", `{"folder":"file:///repo/a"}`)
	mainPath := writeVSCodeSessionLog(t, hash, "parent", `{"type":"llm_request","ts":1000,"attrs":{"model":"model-main"}}
	{"type":"child_session_ref","ts":1001,"attrs":{"label":"Explore","childLogFile":"missing.jsonl"}}
	{"type":"child_session_ref","ts":1002,"attrs":{"label":"Explore","childLogFile":"../outside.jsonl"}}
	{"type":"child_session_ref","ts":1003,"attrs":{"label":"Explore","childLogFile":"/outside.jsonl"}}
	{"type":"child_session_ref","ts":1004,"attrs":{"label":"Explore","childLogFile":"models.json"}}
	{"type":"child_session_ref","ts":1005,"attrs":{"label":"Explore","childLogFile":"link.jsonl"}}
	`)
	outside := filepath.Join(filepath.Dir(filepath.Dir(mainPath)), "outside.jsonl")
	if err := os.WriteFile(outside, []byte(`{"type":"llm_request","ts":2000,"attrs":{"model":"model-outside"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(filepath.Dir(mainPath), "link.jsonl")); err != nil {
		t.Fatal(err)
	}
	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Sessions != 1 || stats.LLMRequests != 1 || len(stats.ByModel) != 1 || stats.ByModel[0].Key != "model-main" {
		t.Fatalf("stats = %+v, want only main log usage", stats)
	}
	if len(stats.ByAgent) != 1 || stats.ByAgent[0].Total != 5 {
		t.Fatalf("ByAgent = %+v, want all recorded invocations even without available child usage", stats.ByAgent)
	}
	stats, err = CollectVSCodeStats(VSCodeStatsOptions{Root: root, Agent: []string{"^Explore$"}})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Sessions != 1 || stats.LLMRequests != 0 || stats.ToolCalls != 0 || stats.Turns != 0 ||
		len(stats.ByModel) != 0 || len(stats.ByAgent) != 1 || stats.ByAgent[0].Total != 5 {
		t.Fatalf("filtered unavailable logs = %+v, want recorded invocations without invented usage", stats)
	}
}
