package copilotext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/gitutil"
)

// SessionScope restricts which sessions CollectPermissionStats considers. Both fields
// empty means every session is in scope.
type SessionScope struct {
	// ExactCWD, when non-empty, matches only sessions whose recorded working directory
	// is exactly this path.
	ExactCWD string
	// UnderDir, when non-empty, matches sessions whose recorded working directory is
	// this path or a descendant of it.
	UnderDir string
}

// normalizePath resolves path to an absolute, symlink-free form so that scope comparisons
// are not fooled by relative paths or OS symlinks (e.g. macOS's /tmp -> /private/tmp).
// It falls back to the cleaned absolute path when the path does not exist yet.
func normalizePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %q: %w", path, err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved, nil
	}
	return abs, nil
}

// ResolveSessionScope turns the --all/--cwd/--worktree command-line flags into a
// SessionScope. Exactly one of all, cwd, or worktree is expected to be set by the caller
// (the command enforces this via mutually exclusive flags); when none are set, worktree
// defaults to the current directory.
//
// worktree is resolved to its git worktree root via `git rev-parse --show-toplevel`. When
// worktree is not inside a git working tree, ResolveSessionScope falls back to treating it
// as an exact cwd match instead of failing, since a non-repo directory has no "worktree" to
// scope by.
func ResolveSessionScope(ctx context.Context, all bool, cwd string, worktree string) (SessionScope, error) {
	if all {
		return SessionScope{}, nil
	}
	if cwd != "" {
		normalized, err := normalizePath(cwd)
		if err != nil {
			return SessionScope{}, err
		}
		return SessionScope{ExactCWD: normalized}, nil
	}

	dir := worktree
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return SessionScope{}, fmt.Errorf("failed to determine current directory: %w", err)
		}
		dir = wd
	}

	top, err := gitutil.ClientForDir(dir).ToplevelDir(ctx)
	if err != nil {
		normalized, err := normalizePath(dir)
		if err != nil {
			return SessionScope{}, err
		}
		return SessionScope{ExactCWD: normalized}, nil
	}
	normalized, err := normalizePath(top)
	if err != nil {
		return SessionScope{}, err
	}
	return SessionScope{UnderDir: normalized}, nil
}

// matches reports whether cwd (a session's recorded working directory) falls within scope.
// An empty scope matches every cwd.
func (scope SessionScope) matches(cwd string) bool {
	if scope.ExactCWD == "" && scope.UnderDir == "" {
		return true
	}
	if cwd == "" {
		return false
	}
	normalized, err := normalizePath(cwd)
	if err != nil {
		return false
	}
	if scope.ExactCWD != "" {
		return normalized == scope.ExactCWD
	}
	if normalized == scope.UnderDir {
		return true
	}
	rel, err := filepath.Rel(scope.UnderDir, normalized)
	if err != nil {
		return false
	}
	return rel != ".." && !hasParentPrefix(rel)
}

// hasParentPrefix reports whether rel escapes its base via a leading "../" segment.
func hasParentPrefix(rel string) bool {
	return len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator)
}

// PermissionStatsOptions configures CollectPermissionStats.
type PermissionStatsOptions struct {
	// Root overrides the session-state directory to scan. Empty uses the default
	// (COPILOT_HOME or ~/.copilot)/session-state.
	Root string
	// Since and Until, when non-zero, restrict counted requests to
	// Since <= RequestedAt < Until.
	Since, Until time.Time
	// Scope restricts which sessions are considered.
	Scope SessionScope
	// SessionID, when non-empty, restricts to a single session by ID.
	SessionID string
	// Operations, when non-empty, restricts to requests classified as any of these
	// values ("read" or "write"). A request is "read" when its Kind is "read", or when
	// its Kind is "shell" and every command is read-only; otherwise it is "write".
	Operations []string
	// Kind, when non-empty, restricts to requests whose Kind matches any of these exact
	// values (e.g. "shell", "read", "write").
	Kind []string
	// DecisionSource, when non-empty, restricts to records whose DecisionSource matches
	// any of these exact values (e.g. DecisionSourceHumanResponse, DecisionSourceUnknown).
	DecisionSource []string
	// Command, when non-empty, restricts to requests that include any of these exact
	// command identifiers.
	Command []string
	// Path, when non-empty, restricts to requests with at least one recorded path
	// matching any of these regular expressions (a plain substring is also a valid,
	// unanchored regular expression).
	Path []string
	// URL, when non-empty, restricts to requests with at least one recorded URL matching
	// any of these regular expressions (a plain substring is also a valid, unanchored
	// regular expression).
	URL []string
	// Top, when > 0, keeps only the top N entries (by Total, then Key) of each axis.
	// 0 keeps every entry.
	Top int
}

