package copilotext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeVSCodeWorkspace creates a workspaceStorage/<hash> directory with the given
// workspace.json content (verbatim; empty skips creating the file).
func writeVSCodeWorkspace(t *testing.T, root, hash, workspaceJSON string) string {
	t.Helper()
	dir := filepath.Join(root, hash)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", dir, err)
	}
	if workspaceJSON != "" {
		if err := os.WriteFile(filepath.Join(dir, "workspace.json"), []byte(workspaceJSON), 0o644); err != nil {
			t.Fatalf("WriteFile(workspace.json) error = %v", err)
		}
	}
	return dir
}

// writeVSCodeSessionLog creates hashDir/GitHub.copilot-chat/debug-logs/<sessionID>/main.jsonl
// with the given content.
func writeVSCodeSessionLog(t *testing.T, hashDir, sessionID, mainJSONL string) string {
	t.Helper()
	dir := filepath.Join(hashDir, "GitHub.copilot-chat", "debug-logs", sessionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", dir, err)
	}
	path := filepath.Join(dir, "main.jsonl")
	if err := os.WriteFile(path, []byte(mainJSONL), 0o644); err != nil {
		t.Fatalf("WriteFile(main.jsonl) error = %v", err)
	}
	return path
}

func TestListVSCodeSessions(t *testing.T) {
	root := t.TempDir()

	hashA := writeVSCodeWorkspace(t, root, "hash-a", `{"folder":"file:///repo/a"}`)
	writeVSCodeSessionLog(t, hashA, "session-a", `{"type":"session_start"}`+"\n")

	// Multi-root workspaces record "workspace" instead of "folder"; sessions are still
	// listed, but with an empty Folder.
	hashB := writeVSCodeWorkspace(t, root, "hash-b", `{"workspace":"file:///repo/b.code-workspace"}`)
	writeVSCodeSessionLog(t, hashB, "session-b", `{"type":"session_start"}`+"\n")

	// A workspace with no debug-logs directory at all must be skipped silently.
	writeVSCodeWorkspace(t, root, "hash-c", `{"folder":"file:///repo/c"}`)

	sessions, err := ListVSCodeSessions(root)
	if err != nil {
		t.Fatalf("ListVSCodeSessions() error = %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("ListVSCodeSessions() len = %d, want 2: %+v", len(sessions), sessions)
	}

	byID := make(map[string]VSCodeSession)
	for _, s := range sessions {
		byID[s.ID] = s
	}
	if got := byID["session-a"].Folder; got != "/repo/a" {
		t.Fatalf("session-a Folder = %q, want /repo/a", got)
	}
	if got := byID["session-b"].Folder; got != "" {
		t.Fatalf("session-b (multi-root) Folder = %q, want empty", got)
	}
}

func TestListVSCodeSessionsMissingRoot(t *testing.T) {
	sessions, err := ListVSCodeSessions(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("ListVSCodeSessions() error = %v, want nil", err)
	}
	if sessions != nil {
		t.Fatalf("ListVSCodeSessions() = %+v, want nil", sessions)
	}
}

func TestFolderURIToPath(t *testing.T) {
	got, err := folderURIToPath("file:///Users/me/my%20project")
	if err != nil {
		t.Fatalf("folderURIToPath() error = %v", err)
	}
	if got != "/Users/me/my project" {
		t.Fatalf("folderURIToPath() = %q, want %q", got, "/Users/me/my project")
	}
}

func TestFolderURIToPathUnsupportedScheme(t *testing.T) {
	// Remote workspaces (SSH, dev containers, etc.) use non-"file" schemes and are not
	// resolvable to a local path; this must not be treated as an error.
	got, err := folderURIToPath("vscode-remote://ssh-remote%2Bmy-host/home/me/repo")
	if err != nil {
		t.Fatalf("folderURIToPath() error = %v, want nil", err)
	}
	if got != "" {
		t.Fatalf("folderURIToPath() = %q, want empty", got)
	}
}

func TestReadWorkspaceFolderMissingFile(t *testing.T) {
	folder, err := readWorkspaceFolder(t.TempDir())
	if err != nil {
		t.Fatalf("readWorkspaceFolder() error = %v, want nil", err)
	}
	if folder != "" {
		t.Fatalf("readWorkspaceFolder() = %q, want empty", folder)
	}
}

func TestReadWorkspaceFolderCorruptJSON(t *testing.T) {
	dir := writeVSCodeWorkspace(t, t.TempDir(), "hash", "not json")
	if _, err := readWorkspaceFolder(dir); err == nil {
		t.Fatal("readWorkspaceFolder() with corrupt workspace.json: expected error, got nil")
	}
}

func TestReadVSCodeEventsSkipsUnparsableLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.jsonl")
	content := `{"type":"session_start"}
not valid json
{"type":"user_message"}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var types []string
	if err := readVSCodeEvents(path, func(ev vscodeRawEvent) error {
		types = append(types, ev.Type)
		return nil
	}); err != nil {
		t.Fatalf("readVSCodeEvents() error = %v", err)
	}
	if len(types) != 2 || types[0] != "session_start" || types[1] != "user_message" {
		t.Fatalf("readVSCodeEvents() types = %v, want [session_start user_message]", types)
	}
}

func TestReadVSCodeEventsHandlesLinesLargerThan8MB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.jsonl")

	// bufio.Scanner's conventional max buffer size in this package is 8MB (see
	// readPermissionRecords); readVSCodeEvents must not silently drop lines longer than
	// that, since llm_request/tool_call attrs routinely exceed it.
	huge := strings.Repeat("x", 9*1024*1024)
	content := `{"type":"tool_call","name":"read_file","attrs":{"args":"` + huge + `"}}` + "\n" +
		`{"type":"user_message"}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var types []string
	if err := readVSCodeEvents(path, func(ev vscodeRawEvent) error {
		types = append(types, ev.Type)
		return nil
	}); err != nil {
		t.Fatalf("readVSCodeEvents() error = %v", err)
	}
	if len(types) != 2 || types[0] != "tool_call" || types[1] != "user_message" {
		t.Fatalf("readVSCodeEvents() types = %v, want [tool_call user_message]", types)
	}
}

func TestReadVSCodeEventsHandlesFractionalTimestamps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.jsonl")
	// Some events (e.g. "subagent") have been observed with a fractional epoch-ms "ts",
	// which must not cause the whole line to be dropped.
	content := `{"type":"subagent","ts":1781835067844.9155}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var got []vscodeRawEvent
	if err := readVSCodeEvents(path, func(ev vscodeRawEvent) error {
		got = append(got, ev)
		return nil
	}); err != nil {
		t.Fatalf("readVSCodeEvents() error = %v", err)
	}
	if len(got) != 1 || got[0].Type != "subagent" {
		t.Fatalf("readVSCodeEvents() got = %+v, want one subagent event", got)
	}
}

func TestReadVSCodeEventsMissingFile(t *testing.T) {
	err := readVSCodeEvents(filepath.Join(t.TempDir(), "does-not-exist.jsonl"), func(vscodeRawEvent) error {
		t.Fatal("callback should not be invoked for a missing file")
		return nil
	})
	if err != nil {
		t.Fatalf("readVSCodeEvents() error = %v, want nil", err)
	}
}
