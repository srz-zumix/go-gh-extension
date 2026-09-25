package copilotext

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// VSCodeStatsOptions configures CollectVSCodeStats.
type VSCodeStatsOptions struct {
	// Root overrides the workspaceStorage directory to scan. Empty uses the default
	// (~/Library/Application Support/Code/User/workspaceStorage).
	Root string
	// Since and Until, when non-zero, restrict counted events to Since <= ts < Until.
	// A session whose log file was last modified before Since is skipped without being
	// opened, since VS Code only ever appends to main.jsonl.
	Since, Until time.Time
	// Scope restricts which sessions are considered, matching against the owning
	// workspace's recorded folder path.
	Scope SessionScope
	// SessionID, when non-empty, restricts to a single session by ID.
	SessionID string
	// Tool, when non-empty, restricts tool usage stats to tool_call events whose name
	// matches any of these regular expressions (a plain substring is also a valid,
	// unanchored regular expression).
	Tool []string
	// Model, when non-empty, restricts LLM request stats to llm_request events whose
	// model matches any of these regular expressions.
	Model []string
	// Agent, when non-empty, restricts subagent stats to child_session_ref events whose
	// derived agent name matches any of these regular expressions.
	Agent []string
	// Top, when > 0, keeps only the top N entries (by their primary count, then Key) of
	// each axis. 0 keeps every entry.
	Top int
}

// VSCodeToolCount is one entry of VSCodeStats.ByTool: a tool name and how often it was
// called, broken down by outcome and duration.
type VSCodeToolCount struct {
	Key     string
	Total   int
	OK      int
	Error   int
	TotalMs float64
	AvgMs   float64
}

// VSCodeModelCount is one entry of VSCodeStats.ByModel: an LLM model and its token and
// usage totals.
type VSCodeModelCount struct {
	Key          string
	Requests     int
	InputTokens  int64
	OutputTokens int64
	CachedTokens int64
	AvgTTFTMs    float64
	UsageAIU     float64
}

// VSCodeAgentCount is one entry of VSCodeStats.ByAgent: a subagent name and how many
// times it was invoked.
type VSCodeAgentCount struct {
	Key   string
	Total int
}

// VSCodeWorkspaceCount is one entry of VSCodeStats.ByWorkspace: a workspace folder and its
// activity totals.
type VSCodeWorkspaceCount struct {
	Key         string
	ToolCalls   int
	LLMRequests int
	Turns       int
}

// VSCodeStats summarizes the VS Code Copilot Chat activity collected by
// CollectVSCodeStats.
type VSCodeStats struct {
	Sessions    int
	Turns       int
	LLMRequests int
	ToolCalls   int
	Since       time.Time
	Until       time.Time
	Scope       SessionScope

	ByTool      []VSCodeToolCount
	ByModel     []VSCodeModelCount
	ByAgent     []VSCodeAgentCount
	ByWorkspace []VSCodeWorkspaceCount
}

// vscodeAgentName derives a subagent axis key from a child_session_ref event's
// attrs.label, or reports ok=false when label does not represent a subagent invocation
// (e.g. VS Code's internal "title" child session, used to generate a chat title).
func vscodeAgentName(label string) (name string, ok bool) {
	if label == "" || label == "title" {
		return "", false
	}
	if rest, found := strings.CutPrefix(label, "runSubagent-"); found {
		return rest, true
	}
	return label, true
}

// vscodeFilters is the compiled form of VSCodeStatsOptions' Tool/Model/Agent filters.
type vscodeFilters struct {
	tools  []*regexp.Regexp
	models []*regexp.Regexp
	agents []*regexp.Regexp
}

func newVSCodeFilters(opts VSCodeStatsOptions) (*vscodeFilters, error) {
	f := &vscodeFilters{}
	compileAll := func(patterns []string, field string) ([]*regexp.Regexp, error) {
		var res []*regexp.Regexp
		for _, p := range patterns {
			re, err := regexp.Compile(p)
			if err != nil {
				return nil, fmt.Errorf("invalid %s pattern %q: %w", field, p, err)
			}
			res = append(res, re)
		}
		return res, nil
	}
	var err error
	if f.tools, err = compileAll(opts.Tool, "tool"); err != nil {
		return nil, err
	}
	if f.models, err = compileAll(opts.Model, "model"); err != nil {
		return nil, err
	}
	if f.agents, err = compileAll(opts.Agent, "agent"); err != nil {
		return nil, err
	}
	return f, nil
}

