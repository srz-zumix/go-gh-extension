package copilotext

import (
	"encoding/json"
	"math"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// VSCodeQuestionStats counts completed tool calls, including multi-question batches.
// In-flight questions are not recorded by these debug-log events.
type VSCodeQuestionStats struct {
	Sessions       int
	Calls          int
	Answered       int
	Skipped        int
	Cancelled      int
	Failed         int
	Unknown        int
	TimedAnswers   int
	UntimedAnswers int
	TotalWaitMs    float64
	AvgWaitMs      *float64
	MaxWaitMs      *float64
}

func vscodeQuestionStats(q *AskUserStats) *VSCodeQuestionStats {
	return &VSCodeQuestionStats{
		Sessions: q.Sessions, Calls: q.Questions, Answered: q.Answered, Skipped: q.Skipped,
		Cancelled: q.Cancelled, Failed: q.Failed, Unknown: q.Unknown,
		TimedAnswers: q.TimedAnswers, UntimedAnswers: q.UntimedAnswers,
		TotalWaitMs: q.TotalWaitMs, AvgWaitMs: q.AvgWaitMs, MaxWaitMs: q.MaxWaitMs,
	}
}

func vscodeQuestionOutcome(ev vscodeRawEvent) string {
	var attrs struct {
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	err := json.Unmarshal(ev.Attrs, &attrs)
	if ev.Status == "error" {
		var failure string
		if err == nil && json.Unmarshal(attrs.Error, &failure) == nil &&
			(failure == "Canceled" || failure == "Cancelled" || failure == "CancellationError") {
			return "cancelled"
		}
		return "failed"
	}
	if ev.Status != "ok" || err != nil {
		return "unknown"
	}
	var resultText string
	if json.Unmarshal(attrs.Result, &resultText) != nil {
		return "unknown"
	}
	var result struct {
		Answers map[string]struct {
			Selected []string `json:"selected"`
			FreeText *string  `json:"freeText"`
			Skipped  *bool    `json:"skipped"`
		} `json:"answers"`
	}
	if json.Unmarshal([]byte(resultText), &result) != nil || len(result.Answers) == 0 {
		return "unknown"
	}
	answered := false
	for _, answer := range result.Answers {
		if answer.Skipped == nil {
			return "unknown"
		}
		if *answer.Skipped {
			continue
		}
		if len(answer.Selected) == 0 && (answer.FreeText == nil || *answer.FreeText == "") {
			return "unknown"
		}
		answered = true
	}
	if answered {
		return "answered"
	}
	return "skipped"
}

func vscodeQuestion(ev vscodeRawEvent, path string) (string, *float64) {
	outcome := vscodeQuestionOutcome(ev)
	if outcome == "unknown" {
		logger.Warn("vscode question has an unrecognized outcome", "path", path, "spanId", ev.SpanID)
	}
	wait := ev.Dur
	if outcome == "answered" && (wait == nil || *wait < 0 || math.IsNaN(*wait) || math.IsInf(*wait, 0)) {
		logger.Warn("vscode question answer has missing or invalid duration", "path", path, "spanId", ev.SpanID)
		wait = nil
	}
	return outcome, wait
}
