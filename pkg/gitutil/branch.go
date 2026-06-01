package gitutil

import (
	"fmt"
	"strings"
)

// ValidateBranchName reports an error if branch is not a safe Git ref name.
//
// The rules enforced here are a strict subset of git-check-ref-format(1):
//   - May only contain [a-zA-Z0-9._/-].
//   - Must not start or end with /.
//   - Must not end with .
//   - Must not contain // or ..
//   - No path component may start with . or end with .lock.
//
// The character allowlist also ensures the name is safe to embed inside a
// shell script (no $, backticks, semicolons, etc.).
func ValidateBranchName(branch string) error {
	if branch == "" {
		return fmt.Errorf("branch name must not be empty")
	}
	if strings.HasPrefix(branch, "/") || strings.HasSuffix(branch, "/") {
		return fmt.Errorf("branch name must not start or end with /")
	}
	if strings.HasSuffix(branch, ".") {
		return fmt.Errorf("branch name must not end with a dot")
	}
	if strings.Contains(branch, "//") || strings.Contains(branch, "..") {
		return fmt.Errorf("branch name must not contain empty path components or consecutive dots")
	}
	for _, part := range strings.Split(branch, "/") {
		if strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".lock") {
			return fmt.Errorf("branch path component %q is invalid: must not start with a dot or end with .lock", part)
		}
	}
	for i, c := range branch {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '-' || c == '_' || c == '.' || c == '/':
		default:
			return fmt.Errorf("branch name contains invalid character %q at position %d; only [a-zA-Z0-9._/-] are allowed", c, i)
		}
	}
	return nil
}
