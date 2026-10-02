package copilotext

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// Permission result kinds, as recorded in a "permission.completed" event's
// data.result.kind field. Unresolved is synthesized locally for requests that never
// received a matching completed event (e.g. an in-progress or interrupted session).
const (
	PermissionApproved            = "approved"
	PermissionDenied              = "denied"
	PermissionApprovedForLocation = "approved-for-location"
	PermissionUnresolved          = "unresolved"
)

// Decision source values, as recorded in a "permission.completed" event's
// data.decisionSource field. CLI versions older than 1.0.84-5 never set this field, so
// DecisionSourceUnknown is synthesized locally to mean "the CLI did not record who or what
// decided" rather than any particular decision maker.
const (
	DecisionSourceHumanResponse      = "human_response"
	DecisionSourceUnattendedFallback = "unattended_fallback"
	DecisionSourceUnknown            = "unknown"
)

// Session describes a single Copilot CLI session recorded under session-state.
type Session struct {
	ID         string
	Dir        string
	CWD        string
	ClientName string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PermissionCommand identifies one shell command segment covered by a permission request.
type PermissionCommand struct {
	Identifier string
	ReadOnly   bool
}

// PermissionRequest is the data recorded by a "permission.requested" event.
type PermissionRequest struct {
	RequestID       string
	ToolCallID      string
	Kind            string // e.g. "shell", "read"
	FullCommandText string
	Intention       string
	Commands        []PermissionCommand
	Paths           []string
	URLs            []string
	RequestedAt     time.Time
}

// PermissionRecord pairs a PermissionRequest with its outcome (Result), regardless of
// whether a matching "permission.completed" event was found.
type PermissionRecord struct {
	Session Session
	Request PermissionRequest
	Result  string
	// DecisionSource is the CLI's own record of who or what produced Result (e.g.
	// DecisionSourceHumanResponse), or DecisionSourceUnknown when the CLI version that
	// wrote this event never recorded it. It does not distinguish a rule-based
	// auto-approval from any other non-human source; as of CLI 1.0.88, rule-based
	// auto-approvals have not been observed to emit a "permission.completed" event at
	// all, so Result==PermissionApproved records collected so far are all
	// DecisionSourceHumanResponse or DecisionSourceUnknown.
	DecisionSource string
	DecidedAt      time.Time
}

// sessionStateRoot returns the directory containing one subdirectory per Copilot CLI
// session.
func sessionStateRoot() (string, error) {
	home, err := copilotHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "session-state"), nil
}

// workspaceMetadata mirrors the fields read from a session's workspace.yaml.
type workspaceMetadata struct {
	ID         string    `yaml:"id"`
	CWD        string    `yaml:"cwd"`
	ClientName string    `yaml:"client_name"`
	CreatedAt  time.Time `yaml:"created_at"`
	UpdatedAt  time.Time `yaml:"updated_at"`
}

// ListSessions returns every session found under root, one directory entry per session.
// Entries without a readable, parseable workspace.yaml are skipped with a warning, since
// they cannot be attributed to a working directory.
func ListSessions(root string) ([]Session, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list session state directory %q: %w", root, err)
	}

	sessions := make([]Session, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		s, err := readSession(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Warn("skipping unreadable session", "dir", dir, "error", err)
			}
			continue
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// readSession parses dir's workspace.yaml into a Session.
func readSession(dir string) (Session, error) {
	data, err := os.ReadFile(filepath.Join(dir, "workspace.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}, err
		}
		return Session{}, fmt.Errorf("failed to read workspace.yaml: %w", err)
	}
	var m workspaceMetadata
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Session{}, fmt.Errorf("failed to parse workspace.yaml: %w", err)
	}
	return Session{
		ID:         m.ID,
		Dir:        dir,
		CWD:        m.CWD,
		ClientName: m.ClientName,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}, nil
}

