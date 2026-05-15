package gh

import (
	"bufio"
	"context"
	"path"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// isLinguistGeneratedPattern reports whether a gitattributes attribute token
// sets linguist-generated to true. It accepts both "linguist-generated" and
// "linguist-generated=true" forms.
func isLinguistGeneratedPattern(attr string) bool {
	attr = strings.TrimSpace(attr)
	return attr == "linguist-generated" || strings.EqualFold(attr, "linguist-generated=true")
}

// matchGitattributesPattern reports whether filePath matches a gitattributes glob pattern.
// Patterns without a '/' are matched against the file's base name only.
// Patterns with '/' are matched against the full path using path.Match.
// '**' is handled by splitting the pattern around '**' and checking prefix/suffix.
func matchGitattributesPattern(pattern, filePath string) bool {
	if !strings.Contains(pattern, "/") {
		// Match against base name only
		base := path.Base(filePath)
		matched, err := path.Match(pattern, base)
		return err == nil && matched
	}
	// Handle ** by splitting; e.g. "vendor/**" → prefix "vendor/"
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

// GetLinguistGenerated returns the subset of files that are marked as
// linguist-generated in the repository's .gitattributes file at the specified ref.
func GetLinguistGenerated(ctx context.Context, g *GitHubClient, repo repository.Repository, ref string, prFiles []*github.CommitFile) ([]*github.CommitFile, error) {
	content, err := g.GetRepositoryFileContent(ctx, repo.Owner, repo.Name, ".gitattributes", ref)
	if err != nil {
		// .gitattributes not found → no linguist-generated files
		return nil, nil //nolint:nilerr
	}

	// Parse gitattributes for linguist-generated patterns
	var generatedPatterns []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pattern := fields[0]
		for _, attr := range fields[1:] {
			if isLinguistGeneratedPattern(attr) {
				generatedPatterns = append(generatedPatterns, pattern)
				break
			}
		}
	}

	if len(generatedPatterns) == 0 {
		return nil, nil
	}

	var result []*github.CommitFile
	for _, f := range prFiles {
		filePath := f.GetFilename()
		for _, pattern := range generatedPatterns {
			if matchGitattributesPattern(pattern, filePath) {
				result = append(result, f)
				break
			}
		}
	}
	return result, nil
}
