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
	case "dot":
		r.RenderDotGraphEdge(edges)
		return
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

// RenderDotGraphEdge renders dependency edges as a Graphviz DOT digraph
func (r *Renderer) RenderDotGraphEdge(edges []gh.GraphEdge) {
	if r.exporter != nil {
		r.RenderExportedData(edges)
		return
	}

	r.writeLine("digraph {")
	seen := make(map[string]bool)
	for _, edge := range edges {
		from := edge.GetFromName()
		to := edge.GetToName()
		edgeKey := fmt.Sprintf("%s->%s", from, to)
		if seen[edgeKey] {
			continue
		}
		seen[edgeKey] = true
		r.writeLine(fmt.Sprintf("    %s -> %s",
			dotQuote(from),
			dotQuote(to),
		))
	}
	r.writeLine("}")
}

// mermaidNodeID creates a collision-free Mermaid node identifier from a string.
// All non-alphanumeric characters are hex-encoded using their Unicode code point
// in lowercase hexadecimal, prefixed with '_' (e.g. "ci-test" and "ci_test" produce
// different IDs: "ci_2dtest" vs "ci_5ftest").
func mermaidNodeID(name string) string {
	var b strings.Builder
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			b.WriteRune(c)
		} else {
			fmt.Fprintf(&b, "_%02x", c)
		}
	}
	return b.String()
}

// dotQuote returns a DOT-safe quoted string by escaping backslashes and double quotes.
func dotQuote(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}
