package gh

import (
	"testing"

	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

func TestExpandFilteredDependencies_LocalReusableWorkflow(t *testing.T) {
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/release-drafter.yml",
			Name:   "Release Drafter",
			Actions: []parser.ActionReference{
				{Raw: "release-drafter/release-drafter@sha", Owner: "release-drafter", Repo: "release-drafter", Ref: "sha"},
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
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}

	// Filter only release-drafter.yml
	filtered := []parser.WorkflowDependency{allDeps[0]}

	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 2 {
		t.Fatalf("expected 2 deps after expansion, got %d", len(expanded))
	}

	// Verify release.yml is included
	found := false
	for _, dep := range expanded {
		if dep.Source == ".github/workflows/release.yml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected .github/workflows/release.yml to be included after expansion")
	}
}

func TestExpandFilteredDependencies_RemoteReusableWorkflow(t *testing.T) {
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "org/repo/.github/workflows/shared.yml@main", Owner: "org", Repo: "repo", Path: ".github/workflows/shared.yml", Ref: "main"},
			},
		},
		{
			Source: "org/repo:.github/workflows/shared.yml",
			Name:   "Shared",
			Actions: []parser.ActionReference{
				{Raw: "actions/setup-node@v4", Owner: "actions", Repo: "setup-node", Ref: "v4"},
			},
		},
	}

	filtered := []parser.WorkflowDependency{allDeps[0]}
	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 2 {
		t.Fatalf("expected 2 deps after expansion, got %d", len(expanded))
	}

	found := false
	for _, dep := range expanded {
		if dep.Source == "org/repo:.github/workflows/shared.yml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected remote reusable workflow to be included after expansion")
	}
}

func TestExpandFilteredDependencies_ActionRepository(t *testing.T) {
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "custom-org/custom-action@v1", Owner: "custom-org", Repo: "custom-action", Ref: "v1"},
			},
		},
		{
			Source: "custom-org/custom-action:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}

	filtered := []parser.WorkflowDependency{allDeps[0]}
	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 2 {
		t.Fatalf("expected 2 deps after expansion, got %d", len(expanded))
	}

	found := false
	for _, dep := range expanded {
		if dep.Source == "custom-org/custom-action:action.yml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected child action repo to be included after expansion")
	}
}

func TestExpandFilteredDependencies_ActionRepositorySubdirectory(t *testing.T) {
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "custom-org/custom-action/path/to/action@v1", Owner: "custom-org", Repo: "custom-action", Path: "path/to/action", Ref: "v1"},
			},
		},
		{
			Source: "custom-org/custom-action:path/to/action/action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}

	filtered := []parser.WorkflowDependency{allDeps[0]}
	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 2 {
		t.Fatalf("expected 2 deps after expansion, got %d", len(expanded))
	}

	found := false
	for _, dep := range expanded {
		if dep.Source == "custom-org/custom-action:path/to/action/action.yml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected child action repo subdirectory action to be included after expansion")
	}
}

func TestExpandFilteredDependencies_TransitiveChain(t *testing.T) {
	// A -> B (local reusable) -> C (action repo)
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/a.yml",
			Name:   "A",
			Actions: []parser.ActionReference{
				{Raw: "./.github/workflows/b.yml", IsLocal: true},
			},
		},
		{
			Source: ".github/workflows/b.yml",
			Name:   "B",
			Actions: []parser.ActionReference{
				{Raw: "org/action-c@v1", Owner: "org", Repo: "action-c", Ref: "v1"},
			},
		},
		{
			Source: "org/action-c:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}

	filtered := []parser.WorkflowDependency{allDeps[0]}
	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 3 {
		t.Fatalf("expected 3 deps after expansion (transitive chain), got %d", len(expanded))
	}
}

func TestExpandFilteredDependencies_NoCycle(t *testing.T) {
	// A -> B -> A (cycle should not cause infinite loop)
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/a.yml",
			Name:   "A",
			Actions: []parser.ActionReference{
				{Raw: "./.github/workflows/b.yml", IsLocal: true},
			},
		},
		{
			Source: ".github/workflows/b.yml",
			Name:   "B",
			Actions: []parser.ActionReference{
				{Raw: "./.github/workflows/a.yml", IsLocal: true},
			},
		},
	}

	filtered := []parser.WorkflowDependency{allDeps[0]}
	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 2 {
		t.Fatalf("expected 2 deps after expansion (with cycle), got %d", len(expanded))
	}
}

