package render

import (
	"strings"
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/srz-zumix/go-gh-extension/pkg/copilotext"
)

func TestRenderVSCodeStatsTable(t *testing.T) {
	sr := NewStringRenderer(nil)
	stats := &copilotext.VSCodeStats{
		Sessions:    2,
		Turns:       3,
		LLMRequests: 4,
		ToolCalls:   5,
		ByTool: []copilotext.VSCodeToolCount{
			{Key: "read_file", Total: 3, OK: 3, TotalMs: 30, AvgMs: 10},
		},
		ByModel: []copilotext.VSCodeModelCount{
			{Key: "gpt-5.5", Requests: 4, InputTokens: 100, OutputTokens: 20, UsageAIU: 1.5},
		},
	}

	if err := sr.Renderer.RenderVSCodeStats(stats); err != nil {
		t.Fatalf("RenderVSCodeStats() error = %v", err)
	}

	out := sr.Stdout.String()
	if !strings.Contains(out, "SUMMARY") || !strings.Contains(out, "sessions: 2, turns: 3, llm_requests: 4, tool_calls: 5") {
		t.Fatalf("output missing SUMMARY line: %q", out)
	}
	if !strings.Contains(out, "TOOL") || !strings.Contains(out, "read_file") {
		t.Fatalf("output missing TOOL table: %q", out)
	}
	if !strings.Contains(out, "MODEL") || !strings.Contains(out, "gpt-5.5") {
		t.Fatalf("output missing MODEL table: %q", out)
	}
	// Empty axes (AGENT, WORKSPACE) must not print a section heading.
	if strings.Contains(out, "AGENT") || strings.Contains(out, "WORKSPACE") {
		t.Fatalf("output should omit empty AGENT/WORKSPACE sections: %q", out)
	}
}

func TestRenderVSCodeStatsJSON(t *testing.T) {
	sr := NewStringRenderer(cmdutil.NewJSONExporter())
	stats := &copilotext.VSCodeStats{Sessions: 1, Turns: 1}

	if err := sr.Renderer.RenderVSCodeStats(stats); err != nil {
		t.Fatalf("RenderVSCodeStats() error = %v", err)
	}

	out := sr.Stdout.String()
	if !strings.Contains(out, `"Sessions":1`) {
		t.Fatalf("expected JSON export of VSCodeStats, got %q", out)
	}
}
