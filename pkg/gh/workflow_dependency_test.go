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
