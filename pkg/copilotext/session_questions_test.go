package copilotext

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"
)

func questionEvent(t *testing.T, collector *questionCollector, line string) {
	t.Helper()
	var ev rawEvent
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		t.Fatal(err)
	}
	collector.handle(ev)
}

func TestQuestionOutcomes(t *testing.T) {
	cases := []struct{ data, want string }{
		{`{"success":true,"result":{"content":"User selected: Yes"}}`, "answered"},
		{`{"success":true,"result":{"content":"User responded: Cancel the deployment"}}`, "answered"},
		{`{"success":true,"result":{"content":"User cancelled the question"}}`, "cancelled"},
		{`{"success":true,"result":{"content":"User canceled the question"}}`, "cancelled"},
		{`{"success":false,"error":{"code":"CANCELLED"}}`, "cancelled"},
		{`{"success":false,"error":{"name":"AbortError"}}`, "cancelled"},
		{`{"success":false,"error":{"code":"UNAVAILABLE"}}`, "failed"},
		{`{"success":false,"error":"tool execution failed","result":{"content":[]}}`, "failed"},
		{`{"success":true,"result":{"content":"Making best guess on autopilot"}}`, "unknown"},
		{`{"result":{"content":"User selected: Yes"}}`, "unknown"},
		{`{"success":true}`, "unknown"},
		{`{"success":true,"result":{"content":[]}}`, "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.want+"/"+tc.data, func(t *testing.T) {
			var data rawQuestionToolEvent
			if err := json.Unmarshal([]byte(tc.data), &data); err != nil {
				t.Fatal(err)
			}
			if got := questionOutcome(data); got != tc.want {
				t.Fatalf("outcome = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestQuestionCollectorIdentitiesReplayAndOrdering(t *testing.T) {
	qc := newQuestionCollector(Session{})
	start := `{"type":"tool.execution_start","timestamp":"2026-01-01T00:00:00Z","data":{"toolName":"ask_user","toolCallId":"same"}}`
	end := `{"type":"tool.execution_complete","timestamp":"2026-01-01T00:00:01.250Z","data":{"toolCallId":"same","success":true,"result":{"content":"User selected: Yes"}}}`
	questionEvent(t, qc, end)
	questionEvent(t, qc, start)
	questionEvent(t, qc, start)
	questionEvent(t, qc, end)
	questionEvent(t, qc, `{"type":"tool.execution_start","agentId":"child","timestamp":"2026-01-01T00:00:00Z","data":{"toolName":"ask_user","toolCallId":"same"}}`)
	questionEvent(t, qc, `{"type":"tool.execution_start","timestamp":"2026-01-01T00:00:00Z","data":{"toolName":"view","toolCallId":"unrelated"}}`)
	questionEvent(t, qc, `{"type":"tool.execution_complete","timestamp":"2026-01-01T00:00:02Z","data":{"toolCallId":"unrelated","success":true,"result":{"content":["not a question"]}}}`)
	questionEvent(t, qc, `{"type":"tool.execution_start","timestamp":"2026-01-01T00:00:00Z","data":{"toolName":"ask_user"}}`)
	questionEvent(t, qc, `{"type":"tool.execution_complete","timestamp":"2026-01-01T00:00:01Z","data":{"toolCallId":"orphan","success":true}}`)
	records := qc.records()
	if len(records) != 2 || records[0].outcome != "answered" || records[1].outcome != "unresolved" {
		t.Fatalf("records = %+v", records)
	}
	var stats AskUserStats
	for _, rec := range records {
		stats.add(rec)
	}
	if stats.Questions != 2 || stats.Answered != 1 || stats.Unresolved != 1 ||
		stats.TimedAnswers != 1 || stats.TotalWaitMs != 1250 ||
		stats.AvgWaitMs == nil || *stats.AvgWaitMs != 1250 || *stats.MaxWaitMs != 1250 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestQuestionTimingUsesOnlyValidAnsweredSamples(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var stats AskUserStats
	for _, rec := range []questionRecord{
		{startedAt: start, endedAt: start.Add(10 * time.Second), outcome: "answered"},
		{startedAt: start, endedAt: start.Add(20 * time.Second), outcome: "answered"},
		{startedAt: start, endedAt: start, outcome: "answered"},
		{startedAt: start, endedAt: start.Add(-time.Second), outcome: "answered"},
		{startedAt: start, outcome: "answered"},
		{endedAt: start, outcome: "answered"},
		{startedAt: start, endedAt: start.Add(time.Hour), outcome: "cancelled"},
		{startedAt: start, endedAt: start.Add(time.Hour), outcome: "failed"},
		{startedAt: start, outcome: "unresolved"},
		{startedAt: start, endedAt: start.Add(time.Hour), outcome: "unknown"},
	} {
		stats.add(rec)
	}
	if stats.Questions != 10 || stats.Answered != 6 || stats.TimedAnswers != 3 ||
		stats.UntimedAnswers != 3 || stats.Cancelled != 1 || stats.Failed != 1 ||
		stats.Unknown != 1 || stats.Unresolved != 1 || stats.TotalWaitMs != 30000 ||
		*stats.AvgWaitMs != 10000 || *stats.MaxWaitMs != 20000 {
		t.Fatalf("stats = %+v", stats)
	}
	var empty AskUserStats
	empty.add(questionRecord{outcome: "unresolved"})
	if empty.AvgWaitMs != nil || empty.MaxWaitMs != nil {
		t.Fatalf("unresolved question invented timing: %+v", empty)
	}
}

func TestCollectQuestionStatsScopeTimeAndPermissionIndependence(t *testing.T) {
	root := t.TempDir()
	events := ""
	for i, wait := range []float64{9.866, 191.859, 452.502} {
		start := time.Date(2026, 1, 1, 0, i, 0, 0, time.UTC)
		end := start.Add(time.Duration(math.Round(wait * 1e9)))
		events += fmt.Sprintf("{\"type\":\"tool.execution_start\",\"timestamp\":%q,\"data\":{\"toolName\":\"ask_user\",\"toolCallId\":%q}}\n", start.Format(time.RFC3339Nano), fmt.Sprint(i))
		events += fmt.Sprintf("{\"type\":\"tool.execution_complete\",\"timestamp\":%q,\"data\":{\"toolCallId\":%q,\"success\":true,\"result\":{\"content\":\"User selected: Yes\"}}}\n", end.Format(time.RFC3339Nano), fmt.Sprint(i))
	}
	for _, s := range []struct{ id, cwd string }{{"a", "/repo/a"}, {"b", "/repo/b"}} {
		writeSessionDir(t, root, s.id, fmt.Sprintf("id: %s\ncwd: %s\n", s.id, s.cwd), events)
	}
	opts := PermissionStatsOptions{Root: root, Scope: SessionScope{ExactCWD: "/repo/a"}, Kind: []string{"never"}, Top: 1}
	stats, err := CollectPermissionStats(opts)
	if err != nil {
		t.Fatal(err)
	}
	q := stats.AskUser
	if stats.Sessions != 0 || stats.Requests != 0 || q.Sessions != 1 || q.Questions != 3 || q.Answered != 3 ||
		math.Abs(q.TotalWaitMs-654227) > 1e-6 || math.Abs(*q.AvgWaitMs-218075.66666666666) > 1e-6 || *q.MaxWaitMs != 452502 {
		t.Fatalf("stats = %+v, questions = %+v", stats, q)
	}
	opts.Since = time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)
	opts.Until = time.Date(2026, 1, 1, 0, 2, 0, 0, time.UTC)
	stats, err = CollectPermissionStats(opts)
	if err != nil || stats.AskUser.Questions != 1 || stats.AskUser.TotalWaitMs != 191859 {
		t.Fatalf("start-time bounds lost an answer completed after Until: %+v, %v", stats, err)
	}
	opts.Scope = SessionScope{}
	opts.SessionID = "b"
	stats, err = CollectPermissionStats(opts)
	if err != nil || stats.AskUser.Sessions != 1 || stats.AskUser.Questions != 1 {
		t.Fatalf("session filter failed: %+v, %v", stats, err)
	}
	opts.SessionID = "absent"
	stats, err = CollectPermissionStats(opts)
	if err != nil || stats.AskUser.Questions != 0 || stats.AskUser.AvgWaitMs != nil {
		t.Fatalf("empty result should retain known zero counts: %+v, %v", stats, err)
	}
}
