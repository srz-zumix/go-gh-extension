package parser

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ActionReference represents a parsed reference to a GitHub Action or reusable workflow
type ActionReference struct {
	Raw     string `json:"raw" yaml:"raw"`                       // Original uses string, e.g. "actions/checkout@v4"
	Owner   string `json:"owner" yaml:"owner"`                   // e.g. "actions"
	Repo    string `json:"repo" yaml:"repo"`                     // e.g. "checkout"
	Path    string `json:"path" yaml:"path"`                     // Subdirectory path, e.g. "subdir" for "owner/repo/subdir@v1"
	Ref     string `json:"ref" yaml:"ref"`                       // Version/ref, e.g. "v4"
	IsLocal bool   `json:"isLocal" yaml:"isLocal"`               // true for "./local-action" references
	Using   string `json:"using,omitempty" yaml:"using,omitempty"` // runs.using value of the referenced action, e.g. "node20", "composite"
}

// Name returns a human-readable name for the action reference
func (a ActionReference) Name() string {
	if a.IsLocal {
		return a.Raw
	}
	if a.Owner == "" {
		return a.Raw
	}
	name := a.Owner + "/" + a.Repo
	if a.Path != "" {
		name += "/" + a.Path
	}
	return name
}

// VersionedName returns name@ref string
func (a ActionReference) VersionedName() string {
	name := a.Name()
	if a.Ref != "" {
		name += "@" + a.Ref
	}
	return name
}

// IsReusableWorkflow returns true if the reference points to a reusable workflow file
// (i.e. the path ends with .yml or .yaml)
func (a ActionReference) IsReusableWorkflow() bool {
	if a.IsLocal {
		raw := strings.TrimPrefix(a.Raw, "./")
		return strings.HasSuffix(raw, ".yml") || strings.HasSuffix(raw, ".yaml")
	}
	return strings.HasSuffix(a.Path, ".yml") || strings.HasSuffix(a.Path, ".yaml")
}

// LocalPath returns the resolved file path for a local reference (strips leading "./")
func (a ActionReference) LocalPath() string {
	if !a.IsLocal {
		return ""
	}
	return strings.TrimPrefix(a.Raw, "./")
}

// ResolveActionDepSource resolves an action reference to a matching dependency source key.
// The hasSource function checks if a given source key exists in the set of known dep sources.
// Returns the matching source key, or empty string if no match was found.
//
// This maps action references (e.g., "./my-action", "owner/repo@v1") to their
// corresponding WorkflowDependency.Source (e.g., "my-action/action.yml", "owner/repo:action.yml").
func ResolveActionDepSource(action ActionReference, hasSource func(string) bool) string {
	if action.IsLocal && action.IsReusableWorkflow() {
		// Local reusable workflow (e.g. "./.github/workflows/release.yml" -> ".github/workflows/release.yml")
		localPath := action.LocalPath()
		if hasSource(localPath) {
			return localPath
		}
	} else if action.IsLocal && !action.IsReusableWorkflow() {
		if action.Owner != "" && action.Repo != "" {
			// Checkout-resolved local action (e.g. "./tools/lint" resolved to "owner/repo:lint/action.yml")
			for _, fname := range []string{"action.yml", "action.yaml"} {
				actionFile := fname
				if action.Path != "" {
					actionFile = action.Path + "/" + fname
				}
				key := fmt.Sprintf("%s/%s:%s", action.Owner, action.Repo, actionFile)
				if hasSource(key) {
					return key
				}
			}
		} else {
			// Local action in current repo (e.g. "./my-action" -> "my-action/action.yml")
			localPath := action.LocalPath()
			for _, fname := range []string{"action.yml", "action.yaml"} {
				key := localPath + "/" + fname
				if hasSource(key) {
					return key
				}
			}
		}
	} else if action.Owner != "" && action.Repo != "" {
		if action.IsReusableWorkflow() {
			// Remote reusable workflow (e.g. "owner/repo/.github/workflows/w.yml@ref")
			key := fmt.Sprintf("%s/%s:%s", action.Owner, action.Repo, action.Path)
			if hasSource(key) {
				return key
			}
		} else {
			// Remote action repository (e.g. "owner/repo@v1" -> "owner/repo:action.yml")
			// Also supports subdirectory actions (e.g. "owner/repo/path@v1" -> "owner/repo:path/action.yml")
			for _, fname := range []string{"action.yml", "action.yaml"} {
				actionFile := fname
				if action.Path != "" {
					actionFile = action.Path + "/" + fname
				}
				key := fmt.Sprintf("%s/%s:%s", action.Owner, action.Repo, actionFile)
				if hasSource(key) {
					return key
				}
			}
		}
	}
	return ""
}

