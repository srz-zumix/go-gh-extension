package render

import (
	"testing"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/assert"
)

func TestProjectV2ViewField(t *testing.T) {
	v := &client.ProjectV2View{
		ID:                    "PVTV_1",
		Number:                2,
		Name:                  "Board",
		Layout:                "BOARD_LAYOUT",
		Filter:                "is:open",
		VisibleFields:         []string{"Title", "Status"},
		GroupByFields:         []string{"Status"},
		VerticalGroupByFields: []string{"Priority"},
		SortBy: []client.ProjectV2ViewSortBy{
			{FieldName: "Priority", Direction: "DESC"},
			{FieldName: "Title", Direction: "ASC"},
		},
	}

	assert.Equal(t, "PVTV_1", projectV2ViewField(v, "ID"))
	assert.Equal(t, "2", projectV2ViewField(v, "NUMBER"))
	assert.Equal(t, "Board", projectV2ViewField(v, "NAME"))
	// The _LAYOUT suffix is stripped for readability.
	assert.Equal(t, "BOARD", projectV2ViewField(v, "LAYOUT"))
	assert.Equal(t, "is:open", projectV2ViewField(v, "FILTER"))
	assert.Equal(t, "Status", projectV2ViewField(v, "GROUPBY"))
	assert.Equal(t, "Priority", projectV2ViewField(v, "VERTICALGROUPBY"))
	assert.Equal(t, "Priority (DESC), Title (ASC)", projectV2ViewField(v, "SORTBY"))
	assert.Equal(t, "Title, Status", projectV2ViewField(v, "VISIBLEFIELDS"))
	assert.Equal(t, "", projectV2ViewField(v, "UNKNOWN"))
}

func TestRenderProjectV2Views(t *testing.T) {
	r := NewStringRenderer(nil)
	views := []client.ProjectV2View{
		{Number: 1, Name: "Table", Layout: "TABLE_LAYOUT", Filter: "is:issue"},
	}

	assert.NoError(t, r.Renderer.RenderProjectV2Views(views, nil))

	out := r.Stdout.String()
	assert.Contains(t, out, "Table")
	assert.Contains(t, out, "TABLE")
	assert.Contains(t, out, "is:issue")
}

func TestRenderProjectV2ViewsEmpty(t *testing.T) {
	r := NewStringRenderer(nil)
	assert.NoError(t, r.Renderer.RenderProjectV2Views(nil, nil))
	assert.Contains(t, r.Stdout.String(), "No views.")
}
