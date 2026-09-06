package render

import (
	"testing"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
)

func TestRunnerFieldGetters(t *testing.T) {
	getter := NewRunnerFieldGetters()

	runner := &github.Runner{
		ID:            github.Ptr(int64(12)),
		Name:          github.Ptr("build-01"),
		OS:            github.Ptr("linux"),
		Status:        github.Ptr("online"),
		Busy:          github.Ptr(true),
		RunnerGroupID: github.Ptr(int64(3)),
		Labels: []*github.RunnerLabels{
			{Name: github.Ptr("self-hosted")},
			{Name: github.Ptr("linux")},
		},
	}

	assert.Equal(t, "12", getter.GetField(runner, "ID"))
	assert.Equal(t, "build-01", getter.GetField(runner, "NAME"))
	assert.Equal(t, "linux", getter.GetField(runner, "OS"))
	assert.Equal(t, "online", getter.GetField(runner, "STATUS"))
	// Booleans render as YES/NO for consistency with other fields.
	assert.Equal(t, "YES", getter.GetField(runner, "BUSY"))
	assert.Equal(t, "3", getter.GetField(runner, "GROUP"))
	assert.Equal(t, "self-hosted, linux", getter.GetField(runner, "LABELS"))
	assert.Equal(t, "", getter.GetField(runner, "UNKNOWN"))
}

func TestRunnerFieldGettersCustomField(t *testing.T) {
	getter := NewRunnerFieldGetters()
	getter.Func["EPHEMERAL"] = func(r *github.Runner) string {
		return ToString(r.Ephemeral)
	}

	assert.Contains(t, getter.Fields(), "EPHEMERAL")
	assert.NotContains(t, RunnerFields(), "EPHEMERAL")
	assert.Equal(t, "YES", getter.GetField(&github.Runner{Ephemeral: github.Ptr(true)}, "ephemeral"))
}
