package copilotext

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveSessionScopeAll(t *testing.T) {
	scope, err := ResolveSessionScope(context.Background(), true, "/some/cwd", "/some/worktree")
	if err != nil {
		t.Fatalf("ResolveSessionScope() error = %v", err)
	}
	if scope != (SessionScope{}) {
		t.Fatalf("ResolveSessionScope() = %+v, want empty scope", scope)
	}
	if !scope.matches("/anything") {
		t.Fatal("empty scope should match every cwd")
	}
}

func TestResolveSessionScopeExactCWD(t *testing.T) {
	dir := t.TempDir()
	scope, err := ResolveSessionScope(context.Background(), false, dir, "")
	if err != nil {
		t.Fatalf("ResolveSessionScope() error = %v", err)
	}
	normalized, err := normalizePath(dir)
	if err != nil {
		t.Fatalf("normalizePath() error = %v", err)
	}
	if scope.ExactCWD != normalized || scope.UnderDir != "" {
		t.Fatalf("ResolveSessionScope() = %+v, want ExactCWD=%q", scope, normalized)
	}
}

func TestResolveSessionScopeWorktree(t *testing.T) {
	// This repository's own working tree root is a stable, always-available git
	// worktree to resolve against.
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("filepath.Abs() error = %v", err)
	}
	scope, err := ResolveSessionScope(context.Background(), false, "", repoRoot)
	if err != nil {
		t.Fatalf("ResolveSessionScope() error = %v", err)
	}
	if scope.UnderDir == "" {
		t.Fatalf("ResolveSessionScope() = %+v, want a non-empty UnderDir", scope)
	}
	if !scope.matches(filepath.Join(repoRoot, "pkg", "copilotext")) {
		t.Fatalf("scope %+v should match a subdirectory of the resolved worktree", scope)
	}
}

func TestResolveSessionScopeWorktreeFallsBackOutsideGit(t *testing.T) {
	dir := t.TempDir()
	scope, err := ResolveSessionScope(context.Background(), false, "", dir)
	if err != nil {
		t.Fatalf("ResolveSessionScope() error = %v", err)
	}
	normalized, err := normalizePath(dir)
	if err != nil {
		t.Fatalf("normalizePath() error = %v", err)
	}
	if scope.ExactCWD != normalized {
		t.Fatalf("ResolveSessionScope() outside git = %+v, want ExactCWD=%q (fallback)", scope, normalized)
	}
}

func TestSessionScopeMatchesUnderDirDoesNotMatchSiblingWithSharedPrefix(t *testing.T) {
	scope := SessionScope{UnderDir: "/a/b"}
	if scope.matches("/a/bc") {
		t.Fatal("scope under /a/b should not match /a/bc")
	}
	if !scope.matches("/a/b/c") {
		t.Fatal("scope under /a/b should match /a/b/c")
	}
	if !scope.matches("/a/b") {
		t.Fatal("scope under /a/b should match /a/b itself")
	}
}

