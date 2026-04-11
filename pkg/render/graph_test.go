package render

import (
	"strings"
	"testing"

	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
	"github.com/stretchr/testify/assert"
)

func TestDotQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "abc", `"abc"`},
		{"with slash", "owner/repo", `"owner/repo"`},
		{"with quote", `say "hello"`, `"say \"hello\""`},
		{"with backslash", `a\b`, `"a\\b"`},
		{"empty", "", `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, dotQuote(tt.input))
		})
	}
}

func TestRenderDotGraphEdge(t *testing.T) {
	sr := NewStringRenderer(nil)
	edges := []gh.GraphEdge{
		{From: parser.ActionReference{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"}, To: parser.ActionReference{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5"}},
		{From: parser.ActionReference{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"}, To: parser.ActionReference{Raw: "actions/cache@v3", Owner: "actions", Repo: "cache", Ref: "v3"}},
		// duplicate edge should be deduplicated
		{From: parser.ActionReference{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"}, To: parser.ActionReference{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5"}},
	}
	sr.Renderer.RenderDotGraphEdge(edges)
	got := sr.Stdout.String()
	assert.Contains(t, got, "digraph {")
	assert.Contains(t, got, `"actions/checkout" -> "actions/setup-go"`)
	assert.Contains(t, got, `"actions/checkout" -> "actions/cache"`)
	assert.Contains(t, got, "}")
	// verify deduplication: only 2 edges + digraph header + closing brace
	lines := countNonEmptyLines(got)
	assert.Equal(t, 4, lines)
}

func TestRenderDotGraphEdge_Empty(t *testing.T) {
	sr := NewStringRenderer(nil)
	sr.Renderer.RenderDotGraphEdge(nil)
	got := sr.Stdout.String()
	assert.Contains(t, got, "digraph {")
	assert.Contains(t, got, "}")
}

func TestRenderDrawioGraphEdge(t *testing.T) {
	sr := NewStringRenderer(nil)
	edges := []gh.GraphEdge{
		{From: parser.ActionReference{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"}, To: parser.ActionReference{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5"}},
		{From: parser.ActionReference{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"}, To: parser.ActionReference{Raw: "actions/cache@v3", Owner: "actions", Repo: "cache", Ref: "v3"}},
		// duplicate edge should be deduplicated
		{From: parser.ActionReference{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"}, To: parser.ActionReference{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5"}},
	}
	sr.Renderer.RenderDrawioGraphEdge(edges)
	got := sr.Stdout.String()
	assert.Contains(t, got, `<mxfile host="gh-deps-kit">`)
	assert.Contains(t, got, `</mxfile>`)
	// 3 unique nodes: checkout, setup-go, cache
	assert.Contains(t, got, `actions/checkout`)
	assert.Contains(t, got, `actions/setup-go`)
	assert.Contains(t, got, `actions/cache`)
	// 2 edges (deduplicated)
	assert.Contains(t, got, `source="2" target="3"`)
	assert.Contains(t, got, `source="2" target="4"`)
}

func TestRenderDrawioGraphEdge_Empty(t *testing.T) {
	sr := NewStringRenderer(nil)
	sr.Renderer.RenderDrawioGraphEdge(nil)
	got := sr.Stdout.String()
	assert.Contains(t, got, `<mxfile host="gh-deps-kit">`)
	assert.Contains(t, got, `</mxfile>`)
}

func TestWriteDrawioGraph_Tooltip(t *testing.T) {
	// "node-a" has a tooltip with XML-special characters that must be escaped.
	// "node-b" has no tooltip; its mxCell must not carry a tooltip attribute.
	edges := [][2]string{
		{"node-a", "node-b"},
	}
	nodeTooltips := map[string]string{
		"node-a": `runs: composite <action> with "quotes" & ampersand`,
	}

	sr := NewStringRenderer(nil)
	err := sr.Renderer.writeDrawioGraph(edges, nil, nil, nodeTooltips)
	assert.NoError(t, err)
	got := sr.Stdout.String()

	// Tooltip value must have XML special characters HTML-escaped.
	assert.Contains(t, got, `tooltip="runs: composite &lt;action&gt; with &#34;quotes&#34; &amp; ampersand"`,
		"tooltip attribute must appear with HTML-escaped value for node-a")

	// The tooltip attribute must only be present for the node that has one.
	// Split output into per-node cell lines and verify node-b has no tooltip attr.
	for _, line := range splitLines(got) {
		if strings.Contains(line, "node-b") && strings.Contains(line, `vertex="1"`) {
			assert.NotContains(t, line, "tooltip=",
				"node-b has no tooltip so tooltip attribute must be absent")
		}
	}
}

func TestWorkflowDepSourceURL(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		host     string
		ref      string
		owner    string
		repo     string
		expected string
	}{
		{"remote action", "actions/checkout:action.yml", "github.com", "", "", "",
			"https://github.com/actions/checkout/blob/HEAD/action.yml"},
		{"remote action with ref", "actions/checkout:action.yml", "github.com", "v4", "", "",
			"https://github.com/actions/checkout/blob/v4/action.yml"},
		{"remote workflow", "owner/repo:.github/workflows/ci.yml", "github.com", "", "", "",
			"https://github.com/owner/repo/blob/HEAD/.github/workflows/ci.yml"},
		{"remote workflow with ref", "owner/repo:.github/workflows/ci.yml", "github.com", "main", "", "",
			"https://github.com/owner/repo/blob/main/.github/workflows/ci.yml"},
		{"GHES remote", "actions/checkout:action.yml", "ghes.example.com", "", "", "",
			"https://ghes.example.com/actions/checkout/blob/HEAD/action.yml"},
		{"local workflow no host", ".github/workflows/ci.yml", "", "", "", "",
			""},
		{"local workflow with host and owner/repo", ".github/workflows/ci.yml", "github.com", "", "myorg", "myrepo",
			"https://github.com/myorg/myrepo/blob/HEAD/.github/workflows/ci.yml"},
		{"local workflow with host and ref", ".github/workflows/ci.yml", "github.com", "main", "myorg", "myrepo",
			"https://github.com/myorg/myrepo/blob/main/.github/workflows/ci.yml"},
		{"local action with host and owner/repo", "my-action/action.yml", "github.com", "", "myorg", "myrepo",
			"https://github.com/myorg/myrepo/blob/HEAD/my-action/action.yml"},
		{"local workflow with host but no owner/repo", ".github/workflows/ci.yml", "github.com", "", "", "",
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, workflowDepSourceURL(tt.source, tt.host, tt.ref, tt.owner, tt.repo))
		})
	}
}

func TestActionReferenceURL(t *testing.T) {
	tests := []struct {
		name     string
		action   parser.ActionReference
		expected string
	}{
		{"remote action", parser.ActionReference{Owner: "actions", Repo: "checkout", Host: "github.com"},
			"https://github.com/actions/checkout"},
		{"remote action with path", parser.ActionReference{Owner: "owner", Repo: "repo", Path: "subdir", Host: "github.com"},
			"https://github.com/owner/repo/blob/HEAD/subdir"},
		{"remote action with path and ref", parser.ActionReference{Owner: "owner", Repo: "repo", Path: "subdir", Ref: "v1", Host: "github.com"},
			"https://github.com/owner/repo/blob/v1/subdir"},
		{"GHES remote", parser.ActionReference{Owner: "owner", Repo: "repo", Path: ".github/workflows/ci.yml", Host: "ghes.example.com"},
			"https://ghes.example.com/owner/repo/blob/HEAD/.github/workflows/ci.yml"},
		{"GHES remote with ref", parser.ActionReference{Owner: "owner", Repo: "repo", Path: ".github/workflows/ci.yml", Ref: "abc123", Host: "ghes.example.com"},
			"https://ghes.example.com/owner/repo/blob/abc123/.github/workflows/ci.yml"},
		{"local action", parser.ActionReference{IsLocal: true, Raw: "./my-action", Host: "github.com"},
			""},
		{"no host", parser.ActionReference{Owner: "actions", Repo: "checkout"},
			""},
		{"no owner/repo", parser.ActionReference{Raw: "something", Host: "github.com"},
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, actionReferenceURL(tt.action))
		})
	}
}

func countNonEmptyLines(s string) int {
	count := 0
	for _, line := range splitLines(s) {
		if len(line) > 0 {
			count++
		}
	}
	return count
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func TestMermaidNodeID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "alphanumeric only",
			input:    "abc123",
			expected: "abc123",
		},
		{
			name:     "hyphen encoded",
			input:    "ci-test",
			expected: "ci_2dtest",
		},
		{
			name:     "underscore encoded",
			input:    "ci_test",
			expected: "ci_5ftest",
		},
		{
			name:     "dot encoded",
			input:    "ci.yml",
			expected: "ci_2eyml",
		},
		{
			name:     "slash encoded",
			input:    ".github/workflows/ci.yml",
			expected: "_2egithub_2fworkflows_2fci_2eyml",
		},
		{
			name:     "colon encoded",
			input:    "owner/repo:action.yml",
			expected: "owner_2frepo_3aaction_2eyml",
		},
		{
			name:     "at sign encoded",
			input:    "actions/checkout@v4",
			expected: "actions_2fcheckout_40v4",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mermaidNodeID(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMermaidNodeID_NoCollision(t *testing.T) {
	// These pairs must produce different IDs
	pairs := [][2]string{
		{"ci-test", "ci_test"},
		{"ci-test.yml", "ci_test.yml"},
		{".github/workflows/ci-test.yml", ".github/workflows/ci_test.yml"},
		{"a-b", "a_b"},
		{"a.b", "a_b"},
	}

	for _, pair := range pairs {
		id1 := mermaidNodeID(pair[0])
		id2 := mermaidNodeID(pair[1])
		assert.NotEqual(t, id1, id2,
			"mermaidNodeID(%q) == mermaidNodeID(%q) == %q", pair[0], pair[1], id1)
	}
}
