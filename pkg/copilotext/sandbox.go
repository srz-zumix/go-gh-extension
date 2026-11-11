package copilotext

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// Sandbox states recorded for a tool execution. SandboxUnknown means no sandbox outcome
// was recorded (e.g. an older client, an unexecuted or interrupted call).
const (
	SandboxApplied    = "applied"
	SandboxNotApplied = "not-applied"
	SandboxUnknown    = "unknown"
)

// SandboxStates lists every valid sandbox state in display order.
var SandboxStates = []string{SandboxApplied, SandboxNotApplied, SandboxUnknown}

// SandboxCount is one entry of a SandboxStats axis: a key (e.g. a tool name) and the
// number of executions observed for it, broken down by sandbox state.
type SandboxCount struct {
	Key        string
	Total      int
	Applied    int
	NotApplied int
	Unknown    int
}

// SandboxStats summarizes whether tool executions ran inside the terminal sandbox.
type SandboxStats struct {
	Sessions   int
	Executions int
	Applied    int
	NotApplied int
	Unknown    int
	// OptOutRequested counts executions whose caller asked to run outside the sandbox.
	// Only the Copilot CLI records it.
	OptOutRequested int `json:",omitempty"`
	ByTool          []SandboxCount
}

// IsEmpty reports whether no execution was observed, so renderers can skip the section.
func (s *SandboxStats) IsEmpty() bool {
	return s == nil || s.Executions == 0
}

func (s *SandboxStats) add(tool, state string, optOut bool, byTool *sandboxCounter) {
	s.Executions++
	switch state {
	case SandboxApplied:
		s.Applied++
	case SandboxNotApplied:
		s.NotApplied++
	default:
		s.Unknown++
	}
	if optOut {
		s.OptOutRequested++
	}
	byTool.add(tool, state)
}

// sandboxKey normalizes an empty state to SandboxUnknown.
func sandboxKey(state string) string {
	if state == "" {
		return SandboxUnknown
	}
	return state
}

// newSandboxFilter validates states and returns them as a set, or nil when empty.
func newSandboxFilter(states []string) (map[string]bool, error) {
	if len(states) == 0 {
		return nil, nil
	}
	set := make(map[string]bool, len(states))
	for _, state := range states {
		switch state {
		case SandboxApplied, SandboxNotApplied, SandboxUnknown:
			set[state] = true
		default:
			return nil, fmt.Errorf("invalid sandbox state %q: want %q, %q, or %q", state, SandboxApplied, SandboxNotApplied, SandboxUnknown)
		}
	}
	return set, nil
}

// sandboxCounter accumulates SandboxCount entries keyed by an arbitrary string.
type sandboxCounter struct {
	order []string
	byKey map[string]*SandboxCount
}

func newSandboxCounter() *sandboxCounter {
	return &sandboxCounter{byKey: make(map[string]*SandboxCount)}
}

func (c *sandboxCounter) add(key, state string) {
	entry, ok := c.byKey[key]
	if !ok {
		entry = &SandboxCount{Key: key}
		c.byKey[key] = entry
		c.order = append(c.order, key)
	}
	entry.Total++
	switch state {
	case SandboxApplied:
		entry.Applied++
	case SandboxNotApplied:
		entry.NotApplied++
	default:
		entry.Unknown++
	}
}

// finish returns the accumulated counts sorted by Total descending, then Key ascending,
// truncated to the top entries when top > 0.
func (c *sandboxCounter) finish(top int) []SandboxCount {
	counts := make([]SandboxCount, 0, len(c.order))
	for _, key := range c.order {
		counts = append(counts, *c.byKey[key])
	}
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Total != counts[j].Total {
			return counts[i].Total > counts[j].Total
		}
		return counts[i].Key < counts[j].Key
	})
	if top > 0 && len(counts) > top {
		counts = counts[:top]
	}
	return counts
}

// sandboxShellTools are counted even when no sandbox telemetry was recorded, so older
// shell executions appear as SandboxUnknown instead of disappearing.
var sandboxShellTools = map[string]bool{"bash": true, "powershell": true}

type toolExecutionKey struct {
	agentID    string
	toolCallID string
}

type toolExecutionCompletion struct {
	state  string
	optOut bool
	at     time.Time
}

