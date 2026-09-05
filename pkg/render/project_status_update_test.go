package render

import (
	"testing"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/assert"
)

func TestFirstLine(t *testing.T) {
	assert.Equal(t, "first", firstLine("first\nsecond"))
	assert.Equal(t, "first", firstLine("\n  \n  first  \nsecond"))
	assert.Equal(t, "", firstLine(""))
	assert.Equal(t, "", firstLine("\n\n"))
}

func TestFormatRFC3339(t *testing.T) {
	assert.Equal(t, "", formatRFC3339(""))
	// Unparsable values are passed through unchanged.
	assert.Equal(t, "not-a-time", formatRFC3339("not-a-time"))
	assert.Contains(t, formatRFC3339("2026-06-15T12:00:00Z"), "2026-06-15")
}

func TestRenderProjectV2StatusUpdates(t *testing.T) {
	r := NewStringRenderer(nil)
	updates := []client.ProjectV2StatusUpdate{
		{
			Status:     client.ProjectV2StatusUpdateStatusOnTrack,
			StartDate:  "2026-06-01",
			TargetDate: "2026-06-30",
			Creator:    "octocat",
			Body:       "\nEverything is fine\nmore details",
			CreatedAt:  "2026-06-15T12:00:00Z",
		},
	}

	assert.NoError(t, r.Renderer.RenderProjectV2StatusUpdates(updates))

	out := r.Stdout.String()
	assert.Contains(t, out, "ON_TRACK")
	assert.Contains(t, out, "2026-06-01")
	assert.Contains(t, out, "2026-06-30")
	assert.Contains(t, out, "octocat")
	assert.Contains(t, out, "Everything is fine")
	assert.NotContains(t, out, "more details")
}

func TestRenderProjectV2StatusUpdatesEmpty(t *testing.T) {
	r := NewStringRenderer(nil)
	assert.NoError(t, r.Renderer.RenderProjectV2StatusUpdates(nil))
	assert.Contains(t, r.Stdout.String(), "No status updates.")
}