// PopulateActionUsing sets the Using field on each ActionReference in deps
// by looking up the corresponding source key in usingBySource.
// usingBySource maps dep source keys (e.g. "actions/checkout:action.yml") to
// runs.using values (e.g. "node20", "composite").
func PopulateActionUsing(deps []WorkflowDependency, usingBySource map[string]string) {
	hasSource := func(key string) bool {
		_, ok := usingBySource[key]
		return ok
	}
	for i := range deps {
		for j := range deps[i].Actions {
			action := &deps[i].Actions[j]
			sourceKey := ResolveActionDepSource(*action, hasSource)
			if sourceKey != "" {
				action.Using = usingBySource[sourceKey]
			}
		}
	}
}

// WorkflowDependency represents dependencies found in a single workflow or action file
type WorkflowDependency struct {
	Source  string            `json:"source"`  // File path, e.g. ".github/workflows/ci.yml"
	Name    string            `json:"name"`    // Workflow name from the YAML name field
	Actions []ActionReference `json:"actions"` // Action references found in the file
}

// workflowYAML represents the structure of a GitHub Actions workflow YAML file
type workflowYAML struct {
	Name string                 `yaml:"name"`
	Jobs map[string]workflowJob `yaml:"jobs"`
}

type workflowJob struct {
	Uses  string         `yaml:"uses"`
	Steps []workflowStep `yaml:"steps"`
}

type workflowStep struct {
	Uses string         `yaml:"uses"`
	With map[string]any `yaml:"with"`
}

// CheckoutPath represents a mapping of a checkout destination path to a repository.
// This is used to resolve local action references (e.g. "./my-tool") that depend on
// a repository checked out to a specific path via actions/checkout.
type CheckoutPath struct {
	Repository string `json:"repository" yaml:"repository"`             // e.g. "owner/repo"
	Path       string `json:"path" yaml:"path"`                         // checkout destination path relative to workspace root
	Ref        string `json:"ref,omitempty" yaml:"ref,omitempty"` // git ref for checkout
}

// isCheckoutAction returns true if the uses string refers to actions/checkout
func isCheckoutAction(uses string) bool {
	ref := ParseActionReference(uses)
	return ref.Owner == "actions" && ref.Repo == "checkout"
}

// getWithString extracts a string value from a step's "with" map
func getWithString(with map[string]any, key string) string {
	if with == nil {
		return ""
	}
	v, ok := with[key]
	if !ok {
		return ""
	}
	if v == nil {
		return ""
	}

	switch vv := v.(type) {
	case string:
		return vv
	case *yaml.Node:
		// Only accept scalar nodes as string values
		if vv.Kind == yaml.ScalarNode {
			return vv.Value
		}
		return ""
	case yaml.Node:
		// Handle non-pointer yaml.Node as well
		if vv.Kind == yaml.ScalarNode {
			return vv.Value
		}
		return ""
	default:
		// Non-string types are not treated as valid string values
		return ""
	}
}

// resolveLocalActionByCheckout resolves a local action reference using checkout path mappings.
// If the local action path matches a checkout path prefix, it sets the Owner, Repo, Path, and Ref
// fields on the ActionReference based on the checkout mapping.
func resolveLocalActionByCheckout(ref *ActionReference, checkouts []CheckoutPath) {
	if !ref.IsLocal || ref.IsReusableWorkflow() {
		return
	}
	localPath := ref.LocalPath()
	var bestMatch *CheckoutPath
	for i := range checkouts {
		cp := &checkouts[i]
		if localPath == cp.Path || strings.HasPrefix(localPath, cp.Path+"/") {
			if bestMatch == nil || len(cp.Path) > len(bestMatch.Path) {
				bestMatch = cp
			}
		}
	}
	if bestMatch == nil {
		return
	}
	// Empty Repository means the checkout is for the current (self) repository.
	// No resolution needed since the local action path already refers to the current repo.
	if bestMatch.Repository == "" {
		return
	}
	parts := strings.SplitN(bestMatch.Repository, "/", 2)
	if len(parts) != 2 {
		return
	}
	ref.Owner = parts[0]
	ref.Repo = parts[1]
	if localPath != bestMatch.Path {
		ref.Path = strings.TrimPrefix(localPath, bestMatch.Path+"/")
	}
	if bestMatch.Ref != "" {
		ref.Ref = bestMatch.Ref
	}
}

// actionYAML represents the structure of an action.yml/action.yaml file
type actionYAML struct {
	Runs actionRuns `yaml:"runs"`
}

type actionRuns struct {
	Using string         `yaml:"using"`
	Steps []workflowStep `yaml:"steps"`
}

