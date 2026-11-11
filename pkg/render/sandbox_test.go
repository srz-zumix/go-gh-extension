package render

import (
	"strings"
	"testing"

	"github.com/srz-zumix/go-gh-extension/pkg/copilotext"
)

func TestRenderCopilotPermissionStatsSandbox(t *testing.T) {
	sr := NewStringRenderer(nil)
	stats := &copilotext.PermissionStats{
		Requests:  2,
		BySandbox: []copilotext.Count{{Key: copilotext.SandboxApplied, Total: 2, Approved: 2}},
		Sandbox: &copilotext.SandboxStats{
			Sessions: 1, Executions: 3, Applied: 2, NotApplied: 1, OptOutRequested: 1,
			ByTool: []copilotext.SandboxCount{{Key: "bash", Total: 3, Applied: 2, NotApplied: 1}},
		},
	}
	if err := sr.Renderer.RenderCopilotPermissionStats(stats); err != nil {
		t.Fatalf("RenderCopilotPermissionStats() error = %v", err)
	}
	out := sr.Stdout.String()
	for _, want := range []string{
		"SANDBOX\n", "executions: 3, applied: 2, not_applied: 1, unknown: 0, opt_out_requested: 1",
		"SANDBOX_TOOL", "SANDBOX_REQUEST",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %q", want, out)
		}
	}
}

func TestRenderVSCodeStatsSandbox(t *testing.T) {
	sr := NewStringRenderer(nil)
	stats := &copilotext.VSCodeStats{Sandbox: &copilotext.SandboxStats{
		Sessions: 1, Executions: 1, Applied: 1,
		ByTool: []copilotext.SandboxCount{{Key: "run_in_terminal", Total: 1, Applied: 1}},
	}}
	if err := sr.Renderer.RenderVSCodeStats(stats); err != nil {
		t.Fatalf("RenderVSCodeStats() error = %v", err)
	}
	out := sr.Stdout.String()
	if !strings.Contains(out, "executions: 1, applied: 1, not_applied: 0, unknown: 0\n") || strings.Contains(out, "opt_out") {
		t.Fatalf("output has unexpected SANDBOX summary: %q", out)
	}
	if !strings.Contains(out, "run_in_terminal") {
		t.Fatalf("output missing SANDBOX_TOOL table: %q", out)
	}

	sr = NewStringRenderer(nil)
	if err := sr.Renderer.RenderVSCodeStats(&copilotext.VSCodeStats{Sandbox: &copilotext.SandboxStats{}}); err != nil {
		t.Fatalf("RenderVSCodeStats() error = %v", err)
	}
	if strings.Contains(sr.Stdout.String(), "SANDBOX") {
		t.Fatalf("empty sandbox stats should be omitted: %q", sr.Stdout.String())
	}
}
