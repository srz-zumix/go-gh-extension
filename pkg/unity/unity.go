package unity

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// UnityManifest represents the structure of a Unity Packages/manifest.json file.
type UnityManifest struct {
	Dependencies map[string]string `json:"dependencies"`
}

// UnityPackage represents a normalized dependency entry in a Unity manifest.
// Version holds the parsed version or reference:
//   - for standard dependencies, it is the version string from manifest.json
//   - for URL dependencies with a fragment, it is the fragment part after '#'
//   - for file: dependencies or URL dependencies without a fragment, it is empty
// Path is set to the local path when the manifest value starts with "file:".
// URL is set when the manifest value is recognized as a URL; for values with a
// fragment, URL stores the base URL without the fragment.
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
	slices.SortFunc(packages, func(a, b UnityPackage) int {
		return cmp.Compare(a.Name, b.Name)
	})
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

// isNotFoundError returns true if err is a GitHub 404 response.
func isNotFoundError(err error) bool {
	var errResp *github.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusNotFound
}

// resolvePathGitInfo walks up the directory tree from dir toward the repository root,
// checking each level. If dir itself or any ancestor is a git submodule entry,
// it returns that submodule's commit SHA and remote git URL.
// When no submodule is found, it falls back to the latest commit SHA for the path
// via GetLatestCommitForPath. The submoduleURL is empty in the fallback case.
// Non-404 API errors (including context cancellation) are returned immediately.
func resolvePathGitInfo(ctx context.Context, g *client.GitHubClient, repo repository.Repository, dir string, ref *string) (sha, submoduleURL string, err error) {
	for d := dir; ; {
		if err := ctx.Err(); err != nil {
			return "", "", err
		}
		content, _, apiErr := g.GetRepositoryContent(ctx, repo.Owner, repo.Name, d, ref)
		if apiErr != nil {
			if !isNotFoundError(apiErr) {
				return "", "", apiErr
			}
		} else if content != nil && content.GetType() == "submodule" {
			return content.GetSHA(), content.GetSubmoduleGitURL(), nil
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
	commit, commitErr := g.GetLatestCommitForPath(ctx, repo.Owner, repo.Name, dir, refStr)
	if commitErr != nil && !isNotFoundError(commitErr) {
		return "", "", commitErr
	}
	if commit != nil {
		sha = commit.GetSHA()
	}
	return sha, "", nil
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

		if path.IsAbs(pkgDir) || pkgDir == ".." || strings.HasPrefix(pkgDir, "../") {
			continue
		}

		// Resolve the git SHA and optional submodule remote URL for this path
		sha, submoduleURL, err := resolvePathGitInfo(ctx, g, repo, pkgDir, refPtr)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve git info for %s: %w", pkgDir, err)
		}
		if submoduleURL != "" {
			result[i].URL = submoduleURL
		}
		if sha != "" {
			result[i].SHA = sha
		}

		// Try to read package.json for the version field
		var pkgVersion string
		pkgJSONPath := pkgDir + "/package.json"
		fc, _, ferr := g.GetRepositoryContent(ctx, repo.Owner, repo.Name, pkgJSONPath, refPtr)
		if ferr != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("failed to fetch package.json for %s: %w", pkgDir, ctx.Err())
			}
			if !isNotFoundError(ferr) {
				return nil, fmt.Errorf("failed to fetch package.json for %s: %w", pkgDir, ferr)
			}
		} else if fc != nil {
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
