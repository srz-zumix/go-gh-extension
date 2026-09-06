package ghexec

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// envMap converts KEY=VALUE entries into a map for order-independent assertions.
func envMap(entries []string) map[string]string {
	m := make(map[string]string, len(entries))
	for _, kv := range entries {
		key, value, _ := strings.Cut(kv, "=")
		m[key] = value
	}
	return m
}

func TestEnvDropsHostAndRepoOverrides(t *testing.T) {
	t.Setenv("GH_HOST", "ghe.example.com")
	t.Setenv("GH_REPO", "octo/hello")
	// A unique sentinel proves unrelated variables are preserved.
	t.Setenv("GHEXEC_TEST_SENTINEL", "keep-me")
	// A similarly-prefixed key must NOT be filtered (exact-match, not prefix).
	t.Setenv("GH_HOSTNAME", "should-remain")

	got := envMap(env())

	assert.NotContains(t, got, "GH_HOST")
	assert.NotContains(t, got, "GH_REPO")
	assert.Equal(t, "keep-me", got["GHEXEC_TEST_SENTINEL"])
	assert.Equal(t, "should-remain", got["GH_HOSTNAME"])
}
