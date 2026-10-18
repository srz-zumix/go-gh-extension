package copilotext

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPermissionWaitTimingAndProvenance(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	stats := &PermissionWaitStats{}
	for _, tc := range []struct {
		source     string
		start, end time.Time
		completed  bool
	}{
		{DecisionSourceHumanResponse, start, start.Add(1250 * time.Millisecond), true},
		{DecisionSourceHumanResponse, start, start, true},
		{DecisionSourceHumanResponse, start, start.Add(-time.Second), true},
		{DecisionSourceHumanResponse, time.Time{}, start, true},
		{DecisionSourceHumanResponse, start, time.Time{}, true},
		{DecisionSourceUnattendedFallback, start, start.Add(time.Second), true},
		{"extension_policy", start, start.Add(2 * time.Second), true},
		{DecisionSourceUnknown, start, start.Add(3 * time.Second), true},
		{"", start, start.Add(4 * time.Second), true},
		{DecisionSourceHumanResponse, start, start.Add(time.Hour), false},
	} {
		stats.add(PermissionRecord{
			Request:   PermissionRequest{RequestedAt: tc.start},
			DecidedAt: tc.end, Completed: tc.completed, DecisionSource: tc.source,
		})
	}
	h := stats.HumanResponses
	if h.Completed != 5 || h.Timed != 2 || h.Untimed != 3 || h.TotalWaitMs != 1250 ||
		*h.AvgWaitMs != 625 || *h.MaxWaitMs != 1250 {
		t.Fatalf("human wait = %+v", h)
	}
	n, u := stats.NonHumanDecisions, stats.UnknownSourceDecisions
	if n.Completed != 2 || n.TotalWaitMs != 3000 || u.Completed != 2 || u.TotalWaitMs != 7000 || stats.Unresolved != 1 {
		t.Fatalf("other waits = %+v", stats)
	}
	empty := &PermissionWaitStats{}
	empty.add(PermissionRecord{Request: PermissionRequest{RequestedAt: start}})
	if empty.HumanResponses.AvgWaitMs != nil || empty.HumanResponses.MaxWaitMs != nil {
		t.Fatalf("unresolved request invented timing: %+v", empty)
	}
}

func TestPermissionPairingPreservesIdentityFirstEventsAndCompletionOrder(t *testing.T) {
	pc := newPermissionCollector(Session{})
	events := []string{
		`{"type":"permission.completed","timestamp":"2026-01-01T00:00:10Z","data":{"requestId":"same","result":{"kind":"approved"},"decisionSource":"human_response"}}`,
		`{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"same","permissionRequest":{"kind":"read"}}}`,
		`{"type":"permission.requested","timestamp":"2026-01-01T00:00:05Z","data":{"requestId":"same","permissionRequest":{"kind":"read"}}}`,
		`{"type":"permission.completed","timestamp":"2026-01-01T00:00:20Z","data":{"requestId":"same","result":{"kind":"approved"},"decisionSource":"human_response"}}`,
		`{"type":"permission.requested","agentId":"child","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"same","permissionRequest":{"kind":"read"}}}`,
		`{"type":"permission.completed","agentId":"child","timestamp":"2026-01-01T00:00:30Z","data":{"requestId":"same","result":{"kind":"denied"},"decisionSource":"human_response"}}`,
		`{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"permissionRequest":{"kind":"read"}}}`,
		`{"type":"permission.completed","timestamp":"2026-01-01T00:00:10Z","data":{"result":{"kind":"approved"}}}`,
		`{"type":"permission.completed","timestamp":"2026-01-01T00:00:10Z","data":{"requestId":"orphan","result":{"kind":"approved"}}}`,
		`{"type":"permission.requested","timestamp":"2026-01-01T00:00:00Z","data":{"requestId":"malformed","permissionRequest":{"kind":"read"}}}`,
		`{"type":"permission.completed","timestamp":"2026-01-01T00:00:10Z","data":{"requestId":"malformed","result":{}}}`,
	}
	for _, line := range events {
		var ev rawEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatal(err)
		}
		pc.handle(ev)
	}
	records := pc.records()
	if len(records) != 3 || !records[0].Completed || !records[1].Completed || records[2].Completed {
		t.Fatalf("paired records = %+v", records)
	}
	w := &PermissionWaitStats{}
	for _, rec := range records {
		w.add(rec)
	}
	if w.HumanResponses.Timed != 2 || w.HumanResponses.TotalWaitMs != 40000 || *w.HumanResponses.MaxWaitMs != 30000 || w.Unresolved != 1 {
		t.Fatalf("replayed events or agent identity changed timing: %+v", w)
	}
}

func permissionWaitFixture(t *testing.T, id string, minute int, kind, source, result, command, path, url string, duration time.Duration) string {
	t.Helper()
	start := time.Date(2026, 1, 1, 0, minute, 0, 0, time.UTC)
	pr := map[string]any{"kind": kind, "possiblePaths": []string{path}, "possibleUrls": []string{url}}
	if command != "" {
		pr["commands"] = []map[string]any{{"identifier": command, "readOnly": command == "list"}}
	}
	data, err := json.Marshal(map[string]any{"requestId": id, "permissionRequest": pr})
	if err != nil {
		t.Fatal(err)
	}
	line := fmt.Sprintf("{\"type\":\"permission.requested\",\"timestamp\":%q,\"data\":%s}\n", start.Format(time.RFC3339Nano), data)
	if result != "" {
		line += fmt.Sprintf("{\"type\":\"permission.completed\",\"timestamp\":%q,\"data\":{\"requestId\":%q,\"result\":{\"kind\":%q},\"decisionSource\":%q}}\n",
			start.Add(duration).Format(time.RFC3339Nano), id, result, source)
	}
	return line
}

