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

	assert.Equal(t, "actions/checkout@v4", refs[0].Raw)
	assert.Equal(t, "actions", refs[0].Owner)
	assert.Equal(t, "checkout", refs[0].Repo)
	assert.Equal(t, "v4", refs[0].Ref)

	// Check reusable workflow reference is present
	found := false
	for _, ref := range refs {
		if ref.Owner == "octo-org" && ref.Repo == "this-repo" {
			found = true
			assert.Equal(t, ".github/workflows/workflow-1.yml", ref.Path)
			assert.Equal(t, "main", ref.Ref)
		}
	}
	assert.True(t, found, "reusable workflow reference should be found")
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
