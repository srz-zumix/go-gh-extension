package render

import (
	"testing"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
)

func TestRunnerGroupFieldGetters(t *testing.T) {
	getter := NewRunnerGroupFieldGetters()
	group := &github.RunnerGroup{
		ID:                       github.Ptr(int64(7)),
		Name:                     github.Ptr("prod-group"),
		Visibility:               github.Ptr("selected"),
		Default:                  github.Ptr(false),
		Inherited:                github.Ptr(true),
		AllowsPublicRepositories: github.Ptr(true),
	}

	assert.Equal(t, "7", getter.GetField(group, "ID"))
	assert.Equal(t, "prod-group", getter.GetField(group, "NAME"))
	assert.Equal(t, "selected", getter.GetField(group, "VISIBILITY"))
	// Booleans render as YES/NO.
	assert.Equal(t, "NO", getter.GetField(group, "DEFAULT"))
	assert.Equal(t, "YES", getter.GetField(group, "INHERITED"))
	assert.Equal(t, "YES", getter.GetField(group, "PUBLIC_REPOSITORIES"))
	// Field lookup is case-insensitive.
	assert.Equal(t, "prod-group", getter.GetField(group, "name"))
	// Unknown fields render as empty.
	assert.Equal(t, "", getter.GetField(group, "UNKNOWN"))
}

func TestRunnerGroupFieldsSorted(t *testing.T) {
	fields := RunnerGroupFields()
	assert.Equal(t, []string{
		"DEFAULT",
		"ID",
		"INHERITED",
		"NAME",
		"PUBLIC_REPOSITORIES",
		"RESTRICTED_TO_WORKFLOWS",
		"VISIBILITY",
	}, fields)
}

func TestRenderRunnerGroupsDefaultHeaders(t *testing.T) {
	r := NewStringRenderer(nil)
	groups := []*github.RunnerGroup{
		{
			ID:         github.Ptr(int64(1)),
			Name:       github.Ptr("default-group"),
			Visibility: github.Ptr("all"),
			Default:    github.Ptr(true),
			Inherited:  github.Ptr(false),
		},
	}

	assert.NoError(t, r.Renderer.RenderRunnerGroups(groups, nil))

	out := r.Stdout.String()
	// Default headers.
	for _, h := range []string{"ID", "NAME", "VISIBILITY", "DEFAULT", "INHERITED"} {
		assert.Contains(t, out, h)
	}
	// Representative rendered values, including YES/NO booleans.
	assert.Contains(t, out, "default-group")
	assert.Contains(t, out, "all")
	assert.Contains(t, out, "YES")
	assert.Contains(t, out, "NO")
}

func TestRenderRunnerGroupNil(t *testing.T) {
	r := NewStringRenderer(nil)

	var err error
	assert.NotPanics(t, func() {
		err = r.Renderer.RenderRunnerGroup(nil, nil)
	})
	assert.NoError(t, err)
	assert.Empty(t, r.Stdout.String())
}

func TestRenderRunnerGroupsEmpty(t *testing.T) {
	r := NewStringRenderer(nil)
	assert.NoError(t, r.Renderer.RenderRunnerGroups(nil, nil))
	assert.Contains(t, r.Stdout.String(), "No runner groups.")
}