func TestCollectPermissionWaitFollowsFiltersAndRequestStartBounds(t *testing.T) {
	root := t.TempDir()
	events := permissionWaitFixture(t, "a", 0, "shell", DecisionSourceHumanResponse, PermissionApproved, "list", "/repo/a/file", "https://example.com/a", 10500*time.Millisecond) +
		permissionWaitFixture(t, "b", 1, "read", "", PermissionApprovedForLocation, "", "/repo/a/other", "", 70*time.Second) +
		permissionWaitFixture(t, "c", 2, "shell", DecisionSourceUnattendedFallback, "denied-no-rule", "mutate", "/repo/a/other", "", 0) +
		permissionWaitFixture(t, "d", 3, "shell", DecisionSourceHumanResponse, "denied-interactively-by-user", "mutate", "/repo/a/other", "", 30*time.Second) +
		permissionWaitFixture(t, "e", 4, "write", "", "", "", "/repo/a/other", "", 0)
	writeSessionDir(t, root, "a", "id: a\ncwd: /repo/a\n", events)
	writeSessionDir(t, root, "b", "id: b\ncwd: /repo/b\n",
		permissionWaitFixture(t, "a", 0, "fetch", DecisionSourceHumanResponse, PermissionApproved, "", "/repo/b/file", "https://example.com/b", 0))
	all, err := CollectPermissionStats(PermissionStatsOptions{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	w := all.PermissionWait
	if all.Requests != 6 || w.HumanResponses.Completed != 3 || w.HumanResponses.TotalWaitMs != 40500 ||
		*w.HumanResponses.AvgWaitMs != 13500 || w.UnknownSourceDecisions.TotalWaitMs != 70000 ||
		w.NonHumanDecisions.Completed != 1 || w.Unresolved != 1 {
		t.Fatalf("requests = %d, wait = %+v", all.Requests, w)
	}
	for _, tc := range []struct {
		name                                        string
		opts                                        PermissionStatsOptions
		requests, human, unknown, other, unresolved int
		humanMs                                     float64
	}{
		{"top", PermissionStatsOptions{Top: 1}, 6, 3, 1, 1, 1, 40500},
		{"cwd", PermissionStatsOptions{Scope: SessionScope{ExactCWD: "/repo/a"}}, 5, 2, 1, 1, 1, 40500},
		{"session", PermissionStatsOptions{SessionID: "b"}, 1, 1, 0, 0, 0, 0},
		{"absent", PermissionStatsOptions{SessionID: "absent"}, 0, 0, 0, 0, 0, 0},
		{"operation", PermissionStatsOptions{Operations: []string{"write"}}, 4, 2, 0, 1, 1, 30000},
		{"kind", PermissionStatsOptions{Kind: []string{"read"}}, 1, 0, 1, 0, 0, 0},
		{"command", PermissionStatsOptions{Command: []string{"list"}}, 1, 1, 0, 0, 0, 10500},
		{"path", PermissionStatsOptions{Path: []string{"^/repo/a/file$"}}, 1, 1, 0, 0, 0, 10500},
		{"url", PermissionStatsOptions{URL: []string{"example.com/b"}}, 1, 1, 0, 0, 0, 0},
		{"human", PermissionStatsOptions{DecisionSource: []string{DecisionSourceHumanResponse}}, 3, 3, 0, 0, 0, 40500},
		{"unknown", PermissionStatsOptions{DecisionSource: []string{DecisionSourceUnknown}}, 2, 0, 1, 0, 1, 0},
		{"starts", PermissionStatsOptions{Since: time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC), Until: time.Date(2026, 1, 1, 0, 2, 0, 0, time.UTC)}, 1, 0, 1, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts.Root = root
			stats, err := CollectPermissionStats(tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			w := stats.PermissionWait
			if stats.Requests != tc.requests || w.HumanResponses.Completed != tc.human ||
				w.UnknownSourceDecisions.Completed != tc.unknown || w.NonHumanDecisions.Completed != tc.other ||
				w.Unresolved != tc.unresolved || w.HumanResponses.TotalWaitMs != tc.humanMs {
				t.Fatalf("requests = %d, wait = %+v", stats.Requests, w)
			}
			if tc.name == "starts" && w.UnknownSourceDecisions.TotalWaitMs != 70000 {
				t.Fatalf("completion after Until was lost: %+v", w)
			}
			if tc.name == "top" && !reflect.DeepEqual(w, all.PermissionWait) {
				t.Fatalf("Top truncated wait: %+v", w)
			}
			exported, err := json.Marshal(w)
			if err != nil || strings.Contains(string(exported), "/repo") {
				t.Fatalf("wait should export accounting only: %s, %v", exported, err)
			}
		})
	}
}
