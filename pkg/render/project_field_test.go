package render

import (
	"testing"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/assert"
)

func TestProjectV2FieldOptionsSummary(t *testing.T) {
	selectField := &client.ProjectV2Field{
		Name:     "Status",
		DataType: "SINGLE_SELECT",
		Options: []client.ProjectV2SelectOption{
			{Name: "Todo"}, {Name: "Done"},
		},
	}
	assert.Equal(t, "Todo, Done", projectV2FieldOptionsSummary(selectField))

	iterationField := &client.ProjectV2Field{
		Name:                "Sprint",
		DataType:            "ITERATION",
		Iterations:          []client.ProjectV2IterationOption{{Title: "Sprint 3"}},
		CompletedIterations: []client.ProjectV2IterationOption{{Title: "Sprint 1"}, {Title: "Sprint 2"}},
	}
	assert.Equal(t, "3 iterations (2 completed)", projectV2FieldOptionsSummary(iterationField))

	assert.Equal(t, "", projectV2FieldOptionsSummary(&client.ProjectV2Field{Name: "Title", DataType: "TITLE"}))
}

func TestProjectV2OptionAndIterationDetail(t *testing.T) {
	assert.Equal(t, "green - shipped", projectV2OptionDetail(client.ProjectV2SelectOption{Color: "GREEN", Description: "shipped"}))
	assert.Equal(t, "green", projectV2OptionDetail(client.ProjectV2SelectOption{Color: "GREEN"}))
	assert.Equal(t, "", projectV2OptionDetail(client.ProjectV2SelectOption{}))

	it := client.ProjectV2IterationOption{Title: "Sprint 1", StartDate: "2026-01-05", Duration: 14}
	assert.Equal(t, "2026-01-05 +14d", projectV2IterationDetail(it, false))
	assert.Equal(t, "2026-01-05 +14d (completed)", projectV2IterationDetail(it, true))
}

func TestRenderProjectV2Fields(t *testing.T) {
	r := NewStringRenderer(nil)
	fields := []client.ProjectV2Field{
		{Name: "Status", DataType: "SINGLE_SELECT", Options: []client.ProjectV2SelectOption{{Name: "Todo"}}},
	}

	assert.NoError(t, r.Renderer.RenderProjectV2Fields(fields))

	out := r.Stdout.String()
	assert.Contains(t, out, "Status")
	assert.Contains(t, out, "SINGLE_SELECT")
	assert.Contains(t, out, "Todo")
}

func TestRenderProjectV2FieldOptions(t *testing.T) {
	r := NewStringRenderer(nil)
	fields := []client.ProjectV2Field{
		{Name: "Title", DataType: "TITLE"},
		{Name: "Status", DataType: "SINGLE_SELECT", Options: []client.ProjectV2SelectOption{{Name: "Todo", Color: "GRAY"}}},
		{
			Name:                "Sprint",
			DataType:            "ITERATION",
			Iterations:          []client.ProjectV2IterationOption{{Title: "Sprint 2", StartDate: "2026-01-19", Duration: 14}},
			CompletedIterations: []client.ProjectV2IterationOption{{Title: "Sprint 1", StartDate: "2026-01-05", Duration: 14}},
		},
	}

	assert.NoError(t, r.Renderer.RenderProjectV2FieldOptions(fields))

	out := r.Stdout.String()
	assert.Contains(t, out, "Todo")
	assert.Contains(t, out, "gray")
	assert.Contains(t, out, "Sprint 1")
	assert.Contains(t, out, "(completed)")
	assert.Contains(t, out, "Sprint 2")
	// Fields without options or iterations produce no rows.
	assert.NotContains(t, out, "TITLE")
}

func TestRenderProjectV2FieldsEmpty(t *testing.T) {
	r := NewStringRenderer(nil)
	assert.NoError(t, r.Renderer.RenderProjectV2Fields(nil))
	assert.Contains(t, r.Stdout.String(), "No fields.")

	r2 := NewStringRenderer(nil)
	assert.NoError(t, r2.Renderer.RenderProjectV2FieldOptions([]client.ProjectV2Field{{Name: "Title", DataType: "TITLE"}}))
	assert.Contains(t, r2.Stdout.String(), "No field options.")
}
