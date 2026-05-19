// Package gitglob provides utilities for matching gitattributes/gitignore
// glob patterns against file paths.
package gitglob

import (
	"path"
	"regexp"
	"strings"
)

// MatchPattern reports whether filePath matches a gitattributes/gitignore glob pattern.
//
// Rules (following gitattributes/gitignore semantics):
//   - A leading '/' anchors the pattern to the repository root. It is stripped
//     before matching because filePaths have no leading slash.
//   - A pattern with no '/' (and no leading '/') is matched against the file's
//     base name only, so it applies at any depth in the tree.
//   - All other patterns (containing '/', or anchored with a leading '/') are
//     matched against the full path.
//   - '**' matches any number of path components (including zero).
//   - Multiple '**' occurrences are supported.
func MatchPattern(pattern, filePath string) bool {
	// A leading '/' anchors the pattern to the root. Strip it so the remaining
	// pattern can be matched against a path that has no leading slash.
	anchored := strings.HasPrefix(pattern, "/")
	if anchored {
		pattern = pattern[1:]
	}

	// A trailing '/' marks a directory pattern: match everything inside it.
	if strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/") + "/**"
	}

	// A pattern with no '/' (and not explicitly root-anchored) is matched against
	// the base name only, so "*.go" matches any .go file at any depth.
	if !anchored && !strings.Contains(pattern, "/") {
		base := path.Base(filePath)
		// path.Match uses '[^' for negation; normalise gitignore-style '[!' to '[^'.
		p := strings.ReplaceAll(pattern, "[!", "[^")
		matched, err := path.Match(p, base)
		return err == nil && matched
	}

	// Full-path match: compile to regex for correct globstar ('**') semantics.
	re, err := globToRegexp(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(filePath)
}

// globToRegexp translates a gitattributes/gitignore glob pattern into a
// compiled regular expression. The pattern must already have any leading '/'
// stripped.
//
// Translation rules:
//   - '/**/' → matches '/' or '/anything/' (zero or more intermediate components)
//   - '**/'  at the start → matches empty or any leading path components
//   - '**'   elsewhere (end or bare) → matches any characters including '/'
//   - '*'    → matches any characters except '/'
//   - '?'    → matches exactly one character except '/'
//   - '[…]'  → character class; '[!…]' negation is rewritten to '[^…]'
//   - other  → literal (regexp-escaped)
func globToRegexp(pattern string) (*regexp.Regexp, error) {
	var sb strings.Builder
	sb.WriteString(`^`)

	i := 0
	for i < len(pattern) {
		if pattern[i] == '*' && i+1 < len(pattern) && pattern[i+1] == '*' {
			leadSlash := i > 0 && pattern[i-1] == '/'
			trailSlash := i+2 < len(pattern) && pattern[i+2] == '/'

			switch {
			case leadSlash && trailSlash:
				// "/**/" — the leading '/' is already emitted; match zero or more
				// intermediate path components followed by '/'.
				sb.WriteString(`(?:.+/)?`)
				i += 3 // consume "**/"
			case !leadSlash && trailSlash:
				// "**/" at start — match zero or more leading path components.
				sb.WriteString(`(?:.*/)?`)
				i += 3 // consume "**/"
			default:
				// "**" at end or bare — match anything.
				sb.WriteString(`.*`)
				i += 2
			}
			continue
		}

		switch pattern[i] {
		case '*':
			sb.WriteString(`[^/]*`)
			i++
		case '?':
			sb.WriteString(`[^/]`)
			i++
		case '[':
			// Find the closing ']', honouring the gitignore rules for a literal ']'
			// at the start of the class and '[!' negation.
			j := i + 1
			if j < len(pattern) && (pattern[j] == '!' || pattern[j] == '^') {
				j++
			}
			if j < len(pattern) && pattern[j] == ']' {
				j++ // literal ']' as the first char of the class
			}
			for j < len(pattern) && pattern[j] != ']' {
				j++
			}
			if j < len(pattern) {
				class := pattern[i : j+1]
				// Rewrite gitignore negation '[!…]' to regex '[^…]'.
				if len(class) > 1 && class[1] == '!' {
					class = "[^" + class[2:]
				}
				sb.WriteString(class)
				i = j + 1
			} else {
				// No closing bracket; treat '[' as a literal.
				sb.WriteString(`\[`)
				i++
			}
		default:
			sb.WriteString(regexp.QuoteMeta(string(pattern[i])))
			i++
		}
	}

	sb.WriteString(`$`)
	return regexp.Compile(sb.String())
}
