package render

import (
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
