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
			sourceID, mermaidLabel(from),
			targetID, mermaidLabel(to),
		))
	}
}

// mermaidNodeID creates a safe Mermaid node identifier by replacing all
// non-alphanumeric/underscore characters with '_' and prefixing with '_'
// if the name starts with a digit.
func mermaidNodeID(name string) string {
	var b strings.Builder
	for i, ch := range name {
		switch {
		case ch >= 'A' && ch <= 'Z', ch >= 'a' && ch <= 'z', ch == '_':
			b.WriteRune(ch)
		case ch >= '0' && ch <= '9':
			if i == 0 {
				b.WriteRune('_')
			}
			b.WriteRune(ch)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

// mermaidLabel escapes a string for use inside Mermaid ["..."] label blocks.
func mermaidLabel(label string) string {
	return mermaidLabelReplacer.Replace(label)
}

var mermaidLabelReplacer = strings.NewReplacer(
	"\\", "\\\\",
	"\"", "&quot;",
	"\r\n", " ",
	"\r", "",
	"\n", " ",
)
