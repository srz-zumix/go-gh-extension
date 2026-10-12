package copilotext

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

func vscodeQuestionEvent(t *testing.T, status, result, failure string, duration *float64) vscodeRawEvent {
	t.Helper()
	attrs, err := json.Marshal(map[string]string{"result": result, "error": failure})
	if err != nil {
		t.Fatal(err)
	}
	return vscodeRawEvent{Name: "vscode_askQuestions", Type: "tool_call", Status: status, Attrs: attrs, Dur: duration}
}

func TestVSCodeQuestionOutcomes(t *testing.T) {
	for _, tc := range []struct{ status, result, failure, want string }{
		{"ok", `{"answers":{"one":{"selected":["Yes"],"freeText":null,"skipped":false}}}`, "", "answered"},
		{"ok", `{"answers":{"one":{"selected":[],"freeText":"Cancel the deployment","skipped":false}}}`, "", "answered"},
		{"ok", `{"answers":{"one":{"skipped":true},"two":{"skipped":false,"selected":["Yes"]}}}`, "", "answered"},
		{"ok", `{"answers":{"one":{"skipped":true},"two":{"skipped":true}}}`, "", "skipped"},
		{"error", "", "Canceled", "cancelled"},
		{"error", "", "Cancelled", "cancelled"},
		{"error", "", "Cannot find question UI", "failed"},
		{"error", "{", "Other error", "failed"},
		{"ok", `{"answers":{}}`, "", "unknown"},
		{"ok", `{"answers":{"one":{"selected":["Yes"]}}}`, "", "unknown"},
		{"ok", `{"answers":{"one":{"skipped":false,"selected":[],"freeText":""}}}`, "", "unknown"},
		{"ok", `{"answers":{"one":null}}`, "", "unknown"},
		{"ok", "{", "", "unknown"},
		{"", `{"answers":{"one":{"skipped":false,"selected":["Yes"]}}}`, "", "unknown"},
	} {
		ev := vscodeQuestionEvent(t, tc.status, tc.result, tc.failure, nil)
		if got := vscodeQuestionOutcome(ev); got != tc.want {
			t.Fatalf("outcome %s, want %s for %+v", got, tc.want, tc)
		}
	}
}

func TestVSCodeQuestionFailureWithDifferentResultShape(t *testing.T) {
	for _, tc := range []struct{ failure, want string }{{"Canceled", "cancelled"}, {"Unavailable", "failed"}} {
		ev := vscodeRawEvent{Status: "error", Attrs: json.RawMessage(`{"result":{},"error":"` + tc.failure + `"}`)}
		if got := vscodeQuestionOutcome(ev); got != tc.want {
			t.Fatalf("outcome = %s, want %s", got, tc.want)
		}
	}
}

func TestVSCodeQuestionDurations(t *testing.T) {
	answer := `{"answers":{"one":{"selected":["Yes"],"skipped":false}}}`
	q := &AskUserStats{}
	for _, ms := range []float64{10.25, 0, -1, math.NaN(), math.Inf(1)} {
		ev := vscodeQuestionEvent(t, "ok", answer, "", &ms)
		outcome, wait := vscodeQuestion(ev, "test")
		q.addOutcome(outcome, wait)
	}
	outcome, wait := vscodeQuestion(vscodeQuestionEvent(t, "ok", answer, "", nil), "test")
	q.addOutcome(outcome, wait)
	if q.Answered != 6 || q.TimedAnswers != 2 || q.UntimedAnswers != 4 ||
		q.TotalWaitMs != 10.25 || *q.AvgWaitMs != 5.125 || *q.MaxWaitMs != 10.25 {
		t.Fatalf("timing = %+v", q)
	}
}

func questionLog(t *testing.T, id string, ts int, duration *float64, status, result, failure string) string {
	t.Helper()
	ev := vscodeQuestionEvent(t, status, result, failure, duration)
	ev.SpanID, ev.TS = id, float64(ts)
	line, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	return string(line) + "\n"
}