func TestCollectPermissionStats(t *testing.T) {
	root := t.TempDir()

	eventsA := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"req-1","permissionRequest":{"kind":"shell","toolCallId":"tc-1","commands":[{"identifier":"ls","readOnly":true}],"possiblePaths":["/repo/a"],"possibleUrls":["https://example.com"]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:01Z","data":{"requestId":"req-1","result":{"kind":"approved"}}}
{"type":"permission.requested","timestamp":"2026-01-02T00:00:00Z","data":{"requestId":"req-2","permissionRequest":{"kind":"shell","toolCallId":"tc-2","commands":[{"identifier":"ls","readOnly":true},{"identifier":"ls","readOnly":true}],"possiblePaths":["/repo/a","/repo/a"],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-02T00:00:01Z","data":{"requestId":"req-2","result":{"kind":"denied"}}}
`
	writeSessionDir(t, root, "session-a",
		"id: session-a\ncwd: /repo/a\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n",
		eventsA,
	)

	eventsB := `{"type":"permission.requested","timestamp":"2026-01-03T00:00:00Z","data":{"requestId":"req-3","permissionRequest":{"kind":"read","toolCallId":"tc-3","commands":[],"possiblePaths":["/repo/b"],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-03T00:00:01Z","data":{"requestId":"req-3","result":{"kind":"approved-for-location"}}}
`
	writeSessionDir(t, root, "session-b",
		"id: session-b\ncwd: /repo/b\ncreated_at: 2026-01-03T00:00:00Z\nupdated_at: 2026-01-03T00:00:00Z\n",
		eventsB,
	)

	// A pending, not-yet-started session must be silently excluded, not error out.
	writeSessionDir(t, root, "pending-session:draft:xyz", "", "")

	stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root})
	if err != nil {
		t.Fatalf("CollectPermissionStats() error = %v", err)
	}

	if stats.Sessions != 2 {
		t.Fatalf("Sessions = %d, want 2", stats.Sessions)
	}
	if stats.Requests != 3 {
		t.Fatalf("Requests = %d, want 3", stats.Requests)
	}

	// req-2 lists "ls" twice and "/repo/a" twice; both must be counted once per request.
	if len(stats.ByCommand) != 1 || stats.ByCommand[0].Key != "ls" || stats.ByCommand[0].Total != 2 {
		t.Fatalf("ByCommand = %+v, want single entry ls total=2", stats.ByCommand)
	}
	byPath := map[string]int{}
	for _, c := range stats.ByPath {
		byPath[c.Key] = c.Total
	}
	if byPath["/repo/a"] != 2 || byPath["/repo/b"] != 1 {
		t.Fatalf("ByPath = %+v, want /repo/a=2 /repo/b=1", stats.ByPath)
	}

	byResult := map[string]int{}
	for _, c := range stats.ByResult {
		byResult[c.Key] = c.Total
	}
	if byResult[PermissionApproved] != 1 || byResult[PermissionDenied] != 1 || byResult[PermissionApprovedForLocation] != 1 {
		t.Fatalf("ByResult = %+v, want approved=1 denied=1 approved-for-location=1", stats.ByResult)
	}
}

func TestCollectPermissionStatsSince(t *testing.T) {
	root := t.TempDir()
	events := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"req-old","permissionRequest":{"kind":"shell","toolCallId":"tc-1","commands":[{"identifier":"old","readOnly":true}],"possiblePaths":[],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:01Z","data":{"requestId":"req-old","result":{"kind":"approved"}}}
{"type":"permission.requested","timestamp":"2026-02-01T00:00:00Z","data":{"requestId":"req-new","permissionRequest":{"kind":"shell","toolCallId":"tc-2","commands":[{"identifier":"new","readOnly":true}],"possiblePaths":[],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-02-01T00:00:01Z","data":{"requestId":"req-new","result":{"kind":"approved"}}}
`
	writeSessionDir(t, root, "session-a",
		"id: session-a\ncwd: /repo/a\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n",
		events,
	)

	stats, err := CollectPermissionStats(PermissionStatsOptions{
		Root:  root,
		Since: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("CollectPermissionStats() error = %v", err)
	}
	if stats.Requests != 1 {
		t.Fatalf("Requests = %d, want 1", stats.Requests)
	}
	if len(stats.ByCommand) != 1 || stats.ByCommand[0].Key != "new" {
		t.Fatalf("ByCommand = %+v, want single entry new", stats.ByCommand)
	}
}

func TestCollectPermissionStatsSessionIDFilter(t *testing.T) {
	root := t.TempDir()
	events := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"req-1","permissionRequest":{"kind":"shell","toolCallId":"tc-1","commands":[{"identifier":"ls","readOnly":true}],"possiblePaths":[],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:01Z","data":{"requestId":"req-1","result":{"kind":"approved"}}}
`
	writeSessionDir(t, root, "session-a",
		"id: session-a\ncwd: /repo/a\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n",
		events,
	)
	writeSessionDir(t, root, "session-b",
		"id: session-b\ncwd: /repo/b\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n",
		events,
	)

	stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, SessionID: "session-b"})
	if err != nil {
		t.Fatalf("CollectPermissionStats() error = %v", err)
	}
	if stats.Sessions != 1 || stats.Requests != 1 {
		t.Fatalf("Sessions/Requests = %d/%d, want 1/1", stats.Sessions, stats.Requests)
	}
}

func TestCollectPermissionStatsFilters(t *testing.T) {
	root := t.TempDir()
	events := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"req-read","permissionRequest":{"kind":"read","toolCallId":"tc-1","commands":[],"possiblePaths":["/repo/a/file.go"],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:01Z","data":{"requestId":"req-read","result":{"kind":"approved"}}}
{"type":"permission.requested","timestamp":"2026-01-01T00:00:02Z","data":{"requestId":"req-shell-ro","permissionRequest":{"kind":"shell","toolCallId":"tc-2","commands":[{"identifier":"ls","readOnly":true}],"possiblePaths":[],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:03Z","data":{"requestId":"req-shell-ro","result":{"kind":"approved"}}}
{"type":"permission.requested","timestamp":"2026-01-01T00:00:04Z","data":{"requestId":"req-shell-rw","permissionRequest":{"kind":"shell","toolCallId":"tc-3","commands":[{"identifier":"rm","readOnly":false}],"possiblePaths":[],"possibleUrls":["https://example.com/asset"]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:05Z","data":{"requestId":"req-shell-rw","result":{"kind":"approved"}}}
`
	writeSessionDir(t, root, "session-a",
		"id: session-a\ncwd: /repo/a\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n",
		events,
	)

	t.Run("Operations read", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Operations: []string{"read"}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 2 {
			t.Fatalf("Requests = %d, want 2 (req-read, req-shell-ro)", stats.Requests)
		}
	})

	t.Run("Operations write", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Operations: []string{"write"}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 1 {
			t.Fatalf("Requests = %d, want 1 (req-shell-rw)", stats.Requests)
		}
	})

	t.Run("Operations read and write", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Operations: []string{"read", "write"}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 3 {
			t.Fatalf("Requests = %d, want 3 (no filtering)", stats.Requests)
		}
	})

	t.Run("Operations invalid", func(t *testing.T) {
		_, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Operations: []string{"bogus"}})
		if err == nil {
			t.Fatal("CollectPermissionStats() error = nil, want error for invalid operation")
		}
	})

	t.Run("Kind", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Kind: []string{"read"}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 1 {
			t.Fatalf("Requests = %d, want 1 (req-read)", stats.Requests)
		}
	})

	t.Run("Kind multiple", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Kind: []string{"read", "shell"}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 3 {
			t.Fatalf("Requests = %d, want 3", stats.Requests)
		}
	})

	t.Run("Command", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Command: []string{"rm"}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 1 {
			t.Fatalf("Requests = %d, want 1 (req-shell-rw)", stats.Requests)
		}
	})

	t.Run("Command multiple", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Command: []string{"rm", "ls"}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 2 {
			t.Fatalf("Requests = %d, want 2 (req-shell-ro, req-shell-rw)", stats.Requests)
		}
	})

	t.Run("Path substring", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Path: []string{"file.go"}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 1 {
			t.Fatalf("Requests = %d, want 1 (req-read)", stats.Requests)
		}
	})

	t.Run("Path regex", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Path: []string{`\.go$`}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 1 {
			t.Fatalf("Requests = %d, want 1 (req-read)", stats.Requests)
		}
	})

	t.Run("Path invalid regex", func(t *testing.T) {
		_, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Path: []string{"("}})
		if err == nil {
			t.Fatal("CollectPermissionStats() error = nil, want error for invalid path pattern")
		}
	})

	t.Run("URL substring", func(t *testing.T) {
		stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, URL: []string{"example.com"}})
		if err != nil {
			t.Fatalf("CollectPermissionStats() error = %v", err)
		}
		if stats.Requests != 1 {
			t.Fatalf("Requests = %d, want 1 (req-shell-rw)", stats.Requests)
		}
	})
}

