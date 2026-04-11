package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/ddddddO/gtree"
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
var WorkflowDependencyFields = []string{"Name", "Version", "Owner", "Repo", "Path", "Raw", "Using", "Node_Version", "Job"}

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
			"JOB": func(ref *parser.ActionReference) string {
				return ref.JobID
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
func (r *Renderer) renderActionReferencesWithGetter(refs []parser.ActionReference, headers []string, getter *WorkflowDependencyFieldGetters) error {
	if len(refs) == 0 {
		r.writeLine("No action references.")
		return nil
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
	return table.Render()
}

// RenderActionReferences renders a list of ActionReferences as a table
func (r *Renderer) RenderActionReferences(refs []parser.ActionReference, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(refs)
	}
	return r.renderActionReferencesWithGetter(refs, headers, NewWorkflowDependencyFieldGetters())
}

// RenderWorkflowDependencies renders workflow dependencies grouped by source file
func (r *Renderer) RenderWorkflowDependencies(deps []parser.WorkflowDependency, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(deps)
	}

	if len(deps) == 0 {
		r.writeLine("No workflow dependencies.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"Name", "Version"}
	}

	getter := NewWorkflowDependencyFieldGetters()

	var firstErr error
	for _, dep := range deps {
		r.writeLine(dep.Source)
		if err := r.renderActionReferencesWithGetter(dep.Actions, headers, getter); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			r.WriteError(err)
		}
	}
	return firstErr
}

// RenderWorkflowDependenciesWithFormat renders workflow dependencies in the specified format
func (r *Renderer) RenderWorkflowDependenciesWithFormat(format string, deps []parser.WorkflowDependency, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(deps)
	}

	if format == "" {
		return r.RenderWorkflowDependencies(deps, headers)
	}

	switch strings.ToLower(format) {
	case "dot":
		return r.RenderDotWorkflowDependencies(deps)
	case "drawio":
		return r.RenderDrawioWorkflowDependencies(deps)
	case "markdown":
		return r.RenderMarkdownWorkflowDependencies(deps)
	case "mermaid":
		return r.RenderMermaidWorkflowDependencies(deps)
	case "tree":
		return r.RenderTreeWorkflowDependencies(deps)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// actionTreeLabel returns a human-readable label for an action reference node.
// If the action has a Using field, it is appended in brackets (e.g. "actions/checkout@v4 [node20]").
func actionTreeLabel(action parser.ActionReference) string {
	label := action.VersionedName()
	if action.Using != "" {
		label += " [" + action.Using + "]"
	}
	return label
}

// addActionsToGtreeNode adds action references as children of node, recursing into known dep sources.
func addActionsToGtreeNode(node *gtree.Node, actions []parser.ActionReference, depBySource map[string]*parser.WorkflowDependency, hasSource func(string) bool, visited map[string]bool) {
	seen := make(map[string]bool)
	for _, action := range actions {
		label := actionTreeLabel(action)
		if seen[label] {
			continue
		}
		seen[label] = true

		child := node.Add(label)
		sourceKey := parser.ResolveActionDepSource(action, hasSource)
		if sourceKey != "" && !visited[sourceKey] {
			if childDep, ok := depBySource[sourceKey]; ok {
				visited[sourceKey] = true
				buildWorkflowDepGtreeNode(child, *childDep, depBySource, hasSource, visited)
			}
		}
	}
}

// buildWorkflowDepGtreeNode recursively adds action reference children to a gtree node.
// If the dep's actions have JobID set, actions are grouped under job sub-nodes.
// visited prevents infinite recursion in cyclic dependency graphs.
func buildWorkflowDepGtreeNode(node *gtree.Node, dep parser.WorkflowDependency, depBySource map[string]*parser.WorkflowDependency, hasSource func(string) bool, visited map[string]bool) {
	// Check if any action has a JobID to determine grouping mode
	hasJobIDs := false
	for _, action := range dep.Actions {
		if action.JobID != "" {
			hasJobIDs = true
			break
		}
	}

	if !hasJobIDs {
		addActionsToGtreeNode(node, dep.Actions, depBySource, hasSource, visited)
		return
	}

	// Group actions by JobID preserving document order
	var jobOrder []string
	jobActions := make(map[string][]parser.ActionReference)
	for _, action := range dep.Actions {
		jid := action.JobID
		if _, exists := jobActions[jid]; !exists {
			jobOrder = append(jobOrder, jid)
		}
		jobActions[jid] = append(jobActions[jid], action)
	}
	for _, jid := range jobOrder {
		jobNode := node.Add(jid)
		addActionsToGtreeNode(jobNode, jobActions[jid], depBySource, hasSource, visited)
	}
}

// RenderTreeWorkflowDependencies renders workflow dependencies as a text dependency tree
// using box-drawing characters (├──, └──, │). Root nodes are dependency sources that are
// not referenced by any other dependency. Recursive action deps are shown as subtrees.
func (r *Renderer) RenderTreeWorkflowDependencies(deps []parser.WorkflowDependency) error {
	if r.exporter != nil {
		return r.RenderExportedData(deps)
	}
	if len(deps) == 0 {
		r.writeLine("No workflow dependencies.")
		return nil
	}

	depBySource := make(map[string]*parser.WorkflowDependency)
	for i := range deps {
		depBySource[deps[i].Source] = &deps[i]
	}
	hasSource := func(key string) bool {
		_, ok := depBySource[key]
		return ok
	}

	// Identify sources referenced by other deps (non-root nodes)
	referenced := make(map[string]bool)
	for _, dep := range deps {
		for _, action := range dep.Actions {
			if key := parser.ResolveActionDepSource(action, hasSource); key != "" {
				referenced[key] = true
			}
		}
	}

	var firstErr error
	for _, dep := range deps {
		if referenced[dep.Source] {
			continue // non-root: will appear as a subtree under its parent
		}
		root := gtree.NewRoot(dep.Source)
		visited := make(map[string]bool)
		visited[dep.Source] = true
		buildWorkflowDepGtreeNode(root, dep, depBySource, hasSource, visited)
		if err := gtree.OutputFromRoot(r.IO.Out, root); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			r.WriteError(err)
		}
	}
	return firstErr
}

