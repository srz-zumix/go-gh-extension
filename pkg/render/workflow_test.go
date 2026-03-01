package render

import (
	"testing"

	"github.com/srz-zumix/go-gh-extension/pkg/parser"
	"github.com/stretchr/testify/assert"
)

func TestExtractNodeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"node20", "20"},
		{"node16", "16"},
		{"node12", "12"},
		{"composite", ""},
		{"docker", ""},
		{"", ""},
		{"node", ""},     // no version digits
		{"nodeXX", ""},   // non-numeric suffix
		{"node20a", ""},  // non-numeric suffix
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractNodeVersion(tt.input))
		})
	}
}

func TestWorkflowDependencyFieldGetters_Using(t *testing.T) {
	getter := NewWorkflowDependencyFieldGetters()

	// ActionReference with Using set (populated by PopulateActionUsing)
	checkoutRef := parser.ActionReference{
		Raw:   "actions/checkout@v4",
		Owner: "actions",
		Repo:  "checkout",
		Ref:   "v4",
		Using: "node20",
	}
	assert.Equal(t, "node20", getter.GetField(&checkoutRef, "USING"))
	assert.Equal(t, "20", getter.GetField(&checkoutRef, "NODE_VERSION"))

	// ActionReference pointing to composite action
	compositeRef := parser.ActionReference{
		Raw:   "my-org/composite-action@v1",
		Owner: "my-org",
		Repo:  "composite-action",
		Ref:   "v1",
		Using: "composite",
	}
	assert.Equal(t, "composite", getter.GetField(&compositeRef, "USING"))
	assert.Equal(t, "", getter.GetField(&compositeRef, "NODE_VERSION"))

	// Unknown action (Using not populated)
	unknownRef := parser.ActionReference{
		Raw:   "unknown/action@v1",
		Owner: "unknown",
		Repo:  "action",
		Ref:   "v1",
	}
	assert.Equal(t, "", getter.GetField(&unknownRef, "USING"))
	assert.Equal(t, "", getter.GetField(&unknownRef, "NODE_VERSION"))
}

func TestRenderDotWorkflowDependencies(t *testing.T) {
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
				{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5"},
			},
		},
		{
			Source: ".github/workflows/release.yml",
			Name:   "Release",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}
	sr.Renderer.RenderDotWorkflowDependencies(deps)
	got := sr.Stdout.String()
	assert.Contains(t, got, "digraph {")
	assert.Contains(t, got, `".github/workflows/ci.yml" -> "actions/checkout"`)
	assert.Contains(t, got, `".github/workflows/ci.yml" -> "actions/setup-go"`)
	assert.Contains(t, got, `".github/workflows/release.yml" -> "actions/checkout"`)
	assert.Contains(t, got, "}")
}

func TestRenderDotWorkflowDependencies_Empty(t *testing.T) {
	sr := NewStringRenderer(nil)
	sr.Renderer.RenderDotWorkflowDependencies(nil)
	got := sr.Stdout.String()
	assert.Contains(t, got, "digraph {")
	assert.Contains(t, got, "}")
}

func TestRenderDotWorkflowDependencies_ResolvedSource(t *testing.T) {
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				// Local action that resolves to another dep source
				{Raw: "./my-action", IsLocal: true, Path: "my-action"},
			},
		},
		{
			Source: "my-action/action.yml",
			Name:   "",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}
	sr.Renderer.RenderDotWorkflowDependencies(deps)
	got := sr.Stdout.String()
	// The local action should resolve to the dep source path
	assert.Contains(t, got, `".github/workflows/ci.yml" -> "my-action/action.yml"`)
	assert.Contains(t, got, `"my-action/action.yml" -> "actions/checkout"`)
}
