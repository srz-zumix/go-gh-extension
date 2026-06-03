package parser

import (
	"fmt"
	"strings"
)

// knownTokenPrefixes lists known GitHub token prefixes used to detect
// raw token values accidentally passed as secret names.
var knownTokenPrefixes = []string{
	"ghp_",        // classic personal access token
	"github_pat_", // fine-grained personal access token
	"gho_",        // OAuth app token
	"ghs_",        // GitHub Actions token
	"ghr_",        // refresh token
}

// ValidateTokenSecretName returns an error if name looks like a raw GitHub
// token value rather than a secret name. Flags that expect a secret name
// (e.g. MY_SECRET) should use this to guard against accidentally receiving
// a raw token value.
func ValidateTokenSecretName(name string) error {
	for _, prefix := range knownTokenPrefixes {
		if strings.HasPrefix(name, prefix) {
			return fmt.Errorf("expected a secret name (e.g. MY_SECRET), not a token value; got a value that looks like a %s token", prefix)
		}
	}
	return nil
}