func TestExpandFilteredDependencies_LocalAction(t *testing.T) {
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "./my-action", IsLocal: true},
			},
		},
		{
			Source: "my-action/action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}

	filtered := []parser.WorkflowDependency{allDeps[0]}
	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 2 {
		t.Fatalf("expected 2 deps after expansion, got %d", len(expanded))
	}

	found := false
	for _, dep := range expanded {
		if dep.Source == "my-action/action.yml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected local action dep to be included after expansion")
	}
}

func TestExpandFilteredDependencies_LocalActionTransitiveChain(t *testing.T) {
	// Workflow -> local action -> external action repo
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "./my-action", IsLocal: true},
			},
		},
		{
			Source: "my-action/action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/helper@v1", Owner: "org", Repo: "helper", Ref: "v1"},
			},
		},
		{
			Source: "org/helper:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}

	filtered := []parser.WorkflowDependency{allDeps[0]}
	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 3 {
		t.Fatalf("expected 3 deps after expansion (transitive chain via local action), got %d", len(expanded))
	}
}

func TestParseNodeVersion(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"node16", 16},
		{"node20", 20},
		{"node12", 12},
		{"composite", 0},
		{"", 0},
		{"node", 0},
		{"nodeabc", 0},
		{"node0", 0},
		{"docker", 0},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := parseNodeVersion(tc.input)
			if got != tc.want {
				t.Errorf("parseNodeVersion(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestFilterWorkflowDependenciesByNodeVersion_MinVersionZero(t *testing.T) {
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 0)
	if len(got) != len(deps) {
		t.Fatalf("expected all deps returned when minNodeVersion=0, got %d", len(got))
	}
}

func TestFilterWorkflowDependenciesByNodeVersion_NoOldNodeActions(t *testing.T) {
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20"},
			},
		},
		{
			Source:  "actions/checkout:action.yml",
			Actions: []parser.ActionReference{},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 0 {
		t.Fatalf("expected no deps when all actions meet minNodeVersion, got %d", len(got))
	}
}

func TestFilterWorkflowDependenciesByNodeVersion_OldNodeDirectReference(t *testing.T) {
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source:  "org/old-action:action.yml",
			Actions: []parser.ActionReference{},
		},
		{
			Source: ".github/workflows/other.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20"},
			},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 2 {
		t.Fatalf("expected 2 deps (workflow + old action), got %d", len(got))
	}
	sources := make(map[string]bool)
	for _, d := range got {
		sources[d.Source] = true
	}
	if !sources[".github/workflows/ci.yml"] {
		t.Error("expected ci.yml to be included")
	}
	if !sources["org/old-action:action.yml"] {
		t.Error("expected org/old-action:action.yml to be included")
	}
}

func TestFilterWorkflowDependenciesByNodeVersion_CompositeActionWithOldNodeDep(t *testing.T) {
	// ci.yml -> composite:action.yml (composite) -> old-action:action.yml (node16)
	// ci.yml references composite action which itself uses an old node action.
	// Although ci.yml uses composite (not old node directly), fixpoint expansion
	// should walk reverse edges and include ci.yml as an upstream dependent.
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/composite@v1", Owner: "org", Repo: "composite", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/composite:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source:  "org/old-action:action.yml",
			Actions: []parser.ActionReference{},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 3 {
		t.Fatalf("expected 3 deps (ci.yml + composite + old action), got %d", len(got))
	}
	sources := make(map[string]bool)
	for _, d := range got {
		sources[d.Source] = true
	}
	if !sources[".github/workflows/ci.yml"] {
		t.Error("expected ci.yml to be included (transitively references old node action)")
	}
	if !sources["org/composite:action.yml"] {
		t.Error("expected org/composite:action.yml to be included")
	}
	if !sources["org/old-action:action.yml"] {
		t.Error("expected org/old-action:action.yml to be included")
	}
}

