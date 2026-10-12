package copilotext

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
	// Since and Until, when non-zero, restrict counted events to Since <= ts < Until,
	// including child logs referenced before Since.
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
	// Agent, when non-empty, scopes all activity to child logs whose derived agent name
	// matches any pattern. Main-agent and differently named descendant activity is excluded.
	Agent []string
	// IncludeLogStats exports untruncated log-level statistics for local projection.
	// It cannot be combined with Agent.
	IncludeLogStats bool
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
// activity, token and usage totals. Token and usage totals are summed over the
// llm_request events counted in LLMRequests (i.e. after the Model filter).
type VSCodeWorkspaceCount struct {
	Key          string
	Sessions     int
	ToolCalls    int
	LLMRequests  int
	Turns        int
	InputTokens  int64
	OutputTokens int64
	CachedTokens int64
	UsageAIU     float64
}

// vscodeWorkspaceUsage is the per-session token and usage subtotal folded into a
// VSCodeWorkspaceCount.
type vscodeWorkspaceUsage struct {
	InputTokens  int64
	OutputTokens int64
	CachedTokens int64
	UsageAIU     float64
}

func (u *vscodeWorkspaceUsage) add(attrs vscodeLLMRequestAttrs) {
	u.InputTokens += attrs.InputTokens
	u.OutputTokens += attrs.OutputTokens
	u.CachedTokens += attrs.CachedTokens
	u.UsageAIU += float64(attrs.CopilotUsageNanoAIU) / 1e9
}

// VSCodeStats summarizes the VS Code Copilot Chat activity collected by
// CollectVSCodeStats.
type VSCodeStats struct {
	Sessions      int
	Turns         int
	LLMRequests   int
	ToolCalls     int
	Since         time.Time
	Until         time.Time
	Scope         SessionScope
	UserQuestions *VSCodeQuestionStats

	ByTool      []VSCodeToolCount
	ByModel     []VSCodeModelCount
	ByAgent     []VSCodeAgentCount
	ByWorkspace []VSCodeWorkspaceCount
	LogDetails  *VSCodeLogDataset `json:",omitempty"`
}

// VSCodeLogDataset contains normalized activity and invocation provenance, not raw events.
type VSCodeLogDataset struct {
	Version  int
	Sessions []VSCodeSessionDetails
}

// VSCodeSessionDetails identifies a parent chat and its independently countable logs.
type VSCodeSessionDetails struct {
	ID          string
	Workspace   string
	Logs        []*VSCodeLogStats
	Invocations []VSCodeAgentInvocation
}

// VSCodeAgentInvocation preserves reference order and identity for filtered deduplication.
type VSCodeAgentInvocation struct {
	ID    string
	Agent string
}

// VSCodeLogStats records one log's own activity and all referencing agent names.
type VSCodeLogStats struct {
	ID            string
	Main          bool
	Agents        []string
	Turns         int
	LLMRequests   int
	ToolCalls     int
	ByTool        []VSCodeToolCount
	ByModel       []VSCodeLogModelCount
	UserQuestions *VSCodeQuestionStats
}