// RenderDotWorkflowDependencies renders workflow dependencies as a Graphviz DOT digraph.
// Node attributes include a tooltip with the runs.using value when available.
func (r *Renderer) RenderDotWorkflowDependencies(deps []parser.WorkflowDependency) error {
	if r.exporter != nil {
		return r.RenderExportedData(deps)
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

	// Collect using values per node label to emit node attribute declarations
	nodeUsing := make(map[string]string)
	var edgeLines []string
	seen := make(map[string]bool)
	for _, dep := range deps {
		for _, action := range dep.Actions {
			var targetLabel string
			if resolved := parser.ResolveActionDepSource(action, hasSource); resolved != "" {
				targetLabel = resolved
			} else {
				targetLabel = action.Name()
			}
			if action.Using != "" {
				if _, exists := nodeUsing[targetLabel]; !exists {
					nodeUsing[targetLabel] = action.Using
				}
			}
			edgeKey := dep.Source + "->" + targetLabel
			if seen[edgeKey] {
				continue
			}
			seen[edgeKey] = true
			edgeLines = append(edgeLines, fmt.Sprintf("    %s -> %s",
				dotQuote(dep.Source),
				dotQuote(targetLabel),
			))
		}
	}

	// Emit node attribute declarations before edges, sorted for deterministic output.
	nodeUsingKeys := make([]string, 0, len(nodeUsing))
	for label := range nodeUsing {
		nodeUsingKeys = append(nodeUsingKeys, label)
	}
	sort.Strings(nodeUsingKeys)
	for _, label := range nodeUsingKeys {
		r.writeLine(fmt.Sprintf("    %s [tooltip=%s]",
			dotQuote(label),
			dotQuote(nodeUsing[label]),
		))
	}
	for _, line := range edgeLines {
		r.writeLine(line)
	}
	r.writeLine("}")
	return nil
}

// RenderDrawioWorkflowDependencies renders workflow dependencies as a draw.io XML document
func (r *Renderer) RenderDrawioWorkflowDependencies(deps []parser.WorkflowDependency) error {
	if r.exporter != nil {
		return r.RenderExportedData(deps)
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
	nodeTooltips := make(map[string]string)
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

			// Assign tooltip from using value
			if action.Using != "" {
				if _, ok := nodeTooltips[targetLabel]; !ok {
					nodeTooltips[targetLabel] = action.Using
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
	return r.writeDrawioGraph(edges, nodeURLs, nodeColors, nodeTooltips)
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
func (r *Renderer) RenderMarkdownWorkflowDependencies(deps []parser.WorkflowDependency) error {
	if r.exporter != nil {
		return r.RenderExportedData(deps)
	}
	r.writeLine("```mermaid")
	err := r.RenderMermaidWorkflowDependencies(deps)
	r.writeLine("```")
	return err
}

// RenderMermaidWorkflowDependencies renders workflow dependencies as a Mermaid flowchart.
// When a node has a runs.using value, it is shown as a second line in the node label.
func (r *Renderer) RenderMermaidWorkflowDependencies(deps []parser.WorkflowDependency) error {
	if r.exporter != nil {
		return r.RenderExportedData(deps)
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

	// Collect node labels and using values to emit node definitions before edges
	nodeLabel := make(map[string]string) // id -> display label
	nodeUsing := make(map[string]string) // id -> using value
	type edge struct{ src, dst string }
	var edges []edge
	seen := make(map[string]bool)

	for _, dep := range deps {
		sourceID := mermaidNodeID(dep.Source)
		if _, exists := nodeLabel[sourceID]; !exists {
			nodeLabel[sourceID] = dep.Source
		}
		for _, action := range dep.Actions {
			var targetID, targetName string
			if resolved := parser.ResolveActionDepSource(action, hasSource); resolved != "" {
				targetID = mermaidNodeID(resolved)
				targetName = resolved
			} else {
				targetName = action.Name()
				targetID = mermaidNodeID(targetName)
			}
			if _, exists := nodeLabel[targetID]; !exists {
				nodeLabel[targetID] = targetName
			}
			if action.Using != "" {
				if _, exists := nodeUsing[targetID]; !exists {
					nodeUsing[targetID] = action.Using
				}
			}
			edgeKey := sourceID + "-->" + targetID
			if seen[edgeKey] {
				continue
			}
			seen[edgeKey] = true
			edges = append(edges, edge{sourceID, targetID})
		}
	}

	// Emit node definitions with using as a second label line when available
	emittedNodes := make(map[string]bool)
	emitNode := func(id string) {
		if emittedNodes[id] {
			return
		}
		emittedNodes[id] = true
		label := nodeLabel[id]
		if u := nodeUsing[id]; u != "" {
			label += "<br/>" + u
		}
		r.writeLine(fmt.Sprintf("    %s[\"%s\"]", id, label))
	}

	for _, e := range edges {
		emitNode(e.src)
		emitNode(e.dst)
		r.writeLine(fmt.Sprintf("    %s --> %s", e.src, e.dst))
	}
	return nil
}
