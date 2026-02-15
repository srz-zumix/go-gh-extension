package render

import (
	"fmt"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

// RenderGraphEdge renders dependency edges as a Mermaid flowchart
func (r *Renderer) RenderGraphEdge(format string, edges []gh.GraphEdge) {
	if r.exporter != nil {
		r.RenderExportedData(edges)
		return
	}

	switch strings.ToLower(format) {
	case "markdown":
		r.RenderMarkdownGraphEdge(edges)
		return
	case "mermaid":
		r.RenderMermaidGraphEdge(edges)
		return
	default:
		r.WriteLine(fmt.Sprintf("Unsupported graph format: %s", format))
		return
	}
}

// RenderMarkdownGraphEdge renders dependency edges in a simple Markdown format
func (r *Renderer) RenderMarkdownGraphEdge(edges []gh.GraphEdge) {
	if r.exporter != nil {
		r.RenderExportedData(edges)
		return
	}
	r.WriteLine("```mermaid")
	r.RenderMermaidGraphEdge(edges)
	r.WriteLine("```")
}

// RenderMermaidGraphEdge renders dependency edges as a Mermaid flowchart
func (r *Renderer) RenderMermaidGraphEdge(edges []gh.GraphEdge) {
	if r.exporter != nil {
		r.RenderExportedData(edges)
		return
	}

	r.writeLine("graph LR")
	seen := make(map[string]bool)
	for _, edge := range edges {
		from := edge.GetFromName()
		to := edge.GetToName()
		sourceID := mermaidNodeID(from)
		targetID := mermaidNodeID(to)
		edgeKey := fmt.Sprintf("%s-->%s", sourceID, targetID)
		if seen[edgeKey] {
			continue
		}
		seen[edgeKey] = true
		r.writeLine(fmt.Sprintf("    %s[\"%s\"] --> %s[\"%s\"]",
			sourceID, from,
			targetID, to,
		))
	}
}

// mermaidNodeID creates a safe Mermaid node identifier from a repository
func mermaidNodeID(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}
