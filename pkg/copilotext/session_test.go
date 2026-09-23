package copilotext

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeSessionDir creates dir with a workspace.yaml (unless workspaceYAML is empty) and an
// events.jsonl (unless eventsJSONL is empty).
func writeSessionDir(t *testing.T, root, name, workspaceYAML, eventsJSONL string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", dir, err)
	}
	if workspaceYAML != "" {
		if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(workspaceYAML), 0o644); err != nil {
			t.Fatalf("WriteFile(workspace.yaml) error = %v", err)
		}
	}
	if eventsJSONL != "" {
		if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), []byte(eventsJSONL), 0o644); err != nil {
			t.Fatalf("WriteFile(events.jsonl) error = %v", err)
		}
	}
	return dir
}

func TestListSessions(t *testing.T) {
	root := t.TempDir()
	writeSessionDir(t, root, "session-a",
		"id: session-a\ncwd: /repo/a\nclient_name: cli\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T01:00:00Z\n",
		"",
	)
	// A "pending-session:*" style directory with no workspace.yaml must be skipped
	// silently, not treated as an error.
	writeSessionDir(t, root, "pending-session:draft:123", "", "")

	sessions, err := ListSessions(root)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("ListSessions() len = %d, want 1: %+v", len(sessions), sessions)
	}
	got := sessions[0]
	if got.ID != "session-a" || got.CWD != "/repo/a" || got.ClientName != "cli" {
		t.Fatalf("ListSessions()[0] = %+v, want ID/CWD/ClientName session-a//repo/a/cli", got)
	}
	wantCreated := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !got.CreatedAt.Equal(wantCreated) {
		t.Fatalf("ListSessions()[0].CreatedAt = %v, want %v", got.CreatedAt, wantCreated)
	}
}

func TestListSessionsMissingRoot(t *testing.T) {
	sessions, err := ListSessions(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("ListSessions() error = %v, want nil", err)
	}
	if sessions != nil {
		t.Fatalf("ListSessions() = %+v, want nil", sessions)
	}
}

func TestReadSessionCorruptWorkspaceYAML(t *testing.T) {
	dir := writeSessionDir(t, t.TempDir(), "broken", "not: [valid: yaml", "")
	if _, err := readSession(dir); err == nil {
		t.Fatal("readSession() with corrupt workspace.yaml: expected error, got nil")
	} else if os.IsNotExist(err) {
		t.Fatalf("readSession() with corrupt workspace.yaml: error satisfies os.IsNotExist, want a parse error: %v", err)
	}
}

func TestReadPermissionRecords(t *testing.T) {
	events := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"req-1","permissionRequest":{"kind":"shell","toolCallId":"tc-1","fullCommandText":"ls","intention":"list files","commands":[{"identifier":"ls","readOnly":true}],"possiblePaths":["/repo"],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:01Z","data":{"requestId":"req-1","toolCallId":"tc-1","result":{"kind":"approved"}}}
{"type":"permission.requested","timestamp":"2026-01-01T00:00:02Z","data":{"requestId":"req-2","permissionRequest":{"kind":"shell","toolCallId":"tc-2","fullCommandText":"rm -rf /","intention":"cleanup","commands":[{"identifier":"rm","readOnly":false}],"possiblePaths":[],"possibleUrls":[]}}}
`
	dir := writeSessionDir(t, t.TempDir(), "session-b", "", events)
	s := Session{ID: "session-b", Dir: dir, CWD: "/repo"}

	records, err := readPermissionRecords(s)
	if err != nil {
		t.Fatalf("readPermissionRecords() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("readPermissionRecords() len = %d, want 2: %+v", len(records), records)
	}

	first := records[0]
	if first.Request.RequestID != "req-1" || first.Result != PermissionApproved {
		t.Fatalf("records[0] = %+v, want RequestID req-1, Result approved", first)
	}
	wantDecidedAt := time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC)
	if !first.DecidedAt.Equal(wantDecidedAt) {
		t.Fatalf("records[0].DecidedAt = %v, want %v", first.DecidedAt, wantDecidedAt)
	}

	// req-2 has no matching permission.completed event, so it must be reported as
	// unresolved rather than dropped.
	second := records[1]
	if second.Request.RequestID != "req-2" || second.Result != PermissionUnresolved {
		t.Fatalf("records[1] = %+v, want RequestID req-2, Result unresolved", second)
	}
	if !second.DecidedAt.IsZero() {
		t.Fatalf("records[1].DecidedAt = %v, want zero value", second.DecidedAt)
	}
}

func TestReadPermissionRecordsPossibleURLsObjectShape(t *testing.T) {
	// The Copilot CLI has been observed emitting possibleUrls entries as either bare
	// strings or objects with a "url" field; both shapes must be tolerated.
	events := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"req-1","permissionRequest":{"kind":"fetch","toolCallId":"tc-1","commands":[],"possiblePaths":[],"possibleUrls":["https://example.com/plain",{"url":"https://example.com/asset.tar.gz"}]}}}
`
	dir := writeSessionDir(t, t.TempDir(), "session-d", "", events)
	s := Session{ID: "session-d", Dir: dir}

	records, err := readPermissionRecords(s)
	if err != nil {
		t.Fatalf("readPermissionRecords() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("readPermissionRecords() len = %d, want 1: %+v", len(records), records)
	}
	want := []string{"https://example.com/plain", "https://example.com/asset.tar.gz"}
	got := records[0].Request.URLs
	if len(got) != len(want) {
		t.Fatalf("records[0].Request.URLs = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("records[0].Request.URLs = %+v, want %+v", got, want)
		}
	}
}

func TestReadPermissionRecordsTruncatedTrailingLine(t *testing.T) {
	// The trailing line is missing its closing brace, simulating a session that was
	// still writing events.jsonl when it was read.
	events := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"req-1","permissionRequest":{"kind":"read","toolCallId":"tc-1","commands":[],"possiblePaths":["/repo/file.go"],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:01Z","data":{"requestId":"req-1"`
	dir := writeSessionDir(t, t.TempDir(), "session-c", "", events)
	s := Session{ID: "session-c", Dir: dir}

	records, err := readPermissionRecords(s)
	if err != nil {
		t.Fatalf("readPermissionRecords() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("readPermissionRecords() len = %d, want 1: %+v", len(records), records)
	}
	if records[0].Result != PermissionUnresolved {
		t.Fatalf("records[0].Result = %q, want %q (truncated completed event should be skipped, not applied)", records[0].Result, PermissionUnresolved)
	}
}

func TestReadPermissionRecordsMissingFile(t *testing.T) {
	dir := t.TempDir()
	records, err := readPermissionRecords(Session{ID: "no-events", Dir: dir})
	if err != nil {
		t.Fatalf("readPermissionRecords() error = %v, want nil", err)
	}
	if records != nil {
		t.Fatalf("readPermissionRecords() = %+v, want nil", records)
	}
}
