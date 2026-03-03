package render

import (
	"fmt"
	"html"
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
	case "drawio":
		r.RenderDrawioGraphEdge(edges)
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

// RenderDrawioGraphEdge renders dependency edges as a draw.io XML document
func (r *Renderer) RenderDrawioGraphEdge(edges []gh.GraphEdge) {
	if r.exporter != nil {
		r.RenderExportedData(edges)
		return
	}

	nodeURLs := make(map[string]string)
	var dedupEdges [][2]string
	seen := make(map[string]bool)
	for _, edge := range edges {
		from := edge.GetFromName()
		to := edge.GetToName()
		edgeKey := from + "->" + to
		if seen[edgeKey] {
			continue
		}
		seen[edgeKey] = true
		dedupEdges = append(dedupEdges, [2]string{from, to})

		// Build node URL map from edge's typed objects and host
		if _, ok := nodeURLs[from]; !ok {
			if u := edge.GetFromURL(); u != "" {
				nodeURLs[from] = u
			}
		}
		if _, ok := nodeURLs[to]; !ok {
			if u := edge.GetToURL(); u != "" {
				nodeURLs[to] = u
			}
		}
	}
	r.writeDrawioGraph(dedupEdges, nodeURLs, nil)
}

// writeDrawioGraph writes a draw.io (mxGraph) XML document from directed edges.
// Nodes are laid out using a subtree-based placement so that each parent's
// children are grouped directly below it, avoiding arrows that cross through
// sibling nodes.
// nodeURLs maps node labels to their remote URLs. If nil or a key is missing,
// the node is rendered without a link.
// nodeColors maps node labels to border color hex strings (e.g. "#FF9800").
// If nil or a key is missing, the default border style is used.
func (r *Renderer) writeDrawioGraph(edges [][2]string, nodeURLs map[string]string, nodeColors map[string]string) {
	// Collect unique nodes preserving insertion order
	nodeIndex := make(map[string]int)
	var nodes []string
	for _, e := range edges {
		if _, ok := nodeIndex[e[0]]; !ok {
			nodeIndex[e[0]] = len(nodes)
			nodes = append(nodes, e[0])
		}
		if _, ok := nodeIndex[e[1]]; !ok {
			nodeIndex[e[1]] = len(nodes)
			nodes = append(nodes, e[1])
		}
	}

	// Build adjacency lists
	outgoing := make(map[string][]string)
	inDegree := make(map[string]int)
	for _, n := range nodes {
		inDegree[n] = 0
	}
	for _, e := range edges {
		outgoing[e[0]] = append(outgoing[e[0]], e[1])
		inDegree[e[1]]++
	}

	// Compute layer (depth) for each node using longest-path from sources.
	remaining := make(map[string]int, len(nodes))
	for k, v := range inDegree {
		remaining[k] = v
	}
	depth := make(map[string]int)
	queue := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if remaining[n] == 0 {
			queue = append(queue, n)
			depth[n] = 0
		}
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range outgoing[cur] {
			d := depth[cur] + 1
			if d > depth[next] {
				depth[next] = d
			}
			remaining[next]--
			if remaining[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	// Handle nodes in cycles: assign them to one layer below the deepest processed node
	maxProcessed := 0
	for _, d := range depth {
		if d > maxProcessed {
			maxProcessed = d
		}
	}
	for _, n := range nodes {
		if remaining[n] > 0 {
			depth[n] = maxProcessed + 1
		}
	}

	// Identify root nodes (no incoming edges)
	var roots []string
	for _, n := range nodes {
		if inDegree[n] == 0 {
			roots = append(roots, n)
		}
	}

	// Build a primary-children tree for subtree layout.
	// Each child is assigned to only one parent (first parent encountered in
	// depth-first order from the roots) so subtrees don't overlap.
	primaryChildren := make(map[string][]string)
	owned := make(map[string]bool)
	var assignChildren func(parent string)
	assignChildren = func(parent string) {
		for _, child := range outgoing[parent] {
			if owned[child] {
				continue
			}
			owned[child] = true
			primaryChildren[parent] = append(primaryChildren[parent], child)
			assignChildren(child)
		}
	}
	for _, root := range roots {
		owned[root] = true
	}
	for _, root := range roots {
		assignChildren(root)
	}
	// Any node not yet owned (e.g. cycle-only nodes) becomes a root
	for _, n := range nodes {
		if !owned[n] {
			owned[n] = true
			roots = append(roots, n)
			assignChildren(n)
		}
	}

	const (
		nodeWidth  = 180
		nodeHeight = 60
		xGap       = 60
		yGap       = 120
	)

	// Recursive subtree layout: place each node's primary children contiguously
	// below it, then center the parent above them.
	posX := make(map[string]int)
	posY := make(map[string]int)

	var layoutSubtree func(node string, startX int) int
	layoutSubtree = func(node string, startX int) int {
		posY[node] = depth[node] * (nodeHeight + yGap)

		children := primaryChildren[node]
		if len(children) == 0 {
			posX[node] = startX
			return nodeWidth
		}

		// Layout children left-to-right
		cursor := startX
		for i, child := range children {
			w := layoutSubtree(child, cursor)
			cursor += w
			if i < len(children)-1 {
				cursor += xGap
			}
		}
		subtreeWidth := cursor - startX
		if subtreeWidth < nodeWidth {
			subtreeWidth = nodeWidth
		}

		// Center parent above its children span
		firstChildX := posX[children[0]]
		lastChildX := posX[children[len(children)-1]]
		childrenCenter := (firstChildX + lastChildX + nodeWidth) / 2
		posX[node] = childrenCenter - nodeWidth/2

		return subtreeWidth
	}

	cursor := 0
	for i, root := range roots {
		w := layoutSubtree(root, cursor)
		cursor += w
		if i < len(roots)-1 {
			cursor += xGap
		}
	}

	r.writeLine(`<mxfile host="gh-deps-kit">`)
	r.writeLine(`  <diagram name="Dependencies">`)
	r.writeLine(`    <mxGraphModel>`)
	r.writeLine(`      <root>`)
	r.writeLine(`        <mxCell id="0"/>`)
	r.writeLine(`        <mxCell id="1" parent="0"/>`)

	// Write node cells (IDs start from 2)
	for i, name := range nodes {
		cellID := i + 2
		// Wrap text in a div with word-break:break-all so that long unbroken
		// strings (e.g. file paths) are wrapped at character boundaries.
		var innerHTML string
		if nodeURLs != nil {
			if u, ok := nodeURLs[name]; ok && u != "" {
				innerHTML = fmt.Sprintf(`<a href="%s">%s</a>`, html.EscapeString(u), html.EscapeString(name))
			} else {
				innerHTML = html.EscapeString(name)
			}
		} else {
			innerHTML = html.EscapeString(name)
		}
		nodeValue := html.EscapeString(fmt.Sprintf(`<div style="word-break:break-all">%s</div>`, innerHTML))
		style := "rounded=1;whiteSpace=wrap;html=1;"
		if nodeColors != nil {
			if color, ok := nodeColors[name]; ok && color != "" {
				style += "strokeColor=" + color + ";strokeWidth=2;"
			}
		}
		r.writeLine(fmt.Sprintf(`        <mxCell id="%d" value="%s" style="%s" vertex="1" parent="1">`,
			cellID, nodeValue, style))
		r.writeLine(fmt.Sprintf(`          <mxGeometry x="%d" y="%d" width="%d" height="%d" as="geometry"/>`,
			posX[name], posY[name], nodeWidth, nodeHeight))
		r.writeLine(`        </mxCell>`)
	}

	// Write edge cells
	// Distribute entry/exit points so that multiple arrows to/from the same node
	// don't overlap. Count edges per source (exit from bottom) and per target
	// (enter at top), then assign evenly spaced positions along the node width.
	type portCounter struct {
		total int
		used  int
	}
	exitCount := make(map[string]*portCounter)
	entryCount := make(map[string]*portCounter)
	for _, e := range edges {
		if exitCount[e[0]] == nil {
			exitCount[e[0]] = &portCounter{}
		}
		exitCount[e[0]].total++
		if entryCount[e[1]] == nil {
			entryCount[e[1]] = &portCounter{}
		}
		entryCount[e[1]].total++
	}

	edgeID := len(nodes) + 2
	for _, e := range edges {
		sourceID := nodeIndex[e[0]] + 2
		targetID := nodeIndex[e[1]] + 2

		// Compute exit point on source bottom edge
		ec := exitCount[e[0]]
		var exitX float64
		if ec.total == 1 {
			exitX = 0.5
		} else {
			exitX = (float64(ec.used) + 1) / float64(ec.total+1)
		}
		ec.used++

		// Compute entry point on target top edge
		nc := entryCount[e[1]]
		var entryX float64
		if nc.total == 1 {
			entryX = 0.5
		} else {
			entryX = (float64(nc.used) + 1) / float64(nc.total+1)
		}
		nc.used++

		style := fmt.Sprintf("edgeStyle=orthogonalEdgeStyle;rounded=1;orthogonalLoop=1;jettySize=auto;html=1;exitX=%.4f;exitY=1;exitDx=0;exitDy=0;entryX=%.4f;entryY=0;entryDx=0;entryDy=0;",
			exitX, entryX)
		r.writeLine(fmt.Sprintf(`        <mxCell id="%d" style="%s" edge="1" source="%d" target="%d" parent="1">`,
			edgeID, style, sourceID, targetID))

		// For edges spanning more than one layer, add a waypoint in the gap
		// just below the source layer. This forces the horizontal routing
		// segment into the empty gap rather than the midpoint, which would
		// cross through intermediate-layer sibling nodes.
		sourceDepth := depth[e[0]]
		targetDepth := depth[e[1]]
		layerSpan := targetDepth - sourceDepth
		if layerSpan < 0 {
			layerSpan = -layerSpan
		}
		if layerSpan > 1 {
			waypointY := posY[e[0]] + nodeHeight + yGap/3
			entryXPos := posX[e[1]] + int(entryX*float64(nodeWidth))
			r.writeLine(`          <mxGeometry relative="1" as="geometry">`)
			r.writeLine(`            <Array as="points">`)
			r.writeLine(fmt.Sprintf(`              <mxPoint x="%d" y="%d"/>`, entryXPos, waypointY))
			r.writeLine(`            </Array>`)
			r.writeLine(`          </mxGeometry>`)
		} else {
			r.writeLine(`          <mxGeometry relative="1" as="geometry"/>`)
		}
		r.writeLine(`        </mxCell>`)
		edgeID++
	}

	r.writeLine(`      </root>`)
	r.writeLine(`    </mxGraphModel>`)
	r.writeLine(`  </diagram>`)
	r.writeLine(`</mxfile>`)
}

// dotQuote returns a DOT-safe quoted string by escaping backslashes and double quotes.
func dotQuote(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}


