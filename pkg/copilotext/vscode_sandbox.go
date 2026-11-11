package copilotext

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// vscodeTerminalTool is the VS Code Copilot Chat tool that runs terminal commands.
const vscodeTerminalTool = "run_in_terminal"

// vscodeChatSessionPath returns the chatSessions JSONL file that records s's chat
// transcript. VS Code names it after the same session ID as the debug-logs directory.
func vscodeChatSessionPath(root string, s VSCodeSession) string {
	return filepath.Join(root, s.WorkspaceHash, "chatSessions", s.ID+".jsonl")
}

// vscodeTerminalExecution is one terminal tool invocation recovered from a chat session.
type vscodeTerminalExecution struct {
	tool        string
	state       string
	requestedAt time.Time
}

// vscodeChatPatch is one line of a chatSessions JSONL mutation log: kind 0 replaces the
// whole value, kind 1 sets the value at path k, kind 2 pushes v's items onto the array at
// k (truncating it to i first when i is present), and kind 3 deletes the value at k.
type vscodeChatPatch struct {
	Kind int             `json:"kind"`
	K    []any           `json:"k"`
	V    json.RawMessage `json:"v"`
	I    *int            `json:"i"`
}

// readVSCodeTerminalExecutions replays the chatSessions mutation log at path and returns
// its terminal tool invocations with their sandbox state. A missing file yields nil.
func readVSCodeTerminalExecutions(path string) (executions []vscodeTerminalExecution, rerr error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open %q: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && rerr == nil {
			rerr = fmt.Errorf("failed to close %q: %w", path, cerr)
		}
	}()

	var root any
	// Later serializations of an invocation may omit isSandboxWrapped; keep the latest
	// explicitly recorded value per tool call.
	explicit := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var patch vscodeChatPatch
		if err := json.Unmarshal(line, &patch); err != nil {
			logger.Warn("skipping unparsable chat session patch", "path", path, "error", err)
			continue
		}
		var value any
		if len(patch.V) > 0 {
			if err := json.Unmarshal(patch.V, &value); err != nil {
				logger.Warn("skipping unparsable chat session patch value", "path", path, "error", err)
				continue
			}
		}
		recordSandboxWrapped(value, explicit)
		root = applyVSCodeChatPatch(root, patch, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", path, err)
	}

	requests, _ := jsonField(root, "requests").([]any)
	seen := make(map[string]bool)
	for _, request := range requests {
		var requestedAt time.Time
		if ts, ok := jsonField(request, "timestamp").(float64); ok && ts > 0 {
			requestedAt = time.UnixMilli(int64(ts))
		}
		response, _ := jsonField(request, "response").([]any)
		for _, part := range response {
			callID, _ := jsonField(part, "toolCallId").(string)
			toolID, _ := jsonField(part, "toolId").(string)
			if callID == "" || seen[callID] {
				continue
			}
			wrapped, hasWrapped := sandboxWrapped(part)
			if !hasWrapped {
				wrapped, hasWrapped = explicit[callID]
			}
			if toolID != vscodeTerminalTool && !hasWrapped {
				continue
			}
			seen[callID] = true
			state := SandboxUnknown
			if hasWrapped {
				state = SandboxNotApplied
				if wrapped {
					state = SandboxApplied
				}
			}
			if toolID == "" {
				toolID = "unknown"
			}
			executions = append(executions, vscodeTerminalExecution{tool: toolID, state: state, requestedAt: requestedAt})
		}
	}
	return executions, nil
}

// jsonField returns obj[key] when obj is a JSON object, or nil.
func jsonField(obj any, key string) any {
	m, ok := obj.(map[string]any)
	if !ok {
		return nil
	}
	return m[key]
}

// sandboxWrapped reads toolSpecificData.commandLine.isSandboxWrapped from an invocation.
func sandboxWrapped(invocation any) (wrapped, ok bool) {
	wrapped, ok = jsonField(jsonField(jsonField(invocation, "toolSpecificData"), "commandLine"), "isSandboxWrapped").(bool)
	return wrapped, ok
}

// recordSandboxWrapped walks value and stores every explicitly recorded sandbox state by
// tool call ID.
func recordSandboxWrapped(value any, explicit map[string]bool) {
	switch v := value.(type) {
	case map[string]any:
		if callID, ok := v["toolCallId"].(string); ok && callID != "" {
			if wrapped, ok := sandboxWrapped(v); ok {
				explicit[callID] = wrapped
			}
		}
		for _, child := range v {
			recordSandboxWrapped(child, explicit)
		}
	case []any:
		for _, child := range v {
			recordSandboxWrapped(child, explicit)
		}
	}
}

// applyVSCodeChatPatch applies patch (with decoded value) to root and returns the new root.
// Paths that do not resolve are ignored.
func applyVSCodeChatPatch(root any, patch vscodeChatPatch, value any) any {
	if patch.Kind == 0 {
		return value
	}
	if len(patch.K) == 0 {
		if patch.Kind == 1 {
			return value
		}
		return root
	}
	parent := root
	for _, segment := range patch.K[:len(patch.K)-1] {
		parent = jsonChild(parent, segment)
		if parent == nil {
			return root
		}
	}
	last := patch.K[len(patch.K)-1]
	switch patch.Kind {
	case 1:
		setJSONChild(parent, last, value)
	case 2:
		arr, _ := jsonChild(parent, last).([]any)
		if patch.I != nil && *patch.I >= 0 && *patch.I < len(arr) {
			arr = arr[:*patch.I]
		}
		if items, ok := value.([]any); ok {
			arr = append(arr, items...)
		}
		setJSONChild(parent, last, arr)
	case 3:
		if m, ok := parent.(map[string]any); ok {
			if key, ok := last.(string); ok {
				delete(m, key)
			}
		}
	}
	return root
}

// jsonChild resolves one path segment (an object key or an array index).
func jsonChild(node, segment any) any {
	switch n := node.(type) {
	case map[string]any:
		if key, ok := segment.(string); ok {
			return n[key]
		}
	case []any:
		if index, ok := segment.(float64); ok && index >= 0 && int(index) < len(n) {
			return n[int(index)]
		}
	}
	return nil
}

// setJSONChild assigns value at segment of node when the segment resolves.
func setJSONChild(node, segment, value any) {
	switch n := node.(type) {
	case map[string]any:
		if key, ok := segment.(string); ok {
			n[key] = value
		}
	case []any:
		if index, ok := segment.(float64); ok && index >= 0 && int(index) < len(n) {
			n[int(index)] = value
		}
	}
}

// collectVSCodeSandbox adds the in-window terminal executions recorded at path that match
// filter to stats.
func collectVSCodeSandbox(path string, opts VSCodeStatsOptions, filter map[string]bool, stats *SandboxStats, byTool *sandboxCounter) error {
	executions, err := readVSCodeTerminalExecutions(path)
	if err != nil {
		return err
	}
	counted := false
	for _, exec := range executions {
		if (!opts.Since.IsZero() || !opts.Until.IsZero()) && exec.requestedAt.IsZero() {
			continue
		}
		if (!opts.Since.IsZero() && exec.requestedAt.Before(opts.Since)) ||
			(!opts.Until.IsZero() && !exec.requestedAt.Before(opts.Until)) {
			continue
		}
		if len(filter) > 0 && !filter[exec.state] {
			continue
		}
		stats.add(exec.tool, exec.state, false, byTool)
		counted = true
	}
	if counted {
		stats.Sessions++
	}
	return nil
}