// Count is one entry of a PermissionStats axis: a key (e.g. a command identifier) and the
// number of requests observed for it, broken down by result.
type Count struct {
	Key                 string
	Total               int
	Approved            int
	Denied              int
	ApprovedForLocation int
	Unresolved          int
}

// UsageCount is one entry of PermissionStats.ByCWDUsage: a working directory and its
// aggregated session usage totals, drawn from each in-scope session's "session.shutdown"
// event.
type UsageCount struct {
	Key              string
	Sessions         int
	PremiumRequests  float64
	AIU              float64
	InputTokens      int64
	CacheReadTokens  int64
	CacheWriteTokens int64
	OutputTokens     int64
	APIDurationMs    float64
}

// PermissionStats summarizes the permission requests and session usage collected by
// CollectPermissionStats.
type PermissionStats struct {
	Sessions int
	Requests int
	Since    time.Time
	Until    time.Time
	Scope    SessionScope

	ByResult         []Count
	ByDecisionSource []Count
	ByKind           []Count
	ByReadOnly       []Count
	ByCommand        []Count
	ByPath           []Count
	ByURL            []Count
	ByCWD            []Count

	// UsageSessions is the number of in-scope sessions with a "session.shutdown" event
	// recording usage totals, independent of whether they had any permission requests.
	UsageSessions         int
	UsagePremiumRequests  float64
	UsageAIU              float64
	UsageInputTokens      int64
	UsageCacheReadTokens  int64
	UsageCacheWriteTokens int64
	UsageOutputTokens     int64
	UsageAPIDurationMs    float64
	ByCWDUsage            []UsageCount
}

// counter accumulates Count entries keyed by an arbitrary string, preserving first-seen
// order until the final sort.
type counter struct {
	order []string
	byKey map[string]*Count
}

func newCounter() *counter {
	return &counter{byKey: make(map[string]*Count)}
}

func (c *counter) add(key, result string) {
	entry, ok := c.byKey[key]
	if !ok {
		entry = &Count{Key: key}
		c.byKey[key] = entry
		c.order = append(c.order, key)
	}
	entry.Total++
	// The CLI records outcome variants such as
	// "denied-no-approval-rule-and-could-not-request-from-user"; classify any hyphenated
	// "approved-"/"denied-" variant with its base outcome so real denials are not mistaken
	// for unresolved requests. "approved-for-location" keeps its own dedicated tally and is
	// therefore matched before the generic "approved" case.
	switch {
	case result == PermissionApprovedForLocation:
		entry.ApprovedForLocation++
	case result == PermissionApproved || strings.HasPrefix(result, PermissionApproved+"-"):
		entry.Approved++
	case result == PermissionDenied || strings.HasPrefix(result, PermissionDenied+"-"):
		entry.Denied++
	default:
		entry.Unresolved++
	}
}

// finish returns the accumulated counts sorted by Total descending, then Key ascending,
// truncated to the top entries when top > 0.
func (c *counter) finish(top int) []Count {
	counts := make([]Count, 0, len(c.order))
	for _, key := range c.order {
		counts = append(counts, *c.byKey[key])
	}
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Total != counts[j].Total {
			return counts[i].Total > counts[j].Total
		}
		return counts[i].Key < counts[j].Key
	})
	if top > 0 && len(counts) > top {
		counts = counts[:top]
	}
	return counts
}

// usageCounter accumulates UsageCount entries keyed by an arbitrary string (e.g. a
// session's CWD), preserving first-seen order until the final sort.
type usageCounter struct {
	order []string
	byKey map[string]*UsageCount
}

func newUsageCounter() *usageCounter {
	return &usageCounter{byKey: make(map[string]*UsageCount)}
}

func (c *usageCounter) add(key string, u SessionUsage) {
	entry, ok := c.byKey[key]
	if !ok {
		entry = &UsageCount{Key: key}
		c.byKey[key] = entry
		c.order = append(c.order, key)
	}
	entry.Sessions++
	entry.PremiumRequests += u.PremiumRequests
	entry.AIU += u.AIU
	entry.InputTokens += u.InputTokens
	entry.CacheReadTokens += u.CacheReadTokens
	entry.CacheWriteTokens += u.CacheWriteTokens
	entry.OutputTokens += u.OutputTokens
	entry.APIDurationMs += u.APIDurationMs
}