func TestFilterWorkflowDependenciesByNodeVersion_TransitiveChain(t *testing.T) {
	// ci.yml -> old-action:action.yml (node16) -> checkout:action.yml (node20)
	// Only ci.yml and old-action:action.yml should be included.
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source: "org/old-action:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20"},
			},
		},
		{
			Source:  "actions/checkout:action.yml",
			Actions: []parser.ActionReference{},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 2 {
		t.Fatalf("expected 2 deps (ci.yml + old-action), got %d", len(got))
	}
	sources := make(map[string]bool)
	for _, d := range got {
		sources[d.Source] = true
	}
	if !sources[".github/workflows/ci.yml"] {
		t.Error("expected ci.yml to be included")
	}
	if !sources["org/old-action:action.yml"] {
		t.Error("expected org/old-action:action.yml to be included")
	}
	if sources["actions/checkout:action.yml"] {
		t.Error("checkout:action.yml should not be included (uses node20)")
	}
}

func TestFilterWorkflowDependenciesByNodeVersion_OldActionNotInDeps(t *testing.T) {
	// ci.yml uses an old node action that has no entry in deps (no source resolved)
	// Only ci.yml should be included; no old source to add.
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 1 {
		t.Fatalf("expected 1 dep (ci.yml), got %d", len(got))
	}
	if got[0].Source != ".github/workflows/ci.yml" {
		t.Errorf("expected ci.yml, got %s", got[0].Source)
	}
}

// TestExpandFilteredDependencies_DiamondDependency verifies that a shared dep reachable
// via two distinct paths (diamond shape: W→A→C and W→B→C) is included exactly once
// and its own children are still expanded.
func TestExpandFilteredDependencies_DiamondDependency(t *testing.T) {
	// W -> A -> C and W -> B -> C  (C reachable via two paths)
	// C -> D (D must also be included despite C being "already visited" when reached from B)
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/w.yml",
			Name:   "W",
			Actions: []parser.ActionReference{
				{Raw: "org/action-a@v1", Owner: "org", Repo: "action-a", Ref: "v1"},
				{Raw: "org/action-b@v1", Owner: "org", Repo: "action-b", Ref: "v1"},
			},
		},
		{
			Source: "org/action-a:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/shared@v1", Owner: "org", Repo: "shared", Ref: "v1"},
			},
		},
		{
			Source: "org/action-b:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/shared@v1", Owner: "org", Repo: "shared", Ref: "v1"},
			},
		},
		{
			Source: "org/shared:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/leaf@v1", Owner: "org", Repo: "leaf", Ref: "v1"},
			},
		},
		{
			Source:  "org/leaf:action.yml",
			Actions: []parser.ActionReference{},
		},
	}

	filtered := []parser.WorkflowDependency{allDeps[0]}
	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 5 {
		t.Fatalf("expected 5 deps (diamond + leaf), got %d", len(expanded))
	}
	sources := make(map[string]bool)
	for _, d := range expanded {
		sources[d.Source] = true
	}
	for _, want := range []string{
		".github/workflows/w.yml",
		"org/action-a:action.yml",
		"org/action-b:action.yml",
		"org/shared:action.yml",
		"org/leaf:action.yml",
	} {
		if !sources[want] {
			t.Errorf("expected %s to be included", want)
		}
	}
}

