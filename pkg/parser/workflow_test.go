package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseActionReference(t *testing.T) {
	tests := []struct {
		name     string
		uses     string
		expected ActionReference
	}{
		{
			name: "standard action with tag",
			uses: "actions/checkout@v4",
			expected: ActionReference{
				Raw:   "actions/checkout@v4",
				Owner: "actions",
				Repo:  "checkout",
				Ref:   "v4",
			},
		},
		{
			name: "action with SHA ref",
			uses: "actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f953675",
			expected: ActionReference{
				Raw:   "actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f953675",
				Owner: "actions",
				Repo:  "checkout",
				Ref:   "a81bbbf8298c0fa03ea29cdc473d45769f953675",
			},
		},
		{
			name: "action with subdirectory",
			uses: "github/codeql-action/init@v3",
			expected: ActionReference{
				Raw:   "github/codeql-action/init@v3",
				Owner: "github",
				Repo:  "codeql-action",
				Path:  "init",
				Ref:   "v3",
			},
		},
		{
			name: "local action",
			uses: "./path/to/action",
			expected: ActionReference{
				Raw:     "./path/to/action",
				IsLocal: true,
			},
		},
		{
			name: "docker action",
			uses: "docker://alpine:3.8",
			expected: ActionReference{
				Raw: "docker://alpine:3.8",
			},
		},
		{
			name: "reusable workflow",
			uses: "octo-org/this-repo/.github/workflows/workflow-1.yml@172239021f7ba04fe7327647b213799853a9eb89",
			expected: ActionReference{
				Raw:   "octo-org/this-repo/.github/workflows/workflow-1.yml@172239021f7ba04fe7327647b213799853a9eb89",
				Owner: "octo-org",
				Repo:  "this-repo",
				Path:  ".github/workflows/workflow-1.yml",
				Ref:   "172239021f7ba04fe7327647b213799853a9eb89",
			},
		},
		{
			name:     "empty string",
			uses:     "",
			expected: ActionReference{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseActionReference(tt.uses)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParseWorkflowYAML(t *testing.T) {
	yaml := `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test ./...
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v6
  reusable:
    uses: octo-org/this-repo/.github/workflows/workflow-1.yml@main
`
	name, refs, err := ParseWorkflowYAML([]byte(yaml))
	assert.NoError(t, err)
	assert.Equal(t, "CI", name)
	assert.Len(t, refs, 5)

	// Verify refs are returned in YAML document order (build -> lint -> reusable)
	assert.Equal(t, "actions/checkout@v4", refs[0].Raw)
	assert.Equal(t, "actions", refs[0].Owner)
	assert.Equal(t, "checkout", refs[0].Repo)
	assert.Equal(t, "v4", refs[0].Ref)

	assert.Equal(t, "actions/setup-go@v5", refs[1].Raw)

	assert.Equal(t, "actions/checkout@v4", refs[2].Raw)

	assert.Equal(t, "golangci/golangci-lint-action@v6", refs[3].Raw)

	assert.Equal(t, "octo-org/this-repo/.github/workflows/workflow-1.yml@main", refs[4].Raw)
	assert.Equal(t, "octo-org", refs[4].Owner)
	assert.Equal(t, "this-repo", refs[4].Repo)
	assert.Equal(t, ".github/workflows/workflow-1.yml", refs[4].Path)
	assert.Equal(t, "main", refs[4].Ref)
}

func TestParseActionYAML(t *testing.T) {
	t.Run("composite action", func(t *testing.T) {
		yaml := `
name: My Composite Action
description: A composite action
runs:
  using: composite
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: '20'
    - run: npm install
      shell: bash
`
		refs, err := ParseActionYAML([]byte(yaml))
		assert.NoError(t, err)
		assert.Len(t, refs, 2)
		assert.Equal(t, "actions/checkout@v4", refs[0].Raw)
		assert.Equal(t, "actions/setup-node@v4", refs[1].Raw)
	})

	t.Run("non-composite action", func(t *testing.T) {
		yaml := `
name: My JS Action
description: A JavaScript action
runs:
  using: node20
  main: index.js
`
		refs, err := ParseActionYAML([]byte(yaml))
		assert.NoError(t, err)
		assert.Nil(t, refs)
	})
}

func TestActionReferenceName(t *testing.T) {
	tests := []struct {
		name     string
		ref      ActionReference
		expected string
	}{
		{
			name: "standard action",
			ref: ActionReference{
				Raw:   "actions/checkout@v4",
				Owner: "actions",
				Repo:  "checkout",
				Ref:   "v4",
			},
			expected: "actions/checkout",
		},
		{
			name: "action with path",
			ref: ActionReference{
				Raw:   "github/codeql-action/init@v3",
				Owner: "github",
				Repo:  "codeql-action",
				Path:  "init",
				Ref:   "v3",
			},
			expected: "github/codeql-action/init",
		},
		{
			name: "local action",
			ref: ActionReference{
				Raw:     "./local",
				IsLocal: true,
			},
			expected: "./local",
		},
		{
			name: "docker action",
			ref: ActionReference{
				Raw: "docker://alpine:3.8",
			},
			expected: "docker://alpine:3.8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ref.Name())
		})
	}
}

func TestActionReferenceVersionedName(t *testing.T) {
	ref := ActionReference{
		Raw:   "actions/checkout@v4",
		Owner: "actions",
		Repo:  "checkout",
		Ref:   "v4",
	}
	assert.Equal(t, "actions/checkout@v4", ref.VersionedName())
}

func TestIsCheckoutAction(t *testing.T) {
	tests := []struct {
		uses     string
		expected bool
	}{
		{"actions/checkout@v4", true},
		{"actions/checkout@v3", true},
		{"actions/checkout@abc123", true},
		{"actions/setup-node@v4", false},
		{"owner/repo@v1", false},
		{"./local-action", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.uses, func(t *testing.T) {
			assert.Equal(t, tt.expected, isCheckoutAction(tt.uses))
		})
	}
}

func TestResolveLocalActionByCheckout_MatchesExactPath(t *testing.T) {
	ref := ActionReference{
		Raw:     "./my-tool",
		IsLocal: true,
	}
	checkouts := []CheckoutPath{
		{Repository: "owner/tool", Path: "my-tool", Ref: "v1"},
	}
	resolveLocalActionByCheckout(&ref, checkouts)

	assert.Equal(t, "owner", ref.Owner)
	assert.Equal(t, "tool", ref.Repo)
	assert.Equal(t, "v1", ref.Ref)
	assert.Equal(t, "", ref.Path)
}

func TestResolveLocalActionByCheckout_MatchesSubpath(t *testing.T) {
	ref := ActionReference{
		Raw:     "./tools/lint",
		IsLocal: true,
	}
	checkouts := []CheckoutPath{
		{Repository: "owner/lint-tool", Path: "tools"},
	}
	resolveLocalActionByCheckout(&ref, checkouts)

	assert.Equal(t, "owner", ref.Owner)
	assert.Equal(t, "lint-tool", ref.Repo)
	assert.Equal(t, "lint", ref.Path)
}

func TestResolveLocalActionByCheckout_LongestPrefixWins(t *testing.T) {
	ref := ActionReference{
		Raw:     "./tools/lint/custom",
		IsLocal: true,
	}
	checkouts := []CheckoutPath{
		{Repository: "owner/tools-repo", Path: "tools"},
		{Repository: "owner/lint-repo", Path: "tools/lint"},
	}
	resolveLocalActionByCheckout(&ref, checkouts)

	assert.Equal(t, "owner", ref.Owner)
	assert.Equal(t, "lint-repo", ref.Repo)
	assert.Equal(t, "custom", ref.Path)
}

func TestResolveLocalActionByCheckout_NoMatch(t *testing.T) {
	ref := ActionReference{
		Raw:     "./unrelated-path",
		IsLocal: true,
	}
	checkouts := []CheckoutPath{
		{Repository: "owner/tool", Path: "my-tool"},
	}
	resolveLocalActionByCheckout(&ref, checkouts)

	assert.Equal(t, "", ref.Owner)
	assert.Equal(t, "", ref.Repo)
}

func TestResolveLocalActionByCheckout_SkipsReusableWorkflow(t *testing.T) {
	ref := ActionReference{
		Raw:     "./.github/workflows/reusable.yml",
		IsLocal: true,
	}
	checkouts := []CheckoutPath{
		{Repository: "owner/repo", Path: ".github"},
	}
	resolveLocalActionByCheckout(&ref, checkouts)

	assert.Equal(t, "", ref.Owner, "should not resolve reusable workflows")
}

func TestResolveLocalActionByCheckout_SkipsNonLocal(t *testing.T) {
	ref := ActionReference{
		Raw:   "actions/checkout@v4",
		Owner: "actions",
		Repo:  "checkout",
		Ref:   "v4",
	}
	checkouts := []CheckoutPath{
		{Repository: "owner/repo", Path: "actions"},
	}
	resolveLocalActionByCheckout(&ref, checkouts)

	assert.Equal(t, "actions", ref.Owner, "should not modify non-local references")
	assert.Equal(t, "checkout", ref.Repo)
}

func TestResolveLocalActionByCheckout_NoRefInCheckout(t *testing.T) {
	ref := ActionReference{
		Raw:     "./my-tool",
		IsLocal: true,
	}
	checkouts := []CheckoutPath{
		{Repository: "owner/tool", Path: "my-tool"},
	}
	resolveLocalActionByCheckout(&ref, checkouts)

	assert.Equal(t, "owner", ref.Owner)
	assert.Equal(t, "tool", ref.Repo)
	assert.Equal(t, "", ref.Ref)
}

func TestParseWorkflowYAML_ResolvesCheckoutLocalAction(t *testing.T) {
	yamlContent := `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/checkout@v4
        with:
          repository: owner/my-tool
          path: my-tool
          ref: main
      - uses: ./my-tool
`
	_, refs, err := ParseWorkflowYAML([]byte(yamlContent))
	assert.NoError(t, err)

	var localRef *ActionReference
	for i := range refs {
		if refs[i].Raw == "./my-tool" {
			localRef = &refs[i]
			break
		}
	}
	if assert.NotNil(t, localRef, "should find local action reference ./my-tool") {
		assert.Equal(t, "owner", localRef.Owner)
		assert.Equal(t, "my-tool", localRef.Repo)
		assert.Equal(t, "main", localRef.Ref)
		assert.True(t, localRef.IsLocal)
	}
}

func TestParseWorkflowYAML_NoResolutionWithoutCheckout(t *testing.T) {
	yamlContent := `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ./my-local-action
`
	_, refs, err := ParseWorkflowYAML([]byte(yamlContent))
	assert.NoError(t, err)

	var localRef *ActionReference
	for i := range refs {
		if refs[i].Raw == "./my-local-action" {
			localRef = &refs[i]
			break
		}
	}
	if assert.NotNil(t, localRef) {
		assert.Equal(t, "", localRef.Owner)
		assert.Equal(t, "", localRef.Repo)
	}
}

func TestParseWorkflowYAML_CheckoutWithoutRepositoryNotResolved(t *testing.T) {
	yamlContent := `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          path: subdir
      - uses: ./subdir
`
	_, refs, err := ParseWorkflowYAML([]byte(yamlContent))
	assert.NoError(t, err)

	var localRef *ActionReference
	for i := range refs {
		if refs[i].Raw == "./subdir" {
			localRef = &refs[i]
			break
		}
	}
	if assert.NotNil(t, localRef) {
		assert.Equal(t, "", localRef.Owner, "should not resolve: checkout without repository means self repo")
	}
}

func TestParseWorkflowYAML_CheckoutScopePerJob(t *testing.T) {
	yamlContent := `
name: CI
on: push
jobs:
  job1:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          repository: owner/tool
          path: my-tool
  job2:
    runs-on: ubuntu-latest
    steps:
      - uses: ./my-tool
`
	_, refs, err := ParseWorkflowYAML([]byte(yamlContent))
	assert.NoError(t, err)

	var localRef *ActionReference
	for i := range refs {
		if refs[i].Raw == "./my-tool" {
			localRef = &refs[i]
			break
		}
	}
	if assert.NotNil(t, localRef) {
		assert.Equal(t, "", localRef.Owner, "checkout in different job should not resolve")
	}
}

func TestResolveLocalActionByCheckout_SelfRepoWinsLongestPrefix(t *testing.T) {
	ref := ActionReference{
		Raw:     "./tools/lint/action",
		IsLocal: true,
	}
	checkouts := []CheckoutPath{
		{Repository: "owner/tools-repo", Path: "tools"},
		{Repository: "", Path: "tools/lint"}, // self repo checkout
	}
	resolveLocalActionByCheckout(&ref, checkouts)

	assert.Equal(t, "", ref.Owner, "self-repo checkout should win by longest prefix and skip resolution")
	assert.Equal(t, "", ref.Repo)
}

func TestParseWorkflowYAML_SelfCheckoutPreventsExternalResolution(t *testing.T) {
	yamlContent := `
name: CI
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          repository: other-owner/other-repo
          path: tools
      - uses: actions/checkout@v4
        with:
          path: tools/local
      - uses: ./tools/local/my-action
`
	_, refs, err := ParseWorkflowYAML([]byte(yamlContent))
	assert.NoError(t, err)

	var localRef *ActionReference
	for i := range refs {
		if refs[i].Raw == "./tools/local/my-action" {
			localRef = &refs[i]
			break
		}
	}
	if assert.NotNil(t, localRef) {
		assert.Equal(t, "", localRef.Owner, "self-repo checkout with longer prefix should prevent external resolution")
		assert.Equal(t, "", localRef.Repo)
	}
}

func TestResolveActionDepSource(t *testing.T) {
	tests := []struct {
		name     string
		action   ActionReference
		sources  map[string]bool
		expected string
	}{
		{
			name:     "local reusable workflow",
			action:   ActionReference{Raw: "./.github/workflows/release.yml", IsLocal: true},
			sources:  map[string]bool{".github/workflows/release.yml": true},
			expected: ".github/workflows/release.yml",
		},
		{
			name:     "local action in current repo",
			action:   ActionReference{Raw: "./my-action", IsLocal: true},
			sources:  map[string]bool{"my-action/action.yml": true},
			expected: "my-action/action.yml",
		},
		{
			name:     "local action yaml variant",
			action:   ActionReference{Raw: "./my-action", IsLocal: true},
			sources:  map[string]bool{"my-action/action.yaml": true},
			expected: "my-action/action.yaml",
		},
		{
			name:   "checkout-resolved local action",
			action: ActionReference{Raw: "./tools/lint", IsLocal: true, Owner: "org", Repo: "tools", Path: "lint"},
			sources: map[string]bool{
				"org/tools:lint/action.yml": true,
			},
			expected: "org/tools:lint/action.yml",
		},
		{
			name:   "remote action repository",
			action: ActionReference{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			sources: map[string]bool{
				"actions/checkout:action.yml": true,
			},
			expected: "actions/checkout:action.yml",
		},
		{
			name:   "remote action with subdirectory",
			action: ActionReference{Raw: "github/codeql-action/init@v3", Owner: "github", Repo: "codeql-action", Path: "init", Ref: "v3"},
			sources: map[string]bool{
				"github/codeql-action:init/action.yml": true,
			},
			expected: "github/codeql-action:init/action.yml",
		},
		{
			name:   "remote reusable workflow",
			action: ActionReference{Raw: "org/repo/.github/workflows/w.yml@main", Owner: "org", Repo: "repo", Path: ".github/workflows/w.yml", Ref: "main"},
			sources: map[string]bool{
				"org/repo:.github/workflows/w.yml": true,
			},
			expected: "org/repo:.github/workflows/w.yml",
		},
		{
			name:     "no matching source",
			action:   ActionReference{Raw: "actions/checkout@v4", Owner: "actions", Repo: "checkout", Ref: "v4"},
			sources:  map[string]bool{},
			expected: "",
		},
		{
			name:     "docker action (no resolution)",
			action:   ActionReference{Raw: "docker://alpine:3.8"},
			sources:  map[string]bool{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasSource := func(key string) bool {
				return tt.sources[key]
			}
			got := ResolveActionDepSource(tt.action, hasSource)
			assert.Equal(t, tt.expected, got)
		})
	}
}
