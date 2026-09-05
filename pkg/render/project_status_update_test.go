package render

import (
	"strings"
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
	assert.Equal(t, "2026-06-15 12:00:00", formatRFC3339("2026-06-15T12:00:00Z"))
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

func TestRenderProjectV2StatusUpdatesNewestFirst(t *testing.T) {
	r := NewStringRenderer(nil)
	// Deliberately shuffled input so neither forward nor reverse traversal of
	// the input yields newest-first; only sorting by CreatedAt does.
	updates := []client.ProjectV2StatusUpdate{
		{Creator: "middle", CreatedAt: "2026-06-15T12:00:00Z"},
		{Creator: "newest", CreatedAt: "2026-07-01T09:00:00Z"},
		{Creator: "oldest", CreatedAt: "2026-05-01T08:00:00Z"},
	}
	original := make([]client.ProjectV2StatusUpdate, len(updates))
	copy(original, updates)

	assert.NoError(t, r.Renderer.RenderProjectV2StatusUpdates(updates))

	out := r.Stdout.String()
	iNewest := strings.Index(out, "newest")
	iMiddle := strings.Index(out, "middle")
	iOldest := strings.Index(out, "oldest")
	assert.NotEqual(t, -1, iNewest)
	assert.NotEqual(t, -1, iMiddle)
	assert.NotEqual(t, -1, iOldest)
	assert.Less(t, iNewest, iMiddle, "newest should appear before middle")
	assert.Less(t, iMiddle, iOldest, "middle should appear before oldest")

	// The renderer must not mutate the caller's slice.
	assert.Equal(t, original, updates)
}

func TestRenderProjectV2StatusUpdatesMalformedSortsLast(t *testing.T) {
	r := NewStringRenderer(nil)
	// Unparsable CreatedAt values sort last and keep their relative order.
	updates := []client.ProjectV2StatusUpdate{
		{Creator: "bad-first", CreatedAt: "not-a-time"},
		{Creator: "valid", CreatedAt: "2026-07-01T09:00:00Z"},
		{Creator: "bad-second", CreatedAt: ""},
	}
	assert.NoError(t, r.Renderer.RenderProjectV2StatusUpdates(updates))

	out := r.Stdout.String()
	iValid := strings.Index(out, "valid")
	iBadFirst := strings.Index(out, "bad-first")
	iBadSecond := strings.Index(out, "bad-second")
	assert.NotEqual(t, -1, iValid)
	assert.NotEqual(t, -1, iBadFirst)
	assert.NotEqual(t, -1, iBadSecond)
	assert.Less(t, iValid, iBadFirst, "valid timestamp should sort before malformed ones")
	assert.Less(t, iBadFirst, iBadSecond, "malformed values should keep their relative order")
}

func TestRenderProjectV2StatusUpdatesEmpty(t *testing.T) {
	r := NewStringRenderer(nil)
	assert.NoError(t, r.Renderer.RenderProjectV2StatusUpdates(nil))
	assert.Contains(t, r.Stdout.String(), "No status updates.")
}
