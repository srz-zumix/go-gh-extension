package render

import (
	"strconv"

	"github.com/google/go-github/v79/github"
)

// RenderMilestones renders a table of milestones
func (r *Renderer) RenderMilestones(milestones []*github.Milestone) {
	if r.exporter != nil {
		r.RenderExportedData(milestones)
		return
	}

	if len(milestones) == 0 {
		r.WriteLine("No milestones found")
		return
	}

	headers := []string{"NUMBER", "TITLE", "STATE", "OPEN_ISSUES", "CLOSED_ISSUES"}
	table := r.newTableWriter(headers)

	for _, milestone := range milestones {
		row := []string{
			strconv.Itoa(milestone.GetNumber()),
			truncateString(milestone.GetTitle(), 50),
			milestone.GetState(),
			strconv.Itoa(milestone.GetOpenIssues()),
			strconv.Itoa(milestone.GetClosedIssues()),
		}
		table.Append(row)
	}

	table.Render()
}