// ParseActionReference parses a "uses" string into an ActionReference
func ParseActionReference(uses string) ActionReference {
	uses = strings.TrimSpace(uses)
	if uses == "" {
		return ActionReference{}
	}

	// Local action reference (e.g. "./path/to/action")
	if strings.HasPrefix(uses, "./") {
		return ActionReference{
			Raw:     uses,
			IsLocal: true,
		}
	}

	// Docker reference (e.g. "docker://image:tag")
	if strings.HasPrefix(uses, "docker://") {
		return ActionReference{
			Raw:     uses,
			IsLocal: false,
		}
	}

	ref := ActionReference{Raw: uses}

	// Split by @ to separate ref/version
	atParts := strings.SplitN(uses, "@", 2)
	if len(atParts) == 2 {
		ref.Ref = atParts[1]
	}
	actionPath := atParts[0]

	// Split by / to get owner, repo, and optional path
	parts := strings.SplitN(actionPath, "/", 3)
	if len(parts) >= 2 {
		ref.Owner = parts[0]
		ref.Repo = parts[1]
	}
	if len(parts) == 3 {
		ref.Path = parts[2]
	}

	return ref
}

// ParseWorkflowName extracts the workflow name from a workflow YAML content
func ParseWorkflowName(content []byte) string {
	var wf workflowYAML
	if err := yaml.Unmarshal(content, &wf); err != nil {
		return ""
	}
	return wf.Name
}

// ParseWorkflowYAML parses a GitHub Actions workflow YAML content and extracts action references.
// Jobs and steps are processed in YAML document order to produce deterministic output.
func ParseWorkflowYAML(content []byte) (string, []ActionReference, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return "", nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	// doc is a Document node; the actual mapping is its first child
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return "", nil, fmt.Errorf("unexpected YAML structure")
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return "", nil, fmt.Errorf("expected YAML mapping at root")
	}

	var name string
	var jobsNode *yaml.Node

	// Iterate root mapping keys in document order
	for i := 0; i+1 < len(root.Content); i += 2 {
		key := root.Content[i].Value
		val := root.Content[i+1]
		switch key {
		case "name":
			name = val.Value
		case "jobs":
			jobsNode = val
		}
	}

	if jobsNode == nil || jobsNode.Kind != yaml.MappingNode {
		return name, nil, nil
	}

	var refs []ActionReference
	// Iterate jobs in document order (mapping key/value pairs)
	for i := 0; i+1 < len(jobsNode.Content); i += 2 {
		jobNode := jobsNode.Content[i+1]
		if jobNode.Kind != yaml.MappingNode {
			continue
		}

		// Decode the job into a struct for convenient access
		var job workflowJob
		if err := jobNode.Decode(&job); err != nil {
			continue
		}

		// Reusable workflow reference (jobs.<job_id>.uses)
		if job.Uses != "" {
			ref := ParseActionReference(job.Uses)
			if ref.Raw != "" {
				refs = append(refs, ref)
			}
		}
		// Collect checkout path mappings and step-level action references per job.
		// Checkout mappings are built incrementally so that only checkouts preceding
		// a local action step are considered for resolution.
		var checkouts []CheckoutPath
		for _, step := range job.Steps {
			if step.Uses == "" {
				continue
			}
			// Track checkout steps with path specified
			if isCheckoutAction(step.Uses) {
				repo := getWithString(step.With, "repository")
				path := getWithString(step.With, "path")
				if path != "" {
					checkouts = append(checkouts, CheckoutPath{
						Repository: repo, // empty means self repository
						Path:       path,
						Ref:        getWithString(step.With, "ref"),
					})
				}
			}
			ref := ParseActionReference(step.Uses)
			if ref.Raw != "" {
				// Resolve local non-reusable-workflow action references using checkout mappings
				if ref.IsLocal && !ref.IsReusableWorkflow() {
					resolveLocalActionByCheckout(&ref, checkouts)
				}
				refs = append(refs, ref)
			}
		}
	}
	return name, refs, nil
}

// ParseActionYAML parses an action.yml/action.yaml content and extracts action references.
// It returns the list of action references (non-nil only for composite actions),
// the runs.using value (e.g. "node20", "composite", "docker"), and any parse error.
func ParseActionYAML(content []byte) ([]ActionReference, string, error) {
	var action actionYAML
	if err := yaml.Unmarshal(content, &action); err != nil {
		return nil, "", fmt.Errorf("failed to parse action YAML: %w", err)
	}

	// Only composite actions have steps that can reference other actions
	if action.Runs.Using != "composite" {
		return nil, action.Runs.Using, nil
	}

	var refs []ActionReference
	for _, step := range action.Runs.Steps {
		if step.Uses != "" {
			ref := ParseActionReference(step.Uses)
			if ref.Raw != "" {
				refs = append(refs, ref)
			}
		}
	}
	return refs, action.Runs.Using, nil
}