// finish returns the accumulated usage counts sorted by AIU descending, then Key
// ascending, truncated to the top entries when top > 0.
func (c *usageCounter) finish(top int) []UsageCount {
	counts := make([]UsageCount, 0, len(c.order))
	for _, key := range c.order {
		counts = append(counts, *c.byKey[key])
	}
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].AIU != counts[j].AIU {
			return counts[i].AIU > counts[j].AIU
		}
		return counts[i].Key < counts[j].Key
	})
	if top > 0 && len(counts) > top {
		counts = counts[:top]
	}
	return counts
}

// CollectPermissionStats scans the Copilot CLI session-state directory and aggregates
// permission requests and session usage into a PermissionStats, applying opts' filters.
func CollectPermissionStats(opts PermissionStatsOptions) (*PermissionStats, error) {
	root := opts.Root
	if root == "" {
		r, err := sessionStateRoot()
		if err != nil {
			return nil, err
		}
		root = r
	}

	sessions, err := ListSessions(root)
	if err != nil {
		return nil, err
	}

	filters, err := newPermissionFilters(opts)
	if err != nil {
		return nil, err
	}

	byResult := newCounter()
	byDecisionSource := newCounter()
	byKind := newCounter()
	byReadOnly := newCounter()
	byCommand := newCounter()
	byPath := newCounter()
	byURL := newCounter()
	byCWD := newCounter()
	byCWDUsage := newUsageCounter()

	stats := &PermissionStats{Since: opts.Since, Until: opts.Until, Scope: opts.Scope}

	for _, s := range sessions {
		if opts.SessionID != "" && s.ID != opts.SessionID {
			continue
		}
		if !opts.Scope.matches(s.CWD) {
			continue
		}

		// Usage is aggregated independently of the permission requests below: a session
		// with no permission requests (e.g. one that only ran read-only tools with no
		// approval rules) still has a usage total worth reporting.
		usage, found, err := readSessionUsage(s)
		if err != nil {
			return nil, fmt.Errorf("failed to read usage events for session %q: %w", s.ID, err)
		}
		if found && (opts.Since.IsZero() || !usage.At.Before(opts.Since)) && (opts.Until.IsZero() || usage.At.Before(opts.Until)) {
			stats.UsageSessions++
			stats.UsagePremiumRequests += usage.PremiumRequests
			stats.UsageAIU += usage.AIU
			stats.UsageInputTokens += usage.InputTokens
			stats.UsageCacheReadTokens += usage.CacheReadTokens
			stats.UsageCacheWriteTokens += usage.CacheWriteTokens
			stats.UsageOutputTokens += usage.OutputTokens
			stats.UsageAPIDurationMs += usage.APIDurationMs
			byCWDUsage.add(s.CWD, usage)
		}

		records, err := readPermissionRecords(s)
		if err != nil {
			return nil, fmt.Errorf("failed to read permission events for session %q: %w", s.ID, err)
		}
		if len(records) == 0 {
			continue
		}

		sessionCounted := false
		for _, rec := range records {
			if !opts.Since.IsZero() && rec.Request.RequestedAt.Before(opts.Since) {
				continue
			}
			if !opts.Until.IsZero() && !rec.Request.RequestedAt.Before(opts.Until) {
				continue
			}
			if !filters.matches(rec.Request) {
				continue
			}
			if !filters.matchesDecisionSource(rec.DecisionSource) {
				continue
			}

			sessionCounted = true
			stats.Requests++

			byResult.add(rec.Result, rec.Result)
			byDecisionSource.add(rec.DecisionSource, rec.Result)
			byKind.add(rec.Request.Kind, rec.Result)
			byCWD.add(s.CWD, rec.Result)

			seenCommand := make(map[string]bool)
			seenReadOnly := make(map[bool]bool)
			for _, cmd := range rec.Request.Commands {
				if !seenCommand[cmd.Identifier] {
					seenCommand[cmd.Identifier] = true
					byCommand.add(cmd.Identifier, rec.Result)
				}
				if !seenReadOnly[cmd.ReadOnly] {
					seenReadOnly[cmd.ReadOnly] = true
					byReadOnly.add(readOnlyKey(cmd.ReadOnly), rec.Result)
				}
			}

			seenPath := make(map[string]bool)
			for _, p := range rec.Request.Paths {
				if !seenPath[p] {
					seenPath[p] = true
					byPath.add(p, rec.Result)
				}
			}

			seenURL := make(map[string]bool)
			for _, u := range rec.Request.URLs {
				if !seenURL[u] {
					seenURL[u] = true
					byURL.add(u, rec.Result)
				}
			}
		}

		if sessionCounted {
			stats.Sessions++
		}
	}

	stats.ByResult = byResult.finish(opts.Top)
	stats.ByDecisionSource = byDecisionSource.finish(opts.Top)
	stats.ByKind = byKind.finish(opts.Top)
	stats.ByReadOnly = byReadOnly.finish(opts.Top)
	stats.ByCommand = byCommand.finish(opts.Top)
	stats.ByPath = byPath.finish(opts.Top)
	stats.ByURL = byURL.finish(opts.Top)
	stats.ByCWD = byCWD.finish(opts.Top)
	stats.ByCWDUsage = byCWDUsage.finish(opts.Top)

	return stats, nil
}

