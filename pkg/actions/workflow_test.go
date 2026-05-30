package actions

import "testing"

func TestWorkflowFilePathFromRef(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want string
	}{
		{
			name: "ref with branch",
			ref:  "octo/repo/.github/workflows/ci.yml@refs/heads/main",
			want: ".github/workflows/ci.yml",
		},
		{
			name: "ref with tag",
			ref:  "octo/repo/.github/workflows/release.yaml@refs/tags/v1.0.0",
			want: ".github/workflows/release.yaml",
		},
		{
			name: "empty",
			ref:  "",
			want: "",
		},
		{
			name: "missing path segment",
			ref:  "octo/repo@refs/heads/main",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WorkflowFilePathFromRef(tt.ref); got != tt.want {
				t.Errorf("WorkflowFilePathFromRef(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}
