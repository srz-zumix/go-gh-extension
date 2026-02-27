package parser

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ActionReference represents a parsed reference to a GitHub Action or reusable workflow
type ActionReference struct {
	Raw     string `json:"raw" yaml:"raw"`         // Original uses string, e.g. "actions/checkout@v4"
	Owner   string `json:"owner" yaml:"owner"`     // e.g. "actions"
	Repo    string `json:"repo" yaml:"repo"`       // e.g. "checkout"
	Path    string `json:"path" yaml:"path"`       // Subdirectory path, e.g. "subdir" for "owner/repo/subdir@v1"
	Ref     string `json:"ref" yaml:"ref"`         // Version/ref, e.g. "v4"
	IsLocal bool   `json:"isLocal" yaml:"isLocal"` // true for "./local-action" references
}

// Name returns a human-readable name for the action reference
func (a ActionReference) Name() string {
	if a.IsLocal {
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
	Uses string `yaml:"uses"`
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

// ParseWorkflowYAML parses a GitHub Actions workflow YAML content and extracts action references
func ParseWorkflowYAML(content []byte) (string, []ActionReference, error) {
	var wf workflowYAML
	if err := yaml.Unmarshal(content, &wf); err != nil {
		return "", nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	var refs []ActionReference
	for _, job := range wf.Jobs {
		// Reusable workflow reference (jobs.<job_id>.uses)
		if job.Uses != "" {
			ref := ParseActionReference(job.Uses)
			if ref.Raw != "" {
				refs = append(refs, ref)
			}
		}
		// Step-level action references (jobs.<job_id>.steps[*].uses)
		for _, step := range job.Steps {
			if step.Uses != "" {
				ref := ParseActionReference(step.Uses)
				if ref.Raw != "" {
					refs = append(refs, ref)
				}
			}
		}
	}
	return wf.Name, refs, nil
}

// ParseActionYAML parses an action.yml/action.yaml content and extracts action references
func ParseActionYAML(content []byte) ([]ActionReference, error) {
	var action actionYAML
	if err := yaml.Unmarshal(content, &action); err != nil {
		return nil, fmt.Errorf("failed to parse action YAML: %w", err)
	}

	// Only composite actions have steps
	if action.Runs.Using != "composite" {
		return nil, nil
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
	return refs, nil
}
