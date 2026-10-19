package copilotext

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func TestReadLogSessionName(t *testing.T) {
	for _, tc := range []struct {
		name, ext, content, want string
	}{
		{"cli name", ".yaml", "name: Named task\nsummary: Generated task\n", "Named task"},
		{"cli summary", ".yaml", "summary: Generated task\n", "Generated task"},
		{"json", ".json", `{"customTitle":"Saved task"}`, "Saved task"},
		{"snapshot", ".jsonl", `{"kind":0,"v":{"customTitle":"Initial task"}}`, "Initial task"},
		{"rename", ".jsonl", "{\"kind\":0,\"v\":{\"customTitle\":\"Old task\"}}\n{\"kind\":1,\"k\":[\"customTitle\"],\"v\":\"New task\"}\n", "New task"},
		{"clear title", ".jsonl", "{\"kind\":0,\"v\":{\"customTitle\":\"Old task\"}}\n{\"kind\":1,\"k\":[\"customTitle\"],\"v\":null}\n", ""},
		{"no title", ".jsonl", `{"kind":0,"v":{"requests":[]}}`, ""},
		{"malformed", ".jsonl", `{invalid`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "session"+tc.ext)
			if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
				t.Fatal(err)
			}
			if got := readLogSessionName(path); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

const (
	testSessionA = "11111111-2222-3333-4444-555555555555"
	testSessionB = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
)

func writeLogFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// newLogTree builds a fake Copilot CLI home and VS Code user dir.
func newLogTree(t *testing.T) (home, userDir string) {
	t.Helper()
	root := t.TempDir()
	home = filepath.Join(root, "copilot")
	userDir = filepath.Join(root, "Code", "User")

	writeLogFile(t, filepath.Join(home, "session-state", testSessionA, "events.jsonl"))
	writeLogFile(t, filepath.Join(home, "session-state", ".session-operation-locks", "lock.json"))
	writeLogFile(t, filepath.Join(home, "logs", "session-"+testSessionB+".log"))
	writeLogFile(t, filepath.Join(userDir, "workspaceStorage", "ws1", "chatSessions", testSessionB+".jsonl"))
	writeLogFile(t, filepath.Join(userDir, "workspaceStorage", "ws1", "GitHub.copilot-chat", "debug-logs", testSessionA, "main.jsonl"))
	writeLogFile(t, filepath.Join(userDir, "globalStorage", "emptyWindowChatSessions", "no-id.json"))
	return home, userDir
}

func walkLogs(t *testing.T, opts LogOptions) map[string][]LogFile {
	t.Helper()
	out := map[string][]LogFile{}
	err := WalkLogFiles(opts, func(f LogFile) error {
		out[f.Source] = append(out[f.Source], f)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkLogFiles() error = %v", err)
	}
	return out
}

func TestWalkLogFilesDefaultLocations(t *testing.T) {
	home, userDir := newLogTree(t)
	got := walkLogs(t, LogOptions{CopilotHome: home, VSCodeUserDir: userDir})

	want := map[string]int{
		LogSourceCLISessions: 1, // the .session-operation-locks entry is skipped
		LogSourceCLILogs:     1,
		LogSourceVSCodeChat:  2,
		LogSourceVSCodeDebug: 1,
	}
	for source, n := range want {
		if len(got[source]) != n {
			t.Errorf("%s files = %d, want %d", source, len(got[source]), n)
		}
	}
	if f := got[LogSourceCLISessions][0]; f.SessionID != testSessionA || !filepath.IsAbs(f.Path) || f.ModifiedAt.IsZero() {
		t.Errorf("cli-sessions file = %+v", f)
	}
}

func TestLogWorkspace(t *testing.T) {
	root := t.TempDir()
	cli := filepath.Join(root, "session-state", testSessionA)
	writeLogFile(t, filepath.Join(cli, "workspace.yaml"))
	if err := os.WriteFile(filepath.Join(cli, "workspace.yaml"), []byte("cwd: /example/cli-project\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	vscode := filepath.Join(root, "workspaceStorage", "example-workspace")
	writeLogFile(t, filepath.Join(vscode, "workspace.json"))
	if err := os.WriteFile(filepath.Join(vscode, "workspace.json"), []byte(`{"folder":"file:///example/vscode%20project"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path string
		want string
	}{
		{filepath.Join(cli, "events.jsonl"), "/example/cli-project"},
		{filepath.Join(cli, "files", "nested.txt"), "/example/cli-project"},
		{filepath.Join(vscode, "GitHub.copilot-chat", "debug-logs", testSessionB, "main.jsonl"), "/example/vscode project"},
		{filepath.Join(root, "unknown", "log.txt"), ""},
	}
	for _, test := range tests {
		if got := logWorkspace(test.path); got != test.want {
			t.Errorf("logWorkspace(%q) = %q, want %q", test.path, got, test.want)
		}
	}
}

func TestWalkLogFilesSourceFilter(t *testing.T) {
	home, userDir := newLogTree(t)
	got := walkLogs(t, LogOptions{Sources: []string{LogSourceCLILogs}, CopilotHome: home, VSCodeUserDir: userDir})
	if len(got) != 1 || len(got[LogSourceCLILogs]) != 1 {
		t.Fatalf("got = %v, want only cli-logs", got)
	}
}

func TestWalkLogFilesUnknownSource(t *testing.T) {
	home, userDir := newLogTree(t)
	err := WalkLogFiles(LogOptions{Sources: []string{"bogus"}, CopilotHome: home, VSCodeUserDir: userDir}, func(LogFile) error { return nil })
	if err == nil {
		t.Fatal("expected error for unknown source, got nil")
	}
}

func TestWalkLogFilesSessionPrefixFilter(t *testing.T) {
	home, userDir := newLogTree(t)
	got := walkLogs(t, LogOptions{
		SessionPrefixes: []string{"AAAAAAAA-bbbb"}, // case-insensitive prefix
		CopilotHome:     home, VSCodeUserDir: userDir,
	})
	var sources []string
	for source, files := range got {
		sources = append(sources, source)
		for _, f := range files {
			if f.SessionID != testSessionB {
				t.Errorf("%s file has session %q, want %s", source, f.SessionID, testSessionB)
			}
		}
	}
	sort.Strings(sources)
	if len(sources) != 2 || sources[0] != LogSourceCLILogs || sources[1] != LogSourceVSCodeChat {
		t.Errorf("sources = %v, want cli-logs and vscode-chat", sources)
	}
}

func TestWalkLogFilesSinceFilter(t *testing.T) {
	home, userDir := newLogTree(t)
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(filepath.Join(home, "session-state", testSessionA, "events.jsonl"), old, old); err != nil {
		t.Fatal(err)
	}
	got := walkLogs(t, LogOptions{Since: time.Now().Add(-24 * time.Hour), CopilotHome: home, VSCodeUserDir: userDir})
	if len(got[LogSourceCLISessions]) != 0 {
		t.Errorf("old file was not skipped: %v", got[LogSourceCLISessions])
	}
	if len(got[LogSourceCLILogs]) != 1 {
		t.Errorf("recent file was skipped")
	}
}

func TestWalkLogFilesPathOverridesDefaults(t *testing.T) {
	home, userDir := newLogTree(t)
	custom := t.TempDir()
	writeLogFile(t, filepath.Join(custom, "x", "events.jsonl"))

	got := walkLogs(t, LogOptions{Paths: []string{custom}, CopilotHome: home, VSCodeUserDir: userDir})
	if len(got) != 1 || len(got[LogSourcePath]) != 1 {
		t.Fatalf("got = %v, want only the path source", got)
	}
}

func TestWalkLogFilesMissingPathIsAnError(t *testing.T) {
	err := WalkLogFiles(LogOptions{Paths: []string{filepath.Join(t.TempDir(), "missing")}}, func(LogFile) error { return nil })
	if err == nil {
		t.Fatal("expected error for a missing path, got nil")
	}
}

func TestWalkLogFilesMissingDefaultLocationsAreSkipped(t *testing.T) {
	root := t.TempDir()
	got := walkLogs(t, LogOptions{CopilotHome: filepath.Join(root, "a"), VSCodeUserDir: filepath.Join(root, "b")})
	if len(got) != 0 {
		t.Fatalf("got = %v, want no files", got)
	}
}

func TestSessionIDFromPath(t *testing.T) {
	tests := []struct {
		rel  string
		want string
	}{
		{testSessionA + "/events.jsonl", testSessionA},
		{"session-" + testSessionB + ".log", testSessionB},
		{"pending-session:" + testSessionA + "/x/toolu_1.json", testSessionA},
		{"process-1790779375563-7463.log", ""},
		{"UPPER-AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE.json", testSessionB},
	}
	for _, tt := range tests {
		if got := SessionIDFromPath(filepath.FromSlash(tt.rel)); got != tt.want {
			t.Errorf("SessionIDFromPath(%q) = %q, want %q", tt.rel, got, tt.want)
		}
	}
}
