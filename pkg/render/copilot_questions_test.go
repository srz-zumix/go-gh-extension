package render

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/srz-zumix/go-gh-extension/pkg/copilotext"
)

func TestRenderQuestionStats(t *testing.T) {
	avg, max := 218075.66666666666, 452502.0
	stats := &copilotext.PermissionStats{AskUser: &copilotext.AskUserStats{
		Sessions: 1, Questions: 5, Answered: 3, Cancelled: 1, Unresolved: 1,
		TimedAnswers: 3, TotalWaitMs: 654227, AvgWaitMs: &avg, MaxWaitMs: &max,
	}}
	r := NewStringRenderer(nil)
	if err := r.Renderer.RenderCopilotPermissionStats(stats); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"ASK_USER", "questions: 5", "answered: 3", "cancelled: 1",
		"unresolved: 1", "total_answer_wait_ms: 654227.000", "avg_answer_wait_ms: 218075.667",
		"max_answer_wait_ms: 452502.000", "not unique wall-clock"} {
		if !strings.Contains(r.Stdout.String(), text) {
			t.Fatalf("missing %q: %s", text, r.Stdout.String())
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
	if decoded.AskUser.Questions != 5 || *decoded.AskUser.AvgWaitMs != avg {
		t.Fatalf("question JSON = %+v", decoded.AskUser)
	}
}

func TestRenderQuestionStatsUnavailableTiming(t *testing.T) {
	r := NewStringRenderer(nil)
	if err := r.Renderer.RenderCopilotPermissionStats(&copilotext.PermissionStats{AskUser: &copilotext.AskUserStats{Unresolved: 1}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Stdout.String(), "avg_answer_wait_ms: n/a, max_answer_wait_ms: n/a") {
		t.Fatal(r.Stdout.String())
	}
}

func TestRenderVSCodeQuestionStats(t *testing.T) {
	stats := &copilotext.VSCodeStats{UserQuestions: &copilotext.VSCodeQuestionStats{
		Sessions: 1, Calls: 3, Answered: 1, Skipped: 1, Cancelled: 1, UntimedAnswers: 1,
	}}
	r := NewStringRenderer(nil)
	if err := r.Renderer.RenderVSCodeStats(stats); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"USER_QUESTIONS", "calls: 3", "skipped: 1",
		"avg_answer_wait_ms: n/a, max_answer_wait_ms: n/a", "in-flight questions are unavailable"} {
		if !strings.Contains(r.Stdout.String(), expected) {
			t.Fatalf("missing %q: %s", expected, r.Stdout.String())
		}
	}
	r = NewStringRenderer(cmdutil.NewJSONExporter())
	if err := r.Renderer.RenderVSCodeStats(stats); err != nil {
		t.Fatal(err)
	}
	var decoded copilotext.VSCodeStats
	if err := json.Unmarshal(r.Stdout.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.UserQuestions.Calls != 3 || decoded.UserQuestions.AvgWaitMs != nil {
		t.Fatalf("VS Code question JSON = %+v", decoded.UserQuestions)
	}
}
