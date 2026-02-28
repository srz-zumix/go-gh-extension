package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
