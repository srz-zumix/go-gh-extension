package render

import (
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// Node border color constants for draw.io rendering by action type
const (
	nodeColorReusableWorkflow = "#2196F3" // Blue: reusable workflow
	nodeColorComposite        = "#4CAF50" // Green: composite action
	nodeColorNode             = "#FF9800" // Orange: node runtime action
	nodeColorDocker           = "#9C27B0" // Purple: docker action
)

// WorkflowDependencyFields lists all available field names for ActionReference table rendering.
var WorkflowDependencyFields = []string{"Name", "Version", "Owner", "Repo", "Path", "Raw", "Using", "Node_Version"}

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
	case "drawio":
		r.RenderDrawioWorkflowDependencies(deps)
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

// RenderDrawioWorkflowDependencies renders workflow dependencies as a draw.io XML document
func (r *Renderer) RenderDrawioWorkflowDependencies(deps []parser.WorkflowDependency) {
	if r.exporter != nil {
		r.RenderExportedData(deps)
		return
	}

	// Build a set of dep sources for resolving action references
	depSources := make(map[string]bool)
	for _, dep := range deps {
		depSources[dep.Source] = true
	}
	hasSource := func(key string) bool {
		return depSources[key]
	}

	// Build dep-source-to-repository map for resolved target URL lookup
	depRepoBySource := make(map[string]repository.Repository)
	for _, dep := range deps {
		depRepoBySource[dep.Source] = dep.Repository
	}

	// Build dep-source-to-ref map by finding action references that resolve to dep sources.
	// This allows us to construct URLs with the correct git ref instead of HEAD.
	depRefBySource := make(map[string]string)
	for _, dep := range deps {
		for _, action := range dep.Actions {
			if resolved := parser.ResolveActionDepSource(action, hasSource); resolved != "" {
				if _, ok := depRefBySource[resolved]; !ok && action.Ref != "" {
					depRefBySource[resolved] = action.Ref
				}
			}
		}
	}

	nodeURLs := make(map[string]string)
	nodeColors := make(map[string]string)
	var edges [][2]string
	seen := make(map[string]bool)
	for _, dep := range deps {
		// Build URL for the source node
		if _, ok := nodeURLs[dep.Source]; !ok {
			nodeURLs[dep.Source] = workflowDepSourceURL(dep.Source, dep.Repository.Host, depRefBySource[dep.Source], dep.Repository.Owner, dep.Repository.Name)
		}

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
			edges = append(edges, [2]string{dep.Source, targetLabel})

			// Assign border color based on action type
			if _, ok := nodeColors[targetLabel]; !ok {
				if c := actionNodeColor(action); c != "" {
					nodeColors[targetLabel] = c
				}
			}

			// Build URL for the target node
			if _, ok := nodeURLs[targetLabel]; !ok {
				// If the target is a known dep source, use its repository; otherwise use the action's host
				if r, ok := depRepoBySource[targetLabel]; ok {
					nodeURLs[targetLabel] = workflowDepSourceURL(targetLabel, r.Host, depRefBySource[targetLabel], r.Owner, r.Name)
				} else {
					u := actionReferenceURL(action)
					// Local action not resolved to a dep source: construct URL from parent dep context
					if u == "" && action.IsLocal && dep.Repository.Host != "" && dep.Repository.Owner != "" && dep.Repository.Name != "" {
						localPath := action.LocalPath()
						u = "https://" + dep.Repository.Host + "/" + dep.Repository.Owner + "/" + dep.Repository.Name + "/blob/HEAD/" + localPath
					}
					nodeURLs[targetLabel] = u
				}
			}
		}
	}
	r.writeDrawioGraph(edges, nodeURLs, nodeColors)
}

// actionNodeColor returns a draw.io border color hex string based on the action reference type.
// Returns empty string for unknown or unresolved types (default border).
func actionNodeColor(action parser.ActionReference) string {
	if action.IsReusableWorkflow() {
		return nodeColorReusableWorkflow
	}
	using := strings.ToLower(action.Using)
	switch {
	case using == "composite":
		return nodeColorComposite
	case strings.HasPrefix(using, "node"):
		return nodeColorNode
	case using == "docker":
		return nodeColorDocker
	default:
		return ""
	}
}

// workflowDepSourceURL constructs a URL for a workflow/action dependency source label.
// Labels containing ":" are remote references (e.g. "owner/repo:path"),
// and labels without ":" are local file paths.
// ref is the git reference used to fetch the file (e.g. "v1", a SHA). If empty, "HEAD" is used.
// owner and repo provide the repository context for local sources.
func workflowDepSourceURL(source string, host string, ref string, owner string, repo string) string {
	if host == "" {
		return ""
	}
	idx := strings.Index(source, ":")
	if idx > 0 {
		// Remote: "owner/repo:path"
		ownerRepo := source[:idx]
		path := source[idx+1:]
		if strings.Contains(ownerRepo, "/") {
			if ref == "" {
				ref = "HEAD"
			}
			return "https://" + host + "/" + ownerRepo + "/blob/" + ref + "/" + path
		}
	}
	// Local: use provided owner/repo
	if owner != "" && repo != "" {
		if ref == "" {
			ref = "HEAD"
		}
		return "https://" + host + "/" + owner + "/" + repo + "/blob/" + ref + "/" + source
	}
	return ""
}

// actionReferenceURL constructs a URL from an ActionReference's structured fields and Host.
func actionReferenceURL(action parser.ActionReference) string {
	host := action.Host
	if host == "" {
		return ""
	}
	if action.IsLocal {
		return ""
	}
	if action.Owner == "" || action.Repo == "" {
		return ""
	}
	// Remote action: owner/repo with optional path
	u := "https://" + host + "/" + action.Owner + "/" + action.Repo
	if action.Path != "" {
		ref := action.Ref
		if ref == "" {
			ref = "HEAD"
		}
		u += "/blob/" + ref + "/" + action.Path
	}
	return u
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
