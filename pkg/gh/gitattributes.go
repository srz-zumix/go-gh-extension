package gh

import (
	"bufio"
	"context"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gitglob"
)

// isLinguistGeneratedPattern reports whether a gitattributes attribute token
// sets linguist-generated to true. It accepts both "linguist-generated" and
// "linguist-generated=true" forms.
func isLinguistGeneratedPattern(attr string) bool {
	attr = strings.TrimSpace(attr)
	return attr == "linguist-generated" || strings.EqualFold(attr, "linguist-generated=true")
}

// GetLinguistGenerated returns the subset of files that are marked as
// linguist-generated in the repository's .gitattributes file at the specified ref.
func GetLinguistGenerated(ctx context.Context, g *GitHubClient, repo repository.Repository, ref *string, prFiles []*github.CommitFile) ([]*github.CommitFile, error) {
	content, err := GetRepositoryFileContent(ctx, g, repo, ".gitattributes", ref)
	if err != nil {
		if IsHTTPNotFound(err) {
			// .gitattributes not found → no linguist-generated files
			return nil, nil
		}
		return nil, err
	}

	body, err := content.GetContent()
	if err != nil {
		return nil, err
	}

	// Parse gitattributes for linguist-generated patterns
	var generatedPatterns []string
	scanner := bufio.NewScanner(strings.NewReader(body))
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
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(generatedPatterns) == 0 {
		return nil, nil
	}

	var result []*github.CommitFile
	for _, f := range prFiles {
		filePath := f.GetFilename()
		for _, pattern := range generatedPatterns {
			if gitglob.MatchPattern(pattern, filePath) {
				result = append(result, f)
				break
			}
		}
	}
	return result, nil
}
