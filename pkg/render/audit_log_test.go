package render

import (
	"testing"
	"time"

	"github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/assert"
)

func TestAuditEntryFieldGetters(t *testing.T) {
	entry := &github.AuditEntry{
		Action:    github.Ptr("repo.update_actions_secret"),
		Actor:     github.Ptr("octocat"),
		Timestamp: &github.Timestamp{Time: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)},
		AdditionalFields: map[string]any{
			"repo":             "octo/hello",
			"key":              "MY_SECRET",
			"environment_name": "production",
			"read_only":        true,
		},
	}

	getter := newAuditEntryFieldGetters()
	assert.Equal(t, "repo.update_actions_secret", getter.getField(entry, "ACTION"))
	assert.Equal(t, "octocat", getter.getField(entry, "actor"))
	// Headers without a matching struct field fall back to the additional fields.
	assert.Equal(t, "octo/hello", getter.getField(entry, "REPO"))
	assert.Equal(t, "MY_SECRET", getter.getField(entry, "KEY"))
	assert.Equal(t, "production", getter.getField(entry, "ENVIRONMENT_NAME"))
	assert.Equal(t, "true", getter.getField(entry, "READ_ONLY"))
	assert.Equal(t, "", getter.getField(entry, "UNKNOWN"))
}

func TestRenderAuditEntries(t *testing.T) {
	r := NewStringRenderer(nil)
	entries := []*github.AuditEntry{
		{
			Action:           github.Ptr("environment.create_actions_secret"),
			Actor:            github.Ptr("octocat"),
			AdditionalFields: map[string]any{"key": "MY_SECRET", "environment_name": "production"},
		},
	}

	err := r.Renderer.RenderAuditEntries(entries, []string{"ACTION", "ENVIRONMENT_NAME", "KEY", "ACTOR"})
	assert.NoError(t, err)

	out := r.Stdout.String()
	assert.Contains(t, out, "environment.create_actions_secret")
	assert.Contains(t, out, "production")
	assert.Contains(t, out, "MY_SECRET")
	assert.Contains(t, out, "octocat")
}

func TestRenderAuditEntriesEmpty(t *testing.T) {
	r := NewStringRenderer(nil)
	assert.NoError(t, r.Renderer.RenderAuditEntries(nil, nil))
	assert.Contains(t, r.Stdout.String(), "No audit log entries found.")
}