func TestCollectPermissionStatsTop(t *testing.T) {
	root := t.TempDir()
	events := `{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"req-1","permissionRequest":{"kind":"shell","toolCallId":"tc-1","commands":[{"identifier":"a","readOnly":true}],"possiblePaths":[],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:01Z","data":{"requestId":"req-1","result":{"kind":"approved"}}}
{"type":"permission.requested","timestamp":"2026-01-01T00:00:02Z","data":{"requestId":"req-2","permissionRequest":{"kind":"shell","toolCallId":"tc-2","commands":[{"identifier":"b","readOnly":true}],"possiblePaths":[],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:03Z","data":{"requestId":"req-2","result":{"kind":"approved"}}}
{"type":"permission.requested","timestamp":"2026-01-01T00:00:04Z","data":{"requestId":"req-3","permissionRequest":{"kind":"shell","toolCallId":"tc-3","commands":[{"identifier":"b","readOnly":true}],"possiblePaths":[],"possibleUrls":[]}}}
{"type":"permission.completed","timestamp":"2026-01-01T00:00:05Z","data":{"requestId":"req-3","result":{"kind":"approved"}}}
`
	writeSessionDir(t, root, "session-a",
		"id: session-a\ncwd: /repo/a\ncreated_at: 2026-01-01T00:00:00Z\nupdated_at: 2026-01-01T00:00:00Z\n",
		events,
	)

	stats, err := CollectPermissionStats(PermissionStatsOptions{Root: root, Top: 1})
	if err != nil {
		t.Fatalf("CollectPermissionStats() error = %v", err)
	}
	if len(stats.ByCommand) != 1 || stats.ByCommand[0].Key != "b" {
		t.Fatalf("ByCommand = %+v, want single entry b (highest total)", stats.ByCommand)
	}
}
