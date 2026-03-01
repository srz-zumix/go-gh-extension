package render

import (
	"fmt"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// WorkflowDependencyFieldGetter defines a function to get a field value from parser.ActionReference
type WorkflowDependencyFieldGetter func(ref *parser.ActionReference) string

// WorkflowDependencyFieldGetters holds field getters for ActionReference table rendering.
type WorkflowDependencyFieldGetters struct {
	Func map[string]WorkflowDependencyFieldGetter
}

// NewWorkflowDependencyFieldGetters creates field getters for ActionReference table rendering
func NewWorkflowDependencyFieldGetters() *WorkflowDependencyFieldGetters {
	return &WorkflowDependencyFieldGetters{
		Func: map[string]WorkflowDependencyFieldGetter{
			"NAME": func(ref *parser.ActionReference) string {
				return ref.Name()
			},
			"VERSION": func(ref *parser.ActionReference) string {
				return ref.Ref
			},
			"OWNER": func(ref *parser.ActionReference) string {
				return ref.Owner
			},
			"REPO": func(ref *parser.ActionReference) string {
				return ref.Repo
			},
			"PATH": func(ref *parser.ActionReference) string {
				return ref.Path
			},
			"RAW": func(ref *parser.ActionReference) string {
				return ref.Raw
			},
			"USING": func(ref *parser.ActionReference) string {
				return ref.Using
			},
			"NODE_VERSION": func(ref *parser.ActionReference) string {
				return extractNodeVersion(ref.Using)
			},
		},
	}
}

// extractNodeVersion extracts the numeric version string from a node runtime identifier.
// e.g., "node20" -> "20", "node16" -> "16", "composite" -> "".
func extractNodeVersion(using string) string {
	remainder, ok := strings.CutPrefix(using, "node")
	if !ok || remainder == "" {
		return ""
	}
	for _, c := range remainder {
		if c < '0' || c > '9' {
			return ""
		}
	}
	return remainder
}

func (g *WorkflowDependencyFieldGetters) GetField(ref *parser.ActionReference, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(ref)
	}
	return ""
}

// renderActionReferencesWithGetter renders a list of ActionReferences as a table using the given getter.
func (r *Renderer) renderActionReferencesWithGetter(refs []parser.ActionReference, headers []string, getter *WorkflowDependencyFieldGetters) {
	if len(refs) == 0 {
		r.writeLine("No action references.")
		return
	}
	if len(headers) == 0 {
		headers = []string{"Name", "Version"}
	}
	table := r.newTableWriter(headers)
	for i := range refs {
		row := make([]string, len(headers))
		for j, header := range headers {
			row[j] = getter.GetField(&refs[i], header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderActionReferences renders a list of ActionReferences as a table
func (r *Renderer) RenderActionReferences(refs []parser.ActionReference, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(refs)
		return
	}
	r.renderActionReferencesWithGetter(refs, headers, NewWorkflowDependencyFieldGetters())
}

// RenderWorkflowDependencies renders workflow dependencies grouped by source file
func (r *Renderer) RenderWorkflowDependencies(deps []parser.WorkflowDependency, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(deps)
		return
	}

	if len(deps) == 0 {
		r.writeLine("No workflow dependencies.")
		return
	}

	getter := NewWorkflowDependencyFieldGetters()

	for _, dep := range deps {
		r.writeLine(dep.Source)
		r.renderActionReferencesWithGetter(dep.Actions, headers, getter)
	}
}

// RenderWorkflowDependenciesWithFormat renders workflow dependencies in the specified format
func (r *Renderer) RenderWorkflowDependenciesWithFormat(format string, deps []parser.WorkflowDependency, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(deps)
		return
	}

	if format == "" {
		r.RenderWorkflowDependencies(deps, headers)
		return
	}

	switch strings.ToLower(format) {
	case "dot":
		r.RenderDotWorkflowDependencies(deps)
		return
	case "markdown":
		r.RenderMarkdownWorkflowDependencies(deps)
		return
	case "mermaid":
		r.RenderMermaidWorkflowDependencies(deps)
		return
	default:
		r.writeLine(fmt.Sprintf("Unsupported format: %s", format))
		return
	}
}

// RenderDotWorkflowDependencies renders workflow dependencies as a Graphviz DOT digraph
func (r *Renderer) RenderDotWorkflowDependencies(deps []parser.WorkflowDependency) {
	if r.exporter != nil {
		r.RenderExportedData(deps)
		return
	}

	r.writeLine("digraph {")

	// Build a set of dep sources for resolving action references
	depSources := make(map[string]bool)
	for _, dep := range deps {
		depSources[dep.Source] = true
	}
	hasSource := func(key string) bool {
		return depSources[key]
	}

	seen := make(map[string]bool)
	for _, dep := range deps {
		for _, action := range dep.Actions {
			var targetLabel string
			if resolved := parser.ResolveActionDepSource(action, hasSource); resolved != "" {
				targetLabel = resolved
			} else {
				targetLabel = action.Name()
			}
			edgeKey := dep.Source + "->" + targetLabel
			if seen[edgeKey] {
				continue
			}
			seen[edgeKey] = true
			r.writeLine(fmt.Sprintf("    %s -> %s",
				dotQuote(dep.Source),
				dotQuote(targetLabel),
			))
		}
	}
	r.writeLine("}")
}

// RenderMarkdownWorkflowDependencies renders workflow dependencies in Markdown format
// with an embedded Mermaid flowchart code block
func (r *Renderer) RenderMarkdownWorkflowDependencies(deps []parser.WorkflowDependency) {
	if r.exporter != nil {
		r.RenderExportedData(deps)
		return
	}
	r.writeLine("```mermaid")
	r.RenderMermaidWorkflowDependencies(deps)
	r.writeLine("```")
}

// RenderMermaidWorkflowDependencies renders workflow dependencies as a Mermaid flowchart
func (r *Renderer) RenderMermaidWorkflowDependencies(deps []parser.WorkflowDependency) {
	if r.exporter != nil {
		r.RenderExportedData(deps)
		return
	}

	r.writeLine("graph LR")

	// Build a set of dep sources for resolving action references
	// to their corresponding dep nodes in the graph.
	depSources := make(map[string]bool)
	for _, dep := range deps {
		depSources[dep.Source] = true
	}
	hasSource := func(key string) bool {
		return depSources[key]
	}

	seen := make(map[string]bool)
	for _, dep := range deps {
		sourceID := mermaidNodeID(dep.Source)
		for _, action := range dep.Actions {
			// Resolve action reference to its dep Source for proper graph connectivity.
			// For example, "./my-action" resolves to "my-action/action.yml", and
			// "owner/repo@v1" resolves to "owner/repo:action.yml".
			var targetID, targetLabel string
			if resolved := parser.ResolveActionDepSource(action, hasSource); resolved != "" {
				targetID = mermaidNodeID(resolved)
				targetLabel = resolved
			} else {
				targetName := action.Name()
				targetID = mermaidNodeID(targetName)
				targetLabel = targetName
			}
			edgeKey := sourceID + "-->" + targetID
			if seen[edgeKey] {
				continue
			}
			seen[edgeKey] = true
			r.writeLine(fmt.Sprintf("    %s[\"%s\"] --> %s[\"%s\"]",
				sourceID, dep.Source,
				targetID, targetLabel,
			))
		}
	}
}
