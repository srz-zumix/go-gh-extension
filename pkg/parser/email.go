package parser

import "strings"

// NormalizeEmail trims leading/trailing whitespace and converts the email address to lowercase.
// Returns an empty string if the input is blank after trimming.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
