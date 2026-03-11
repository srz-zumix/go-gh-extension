package unity

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// UnityManifest represents the structure of a Unity Packages/manifest.json file.
type UnityManifest struct {
	Dependencies map[string]string `json:"dependencies"`
}

// UnityPackage represents a single dependency entry in a Unity manifest.
// Version holds the raw value from manifest.json.
// If the value starts with "file:", Path is set to the local path.
// If the value looks like a URL (http/https/git+https etc.), URL is set.
type UnityPackage struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path,omitempty"`
	URL     string `json:"url,omitempty"`
	SHA     string `json:"sha,omitempty"`
}

// GetUnityManifest fetches and parses the Unity Packages/manifest.json from the given repository.
// The manifestPath defaults to "Packages/manifest.json" if empty.
func GetUnityManifest(ctx context.Context, g *client.GitHubClient, repo repository.Repository, manifestPath string, ref string) (*UnityManifest, error) {
	if manifestPath == "" {
		manifestPath = "Packages/manifest.json"
	}
	var refPtr *string
	if ref != "" {
		refPtr = &ref
	}
	fileContent, _, err := g.GetRepositoryContent(ctx, repo.Owner, repo.Name, manifestPath, refPtr)
	if err != nil {
		return nil, err
	}
	if fileContent == nil {
		return nil, fmt.Errorf("file not found: %s", manifestPath)
	}
	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode content: %w", err)
	}
	var manifest UnityManifest
	if err := json.Unmarshal([]byte(content), &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.json: %w", err)
	}
	return &manifest, nil
}

// ToPackages converts a UnityManifest to a sorted slice of UnityPackage.
func (m *UnityManifest) ToPackages() []UnityPackage {
	if m == nil {
		return nil
	}
	packages := make([]UnityPackage, 0, len(m.Dependencies))
	for name, version := range m.Dependencies {
		pkg := UnityPackage{Name: name, Version: version}
		if strings.HasPrefix(version, "file:") {
			pkg.Path = strings.TrimPrefix(version, "file:")
			pkg.Version = ""
		} else if isURL(version) {
			// Extract version from URL fragment; store base URL and fragment separately.
			if idx := strings.LastIndex(version, "#"); idx != -1 {
				pkg.URL = version[:idx]
				pkg.Version = version[idx+1:]
			} else {
				pkg.URL = version
				pkg.Version = ""
			}
		}
		packages = append(packages, pkg)
	}
	// Sort by name for consistent output
	for i := 1; i < len(packages); i++ {
		for j := i; j > 0 && packages[j].Name < packages[j-1].Name; j-- {
			packages[j], packages[j-1] = packages[j-1], packages[j]
		}
	}
	return packages
}

// isURL reports whether the value looks like a URL reference.
func isURL(v string) bool {
	for _, prefix := range []string{"http://", "https://", "git+", "ssh://", "git://"} {
		if strings.HasPrefix(v, prefix) {
			return true
		}
	}
	return false
}

// packageJSON is a minimal representation of a Unity package.json file.
type packageJSON struct {
	Version string `json:"version"`
}

// isCompressedFile reports whether the path appears to be a compressed archive file.
func isCompressedFile(p string) bool {
	lower := strings.ToLower(p)
	for _, ext := range []string{".tgz", ".zip", ".tar.gz", ".tar.bz2", ".tar.xz", ".gz", ".bz2"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// resolvePathGitInfo walks up the directory tree from dir toward the repository root,
// checking each level. If dir itself or any ancestor is a git submodule entry,
// it returns that submodule's commit SHA and remote git URL.
// When no submodule is found, it falls back to the latest commit SHA for the path
// via GetLatestCommitForPath. The submoduleURL is empty in the fallback case.
func resolvePathGitInfo(ctx context.Context, g *client.GitHubClient, repo repository.Repository, dir string, ref *string) (sha, submoduleURL string) {
	for d := dir; ; {
		content, _, err := g.GetRepositoryContent(ctx, repo.Owner, repo.Name, d, ref)
		if err == nil && content != nil && content.GetType() == "submodule" {
			return content.GetSHA(), content.GetSubmoduleGitURL()
		}
		parent := path.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	// Not a submodule: fetch the latest commit SHA that touched this path.
	refStr := ""
	if ref != nil {
		refStr = *ref
	}
	commit, err := g.GetLatestCommitForPath(ctx, repo.Owner, repo.Name, dir, refStr)
	if err == nil && commit != nil {
		sha = commit.GetSHA()
	}
	return sha, ""
}

// ResolveFilePackages enriches UnityPackage entries that have a local file path
// by reading their version and git metadata from the repository. The resolution strategy is:
//  1. Skip compressed archive files (tgz, zip, etc.).
//  2. Walk up the directory tree to detect a git submodule ancestor.
//     If found, the submodule remote URL is recorded in the URL field.
//  3. Resolve the git SHA for the path (from the submodule or latest commit) and
//     store it in the SHA field when available.
//  4. Read package.json inside the directory for the version string and store it
//     in the Version field when available.
//     Version and SHA are kept in separate fields and are not combined.
//
// manifestPath is the path to manifest.json within the repository and is used to
// resolve relative file: paths to repository-root-relative paths.
// Packages that cannot be resolved are returned as-is.
func ResolveFilePackages(ctx context.Context, g *client.GitHubClient, repo repository.Repository, manifestPath string, ref string, packages []UnityPackage) ([]UnityPackage, error) {
	if manifestPath == "" {
		manifestPath = "Packages/manifest.json"
	}
	var refPtr *string
	if ref != "" {
		refPtr = &ref
	}
	manifestDir := path.Dir(manifestPath)

	result := make([]UnityPackage, len(packages))
	copy(result, packages)
	for i, pkg := range result {
		if pkg.Path == "" {
			continue
		}
		// Skip compressed archive files
		if isCompressedFile(pkg.Path) {
			continue
		}
		// Resolve the package directory path relative to the repository root
		pkgDir := path.Clean(path.Join(manifestDir, pkg.Path))

		// Resolve the git SHA and optional submodule remote URL for this path
		sha, submoduleURL := resolvePathGitInfo(ctx, g, repo, pkgDir, refPtr)
		if submoduleURL != "" {
			result[i].URL = submoduleURL
		}
		if sha != "" {
			result[i].SHA = sha
		}

		// Try to read package.json for the version field
		var pkgVersion string
		pkgJSONPath := pkgDir + "/package.json"
		if fc, _, ferr := g.GetRepositoryContent(ctx, repo.Owner, repo.Name, pkgJSONPath, refPtr); ferr == nil && fc != nil {
			if rawContent, cerr := fc.GetContent(); cerr == nil {
				var pkgJSON packageJSON
				if json.Unmarshal([]byte(rawContent), &pkgJSON) == nil {
					pkgVersion = pkgJSON.Version
				}
			}
		}
		if pkgVersion != "" {
			result[i].Version = pkgVersion
		}
	}
	return result, nil
}
