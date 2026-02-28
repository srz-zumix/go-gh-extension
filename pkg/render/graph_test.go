package render

import (
	"testing"
)

func Test_mermaidNodeID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple name", input: "simple", want: "simple"},
		{name: "hyphen replaced", input: "my-action", want: "my_action"},
		{name: "slash replaced", input: ".github/workflows/ci.yml", want: "_github_workflows_ci_yml"},
		{name: "colon replaced", input: "org/repo:action.yml", want: "org_repo_action_yml"},
		{name: "dot replaced", input: "action.yml", want: "action_yml"},
		{name: "starts with digit", input: "1action", want: "_1action"},
		{name: "all digits", input: "123", want: "_123"},
		{name: "underscore preserved", input: "my_action", want: "my_action"},
		{name: "mixed special chars", input: "owner/repo@v1.2.3", want: "owner_repo_v1_2_3"},
		{name: "empty string", input: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mermaidNodeID(tt.input)
			if got != tt.want {
				t.Errorf("mermaidNodeID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func Test_mermaidLabel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple label", input: "simple", want: "simple"},
		{name: "quote escaped", input: `say "hello"`, want: "say &quot;hello&quot;"},
		{name: "newline replaced", input: "line1\nline2", want: "line1 line2"},
		{name: "carriage return removed", input: "line1\r\nline2", want: "line1 line2"},
		{name: "backslash escaped", input: `path\to\file`, want: `path\\to\\file`},
		{name: "workflow path", input: ".github/workflows/ci.yml", want: ".github/workflows/ci.yml"},
		{name: "repo key with colon", input: "org/repo:action.yml", want: "org/repo:action.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mermaidLabel(tt.input)
			if got != tt.want {
				t.Errorf("mermaidLabel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
