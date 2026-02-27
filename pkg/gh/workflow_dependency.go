package gh

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

const (
	workflowsDir = ".github/workflows"
)

// ResolveWorkflowFilePath resolves a selector (workflow ID, name, or filename) to a workflow file path.
// The selector can be:
//   - A numeric workflow ID (resolved via GitHub API)
//   - A filename (with or without .github/workflows/ prefix)
//   - A workflow name (resolved via ListWorkflows API)
func ResolveWorkflowFilePath(ctx context.Context, g *GitHubClient, repo repository.Repository, selector string) (string, error) {
	if selector == "" {
		return "", fmt.Errorf("selector is empty")
	}

	// Try to resolve by workflow ID (numeric)
	if id, err := strconv.ParseInt(selector, 10, 64); err == nil {
		wf, wfErr := g.GetWorkflowByID(ctx, repo.Owner, repo.Name, id)
		if wfErr == nil && wf != nil && wf.Path != nil {
			return *wf.Path, nil
		}
	}

	// Try as a filename (.yml or .yaml extension)
	if strings.HasSuffix(selector, ".yml") || strings.HasSuffix(selector, ".yaml") {
		if !strings.Contains(selector, "/") {
			// Basename only - prepend workflows directory
			return workflowsDir + "/" + selector, nil
		}
		return selector, nil
	}

	// Try to resolve by workflow name via ListWorkflows API
	workflows, err := g.ListWorkflows(ctx, repo.Owner, repo.Name)
	if err == nil {
		for _, wf := range workflows {
			if wf.Name != nil && strings.EqualFold(*wf.Name, selector) {
				if wf.Path != nil {
					return *wf.Path, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no workflow found matching selector: %s", selector)
}

// GetWorkflowFileDependency fetches and parses a single workflow file to extract action dependencies.
// If recursive is true, it also traverses referenced action repositories and reusable workflows.
func GetWorkflowFileDependency(ctx context.Context, g *GitHubClient, repo repository.Repository, filePath string, ref *string, recursive bool) ([]parser.WorkflowDependency, error) {
	content, err := getFileContent(ctx, g, repo, filePath, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content for %s: %w", filePath, err)
	}

	name, refs, err := parser.ParseWorkflowYAML(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow file %s: %w", filePath, err)
	}

	dep := parser.WorkflowDependency{
		Source:  filePath,
		Name:    name,
		Actions: refs,
	}
	deps := []parser.WorkflowDependency{dep}

	if recursive {
		visited := make(map[string]bool)
		visited[parser.GetRepositoryFullName(repo)] = true

		visitedFiles := make(map[string]bool)
		visitedFiles[filePath] = true

		deps = traverseDependencyActions(ctx, g, repo, ref, deps, visited, visitedFiles)
	}
	return deps, nil
}

// GetRepositoryWorkflowDependencies retrieves and parses workflow files from a repository
// to extract action/reusable workflow dependencies.
// If recursive is true, it also traverses referenced action repositories and reusable workflows.
func GetRepositoryWorkflowDependencies(ctx context.Context, g *GitHubClient, repo repository.Repository, ref *string, recursive bool) ([]parser.WorkflowDependency, error) {
	var deps []parser.WorkflowDependency

	// Fetch workflow files from .github/workflows/
	workflowDeps, err := getWorkflowFileDependencies(ctx, g, repo, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow file dependencies for %s: %w", parser.GetRepositoryFullName(repo), err)
	}
	deps = append(deps, workflowDeps...)

	// Fetch action.yml or action.yaml if present
	actionDep, err := getActionFileDependencies(ctx, g, repo, ref)
	if err != nil {
		// action.yml/action.yaml not found is not an error
	} else if actionDep != nil {
		deps = append(deps, *actionDep)
	}

	if recursive {
		visited := make(map[string]bool)
		visited[parser.GetRepositoryFullName(repo)] = true

		visitedFiles := make(map[string]bool)
		for _, dep := range deps {
			visitedFiles[dep.Source] = true
		}

		deps = traverseDependencyActions(ctx, g, repo, ref, deps, visited, visitedFiles)
	}

	return deps, nil
}

// getChildRepositoryDependencies fetches action.yml/action.yaml from a child repository
// and recursively traverses its dependencies. Workflow files in child repos are not fetched
// because they represent the child repo's own CI, not action dependencies.
func getChildRepositoryDependencies(ctx context.Context, g *GitHubClient, repo repository.Repository, visited map[string]bool) ([]parser.WorkflowDependency, error) {
	repoKey := parser.GetRepositoryFullName(repo)
	if visited[repoKey] {
		return nil, nil
	}
	visited[repoKey] = true

	var deps []parser.WorkflowDependency
	actionDep, err := getActionFileDependencies(ctx, g, repo, nil)
	if err == nil && actionDep != nil {
		actionDep.Source = repoKey + ":" + actionDep.Source
		deps = append(deps, *actionDep)
	}

	visitedFiles := make(map[string]bool)
	for _, dep := range deps {
		visitedFiles[dep.Source] = true
	}

	deps = traverseDependencyActions(ctx, g, repo, nil, deps, visited, visitedFiles)
	return deps, nil
}

// traverseDependencyActions traverses action references in deps and recursively fetches
// dependencies from local reusable workflows, remote reusable workflows, and action repositories.
func traverseDependencyActions(ctx context.Context, g *GitHubClient, repo repository.Repository, ref *string, deps []parser.WorkflowDependency, visited map[string]bool, visitedFiles map[string]bool) []parser.WorkflowDependency {
	repoKey := parser.GetRepositoryFullName(repo)
	initialLen := len(deps)
	for i := range initialLen {
		for _, action := range deps[i].Actions {
			// Handle local reusable workflows (e.g. "./.github/workflows/release.yml")
			if action.IsLocal && action.IsReusableWorkflow() {
				localPath := action.LocalPath()
				fileKey := repoKey + ":" + localPath
				if visitedFiles[localPath] || visitedFiles[fileKey] {
					continue
				}
				visitedFiles[localPath] = true
				localDeps, localErr := getReusableWorkflowDependencies(ctx, g, repo, localPath, ref)
				if localErr != nil {
					continue
				}
				deps = append(deps, localDeps...)
				continue
			}

			if action.Owner == "" || action.Repo == "" {
				continue
			}

			// Handle remote reusable workflows (e.g. "owner/repo/.github/workflows/workflow.yml@ref")
			if action.IsReusableWorkflow() {
				childRepo := repository.Repository{
					Host:  repo.Host,
					Owner: action.Owner,
					Name:  action.Repo,
				}
				fileKey := parser.GetRepositoryFullName(childRepo) + ":" + action.Path
				if visitedFiles[fileKey] {
					continue
				}
				visitedFiles[fileKey] = true
				rwDeps, rwErr := getReusableWorkflowDependencies(ctx, g, childRepo, action.Path, nil)
				if rwErr != nil {
					continue
				}
				// Prefix source with repo key for remote workflows
				childKey := parser.GetRepositoryFullName(childRepo)
				for j := range rwDeps {
					rwDeps[j].Source = childKey + ":" + rwDeps[j].Source
				}
				deps = append(deps, rwDeps...)
				continue
			}

			// Handle action repositories (traverse action.yml only)
			childRepo := repository.Repository{
				Host:  repo.Host,
				Owner: action.Owner,
				Name:  action.Repo,
			}
			childKey := parser.GetRepositoryFullName(childRepo)
			if visited[childKey] {
				continue
			}
			// Traverse referenced repository (only action.yml for child repos)
			childDeps, err := getChildRepositoryDependencies(ctx, g, childRepo, visited)
			if err != nil {
				// Skip repos that are not accessible
				continue
			}
			deps = append(deps, childDeps...)
		}
	}

	return deps
}

// getReusableWorkflowDependencies fetches and parses a single reusable workflow file
func getReusableWorkflowDependencies(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, ref *string) ([]parser.WorkflowDependency, error) {
	content, err := getFileContent(ctx, g, repo, path, ref)
	if err != nil {
		return nil, err
	}

	name, refs, err := parser.ParseWorkflowYAML(content)
	if err != nil {
		return nil, err
	}

	if len(refs) == 0 {
		return nil, nil
	}

	return []parser.WorkflowDependency{
		{
			Source:  path,
			Name:    name,
			Actions: refs,
		},
	}, nil
}

// getWorkflowFileDependencies fetches all workflow YAML files from .github/workflows/ and parses them
func getWorkflowFileDependencies(ctx context.Context, g *GitHubClient, repo repository.Repository, ref *string) ([]parser.WorkflowDependency, error) {
	_, dirContent, err := g.GetRepositoryContent(ctx, repo.Owner, repo.Name, workflowsDir, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow directory: %w", err)
	}

	var deps []parser.WorkflowDependency
	for _, entry := range dirContent {
		if entry.Name == nil || entry.Type == nil {
			continue
		}
		name := *entry.Name
		// Only process YAML files
		if *entry.Type != "file" || (!strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml")) {
			continue
		}

		filePath := workflowsDir + "/" + name
		content, err := getFileContent(ctx, g, repo, filePath, ref)
		if err != nil {
			continue // skip files that cannot be read
		}

		name, refs, err := parser.ParseWorkflowYAML(content)
		if err != nil {
			continue // skip files that cannot be parsed
		}

		if len(refs) > 0 {
			deps = append(deps, parser.WorkflowDependency{
				Source:  filePath,
				Name:    name,
				Actions: refs,
			})
		}
	}

	return deps, nil
}

// getActionFileDependencies fetches action.yml or action.yaml from the repository root and parses it
func getActionFileDependencies(ctx context.Context, g *GitHubClient, repo repository.Repository, ref *string) (*parser.WorkflowDependency, error) {
	// Try action.yml first, then action.yaml
	for _, filename := range []string{"action.yml", "action.yaml"} {
		content, err := getFileContent(ctx, g, repo, filename, ref)
		if err != nil {
			continue
		}

		refs, err := parser.ParseActionYAML(content)
		if err != nil {
			continue
		}

		if len(refs) > 0 {
			return &parser.WorkflowDependency{
				Source:  filename,
				Actions: refs,
			}, nil
		}
		// File exists but has no dependencies
		return nil, nil
	}
	return nil, fmt.Errorf("no action.yml or action.yaml found")
}

// getFileContent retrieves the decoded content of a file from the repository
func getFileContent(ctx context.Context, g *GitHubClient, repo repository.Repository, path string, ref *string) ([]byte, error) {
	fileContent, _, err := g.GetRepositoryContent(ctx, repo.Owner, repo.Name, path, ref)
	if err != nil {
		return nil, err
	}
	if fileContent == nil || fileContent.Content == nil {
		return nil, fmt.Errorf("file content is empty for %s", path)
	}

	decoded, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content for %s: %w", path, err)
	}
	return decoded, nil
}

// FlattenWorkflowDependencies extracts all unique ActionReferences from multiple WorkflowDependencies
func FlattenWorkflowDependencies(deps []parser.WorkflowDependency) []parser.ActionReference {
	seen := make(map[string]bool)
	var refs []parser.ActionReference
	for _, dep := range deps {
		for _, ref := range dep.Actions {
			key := ref.VersionedName()
			if !seen[key] {
				seen[key] = true
				refs = append(refs, ref)
			}
		}
	}
	slices.SortFunc(refs, func(a, b parser.ActionReference) int {
		return strings.Compare(a.VersionedName(), b.VersionedName())
	})
	return refs
}

// ExpandFilteredDependencies performs a BFS from the filtered deps to include all
// transitively referenced dependencies (local reusable workflows, remote reusable
// workflows, and child action repositories) from the full deps list.
// This is used in recursive mode after filtering to preserve the full dependency tree.
func ExpandFilteredDependencies(filtered, allDeps []parser.WorkflowDependency) []parser.WorkflowDependency {
	// Build a map of Source -> dep for quick lookup
	depBySource := make(map[string]*parser.WorkflowDependency)
	for i := range allDeps {
		depBySource[allDeps[i].Source] = &allDeps[i]
	}

	included := make(map[string]bool)
	for _, dep := range filtered {
		included[dep.Source] = true
	}

	// BFS: follow action references to include transitively reachable deps
	queue := make([]parser.WorkflowDependency, len(filtered))
	copy(queue, filtered)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, action := range current.Actions {
			var sourceKey string
			if action.IsLocal && action.IsReusableWorkflow() {
				// Local reusable workflow (e.g. "./.github/workflows/release.yml")
				sourceKey = action.LocalPath()
			} else if action.Owner != "" && action.Repo != "" {
				if action.IsReusableWorkflow() {
					// Remote reusable workflow (e.g. "owner/repo/.github/workflows/w.yml@ref")
					sourceKey = fmt.Sprintf("%s/%s:%s", action.Owner, action.Repo, action.Path)
				} else {
					// Action repository (e.g. "actions/checkout@v4" -> "actions/checkout:action.yml")
					for _, fname := range []string{"action.yml", "action.yaml"} {
						key := fmt.Sprintf("%s/%s:%s", action.Owner, action.Repo, fname)
						if _, ok := depBySource[key]; ok {
							sourceKey = key
							break
						}
					}
				}
			}

			if sourceKey != "" && !included[sourceKey] {
				if dep, ok := depBySource[sourceKey]; ok {
					included[sourceKey] = true
					filtered = append(filtered, *dep)
					queue = append(queue, *dep)
				}
			}
		}
	}

	return filtered
}

// FilterWorkflowDependencies filters workflow dependencies by a selector string.
// The selector can match by:
//   - Workflow ID (numeric, resolved via GitHub API)
//   - Workflow name (from the YAML name: field)
//   - Filename (basename or full path under .github/workflows/)
func FilterWorkflowDependencies(ctx context.Context, g *GitHubClient, repo repository.Repository, deps []parser.WorkflowDependency, selector string) []parser.WorkflowDependency {
	if selector == "" {
		return deps
	}

	// Try to resolve by workflow ID (numeric)
	if id, err := strconv.ParseInt(selector, 10, 64); err == nil {
		wf, wfErr := g.GetWorkflowByID(ctx, repo.Owner, repo.Name, id)
		if wfErr == nil && wf != nil && wf.Path != nil {
			selector = *wf.Path
		}
	}

	var filtered []parser.WorkflowDependency
	for _, dep := range deps {
		if matchesWorkflowSelector(dep, selector) {
			filtered = append(filtered, dep)
		}
	}
	return filtered
}

// matchesWorkflowSelector checks if a WorkflowDependency matches the given selector
func matchesWorkflowSelector(dep parser.WorkflowDependency, selector string) bool {
	// Match by full source path (e.g. ".github/workflows/ci.yml")
	if strings.EqualFold(dep.Source, selector) {
		return true
	}
	// Match by filename basename (e.g. "ci.yml")
	if strings.EqualFold(filepath.Base(dep.Source), selector) {
		return true
	}
	// Match by workflow name (from YAML name: field)
	if dep.Name != "" && strings.EqualFold(dep.Name, selector) {
		return true
	}
	return false
}
