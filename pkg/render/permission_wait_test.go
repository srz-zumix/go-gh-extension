package render

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/srz-zumix/go-gh-extension/pkg/copilotext"
)

func TestRenderPermissionWait(t *testing.T) {
	avg, max := 5.125, 10.25
	stats := &copilotext.PermissionStats{
		Requests: 4,
		PermissionWait: &copilotext.PermissionWaitStats{
			HumanResponses: copilotext.PermissionDecisionWait{
				Completed: 2, Timed: 2, TotalWaitMs: 10.25, AvgWaitMs: &avg, MaxWaitMs: &max,
			},
			UnknownSourceDecisions: copilotext.PermissionDecisionWait{Completed: 1, Untimed: 1},
			Unresolved:             1,
		},
	}
	r := NewStringRenderer(nil)
	if err := r.Renderer.RenderCopilotPermissionStats(stats); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"PERMISSION_WAIT", "unresolved: 1",
		"human_response: completed: 2, timed: 2, untimed: 0, total_wait_ms: 10.250, avg_wait_ms: 5.125, max_wait_ms: 10.250",
		"unknown_source: completed: 1, timed: 0, untimed: 1, total_wait_ms: 0.000, avg_wait_ms: n/a, max_wait_ms: n/a",
		"only human_response is confirmed user wait", "unresolved requests have no duration",
	} {
		if !strings.Contains(r.Stdout.String(), expected) {
			t.Fatalf("missing %q: %s", expected, r.Stdout.String())
		}
	}
	r = NewStringRenderer(cmdutil.NewJSONExporter())
	if err := r.Renderer.RenderCopilotPermissionStats(stats); err != nil {
		t.Fatal(err)
	}
	var decoded copilotext.PermissionStats
	if err := json.Unmarshal(r.Stdout.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if *decoded.PermissionWait.HumanResponses.AvgWaitMs != avg ||
		decoded.PermissionWait.UnknownSourceDecisions.AvgWaitMs != nil {
		t.Fatalf("permission timing precision/nulls were lost: %+v", decoded.PermissionWait)
	}
}

func TestRenderPermissionWaitAbsentAndEmpty(t *testing.T) {
	for _, stats := range []*copilotext.PermissionStats{
		{Requests: 1},
		{PermissionWait: &copilotext.PermissionWaitStats{}},
	} {
		r := NewStringRenderer(nil)
		if err := r.Renderer.RenderCopilotPermissionStats(stats); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(r.Stdout.String(), "PERMISSION_WAIT") {
			t.Fatalf("absent/empty stats should omit text timing: %s", r.Stdout.String())
		}
	}
}