func anyRegexMatches(patterns []*regexp.Regexp, value string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, re := range patterns {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func (f *vscodeFilters) matchesTool(name string) bool  { return anyRegexMatches(f.tools, name) }
func (f *vscodeFilters) matchesModel(name string) bool { return anyRegexMatches(f.models, name) }
func (f *vscodeFilters) matchesAgent(name string) bool { return anyRegexMatches(f.agents, name) }

// vscodeAggregator accumulates the per-axis counts collected while scanning sessions.
type vscodeAggregator struct {
	tools      map[string]*VSCodeToolCount
	toolOrder  []string
	models     map[string]*VSCodeModelCount
	modelOrder []string
	agents     map[string]*VSCodeAgentCount
	agentOrder []string
	workspaces map[string]*VSCodeWorkspaceCount
	wsOrder    []string

	modelTTFTSum   map[string]float64
	modelTTFTCount map[string]int
}

func newVSCodeAggregator() *vscodeAggregator {
	return &vscodeAggregator{
		tools:          make(map[string]*VSCodeToolCount),
		models:         make(map[string]*VSCodeModelCount),
		agents:         make(map[string]*VSCodeAgentCount),
		workspaces:     make(map[string]*VSCodeWorkspaceCount),
		modelTTFTSum:   make(map[string]float64),
		modelTTFTCount: make(map[string]int),
	}
}

func (a *vscodeAggregator) addTool(name, status string, dur float64) {
	entry, ok := a.tools[name]
	if !ok {
		entry = &VSCodeToolCount{Key: name}
		a.tools[name] = entry
		a.toolOrder = append(a.toolOrder, name)
	}
	entry.Total++
	entry.TotalMs += dur
	if status == "error" {
		entry.Error++
	} else {
		entry.OK++
	}
}

func (a *vscodeAggregator) addModel(model string, attrs vscodeLLMRequestAttrs) {
	entry, ok := a.models[model]
	if !ok {
		entry = &VSCodeModelCount{Key: model}
		a.models[model] = entry
		a.modelOrder = append(a.modelOrder, model)
	}
	entry.Requests++
	entry.InputTokens += attrs.InputTokens
	entry.OutputTokens += attrs.OutputTokens
	entry.CachedTokens += attrs.CachedTokens
	entry.UsageAIU += float64(attrs.CopilotUsageNanoAIU) / 1e9
	if attrs.TTFT > 0 {
		a.modelTTFTSum[model] += attrs.TTFT
		a.modelTTFTCount[model]++
	}
}

func (a *vscodeAggregator) addAgent(name string) {
	entry, ok := a.agents[name]
	if !ok {
		entry = &VSCodeAgentCount{Key: name}
		a.agents[name] = entry
		a.agentOrder = append(a.agentOrder, name)
	}
	entry.Total++
}

func (a *vscodeAggregator) addWorkspace(key string, toolCalls, llmRequests, turns int) {
	entry, ok := a.workspaces[key]
	if !ok {
		entry = &VSCodeWorkspaceCount{Key: key}
		a.workspaces[key] = entry
		a.wsOrder = append(a.wsOrder, key)
	}
	entry.ToolCalls += toolCalls
	entry.LLMRequests += llmRequests
	entry.Turns += turns
}

func (a *vscodeAggregator) finishTools(top int) []VSCodeToolCount {
	out := make([]VSCodeToolCount, 0, len(a.toolOrder))
	for _, k := range a.toolOrder {
		c := *a.tools[k]
		if c.Total > 0 {
			c.AvgMs = c.TotalMs / float64(c.Total)
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}
		return out[i].Key < out[j].Key
	})
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out
}

func (a *vscodeAggregator) finishModels(top int) []VSCodeModelCount {
	out := make([]VSCodeModelCount, 0, len(a.modelOrder))
	for _, k := range a.modelOrder {
		c := *a.models[k]
		if n := a.modelTTFTCount[k]; n > 0 {
			c.AvgTTFTMs = a.modelTTFTSum[k] / float64(n)
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Requests != out[j].Requests {
			return out[i].Requests > out[j].Requests
		}
		return out[i].Key < out[j].Key
	})
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out
}

func (a *vscodeAggregator) finishAgents(top int) []VSCodeAgentCount {
	out := make([]VSCodeAgentCount, 0, len(a.agentOrder))
	for _, k := range a.agentOrder {
		out = append(out, *a.agents[k])
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}
		return out[i].Key < out[j].Key
	})
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out
}