func TestCollectVSCodeQuestionsFiltersAndLogDeduplication(t *testing.T) {
	root := t.TempDir()
	answer := `{"answers":{"one":{"skipped":false,"selected":["Yes"]}}}`
	ten, hundred, zero := 10.0, 100.0, 0.0
	hash := writeVSCodeWorkspace(t, root, "a", `{"folder":"file:///repo/a"}`)
	main := writeVSCodeSessionLog(t, hash, "a", questionLog(t, "m", 1000, &ten, "ok", answer, "")+
		`{"type":"child_session_ref","ts":1001,"attrs":{"label":"Other","childLogFile":"shared.jsonl"}}`+"\n"+
		`{"type":"child_session_ref","ts":1002,"attrs":{"label":"Explore","childLogFile":"shared.jsonl"}}`+"\n"+
		`{"type":"child_session_ref","ts":1003,"attrs":{"label":"title","childLogFile":"title.jsonl"}}`+"\n")
	child := questionLog(t, "s", 2000, &hundred, "ok", answer, "")
	writeVSCodeChildLog(t, main, "shared.jsonl", child+child+
		`{"type":"child_session_ref","ts":2001,"attrs":{"label":"Other","childLogFile":"other.jsonl"}}`+"\n"+
		`{"type":"child_session_ref","ts":2001,"attrs":{"label":"Explore","childLogFile":"main.jsonl"}}`+"\n")
	writeVSCodeChildLog(t, main, "other.jsonl", questionLog(t, "s", 2002, &hundred, "error", "", "Canceled"))
	writeVSCodeChildLog(t, main, "title.jsonl", questionLog(t, "t", 2000, &hundred, "ok", answer, ""))
	writeVSCodeChildLog(t, main, "orphan.jsonl", questionLog(t, "o", 2000, &hundred, "ok", answer, ""))
	hashB := writeVSCodeWorkspace(t, root, "b", `{"folder":"file:///repo/b"}`)
	writeVSCodeSessionLog(t, hashB, "b", questionLog(t, "b", 3000, &zero, "ok", answer, ""))

	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root, Top: 1, IncludeLogStats: true})
	if err != nil {
		t.Fatal(err)
	}
	q := stats.UserQuestions
	if q.Sessions != 2 || q.Calls != 4 || q.Answered != 3 || q.Cancelled != 1 ||
		q.TimedAnswers != 3 || q.TotalWaitMs != 110 || *q.MaxWaitMs != 100 {
		t.Fatalf("questions = %+v", q)
	}
	for _, parent := range stats.LogDetails.Sessions {
		for _, log := range parent.Logs {
			if log.UserQuestions == nil {
				t.Fatalf("missing normalized question data: %+v", log)
			}
		}
	}
	cases := []struct {
		name            string
		opts            VSCodeStatsOptions
		calls, sessions int
		total           float64
	}{
		{"agent", VSCodeStatsOptions{Agent: []string{"^Explore$"}}, 1, 1, 100},
		{"aliases", VSCodeStatsOptions{Agent: []string{"Explore", "Other"}}, 2, 1, 100},
		{"scope", VSCodeStatsOptions{Scope: SessionScope{ExactCWD: "/repo/a"}}, 3, 1, 110},
		{"session", VSCodeStatsOptions{SessionID: "b"}, 1, 1, 0},
		{"absent", VSCodeStatsOptions{SessionID: "absent"}, 0, 0, 0},
		{"tool", VSCodeStatsOptions{Tool: []string{"^read_file$"}}, 0, 0, 0},
		{"model", VSCodeStatsOptions{Model: []string{"^Never$"}}, 4, 2, 110},
		{"start bounds", VSCodeStatsOptions{Since: time.UnixMilli(2000), Until: time.UnixMilli(2001)}, 1, 1, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts.Root = root
			actual, err := CollectVSCodeStats(tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			q := actual.UserQuestions
			if q.Calls != tc.calls || q.Sessions != tc.sessions || q.TotalWaitMs != tc.total {
				t.Fatalf("questions = %+v", q)
			}
			if q.TimedAnswers == 0 && (q.AvgWaitMs != nil || q.MaxWaitMs != nil) {
				t.Fatalf("invented timing: %+v", q)
			}
		})
	}
}

func TestCollectVSCodeQuestionSkippedUnknownAndMissingDuration(t *testing.T) {
	root := t.TempDir()
	hash := writeVSCodeWorkspace(t, root, "a", `{"folder":"file:///repo/a"}`)
	zero := 0.0
	writeVSCodeSessionLog(t, hash, "a",
		questionLog(t, "skip", 1, &zero, "ok", `{"answers":{"one":{"skipped":true}}}`, "")+
			questionLog(t, "unknown", 2, &zero, "ok", "{}", "")+
			questionLog(t, "failed", 3, &zero, "error", "", "Unavailable")+
			questionLog(t, "untimed", 4, nil, "ok", `{"answers":{"one":{"selected":["Yes"],"skipped":false}}}`, ""))
	stats, err := CollectVSCodeStats(VSCodeStatsOptions{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	q := stats.UserQuestions
	if q.Calls != 4 || q.Skipped != 1 || q.Unknown != 1 || q.Failed != 1 ||
		q.Answered != 1 || q.UntimedAnswers != 1 || q.AvgWaitMs != nil || q.MaxWaitMs != nil {
		t.Fatalf("questions = %+v", q)
	}
	data, err := json.Marshal(q)
	if err != nil || strings.Contains(string(data), "Yes") || strings.Contains(string(data), "Unresolved") {
		t.Fatalf("export should contain accounting only: %s, %v", data, err)
	}
	if !strings.Contains(string(data), `"AvgWaitMs":null`) {
		t.Fatalf("unavailable timing not preserved: %s", data)
	}
}