// TestFilterWorkflowDependenciesByNodeVersion_MultipleUpstreamCallers verifies that when
// multiple workflows independently reference the same composite action (which itself uses
// an old Node action), all of them are included via fixpoint expansion.
func TestFilterWorkflowDependenciesByNodeVersion_MultipleUpstreamCallers(t *testing.T) {
	// W1 → composite:action.yml → old:action.yml (node16)
	// W2 → composite:action.yml → old:action.yml (node16)
	// W3 → unrelated:action.yml (node20) — should NOT be included
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/w1.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/composite@v1", Owner: "org", Repo: "composite", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: ".github/workflows/w2.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/composite@v1", Owner: "org", Repo: "composite", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: ".github/workflows/w3.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/unrelated@v1", Owner: "org", Repo: "unrelated", Ref: "v1", Using: "node20"},
			},
		},
		{
			Source: "org/composite:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source:  "org/old-action:action.yml",
			Actions: []parser.ActionReference{},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 4 {
		t.Fatalf("expected 4 deps (w1, w2, composite, old-action), got %d", len(got))
	}
	sources := make(map[string]bool)
	for _, d := range got {
		sources[d.Source] = true
	}
	for _, want := range []string{
		".github/workflows/w1.yml",
		".github/workflows/w2.yml",
		"org/composite:action.yml",
		"org/old-action:action.yml",
	} {
		if !sources[want] {
			t.Errorf("expected %s to be included", want)
		}
	}
	if sources[".github/workflows/w3.yml"] {
		t.Error("w3.yml should not be included (uses node20 only)")
	}
}

// TestFilterWorkflowDependenciesByNodeVersion_DeepCompositeChain verifies that fixpoint
// expansion propagates upstream through an arbitrarily deep chain of composite actions:
//
//	W → C1 → C2 → C3 → old (node16)
//
// All nodes in the chain should be in the result.
func TestFilterWorkflowDependenciesByNodeVersion_DeepCompositeChain(t *testing.T) {
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/w.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/c1@v1", Owner: "org", Repo: "c1", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/c1:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/c2@v1", Owner: "org", Repo: "c2", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/c2:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/c3@v1", Owner: "org", Repo: "c3", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/c3:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source:  "org/old-action:action.yml",
			Actions: []parser.ActionReference{},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 5 {
		t.Fatalf("expected 5 deps (full chain), got %d", len(got))
	}
	sources := make(map[string]bool)
	for _, d := range got {
		sources[d.Source] = true
	}
	for _, want := range []string{
		".github/workflows/w.yml",
		"org/c1:action.yml",
		"org/c2:action.yml",
		"org/c3:action.yml",
		"org/old-action:action.yml",
	} {
		if !sources[want] {
			t.Errorf("expected %s to be included", want)
		}
	}
}

// TestFilterWorkflowDependenciesByNodeVersion_DiamondDependency verifies that when two
// composites both reference the same old-node action (diamond shape), and a single workflow
// references both composites, all nodes are included exactly once.
func TestFilterWorkflowDependenciesByNodeVersion_DiamondDependency(t *testing.T) {
	// W → C1, W → C2; C1 → old (node16); C2 → old (node16)
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/w.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/c1@v1", Owner: "org", Repo: "c1", Ref: "v1", Using: "composite"},
				{Raw: "org/c2@v1", Owner: "org", Repo: "c2", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/c1:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source: "org/c2:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source:  "org/old-action:action.yml",
			Actions: []parser.ActionReference{},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 4 {
		t.Fatalf("expected 4 deps (w, c1, c2, old-action), got %d", len(got))
	}
	sources := make(map[string]bool)
	for _, d := range got {
		sources[d.Source] = true
	}
	for _, want := range []string{
		".github/workflows/w.yml",
		"org/c1:action.yml",
		"org/c2:action.yml",
		"org/old-action:action.yml",
	} {
		if !sources[want] {
			t.Errorf("expected %s to be included", want)
		}
	}
}

func TestExpandFilteredDependencies_NoExpansionNeeded(t *testing.T) {
	allDeps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Name:   "CI",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			},
		},
	}

	filtered := []parser.WorkflowDependency{allDeps[0]}
	expanded := ExpandFilteredDependencies(filtered, allDeps)

	if len(expanded) != 1 {
		t.Fatalf("expected 1 dep (no expansion needed), got %d", len(expanded))
	}
}

// TestFilterWorkflowDependenciesByNodeVersion_MixedActionsInWorkflow verifies that
// when a workflow uses both an old Node action and a new Node action, only the old
// Node action is kept in its Actions list.
func TestFilterWorkflowDependenciesByNodeVersion_MixedActionsInWorkflow(t *testing.T) {
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20"},
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
				{Raw: "actions/setup-go@v5", Owner: "actions", Repo: "setup-go", Ref: "v5", Using: "node20"},
			},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 1 {
		t.Fatalf("expected 1 dep, got %d", len(got))
	}
	if len(got[0].Actions) != 1 {
		t.Fatalf("expected 1 action in ci.yml (old-action only), got %d", len(got[0].Actions))
	}
	if got[0].Actions[0].Repo != "old-action" {
		t.Errorf("expected old-action to remain, got %s", got[0].Actions[0].Repo)
	}
}