// permissionFilters is the compiled form of PermissionStatsOptions' Operations/Kind/
// Command/Path/URL filters, built once per CollectPermissionStats call so that regular
// expressions aren't recompiled per record.
type permissionFilters struct {
	operations      map[string]bool
	kinds           map[string]bool
	decisionSources map[string]bool
	commands        map[string]bool
	paths           []*regexp.Regexp
	urls            []*regexp.Regexp
}

// newPermissionFilters compiles opts' filters, validating Operations values and Path/URL
// regular expressions.
func newPermissionFilters(opts PermissionStatsOptions) (*permissionFilters, error) {
	f := &permissionFilters{}

	if len(opts.Operations) > 0 {
		f.operations = make(map[string]bool, len(opts.Operations))
		for _, op := range opts.Operations {
			if op != "read" && op != "write" {
				return nil, fmt.Errorf("invalid operation %q: want %q or %q", op, "read", "write")
			}
			f.operations[op] = true
		}
	}
	if len(opts.Kind) > 0 {
		f.kinds = toSet(opts.Kind)
	}
	if len(opts.DecisionSource) > 0 {
		f.decisionSources = toSet(opts.DecisionSource)
	}
	if len(opts.Command) > 0 {
		f.commands = toSet(opts.Command)
	}
	for _, p := range opts.Path {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("invalid path pattern %q: %w", p, err)
		}
		f.paths = append(f.paths, re)
	}
	for _, u := range opts.URL {
		re, err := regexp.Compile(u)
		if err != nil {
			return nil, fmt.Errorf("invalid url pattern %q: %w", u, err)
		}
		f.urls = append(f.urls, re)
	}

	return f, nil
}

// toSet turns values into a set for O(1) membership checks.
func toSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}
	return set
}

// matches reports whether req satisfies all configured filters. A filter with no values
// always matches.
func (f *permissionFilters) matches(req PermissionRequest) bool {
	if len(f.operations) > 0 {
		op := "write"
		if requestIsReadOnly(req) {
			op = "read"
		}
		if !f.operations[op] {
			return false
		}
	}
	if len(f.kinds) > 0 && !f.kinds[req.Kind] {
		return false
	}
	if len(f.commands) > 0 && !hasAnyCommand(req, f.commands) {
		return false
	}
	if len(f.paths) > 0 && !anyMatches(f.paths, req.Paths) {
		return false
	}
	if len(f.urls) > 0 && !anyMatches(f.urls, req.URLs) {
		return false
	}
	return true
}

// matchesDecisionSource reports whether decisionSource satisfies the configured
// --decision-source filter. A filter with no values always matches.
func (f *permissionFilters) matchesDecisionSource(decisionSource string) bool {
	if len(f.decisionSources) == 0 {
		return true
	}
	return f.decisionSources[decisionSource]
}

// requestIsReadOnly reports whether req involves no mutation: its Kind is "read", or its
// Kind is "shell" and every command is read-only.
func requestIsReadOnly(req PermissionRequest) bool {
	switch req.Kind {
	case "read":
		return true
	case "shell":
		for _, c := range req.Commands {
			if !c.ReadOnly {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// hasAnyCommand reports whether req includes a command whose identifier is in identifiers.
func hasAnyCommand(req PermissionRequest, identifiers map[string]bool) bool {
	for _, c := range req.Commands {
		if identifiers[c.Identifier] {
			return true
		}
	}
	return false
}

// anyMatches reports whether any of values matches any of patterns.
func anyMatches(patterns []*regexp.Regexp, values []string) bool {
	for _, v := range values {
		for _, re := range patterns {
			if re.MatchString(v) {
				return true
			}
		}
	}
	return false
}

// readOnlyKey renders a command's ReadOnly flag as a stable axis key.
func readOnlyKey(readOnly bool) string {
	if readOnly {
		return "read-only"
	}
	return "read-write"
}
