package copilotext

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// AskUserStats reports outcomes and completed-answer latency, not UI-only thinking time.
type AskUserStats struct {
	Sessions       int
	Questions      int
	Answered       int
	Cancelled      int
	Failed         int
	Unresolved     int
	Unknown        int
	Skipped        int `json:",omitempty"`
	TimedAnswers   int
	UntimedAnswers int
	TotalWaitMs    float64
	AvgWaitMs      *float64
	MaxWaitMs      *float64
}

type questionKey struct {
	agentID    string
	toolCallID string
}

type questionRecord struct {
	startedAt time.Time
	endedAt   time.Time
	outcome   string
}

type questionCollector struct {
	session     Session
	starts      map[questionKey]time.Time
	order       []questionKey
	completions map[questionKey]questionRecord
}

func newQuestionCollector(session Session) *questionCollector {
	return &questionCollector{session: session, starts: make(map[questionKey]time.Time), completions: make(map[questionKey]questionRecord)}
}

type rawQuestionToolEvent struct {
	Success *bool           `json:"success"`
	Result  json.RawMessage `json:"result"`
	Error   json.RawMessage `json:"error"`
}

func questionOutcome(data rawQuestionToolEvent) string {
	var result struct{ Content string }
	var failure struct{ Code, Name string }
	resultErr := json.Unmarshal(data.Result, &result)
	errorErr := json.Unmarshal(data.Error, &failure)
	content := result.Content
	if (errorErr == nil && (failure.Code == "CANCELLED" || failure.Code == "CANCELED" ||
		failure.Code == "ABORT_ERR" || failure.Name == "AbortError")) ||
		strings.HasPrefix(content, "User cancelled") || strings.HasPrefix(content, "User canceled") {
		return "cancelled"
	}
	if data.Success == nil {
		return "unknown"
	}
	if !*data.Success {
		return "failed"
	}
	if resultErr == nil && (strings.HasPrefix(content, "User selected:") || strings.HasPrefix(content, "User responded:")) {
		return "answered"
	}
	return "unknown"
}

func (qc *questionCollector) handle(ev rawEvent) {
	if ev.Type != "tool.execution_start" && ev.Type != "tool.execution_complete" {
		return
	}
	// Decode the envelope separately: unrelated tool results may have different shapes.
	var envelope struct {
		ToolCallID string `json:"toolCallId"`
		ToolName   string `json:"toolName"`
	}
	if err := json.Unmarshal(ev.Data, &envelope); err != nil {
		logger.Warn("skipping unparsable tool event for question accounting", "dir", qc.session.Dir, "error", err)
		return
	}
	key := questionKey{agentID: ev.AgentID, toolCallID: envelope.ToolCallID}
	if ev.Type == "tool.execution_start" {
		if envelope.ToolName != "ask_user" {
			return
		}
		if key.toolCallID == "" {
			logger.Warn("skipping ask_user without a tool call identity", "dir", qc.session.Dir)
			return
		}
		if _, exists := qc.starts[key]; !exists {
			qc.order = append(qc.order, key)
			qc.starts[key] = ev.Timestamp
		}
		return
	}
	if key.toolCallID == "" {
		return
	}
	var data rawQuestionToolEvent
	outcome := "unknown"
	if err := json.Unmarshal(ev.Data, &data); err == nil {
		outcome = questionOutcome(data)
	}
	// Store only accounting metadata, never question or answer text.
	if _, exists := qc.completions[key]; !exists {
		qc.completions[key] = questionRecord{endedAt: ev.Timestamp, outcome: outcome}
	}
}

func (qc *questionCollector) records() []questionRecord {
	records := make([]questionRecord, 0, len(qc.order))
	for _, key := range qc.order {
		rec, found := qc.completions[key]
		if !found {
			rec.outcome = "unresolved"
		}
		rec.startedAt = qc.starts[key]
		if rec.outcome == "unknown" {
			logger.Warn("ask_user completion has an unrecognized outcome", "dir", qc.session.Dir, "toolCallId", key.toolCallID, "agentId", key.agentID)
		}
		records = append(records, rec)
	}
	return records
}

func (stats *AskUserStats) add(rec questionRecord) {
	var wait *float64
	if rec.outcome == "answered" {
		if rec.startedAt.IsZero() || rec.endedAt.IsZero() || rec.endedAt.Before(rec.startedAt) {
			logger.Warn("ask_user answer has missing or invalid timestamps", "startedAt", rec.startedAt, "endedAt", rec.endedAt)
		} else {
			ms := float64(rec.endedAt.Sub(rec.startedAt)) / float64(time.Millisecond)
			wait = &ms
		}
	}
	stats.addOutcome(rec.outcome, wait)
}

func (stats *AskUserStats) addOutcome(outcome string, wait *float64) {
	stats.Questions++
	switch outcome {
	case "answered":
		stats.Answered++
		if wait == nil {
			stats.UntimedAnswers++
			return
		}
		stats.TimedAnswers++
		stats.TotalWaitMs += *wait
		if stats.MaxWaitMs == nil || *wait > *stats.MaxWaitMs {
			stats.MaxWaitMs = wait
		}
		avg := stats.TotalWaitMs / float64(stats.TimedAnswers)
		stats.AvgWaitMs = &avg
	case "cancelled":
		stats.Cancelled++
	case "failed":
		stats.Failed++
	case "unresolved":
		stats.Unresolved++
	case "skipped":
		stats.Skipped++
	default:
		stats.Unknown++
	}
}

// IsEmpty reports whether no question was observed, so renderers can skip the section.
func (stats *AskUserStats) IsEmpty() bool {
	return stats == nil || *stats == AskUserStats{}
}