// TestFilterWorkflowDependenciesByNodeVersion_TransitivelyIncludedActionsFiltered verifies
// that a workflow included only because it references a composite action (which uses an old
// Node action) has its unrelated actions pruned from its Actions list.
func TestFilterWorkflowDependenciesByNodeVersion_TransitivelyIncludedActionsFiltered(t *testing.T) {
	// w.yml uses checkout (node20) + composite (which uses node16). Only composite should remain.
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/w.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20"},
				{Raw: "org/composite@v1", Owner: "org", Repo: "composite", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/composite:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source:  "org/old-action:action.yml",
			Actions: []parser.ActionReference{},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 3 {
		t.Fatalf("expected 3 deps, got %d", len(got))
	}
	for _, dep := range got {
		switch dep.Source {
		case ".github/workflows/w.yml":
			if len(dep.Actions) != 1 {
				t.Errorf("w.yml: expected 1 action (composite only), got %d: %v", len(dep.Actions), dep.Actions)
			} else if dep.Actions[0].Repo != "composite" {
				t.Errorf("w.yml: expected composite action to remain, got %s", dep.Actions[0].Repo)
			}
		case "org/composite:action.yml":
			if len(dep.Actions) != 1 {
				t.Errorf("composite:action.yml: expected 1 action (old-action), got %d", len(dep.Actions))
			}
		}
	}
}

// TestFilterWorkflowDependenciesByNodeVersion_CompositeWithMixedDepsFiltered verifies
// that a composite action that references both an old Node action and a new Node action
// retains only the old Node action in its Actions list.
func TestFilterWorkflowDependenciesByNodeVersion_CompositeWithMixedDepsFiltered(t *testing.T) {
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/w.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/composite@v1", Owner: "org", Repo: "composite", Ref: "v1", Using: "composite"},
			},
		},
		{
			Source: "org/composite:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20"},
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source:  "org/old-action:action.yml",
			Actions: []parser.ActionReference{},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 3 {
		t.Fatalf("expected 3 deps, got %d", len(got))
	}
	for _, dep := range got {
		if dep.Source == "org/composite:action.yml" {
			if len(dep.Actions) != 1 {
				t.Errorf("composite:action.yml: expected 1 action (old-action only), got %d: %v", len(dep.Actions), dep.Actions)
			} else if dep.Actions[0].Repo != "old-action" {
				t.Errorf("composite:action.yml: expected old-action to remain, got %s", dep.Actions[0].Repo)
			}
		}
	}
}

// TestFilterWorkflowDependenciesByNodeVersion_OldActionActionsFiltered verifies that an
// old Node action dep's own Actions (e.g. actions it calls internally) are filtered to
// exclude refs that are not themselves old or transitively marked.
func TestFilterWorkflowDependenciesByNodeVersion_OldActionActionsFiltered(t *testing.T) {
	// ci.yml → old-action (node16) → checkout (node20)
	// old-action's checkout reference should be pruned from its Actions.
	deps := []parser.WorkflowDependency{
		{
			Source: ".github/workflows/ci.yml",
			Actions: []parser.ActionReference{
				{Raw: "org/old-action@v1", Owner: "org", Repo: "old-action", Ref: "v1", Using: "node16"},
			},
		},
		{
			Source: "org/old-action:action.yml",
			Actions: []parser.ActionReference{
				{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4", Using: "node20"},
			},
		},
	}
	got := FilterWorkflowDependenciesByNodeVersion(deps, 20)
	if len(got) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(got))
	}
	for _, dep := range got {
		if dep.Source == "org/old-action:action.yml" {
			if len(dep.Actions) != 0 {
				t.Errorf("old-action:action.yml: expected 0 actions after filtering checkout (node20), got %d: %v", len(dep.Actions), dep.Actions)
			}
		}
	}
}
