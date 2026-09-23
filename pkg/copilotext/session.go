package copilotext

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	Session   Session
	Request   PermissionRequest
	Result    string
	DecidedAt time.Time
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
}

// readPermissionRecords reads s's events.jsonl and returns one PermissionRecord per
// "permission.requested" event, matched with its "permission.completed" outcome when one
// exists. Requests without a matching completion (e.g. from an interrupted session) are
// reported with Result set to PermissionUnresolved.
//
// events.jsonl is read line by line with a growable buffer, rather than json.Decoder's
// token stream, so a single truncated trailing line (from a session still in progress)
// can be skipped without discarding every event that follows it.
func readPermissionRecords(s Session) ([]PermissionRecord, error) {
	path := filepath.Join(s.Dir, "events.jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open %q: %w", path, err)
	}
	defer f.Close()

	requests := make(map[string]PermissionRequest)
	order := make([]string, 0)
	type completion struct {
		result    string
		decidedAt time.Time
	}
	results := make(map[string]completion)

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
		switch ev.Type {
		case "permission.requested":
			var data rawPermissionRequested
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				logger.Warn("skipping unparsable permission.requested event", "dir", s.Dir, "error", err)
				continue
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
			if _, exists := requests[req.RequestID]; !exists {
				order = append(order, req.RequestID)
			}
			requests[req.RequestID] = req
		case "permission.completed":
			var data rawPermissionCompleted
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				logger.Warn("skipping unparsable permission.completed event", "dir", s.Dir, "error", err)
				continue
			}
			results[data.RequestID] = completion{result: data.Result.Kind, decidedAt: ev.Timestamp}
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		logger.Warn("stopped reading events.jsonl early", "dir", s.Dir, "error", err)
	}

	records := make([]PermissionRecord, 0, len(order))
	for _, id := range order {
		result := PermissionUnresolved
		var decidedAt time.Time
		if r, ok := results[id]; ok && r.result != "" {
			result = r.result
			decidedAt = r.decidedAt
		}
		records = append(records, PermissionRecord{
			Session:   s,
			Request:   requests[id],
			Result:    result,
			DecidedAt: decidedAt,
		})
	}
	return records, nil
}