func (a *vscodeAggregator) finishWorkspaces(top int) []VSCodeWorkspaceCount {
	out := make([]VSCodeWorkspaceCount, 0, len(a.wsOrder))
	for _, k := range a.wsOrder {
		out = append(out, *a.workspaces[k])
	}
	sort.Slice(out, func(i, j int) bool {
		total := func(c VSCodeWorkspaceCount) int { return c.ToolCalls + c.LLMRequests + c.Turns }
		if total(out[i]) != total(out[j]) {
			return total(out[i]) > total(out[j])
		}
		return out[i].Key < out[j].Key
	})
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out
}

// CollectVSCodeStats scans the VS Code Copilot Chat debug-logs directory and aggregates
// tool usage, token/usage, turn, and subagent activity into a VSCodeStats, applying
// opts' filters.
func CollectVSCodeStats(opts VSCodeStatsOptions) (*VSCodeStats, error) {
	root := opts.Root
	if root == "" {
		r, err := vscodeStorageRoot()
		if err != nil {
			return nil, err
		}
		root = r
	}

	sessions, err := ListVSCodeSessions(root)
	if err != nil {
		return nil, err
	}

	filters, err := newVSCodeFilters(opts)
	if err != nil {
		return nil, err
	}

	agg := newVSCodeAggregator()
	stats := &VSCodeStats{Since: opts.Since, Until: opts.Until, Scope: opts.Scope}

	for _, s := range sessions {
		if opts.SessionID != "" && s.ID != opts.SessionID {
			continue
		}
		if !opts.Scope.matches(s.Folder) {
			continue
		}
		if !opts.Since.IsZero() && s.ModifiedAt.Before(opts.Since) {
			continue
		}

		sessionCounted := false
		turnIDs := make(map[string]bool)
		var toolCallsInSession, llmRequestsInSession int

		readErr := readVSCodeEvents(s.LogPath, func(ev vscodeRawEvent) error {
			ts := time.UnixMilli(int64(ev.TS))
			if !opts.Since.IsZero() && ts.Before(opts.Since) {
				return nil
			}
			if !opts.Until.IsZero() && !ts.Before(opts.Until) {
				return nil
			}

			switch ev.Type {
			case "turn_start":
				var a vscodeTurnAttrs
				if err := json.Unmarshal(ev.Attrs, &a); err != nil {
					logger.Warn("skipping unparsable turn_start event", "path", s.LogPath, "error", err)
					return nil
				}
				if a.TurnID != "" && !turnIDs[a.TurnID] {
					turnIDs[a.TurnID] = true
					stats.Turns++
					sessionCounted = true
				}
			case "tool_call":
				if !filters.matchesTool(ev.Name) {
					return nil
				}
				agg.addTool(ev.Name, ev.Status, ev.Dur)
				stats.ToolCalls++
				toolCallsInSession++
				sessionCounted = true
			case "llm_request":
				var a vscodeLLMRequestAttrs
				if err := json.Unmarshal(ev.Attrs, &a); err != nil {
					logger.Warn("skipping unparsable llm_request event", "path", s.LogPath, "error", err)
					return nil
				}
				if !filters.matchesModel(a.Model) {
					return nil
				}
				agg.addModel(a.Model, a)
				stats.LLMRequests++
				llmRequestsInSession++
				sessionCounted = true
			case "child_session_ref":
				var a vscodeChildSessionRefAttrs
				if err := json.Unmarshal(ev.Attrs, &a); err != nil {
					logger.Warn("skipping unparsable child_session_ref event", "path", s.LogPath, "error", err)
					return nil
				}
				name, ok := vscodeAgentName(a.Label)
				if !ok || !filters.matchesAgent(name) {
					return nil
				}
				agg.addAgent(name)
				sessionCounted = true
			}
			return nil
		})
		if readErr != nil {
			return nil, fmt.Errorf("failed to read vscode session log %q: %w", s.LogPath, readErr)
		}

		if sessionCounted {
			stats.Sessions++
			agg.addWorkspace(s.Folder, toolCallsInSession, llmRequestsInSession, len(turnIDs))
		}
	}

	stats.ByTool = agg.finishTools(opts.Top)
	stats.ByModel = agg.finishModels(opts.Top)
	stats.ByAgent = agg.finishAgents(opts.Top)
	stats.ByWorkspace = agg.finishWorkspaces(opts.Top)

	return stats, nil
}
