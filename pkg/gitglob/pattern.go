// Package gitglob provides utilities for matching gitattributes/gitignore
// glob patterns against file paths.
package gitglob

import (
	"path"
	"strings"
)

// MatchPattern reports whether filePath matches a gitattributes glob pattern.
//
// Rules (following gitattributes/gitignore semantics):
//   - A leading '/' anchors the pattern to the repository root. It is stripped
//     before matching because filePaths have no leading slash.
//   - A pattern with no '/' (and no leading '/') is matched against the file's
//     base name only, so it applies at any depth in the tree.
//   - All other patterns (containing '/', or anchored with a leading '/') are
//     matched against the full path.
//   - '**' matches any number of path components.
func MatchPattern(pattern, filePath string) bool {
	// A leading '/' anchors the pattern to the root. Strip it so the remaining
	// pattern can be matched against a path that has no leading slash.
	anchored := strings.HasPrefix(pattern, "/")
	if anchored {
		pattern = pattern[1:]
	}

	// A pattern with no '/' (and not explicitly root-anchored) is matched against
	// the base name only, so "*.go" matches any .go file at any depth.
	if !anchored && !strings.Contains(pattern, "/") {
		base := path.Base(filePath)
		matched, err := path.Match(pattern, base)
		return err == nil && matched
	}

	// Handle '**' patterns; e.g. "vendor/**" or "/gen/**/*.go".
	if strings.Contains(pattern, "**") {
		parts := strings.SplitN(pattern, "**", 2)
		prefix := parts[0]
		suffix := parts[1]
		if !strings.HasPrefix(filePath, prefix) {
			return false
		}
		remainder := strings.TrimPrefix(filePath, prefix)
		if suffix == "" || suffix == "/" {
			return true
		}
		matched, err := path.Match(strings.TrimPrefix(suffix, "/"), remainder)
		return err == nil && matched
	}

	matched, err := path.Match(pattern, filePath)
	return err == nil && matched
}