// rawEvent is the outer envelope shared by every line of events.jsonl.
type rawEvent struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// possibleURLEntry unmarshals a possibleUrls element, which the Copilot CLI encodes
// inconsistently as either a bare string or an object with a "url" field.
type possibleURLEntry string

func (e *possibleURLEntry) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*e = possibleURLEntry(s)
		return nil
	}
	var obj struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	*e = possibleURLEntry(obj.URL)
	return nil
}

// rawPermissionRequested is the data payload of a "permission.requested" event.
type rawPermissionRequested struct {
	RequestID         string `json:"requestId"`
	PermissionRequest struct {
		Kind            string `json:"kind"`
		ToolCallID      string `json:"toolCallId"`
		FullCommandText string `json:"fullCommandText"`
		Intention       string `json:"intention"`
		Commands        []struct {
			Identifier string `json:"identifier"`
			ReadOnly   bool   `json:"readOnly"`
		} `json:"commands"`
		PossiblePaths []string           `json:"possiblePaths"`
		PossibleURLs  []possibleURLEntry `json:"possibleUrls"`
	} `json:"permissionRequest"`
}

// rawPermissionCompleted is the data payload of a "permission.completed" event.
type rawPermissionCompleted struct {
	RequestID string `json:"requestId"`
	Result    struct {
		Kind string `json:"kind"`
	} `json:"result"`
	DecisionSource string `json:"decisionSource"`
}

// scanSessionEvents reads s's events.jsonl and calls fn once per parsed event, in file
// order. Lines that fail to parse are skipped with a warning rather than aborting the
// scan, since a session still in progress may have a truncated trailing line.
//
// events.jsonl is read line by line with a growable buffer, rather than json.Decoder's
// token stream, so a single truncated trailing line can be skipped without discarding
// every event that follows it.
func scanSessionEvents(s Session, fn func(rawEvent)) (rerr error) {
	path := filepath.Join(s.Dir, "events.jsonl")
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

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev rawEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			logger.Warn("skipping unparsable event", "dir", s.Dir, "error", err)
			continue
		}
		fn(ev)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read events.jsonl in %q: %w", s.Dir, err)
	}
	return nil
}

// readPermissionRecords reads s's events.jsonl and returns one PermissionRecord per
// "permission.requested" event, matched with its "permission.completed" outcome when one
// exists. Requests without a matching completion (e.g. from an interrupted session) are
// reported with Result set to PermissionUnresolved.
func readPermissionRecords(s Session) ([]PermissionRecord, error) {
	pc := newPermissionCollector(s)
	if err := scanSessionEvents(s, pc.handle); err != nil {
		return nil, err
	}
	return pc.records(), nil
}

// permissionCompletion is the outcome recorded by a "permission.completed" event.
type permissionCompletion struct {
	result         string
	decisionSource string
	decidedAt      time.Time
}

// permissionCollector accumulates "permission.requested" and "permission.completed" events
// of a single session so they can be gathered in a shared events.jsonl scan.
type permissionCollector struct {
	session  Session
	requests map[string]PermissionRequest
	order    []string
	results  map[string]permissionCompletion
}

// newPermissionCollector returns an empty permissionCollector for s.
func newPermissionCollector(s Session) *permissionCollector {
	return &permissionCollector{
		session:  s,
		requests: make(map[string]PermissionRequest),
		results:  make(map[string]permissionCompletion),
	}
}

