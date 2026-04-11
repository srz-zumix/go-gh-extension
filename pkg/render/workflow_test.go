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

func TestWorkflowDependencyFieldGetters_Job(t *testing.T) {
	getter := NewWorkflowDependencyFieldGetters()

	ref := parser.ActionReference{
		Raw:   "actions/checkout@v4",
		Owner: "actions",
		Repo:  "checkout",
		Ref:   "v4",
		JobID: "build",
	}
	assert.Equal(t, "build", getter.GetField(&ref, "JOB"))
	// Case-insensitive lookup must also work.
	assert.Equal(t, "build", getter.GetField(&ref, "job"))

	// ActionReference without a JobID returns an empty string.
	noJobRef := parser.ActionReference{
		Raw:   "actions/setup-go@v5",
		Owner: "actions",
		Repo:  "setup-go",
		Ref:   "v5",
	}
	assert.Equal(t, "", getter.GetField(&noJobRef, "JOB"))

	// Unknown field name returns empty string (regression guard).
	assert.Equal(t, "", getter.GetField(&ref, "UNKNOWN"))
}

func TestRenderActionReferences_JobColumn(t *testing.T) {
	// Verify that requesting the "JOB" header produces a table that includes
	// each action's JobID in the correct column.
	sr := NewStringRenderer(nil)
	refs := []parser.ActionReference{
		{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", JobID: "build"},
		{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5", JobID: "lint"},
		{Raw: "actions/cache@v3", Owner: "actions", Repo: "cache", Ref: "v3"},
	}
	err := sr.Renderer.RenderActionReferences(refs, []string{"Name", "Version", "Job"})
	assert.NoError(t, err)
	got := sr.Stdout.String()

	// Header row must be present.
	assert.Contains(t, got, "NAME")
	assert.Contains(t, got, "VERSION")
	assert.Contains(t, got, "JOB")

	// JobID values must appear in the output.
	assert.Contains(t, got, "build")
	assert.Contains(t, got, "lint")

	// Action names and versions must still be present.
	assert.Contains(t, got, "actions/checkout")
	assert.Contains(t, got, "actions/setup-go")
	assert.Contains(t, got, "actions/cache")
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

func TestRenderTreeWorkflowDependencies_Empty(t *testing.T) {
	sr := NewStringRenderer(nil)
	err := sr.Renderer.RenderTreeWorkflowDependencies(nil)
	assert.NoError(t, err)
	got := sr.Stdout.String()
	assert.Contains(t, got, "No workflow dependencies.")
}

func TestRenderTreeWorkflowDependencies_Basic(t *testing.T) {
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20"},
				{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5"},
			},
		},
	}
	err := sr.Renderer.RenderTreeWorkflowDependencies(deps)
	assert.NoError(t, err)
	got := sr.Stdout.String()
	// Root node label
	assert.Contains(t, got, ".github/workflows/ci.yml")
	// Children with Using appended in brackets
	assert.Contains(t, got, "actions/checkout@v4 [node20]")
	assert.Contains(t, got, "actions/setup-go@v5")
}

func TestRenderTreeWorkflowDependencies_RootDetection(t *testing.T) {
	// release.yml is referenced by ci.yml, so it should not appear as a root.
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "./.github/workflows/release.yml", IsLocal: true},
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
	err := sr.Renderer.RenderTreeWorkflowDependencies(deps)
	assert.NoError(t, err)
	got := sr.Stdout.String()
	// ci.yml is a root — its Source must appear as the top-level tree node.
	assert.Contains(t, got, ".github/workflows/ci.yml")
	// release.yml appears as a child label (local reusable workflow).
	assert.Contains(t, got, ".github/workflows/release.yml")
	// actions/checkout must appear somewhere (grandchild).
	assert.Contains(t, got, "actions/checkout@v4")
}

