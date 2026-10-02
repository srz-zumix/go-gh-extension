package render

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/srz-zumix/go-gh-extension/pkg/copilotext"
)

func TestRenderCopilotPermissionStatsTable(t *testing.T) {
	sr := NewStringRenderer(nil)
	stats := &copilotext.PermissionStats{
		Sessions: 2,
		Requests: 3,
		ByResult: []copilotext.Count{
			{Key: "approved", Total: 2, Approved: 2},
			{Key: "denied", Total: 1, Denied: 1},
		},
		ByDecisionSource: []copilotext.Count{
			{Key: "human_response", Total: 2, Approved: 2},
			{Key: "unattended_fallback", Total: 1, Denied: 1},
		},
		ByCommand: []copilotext.Count{
			{Key: "ls", Total: 2, Approved: 2},
		},
		UsageSessions:        1,
		UsagePremiumRequests: 3.5,
		UsageAIU:             1.5,
		ByCWDUsage: []copilotext.UsageCount{
			{Key: "/repo/a", Sessions: 1, PremiumRequests: 3, AIU: 1.5, InputTokens: 10, OutputTokens: 20},
		},
		ByModelUsage: []copilotext.UsageCount{
			{Key: "model-a", Sessions: 1, Requests: 2, PremiumRequests: 3, AIU: 1.5, InputTokens: 10, OutputTokens: 20},
		},
	}

	if err := sr.Renderer.RenderCopilotPermissionStats(stats); err != nil {
		t.Fatalf("RenderCopilotPermissionStats() error = %v", err)
	}

	out := sr.Stdout.String()
	if !strings.Contains(out, "SUMMARY") || !strings.Contains(out, "sessions: 2, requests: 3") {
		t.Fatalf("output missing SUMMARY line: %q", out)
	}
	if !strings.Contains(out, "usage_sessions: 1, premium_requests: 3.5,") || !strings.Contains(out, "aiu: 1.500") {
		t.Fatalf("output missing usage SUMMARY line: %q", out)
	}
	if !strings.Contains(out, "RESULT") || !strings.Contains(out, "approved") || !strings.Contains(out, "denied") {
		t.Fatalf("output missing RESULT table: %q", out)
	}
	if !strings.Contains(out, "DECISION_SOURCE") || !strings.Contains(out, "human_response") {
		t.Fatalf("output missing DECISION_SOURCE table: %q", out)
	}
	if !strings.Contains(out, "COMMAND") || !strings.Contains(out, "ls") {
		t.Fatalf("output missing COMMAND table: %q", out)
	}
	if !strings.Contains(out, "CWD_USAGE") || !strings.Contains(out, "/repo/a") {
		t.Fatalf("output missing CWD_USAGE table: %q", out)
	}
	if !strings.Contains(out, "MODEL_USAGE") || !strings.Contains(out, "model-a") || !strings.Contains(out, "REQUESTS") {
		t.Fatalf("output missing MODEL_USAGE table: %q", out)
	}
	// Empty axes (KIND, READONLY, PATH, URL, CWD) must not print a section heading.
	if strings.Contains(out, "KIND") {
		t.Fatalf("output should omit empty KIND section: %q", out)
	}
}

func TestRenderCopilotPermissionStatsJSON(t *testing.T) {
	sr := NewStringRenderer(cmdutil.NewJSONExporter())
	stats := &copilotext.PermissionStats{
		Sessions: 1, Requests: 1,
		ByModelUsage: []copilotext.UsageCount{{Key: "model-a", Sessions: 1, Requests: 2, PremiumRequests: 3, AIU: 1.5}},
	}

	if err := sr.Renderer.RenderCopilotPermissionStats(stats); err != nil {
		t.Fatalf("RenderCopilotPermissionStats() error = %v", err)
	}

	out := sr.Stdout.String()
	if !strings.Contains(out, `"Sessions":1`) {
		t.Fatalf("expected JSON export of PermissionStats, got %q", out)
	}
	var decoded copilotext.PermissionStats
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(decoded.ByModelUsage) != 1 || decoded.ByModelUsage[0] != stats.ByModelUsage[0] {
		t.Fatalf("ByModelUsage = %+v, want %+v", decoded.ByModelUsage, stats.ByModelUsage)
	}
}
