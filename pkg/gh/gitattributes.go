package gh

import (
	"bufio"
	"context"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v88/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gitglob"
)

// linguistGeneratedRule pairs a gitattributes glob pattern with whether it
// sets (true) or unsets (false) the linguist-generated attribute.
type linguistGeneratedRule struct {
	pattern   string
	generated bool
}

// linguistGeneratedState returns true/false and whether the attr token
// explicitly addresses the linguist-generated attribute at all.
//
// Setting forms  : "linguist-generated", "linguist-generated=true"
// Unsetting forms: "-linguist-generated", "linguist-generated=false"
func linguistGeneratedState(attr string) (generated bool, ok bool) {
	attr = strings.TrimSpace(attr)
	switch {
	case attr == "linguist-generated" || strings.EqualFold(attr, "linguist-generated=true"):
		return true, true
	case attr == "-linguist-generated" || strings.EqualFold(attr, "linguist-generated=false"):
		return false, true
	}
	return false, false
}

// GetLinguistGenerated returns the subset of files that are effectively marked
// as linguist-generated in the repository's .gitattributes file at the
// specified ref.
//
// Evaluation follows gitattributes semantics: all patterns are applied in
// order and the last matching pattern wins. A -linguist-generated or
// linguist-generated=false rule on a later pattern correctly overrides an
// earlier setting rule for the same file.
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

	// Parse gitattributes for linguist-generated rules (both set and unset forms).
	var rules []linguistGeneratedRule
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
			if generated, ok := linguistGeneratedState(attr); ok {
				rules = append(rules, linguistGeneratedRule{pattern: pattern, generated: generated})
				break
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(rules) == 0 {
		return nil, nil
	}

	// For each file apply all rules in order; the last matching rule wins.
	var result []*github.CommitFile
	for _, f := range prFiles {
		filePath := f.GetFilename()
		effective := false
		matched := false
		for _, rule := range rules {
			if gitglob.MatchPattern(rule.pattern, filePath) {
				effective = rule.generated
				matched = true
			}
		}
		if matched && effective {
			result = append(result, f)
		}
	}
	return result, nil
}