// handle records ev when it is a permission request or completion event.
func (pc *permissionCollector) handle(ev rawEvent) {
	switch ev.Type {
	case "permission.requested":
		var data rawPermissionRequested
		if err := json.Unmarshal(ev.Data, &data); err != nil {
			logger.Warn("skipping unparsable permission.requested event", "dir", pc.session.Dir, "error", err)
			return
		}
		req := PermissionRequest{
			RequestID:       data.RequestID,
			ToolCallID:      data.PermissionRequest.ToolCallID,
			Kind:            data.PermissionRequest.Kind,
			FullCommandText: data.PermissionRequest.FullCommandText,
			Intention:       data.PermissionRequest.Intention,
			Paths:           data.PermissionRequest.PossiblePaths,
			RequestedAt:     ev.Timestamp,
		}
		for _, c := range data.PermissionRequest.Commands {
			req.Commands = append(req.Commands, PermissionCommand{Identifier: c.Identifier, ReadOnly: c.ReadOnly})
		}
		for _, u := range data.PermissionRequest.PossibleURLs {
			req.URLs = append(req.URLs, string(u))
		}
		if _, exists := pc.requests[req.RequestID]; !exists {
			pc.order = append(pc.order, req.RequestID)
		}
		pc.requests[req.RequestID] = req
	case "permission.completed":
		var data rawPermissionCompleted
		if err := json.Unmarshal(ev.Data, &data); err != nil {
			logger.Warn("skipping unparsable permission.completed event", "dir", pc.session.Dir, "error", err)
			return
		}
		pc.results[data.RequestID] = permissionCompletion{result: data.Result.Kind, decisionSource: data.DecisionSource, decidedAt: ev.Timestamp}
	}
}

// records returns one PermissionRecord per collected request, in first-seen order, or nil
// when no request was collected.
func (pc *permissionCollector) records() []PermissionRecord {
	if len(pc.order) == 0 {
		return nil
	}

	records := make([]PermissionRecord, 0, len(pc.order))
	for _, id := range pc.order {
		result := PermissionUnresolved
		decisionSource := DecisionSourceUnknown
		var decidedAt time.Time
		if r, ok := pc.results[id]; ok && r.result != "" {
			result = r.result
			decidedAt = r.decidedAt
			if r.decisionSource != "" {
				decisionSource = r.decisionSource
			}
		}
		records = append(records, PermissionRecord{
			Session:        pc.session,
			Request:        pc.requests[id],
			Result:         result,
			DecisionSource: decisionSource,
			DecidedAt:      decidedAt,
		})
	}
	return records
}

// SessionUsage is the usage summary recorded by a session's "session.shutdown" event: how
// many premium requests it consumed, its total AIU (Agentic/AI Usage Unit) cost, and its
// token and API duration totals.
type SessionUsage struct {
	At               time.Time
	Models           map[string]SessionUsage
	Requests         int
	PremiumRequests  float64
	AIU              float64
	InputTokens      int64
	CacheReadTokens  int64
	CacheWriteTokens int64
	OutputTokens     int64
	APIDurationMs    float64
}

// rawSessionShutdown is the data payload of a "session.shutdown" event. The usage fields are
// pointers so that a payload without usage totals (e.g. written by an older CLI version) can
// be told apart from one that recorded explicit zero values.
type rawSessionShutdown struct {
	ModelMetrics map[string]struct {
		Requests struct {
			Count int     `json:"count"`
			Cost  float64 `json:"cost"`
		} `json:"requests"`
		Usage struct {
			InputTokens      int64 `json:"inputTokens"`
			CacheReadTokens  int64 `json:"cacheReadTokens"`
			CacheWriteTokens int64 `json:"cacheWriteTokens"`
			OutputTokens     int64 `json:"outputTokens"`
		} `json:"usage"`
		rawSessionShutdown
	} `json:"modelMetrics"`
	TotalPremiumRequests *float64 `json:"totalPremiumRequests"`
	TotalNanoAiu         *int64   `json:"totalNanoAiu"`
	TotalAPIDurationMs   *float64 `json:"totalApiDurationMs"`
	TokenDetails         *struct {
		Input struct {
			TokenCount int64 `json:"tokenCount"`
		} `json:"input"`
		CacheRead struct {
			TokenCount int64 `json:"tokenCount"`
		} `json:"cache_read"`
		CacheWrite struct {
			TokenCount int64 `json:"tokenCount"`
		} `json:"cache_write"`
		Output struct {
			TokenCount int64 `json:"tokenCount"`
		} `json:"output"`
	} `json:"tokenDetails"`
}

