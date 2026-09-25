package copilotext

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// vscodeRawEvent is the envelope shared by every line of a VS Code Copilot Chat
// debug-logs main.jsonl file. attrs is deferred as json.RawMessage since its shape
// depends on Type, and some events (e.g. llm_request, tool_call) carry attrs several
// megabytes in size that callers should only unmarshal when the event's Type is one they
// care about.
type vscodeRawEvent struct {
	V int `json:"v"`
	// TS is epoch milliseconds. Most events emit an integer, but some (e.g. subagent)
	// have been observed with a fractional value, so this is float64 rather than int64.
	TS           float64         `json:"ts"`
	Dur          float64         `json:"dur"`
	SID          string          `json:"sid"`
	Type         string          `json:"type"`
	Name         string          `json:"name"`
	SpanID       string          `json:"spanId"`
	ParentSpanID string          `json:"parentSpanId"`
	Status       string          `json:"status"`
	Attrs        json.RawMessage `json:"attrs"`
}

// vscodeLLMRequestAttrs is the attrs payload of a "llm_request" event.
type vscodeLLMRequestAttrs struct {
	Model               string  `json:"model"`
	InputTokens         int64   `json:"inputTokens"`
	OutputTokens        int64   `json:"outputTokens"`
	CachedTokens        int64   `json:"cachedTokens"`
	TTFT                float64 `json:"ttft"`
	CopilotUsageNanoAIU int64   `json:"copilotUsageNanoAiu"`
}

// vscodeChildSessionRefAttrs is the attrs payload of a "child_session_ref" event, emitted
// when the assistant spawns a subagent or another child session (e.g. title generation).
type vscodeChildSessionRefAttrs struct {
	ChildSessionID string `json:"childSessionId"`
	ChildLogFile   string `json:"childLogFile"`
	Label          string `json:"label"`
}

// vscodeTurnAttrs is the attrs payload of "turn_start" and "turn_end" events.
type vscodeTurnAttrs struct {
	TurnID string `json:"turnId"`
}

// readVSCodeEvents streams path (a main.jsonl file) line by line, calling fn once per
// parsed event. Lines that fail to parse are skipped with a warning rather than aborting
// the whole file, since a single truncated trailing line (from a session still in
// progress) is expected.
//
// Lines are read with bufio.Reader.ReadString rather than bufio.Scanner, since a single
// line can exceed several megabytes (e.g. an llm_request's inputMessages) and
// Scanner.Buffer requires a fixed maximum that would otherwise silently drop such lines.
func readVSCodeEvents(path string, fn func(vscodeRawEvent) error) (rerr error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open %q: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && rerr == nil {
			rerr = fmt.Errorf("failed to close %q: %w", path, cerr)
		}
	}()

	reader := bufio.NewReaderSize(f, 64*1024)
	for {
		line, readErr := reader.ReadString('\n')
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			var ev vscodeRawEvent
			if err := json.Unmarshal([]byte(trimmed), &ev); err != nil {
				logger.Warn("skipping unparsable vscode event", "path", path, "error", err)
			} else if err := fn(ev); err != nil {
				return err
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("failed to read %q: %w", path, readErr)
		}
	}
	return nil
}