// VSCodeLogModelCount retains positive TTFT samples for exact cross-log averaging.
type VSCodeLogModelCount struct {
	VSCodeModelCount
	TotalTTFTMs float64
	TTFTSamples int
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

func (a *vscodeAggregator) addWorkspace(key string, toolCalls, llmRequests, turns int, usage vscodeWorkspaceUsage) {
	entry, ok := a.workspaces[key]
	if !ok {
		entry = &VSCodeWorkspaceCount{Key: key}
		a.workspaces[key] = entry
		a.wsOrder = append(a.wsOrder, key)
	}
	entry.Sessions++
	entry.ToolCalls += toolCalls
	entry.LLMRequests += llmRequests
	entry.Turns += turns
	entry.InputTokens += usage.InputTokens
	entry.OutputTokens += usage.OutputTokens
	entry.CachedTokens += usage.CachedTokens
	entry.UsageAIU += usage.UsageAIU
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
	if opts.IncludeLogStats && len(opts.Agent) > 0 {
		return nil, fmt.Errorf("--include-log-stats cannot be combined with --agent")
	}
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
	questions := &AskUserStats{}
	stats := &VSCodeStats{Since: opts.Since, Until: opts.Until, Scope: opts.Scope}
	if opts.IncludeLogStats {
		stats.LogDetails = &VSCodeLogDataset{Version: 1, Sessions: []VSCodeSessionDetails{}}
	}

	for _, s := range sessions {
		if opts.SessionID != "" && s.ID != opts.SessionID {
			continue
		}
		if !opts.Scope.matches(s.Folder) {
			continue
		}
		if !opts.Since.IsZero() && !s.ModifiedAt.IsZero() && s.ModifiedAt.Before(opts.Since) {
			continue
		}
		sessionCounted := false
		questionsBefore := questions.Questions
		var toolCallsInSession, llmRequestsInSession, turnsInSession int
		var usageInSession vscodeWorkspaceUsage
		visitedLogs := make(map[string]bool)
		countedLogs := make(map[string]bool)
		countedAgents := make(map[string]bool)
		detail := VSCodeSessionDetails{ID: s.ID, Workspace: s.Folder}
		logDetails := make(map[string]*VSCodeLogStats)
		logAggregators := make(map[string]*vscodeAggregator)
		logQuestions := make(map[string]*AskUserStats)
		getLogDetail := func(path string) *VSCodeLogStats {
			if !opts.IncludeLogStats {
				return nil
			}
			if entry := logDetails[path]; entry != nil {
				return entry
			}
			entry := &VSCodeLogStats{ID: filepath.Base(path), Main: path == s.LogPath, Agents: []string{}}
			logDetails[path] = entry
			logAggregators[path] = newVSCodeAggregator()
			logQuestions[path] = &AskUserStats{}
			detail.Logs = append(detail.Logs, entry)
			return entry
		}

		var readLog func(string, bool) error
		readLog = func(logPath string, countActivity bool) error {
			if logPath == s.LogPath && len(filters.agents) > 0 {
				countActivity = false
			}
			// A later matching reference may select an already visited log's activity.
			followRefs := !visitedLogs[logPath]
			if !followRefs && (!countActivity || countedLogs[logPath]) {
				return nil
			}
			visitedLogs[logPath] = true
			if countActivity {
				countedLogs[logPath] = true
			}
			logDetail := getLogDetail(logPath)
			turnIDs := make(map[string]bool)
			questionSpans := make(map[string]bool)
			return readVSCodeEvents(logPath, func(ev vscodeRawEvent) error {
				ts := time.UnixMilli(int64(ev.TS))
				inWindow := (opts.Since.IsZero() || !ts.Before(opts.Since)) &&
					(opts.Until.IsZero() || ts.Before(opts.Until))
				// Discover matching descendants even through unselected or older invocations.
				if ev.Type == "child_session_ref" {
					if !followRefs {
						return nil
					}
				} else if !countActivity || !inWindow {
					return nil
				}

				switch ev.Type {
				case "turn_start":
					var a vscodeTurnAttrs
					if err := json.Unmarshal(ev.Attrs, &a); err != nil {
						logger.Warn("skipping unparsable turn_start event", "path", logPath, "error", err)
						return nil
					}
					if a.TurnID != "" && !turnIDs[a.TurnID] {
						turnIDs[a.TurnID] = true
						stats.Turns++
						turnsInSession++
						sessionCounted = true
						if logDetail != nil {
							logDetail.Turns++
						}
					}
				case "tool_call":
					if !filters.matchesTool(ev.Name) {
						return nil
					}
					dur := 0.0
					if ev.Dur != nil {
						dur = *ev.Dur
					}
					agg.addTool(ev.Name, ev.Status, dur)
					stats.ToolCalls++
					toolCallsInSession++
					sessionCounted = true
					if logDetail != nil {
						logDetail.ToolCalls++
						logAggregators[logPath].addTool(ev.Name, ev.Status, dur)
					}
					if ev.Name == "vscode_askQuestions" && (ev.SpanID == "" || !questionSpans[ev.SpanID]) {
						if ev.SpanID != "" {
							questionSpans[ev.SpanID] = true
						}
						outcome, wait := vscodeQuestion(ev, logPath)
						questions.addOutcome(outcome, wait)
						if logDetail != nil {
							logQuestions[logPath].addOutcome(outcome, wait)
						}
					}
				case "llm_request":
					var a vscodeLLMRequestAttrs
					if err := json.Unmarshal(ev.Attrs, &a); err != nil {
						logger.Warn("skipping unparsable llm_request event", "path", logPath, "error", err)
						return nil
					}
					if !filters.matchesModel(a.Model) {
						return nil
					}
					agg.addModel(a.Model, a)
					usageInSession.add(a)
					stats.LLMRequests++
					llmRequestsInSession++
					sessionCounted = true
					if logDetail != nil {
						logDetail.LLMRequests++
						logAggregators[logPath].addModel(a.Model, a)
					}
				case "child_session_ref":
					var a vscodeChildSessionRefAttrs
					if err := json.Unmarshal(ev.Attrs, &a); err != nil {
						logger.Warn("skipping unparsable child_session_ref event", "path", logPath, "error", err)
						return nil
					}
					name, ok := vscodeAgentName(a.Label)
					if !ok {
						return nil
					}
					if a.ChildLogFile != "" {
						if filepath.Base(a.ChildLogFile) != a.ChildLogFile || filepath.Ext(a.ChildLogFile) != ".jsonl" {
							logger.Warn("skipping invalid child log filename", "path", logPath, "filename", a.ChildLogFile)
						} else {
							childPath := filepath.Join(s.Dir, a.ChildLogFile)
							info, err := os.Lstat(childPath)
							if os.IsNotExist(err) {
								logger.Warn("referenced vscode child log is missing", "path", childPath)
							} else if err != nil {
								return fmt.Errorf("failed to stat child log %q: %w", childPath, err)
							} else if !info.Mode().IsRegular() {
								logger.Warn("skipping non-regular vscode child log", "path", childPath)
							} else {
								if child := getLogDetail(childPath); child != nil && !slices.Contains(child.Agents, name) {
									child.Agents = append(child.Agents, name)
								}
								if err := readLog(childPath, filters.matchesAgent(name)); err != nil {
									return fmt.Errorf("failed to read child log %q: %w", childPath, err)
								}
							}
						}
					} else if opts.IncludeLogStats || (len(filters.agents) > 0 && filters.matchesAgent(name)) {
						logger.Warn("referenced vscode child log filename is missing", "path", logPath, "agent", name)
					}
					agentID := a.ChildLogFile
					if agentID == "" {
						agentID = a.ChildSessionID
					}
					if opts.IncludeLogStats && inWindow {
						detail.Invocations = append(detail.Invocations, VSCodeAgentInvocation{ID: agentID, Agent: name})
					}
					if inWindow && filters.matchesAgent(name) && (agentID == "" || !countedAgents[agentID]) {
						agg.addAgent(name)
						if agentID != "" {
							countedAgents[agentID] = true
						}
						sessionCounted = true
					}
				}
				return nil
			})
		}
		if readErr := readLog(s.LogPath, len(filters.agents) == 0); readErr != nil {
			return nil, fmt.Errorf("failed to read vscode session log %q: %w", s.LogPath, readErr)
		}

		if questions.Questions > questionsBefore {
			questions.Sessions++
		}
		if sessionCounted {
			stats.Sessions++
			agg.addWorkspace(s.Folder, toolCallsInSession, llmRequestsInSession, turnsInSession, usageInSession)
			if opts.IncludeLogStats {
				for path, entry := range logDetails {
					la := logAggregators[path]
					entry.ByTool = la.finishTools(0)
					entry.ByModel = []VSCodeLogModelCount{}
					q := logQuestions[path]
					if q.Questions > 0 {
						q.Sessions = 1
					}
					entry.UserQuestions = vscodeQuestionStats(q)
					for _, model := range la.finishModels(0) {
						entry.ByModel = append(entry.ByModel, VSCodeLogModelCount{
							VSCodeModelCount: model, TotalTTFTMs: la.modelTTFTSum[model.Key], TTFTSamples: la.modelTTFTCount[model.Key],
						})
					}
				}
				stats.LogDetails.Sessions = append(stats.LogDetails.Sessions, detail)
			}
		}
	}

	stats.ByTool = agg.finishTools(opts.Top)
	stats.ByModel = agg.finishModels(opts.Top)
	stats.ByAgent = agg.finishAgents(opts.Top)
	stats.ByWorkspace = agg.finishWorkspaces(opts.Top)
	stats.UserQuestions = vscodeQuestionStats(questions)

	return stats, nil
}