func TestRenderTreeWorkflowDependencies_SharedDepAppearsUnderEachParent(t *testing.T) {
	// W → A → C  and  W → B → C.
	// C should appear as a child of both A and B.  However, because visited is
	// shared across the whole root subtree, only the first occurrence (under A)
	// will have C's children expanded; the second (under B) will be a leaf.
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/w.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/action-a@v1", Owner: "org", Repo: "action-a", Ref: "v1", Using: "composite"},
				{Raw: "org/action-b@v1", Owner: "org", Repo: "action-b", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/action-a:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/shared@v1", Owner: "org", Repo: "shared", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/action-b:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/shared@v1", Owner: "org", Repo: "shared", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/shared:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}
	err := sr.Renderer.RenderTreeWorkflowDependencies(deps)
	assert.NoError(t, err)
	got := sr.Stdout.String()
	// Root and direct children must be present.
	assert.Contains(t, got, ".github/workflows/w.yml")
	assert.Contains(t, got, "org/action-a@v1")
	assert.Contains(t, got, "org/action-b@v1")
	// The shared dep appears under action-a (and is labelled under action-b as leaf).
	assert.Contains(t, got, "org/shared@v1")
}

func TestRenderTreeWorkflowDependencies_CyclePrevention(t *testing.T) {
	// W → A → B → A  (cycle back to A)
	// The tree should terminate without infinite recursion; A must appear as a
	// leaf within B's subtree instead of being expanded again.
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/w.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/a@v1", Owner: "org", Repo: "a", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/a:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/b@v1", Owner: "org", Repo: "b", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/b:action.yml",
			Actions: []parser.ActionReference{
				// Back-edge: references A again
				{Raw: "org/a@v1", Owner: "org", Repo: "a", Ref: "v1", Using: "composite"},
			},
		},
	}
	err := sr.Renderer.RenderTreeWorkflowDependencies(deps)
	assert.NoError(t, err)
	got := sr.Stdout.String()
	// All three nodes should appear somewhere in the output.
	assert.Contains(t, got, ".github/workflows/w.yml")
	assert.Contains(t, got, "org/a@v1")
	assert.Contains(t, got, "org/b@v1")
}

func TestRenderTreeWorkflowDependencies_JobIDGrouping(t *testing.T) {
	// Actions are grouped under job sub-nodes when JobID is set.
	// Two jobs: "build" (checkout + setup-go) and "lint" (checkout only).
	// "build" uses a composite action that can be recursed into.
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20", JobID: "build"},
				{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5", Using: "node20", JobID: "build"},
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20", JobID: "lint"},
			},
		},
	}
	err := sr.Renderer.RenderTreeWorkflowDependencies(deps)
	assert.NoError(t, err)
	got := sr.Stdout.String()

	// Root node must be the workflow source.
	assert.Contains(t, got, ".github/workflows/ci.yml")

	// Job IDs must appear as intermediate nodes.
	assert.Contains(t, got, "build")
	assert.Contains(t, got, "lint")

	// Action labels must appear as leaves under their respective jobs.
	assert.Contains(t, got, "actions/checkout@v4 [node20]")
	assert.Contains(t, got, "actions/setup-go@v5 [node20]")
}

func TestRenderTreeWorkflowDependencies_JobIDGrouping_RecursionAndCycle(t *testing.T) {
	// Actions with JobID are grouped under job sub-nodes, and recursion into
	// known dep sources must still work; cycles must still be prevented.
	// ci.yml / build job → composite action → checkout (leaf)
	// ci.yml / deploy job → composite action (already visited: leaf, no re-expansion)
	sr := NewStringRenderer(nil)
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "org/composite@v1", Owner: "org", Repo: "composite", Ref: "v1", Using: "composite", JobID: "build"},
				{Raw: "org/composite@v1", Owner: "org", Repo: "composite", Ref: "v1", Using: "composite", JobID: "deploy"},
			},
		},
		{
			Source: "org/composite:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20"},
			},
		},
	}
	err := sr.Renderer.RenderTreeWorkflowDependencies(deps)
	assert.NoError(t, err)
	got := sr.Stdout.String()

	// Root and job nodes must be present.
	assert.Contains(t, got, ".github/workflows/ci.yml")
	assert.Contains(t, got, "build")
	assert.Contains(t, got, "deploy")

	// The composite action must appear under both jobs.
	assert.Contains(t, got, "org/composite@v1 [composite]")

	// The grandchild (checkout) must appear at least once (expanded under the first job).
	assert.Contains(t, got, "actions/checkout@v4 [node20]")

	// The output must be finite (no panic / stack overflow due to cycle).
}
