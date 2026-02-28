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
