package render

import (
	"fmt"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// WorkflowDependencyFieldGetter defines a function to get a field value from parser.ActionReference
type WorkflowDependencyFieldGetter func(ref *parser.ActionReference) string
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
		},
	}
}

func (g *WorkflowDependencyFieldGetters) GetField(ref *parser.ActionReference, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(ref)
	}
	return ""
}

// RenderActionReferences renders a list of ActionReferences as a table
func (r *Renderer) RenderActionReferences(refs []parser.ActionReference, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(refs)
		return
	}
	if len(refs) == 0 {
		r.writeLine("No action references.")
		return
	}

	if len(headers) == 0 {
		headers = []string{"Name", "Version"}
	}
	getter := NewWorkflowDependencyFieldGetters()
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

	for _, dep := range deps {
		r.writeLine(dep.Source)
		r.RenderActionReferences(dep.Actions, headers)
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

// RenderMarkdownWorkflowDependencies renders workflow dependencies in Markdown format
// with an embedded Mermaid flowchart code block
func (r *Renderer) RenderMarkdownWorkflowDependencies(deps []parser.WorkflowDependency) {
	r.writeLine("```mermaid")
	r.RenderMermaidWorkflowDependencies(deps)
	r.writeLine("```")
}

// RenderMermaidWorkflowDependencies renders workflow dependencies as a Mermaid flowchart
func (r *Renderer) RenderMermaidWorkflowDependencies(deps []parser.WorkflowDependency) {
	r.writeLine("graph LR")
	seen := make(map[string]bool)
	for _, dep := range deps {
		sourceID := mermaidNodeID(dep.Source)
		sourceLabel := dep.Source
		if dep.Name != "" {
			sourceLabel = dep.Name
		}
		for _, action := range dep.Actions {
			targetName := action.Name()
			targetID := mermaidNodeID(targetName)
			edgeKey := sourceID + "-->" + targetID
			if seen[edgeKey] {
				continue
			}
			seen[edgeKey] = true
			r.writeLine(fmt.Sprintf("    %s[\"%s\"] --> %s[\"%s\"]",
				sourceID, sourceLabel,
				targetID, targetName,
			))
		}
	}
}