// toolExecutionRecord is one completed tool execution relevant to sandbox accounting.
type toolExecutionRecord struct {
	tool        string
	state       string
	optOut      bool
	completedAt time.Time
}

// toolExecutionCollector pairs "tool.execution_start" and "tool.execution_complete" events
// of a single session to recover each execution's tool name and sandbox telemetry.
type toolExecutionCollector struct {
	session     Session
	tools       map[toolExecutionKey]string
	completions map[toolExecutionKey]toolExecutionCompletion
	order       []toolExecutionKey
	byCallID    map[string][]toolExecutionKey
}

func newToolExecutionCollector(s Session) *toolExecutionCollector {
	return &toolExecutionCollector{
		session:     s,
		tools:       make(map[toolExecutionKey]string),
		completions: make(map[toolExecutionKey]toolExecutionCompletion),
		byCallID:    make(map[string][]toolExecutionKey),
	}
}

type rawToolExecutionEvent struct {
	ToolCallID    string `json:"toolCallId"`
	ToolName      string `json:"toolName"`
	ToolTelemetry struct {
		Properties map[string]json.RawMessage `json:"properties"`
	} `json:"toolTelemetry"`
}

// telemetryBool decodes a telemetry property recorded as "true"/"false" or a JSON bool.
func telemetryBool(raw json.RawMessage) (value, ok bool) {
	if len(raw) == 0 {
		return false, false
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b, true
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		switch s {
		case "true":
			return true, true
		case "false":
			return false, true
		}
	}
	return false, false
}

func (ec *toolExecutionCollector) handle(ev rawEvent) {
	if ev.Type != "tool.execution_start" && ev.Type != "tool.execution_complete" {
		return
	}
	var data rawToolExecutionEvent
	if err := json.Unmarshal(ev.Data, &data); err != nil {
		logger.Warn("skipping unparsable tool event for sandbox accounting", "dir", ec.session.Dir, "error", err)
		return
	}
	if data.ToolCallID == "" {
		return
	}
	key := toolExecutionKey{agentID: ev.AgentID, toolCallID: data.ToolCallID}
	if ev.Type == "tool.execution_start" {
		if _, exists := ec.tools[key]; !exists {
			ec.tools[key] = data.ToolName
		}
		return
	}
	if _, exists := ec.completions[key]; exists {
		return
	}
	completion := toolExecutionCompletion{state: SandboxUnknown, at: ev.Timestamp}
	if applied, ok := telemetryBool(data.ToolTelemetry.Properties["sandboxApplied"]); ok {
		completion.state = SandboxNotApplied
		if applied {
			completion.state = SandboxApplied
		}
	}
	completion.optOut, _ = telemetryBool(data.ToolTelemetry.Properties["sandboxOptOutRequested"])
	ec.completions[key] = completion
	ec.order = append(ec.order, key)
	ec.byCallID[key.toolCallID] = append(ec.byCallID[key.toolCallID], key)
}

// sandboxFor resolves the sandbox state of the execution a permission request belongs to.
// The exact (agentID, toolCallID) match wins; otherwise a toolCallID completed by exactly
// one agent of the session is used, since requests are sometimes logged without the
// subagent identity that runs the tool.
func (ec *toolExecutionCollector) sandboxFor(agentID, toolCallID string) string {
	if toolCallID == "" {
		return SandboxUnknown
	}
	if c, ok := ec.completions[toolExecutionKey{agentID: agentID, toolCallID: toolCallID}]; ok {
		return c.state
	}
	if keys := ec.byCallID[toolCallID]; len(keys) == 1 {
		return ec.completions[keys[0]].state
	}
	return SandboxUnknown
}

// records returns completed executions of shell tools or with recorded sandbox telemetry,
// in completion order.
func (ec *toolExecutionCollector) records() []toolExecutionRecord {
	var records []toolExecutionRecord
	for _, key := range ec.order {
		c := ec.completions[key]
		tool := ec.tools[key]
		if !sandboxShellTools[tool] && c.state == SandboxUnknown {
			continue
		}
		if tool == "" {
			tool = "unknown"
		}
		records = append(records, toolExecutionRecord{tool: tool, state: c.state, optOut: c.optOut, completedAt: c.at})
	}
	return records
}
