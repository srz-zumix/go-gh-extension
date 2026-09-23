package render

import (
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
		ByCommand: []copilotext.Count{
			{Key: "ls", Total: 2, Approved: 2},
		},
	}

	if err := sr.Renderer.RenderCopilotPermissionStats(stats); err != nil {
		t.Fatalf("RenderCopilotPermissionStats() error = %v", err)
	}

	out := sr.Stdout.String()
	if !strings.Contains(out, "SUMMARY") || !strings.Contains(out, "sessions: 2, requests: 3") {
		t.Fatalf("output missing SUMMARY line: %q", out)
	}
	if !strings.Contains(out, "RESULT") || !strings.Contains(out, "approved") || !strings.Contains(out, "denied") {
		t.Fatalf("output missing RESULT table: %q", out)
	}
	if !strings.Contains(out, "COMMAND") || !strings.Contains(out, "ls") {
		t.Fatalf("output missing COMMAND table: %q", out)
	}
	// Empty axes (KIND, READONLY, PATH, URL, CWD) must not print a section heading.
	if strings.Contains(out, "KIND") {
		t.Fatalf("output should omit empty KIND section: %q", out)
	}
}

func TestRenderCopilotPermissionStatsJSON(t *testing.T) {
	sr := NewStringRenderer(cmdutil.NewJSONExporter())
	stats := &copilotext.PermissionStats{Sessions: 1, Requests: 1}

	if err := sr.Renderer.RenderCopilotPermissionStats(stats); err != nil {
		t.Fatalf("RenderCopilotPermissionStats() error = %v", err)
	}

	out := sr.Stdout.String()
	if !strings.Contains(out, `"Sessions":1`) {
		t.Fatalf("expected JSON export of PermissionStats, got %q", out)
	}
}