// hasUsage reports whether the payload recorded any usage total.
func (d rawSessionShutdown) hasUsage() bool {
	return d.TotalPremiumRequests != nil || d.TotalNanoAiu != nil || d.TotalAPIDurationMs != nil || d.TokenDetails != nil || len(d.ModelMetrics) > 0
}

// readSessionUsage reads s's events.jsonl and returns the usage totals recorded by its
// "session.shutdown" event. found is false when the session has no such event yet (e.g. it
// is still in progress) or the CLI version that wrote it did not record usage totals.
func readSessionUsage(s Session) (usage SessionUsage, found bool, rerr error) {
	uc := &usageCollector{session: s}
	if err := scanSessionEvents(s, uc.handle); err != nil {
		return SessionUsage{}, false, err
	}
	return uc.usage, uc.found, nil
}

// usageCollector captures the usage totals of a single session's "session.shutdown" event
// so they can be gathered in a shared events.jsonl scan.
type usageCollector struct {
	session Session
	usage   SessionUsage
	found   bool
}

// handle records ev's usage totals when it is a "session.shutdown" event that contains them.
func (uc *usageCollector) handle(ev rawEvent) {
	if ev.Type != "session.shutdown" {
		return
	}
	var data rawSessionShutdown
	if err := json.Unmarshal(ev.Data, &data); err != nil {
		logger.Warn("skipping unparsable session.shutdown event", "dir", uc.session.Dir, "error", err)
		return
	}
	if !data.hasUsage() {
		return
	}
	usage := data.sessionUsage(ev.Timestamp)
	uc.usage = usage
	uc.found = true
}

func (data rawSessionShutdown) sessionUsage(at time.Time) SessionUsage {
	usage := SessionUsage{At: at}
	if data.TotalPremiumRequests != nil {
		usage.PremiumRequests = *data.TotalPremiumRequests
	}
	if data.TotalNanoAiu != nil {
		usage.AIU = float64(*data.TotalNanoAiu) / 1e9
	}
	if data.TotalAPIDurationMs != nil {
		usage.APIDurationMs = *data.TotalAPIDurationMs
	}
	if data.TokenDetails != nil {
		usage.InputTokens = data.TokenDetails.Input.TokenCount
		usage.CacheReadTokens = data.TokenDetails.CacheRead.TokenCount
		usage.CacheWriteTokens = data.TokenDetails.CacheWrite.TokenCount
		usage.OutputTokens = data.TokenDetails.Output.TokenCount
	}
	if len(data.ModelMetrics) > 0 {
		usage.Models = make(map[string]SessionUsage, len(data.ModelMetrics))
		for model, metrics := range data.ModelMetrics {
			modelUsage := metrics.rawSessionShutdown.sessionUsage(at)
			modelUsage.Requests = metrics.Requests.Count
			modelUsage.PremiumRequests = metrics.Requests.Cost
			if metrics.TokenDetails == nil {
				modelUsage.InputTokens = metrics.Usage.InputTokens
				modelUsage.CacheReadTokens = metrics.Usage.CacheReadTokens
				modelUsage.CacheWriteTokens = metrics.Usage.CacheWriteTokens
				modelUsage.OutputTokens = metrics.Usage.OutputTokens
			}
			usage.Requests += modelUsage.Requests
			usage.Models[model] = modelUsage
		}
	}
	return usage
}

// readSessionPermissionsAndUsage reads s's events.jsonl in a single pass and returns both
// its permission records (as readPermissionRecords) and its usage totals (as
// readSessionUsage).
func readSessionPermissionsAndUsage(s Session) (records []PermissionRecord, usage SessionUsage, usageFound bool, rerr error) {
	pc := newPermissionCollector(s)
	uc := &usageCollector{session: s}
	err := scanSessionEvents(s, func(ev rawEvent) {
		pc.handle(ev)
		uc.handle(ev)
	})
	if err != nil {
		return nil, SessionUsage{}, false, err
	}
	return pc.records(), uc.usage, uc.found, nil
}
