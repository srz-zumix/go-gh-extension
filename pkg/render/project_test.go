package render

import (
	"testing"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/assert"
)

func TestProjectV2ItemFieldGettersContentFields(t *testing.T) {
	item := &client.ProjectV2Item{
		Content: client.ProjectV2ItemContent{
			Type:      client.ProjectV2ItemTypeIssue,
			Number:    42,
			Title:     "Broken build",
			State:     "CLOSED",
			RepoOwner: "octo",
			RepoName:  "hello",
			Assignees: []string{"alice", "bob"},
			Labels:    []string{"bug", "urgent"},
			Milestone: "v1.0",
			CreatedAt: "2026-01-02T03:04:05Z",
			ClosedAt:  "2026-03-04T05:06:07Z",
		},
	}

	getter := NewProjectV2ItemFieldGetters()
	assert.Equal(t, "CLOSED", getter.GetField(item, "STATE"))
	assert.Equal(t, "octo/hello", getter.GetField(item, "REPOSITORY"))
	assert.Equal(t, "alice, bob", getter.GetField(item, "ASSIGNEES"))
	assert.Equal(t, "bug, urgent", getter.GetField(item, "LABELS"))
	assert.Equal(t, "v1.0", getter.GetField(item, "MILESTONE"))
	assert.Contains(t, getter.GetField(item, "CREATED_AT"), "2026-01-02")
	assert.Contains(t, getter.GetField(item, "CLOSED_AT"), "2026-03-04")
	// Timestamps that were never set render as empty rather than as a zero time.
	assert.Equal(t, "", getter.GetField(item, "UPDATED_AT"))
}

func TestProjectV2ItemFieldGettersDraftIssue(t *testing.T) {
	item := &client.ProjectV2Item{
		Content: client.ProjectV2ItemContent{
			Type:  client.ProjectV2ItemTypeDraftIssue,
			Title: "Spike",
		},
	}

	getter := NewProjectV2ItemFieldGetters()
	assert.Equal(t, "", getter.GetField(item, "STATE"))
	assert.Equal(t, "", getter.GetField(item, "REPOSITORY"))
	assert.Equal(t, "", getter.GetField(item, "ASSIGNEES"))
	assert.Equal(t, "", getter.GetField(item, "LABELS"))
}

func TestRenderProjectV2ItemsWithContentFields(t *testing.T) {
	r := NewStringRenderer(nil)
	items := []client.ProjectV2Item{
		{
			Content: client.ProjectV2ItemContent{
				Type:      client.ProjectV2ItemTypeIssue,
				Number:    42,
				Title:     "Broken build",
				State:     "OPEN",
				RepoOwner: "octo",
				RepoName:  "hello",
				Assignees: []string{"alice"},
			},
		},
	}

	assert.NoError(t, r.Renderer.RenderProjectV2Items(items, []string{"NUMBER", "TITLE", "STATE", "REPOSITORY", "ASSIGNEES"}))

	out := r.Stdout.String()
	assert.Contains(t, out, "Broken build")
	assert.Contains(t, out, "OPEN")
	assert.Contains(t, out, "octo/hello")
	assert.Contains(t, out, "alice")
}
