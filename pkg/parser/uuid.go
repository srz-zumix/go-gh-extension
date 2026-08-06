package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// uuidPattern matches the canonical 8-4-4-4-12 hexadecimal UUID format.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// UUID validates that s is a canonical 8-4-4-4-12 hexadecimal UUID and returns
// it normalized to lowercase. Returns an error if s is not a valid UUID.
func UUID(s string) (string, error) {
	if !uuidPattern.MatchString(s) {
		return "", fmt.Errorf("invalid UUID format: %s", s)
	}
	return strings.ToLower(s), nil
}
