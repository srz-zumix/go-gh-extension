package render

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
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
		{"node", ""},    // no version digits
		{"nodeXX", ""},  // non-numeric suffix
		{"node20a", ""}, // non-numeric suffix
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractNodeVersion(tt.input))
		})
	}
}

func TestWorkflowDependencyFieldGetters_Using(t *testing.T) {
	getter := NewWorkflowDependencyFieldGetters()

	// ActionReference with Using set (resolved during traversal)
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

func TestRenderDrawioWorkflowDependencies(t *testing.T) {
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source:     ".github/workflows/ci.yml",
			Name:       "CI",
			Repository: repository.Repository{Host: "github.com", Owner: "myorg", Name: "myrepo"},
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Host: "github.com"},
				{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5", Host: "github.com"},
			},
		},
	}
	sr.Renderer.RenderDrawioWorkflowDependencies(deps)
	got := sr.Stdout.String()
	assert.Contains(t, got, `<mxfile host="gh-deps-kit">`)
	assert.Contains(t, got, `</mxfile>`)
	// Source node should have a URL for local files using owner/repo
	assert.Contains(t, got, `https://github.com/myorg/myrepo/blob/HEAD/.github/workflows/ci.yml`)
	assert.Contains(t, got, `.github/workflows/ci.yml`)
	// Remote actions have URLs
	assert.Contains(t, got, `https://github.com/actions/checkout`)
	assert.Contains(t, got, `actions/checkout`)
	assert.Contains(t, got, `https://github.com/actions/setup-go`)
	assert.Contains(t, got, `actions/setup-go`)
}

func TestRenderDrawioWorkflowDependencies_GHES(t *testing.T) {
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source:     ".github/workflows/ci.yml",
			Name:       "CI",
			Repository: repository.Repository{Host: "ghes.example.com", Owner: "myorg", Name: "myrepo"},
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Host: "ghes.example.com"},
			},
		},
	}
	sr.Renderer.RenderDrawioWorkflowDependencies(deps)
	got := sr.Stdout.String()
	// GHES host should be used for remote actions
	assert.Contains(t, got, `https://ghes.example.com/actions/checkout`)
	// Local source should use GHES host
	assert.Contains(t, got, `https://ghes.example.com/myorg/myrepo/blob/HEAD/.github/workflows/ci.yml`)
	assert.NotContains(t, got, `github.com`)
}

func TestRenderDrawioWorkflowDependencies_FallbackMixedHosts(t *testing.T) {
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source:     ".github/workflows/ci.yml",
			Name:       "CI",
			Repository: repository.Repository{Host: "ghes.example.com", Owner: "myorg", Name: "myrepo"},
			Actions: []parser.ActionReference{
				// This action was resolved on GHES
				{Raw: "internal/action@v1", Owner: "internal", Repo: "action", Ref: "v1", Host: "ghes.example.com"},
				// This action was resolved via fallback to github.com
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Host: "github.com"},
			},
		},
	}
	sr.Renderer.RenderDrawioWorkflowDependencies(deps)
	got := sr.Stdout.String()
	// Each action should use its own host for URL generation
	assert.Contains(t, got, `https://ghes.example.com/internal/action`)
	assert.Contains(t, got, `https://github.com/actions/checkout`)
	// Local source should use GHES host
	assert.Contains(t, got, `https://ghes.example.com/myorg/myrepo/blob/HEAD/.github/workflows/ci.yml`)
}

func TestRenderDrawioWorkflowDependencies_Empty(t *testing.T) {
	sr := NewStringRenderer(nil)
	sr.Renderer.RenderDrawioWorkflowDependencies(nil)
	got := sr.Stdout.String()
	assert.Contains(t, got, `<mxfile host="gh-deps-kit">`)
	assert.Contains(t, got, `</mxfile>`)
}

func TestRenderDrawioWorkflowDependencies_NodeColors(t *testing.T) {
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source:     ".github/workflows/ci.yml",
			Name:       "CI",
			Repository: repository.Repository{Host: "github.com", Owner: "myorg", Name: "myrepo"},
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Host: "github.com", Using: "node20"},
				{Raw: "actions/composite@v1", Owner: "actions", Repo: "composite", Ref: "v1", Host: "github.com", Using: "composite"},
				{Raw: "actions/docker@v1", Owner: "actions", Repo: "docker", Ref: "v1", Host: "github.com", Using: "docker"},
				{Raw: "org/repo/.github/workflows/reuse.yml@main", Owner: "org", Repo: "repo", Path: ".github/workflows/reuse.yml", Ref: "main", Host: "github.com"},
			},
		},
	}
	sr.Renderer.RenderDrawioWorkflowDependencies(deps)
	got := sr.Stdout.String()
	// node action: orange border
	assert.Contains(t, got, `strokeColor=#FF9800;strokeWidth=2;`)
	// composite action: green border
	assert.Contains(t, got, `strokeColor=#4CAF50;strokeWidth=2;`)
	// docker action: purple border
	assert.Contains(t, got, `strokeColor=#9C27B0;strokeWidth=2;`)
	// reusable workflow: blue border
	assert.Contains(t, got, `strokeColor=#2196F3;strokeWidth=2;`)
	// source workflow: default style (no strokeColor)
	// The source node should exist with the default rounded style
	assert.Contains(t, got, `.github/workflows/ci.yml`)
}

func TestActionNodeColor(t *testing.T) {
	tests := []struct {
		name   string
		action parser.ActionReference
		want   string
	}{
		{
			name:   "reusable workflow",
			action: parser.ActionReference{Raw: "org/repo/.github/workflows/ci.yml@main", Owner: "org", Repo: "repo", Path: ".github/workflows/ci.yml", Ref: "main"},
			want:   "#2196F3",
		},
		{
			name:   "local reusable workflow",
			action: parser.ActionReference{Raw: "./.github/workflows/release.yml", IsLocal: true},
			want:   "#2196F3",
		},
		{
			name:   "composite",
			action: parser.ActionReference{Using: "composite"},
			want:   "#4CAF50",
		},
		{
			name:   "node20",
			action: parser.ActionReference{Using: "node20"},
			want:   "#FF9800",
		},
		{
			name:   "node16",
			action: parser.ActionReference{Using: "node16"},
			want:   "#FF9800",
		},
		{
			name:   "docker",
			action: parser.ActionReference{Using: "docker"},
			want:   "#9C27B0",
		},
		{
			name:   "unknown using",
			action: parser.ActionReference{Using: ""},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := actionNodeColor(tt.action)
			assert.Equal(t, tt.want, got)
		})
	}
}
