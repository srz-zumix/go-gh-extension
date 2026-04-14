package gh

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubTokenRoundTripper implements client.tokenGetter and http.RoundTripper for tests.
type stubTokenRoundTripper struct {
	token string
}

func (s stubTokenRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, nil
}

func (s stubTokenRoundTripper) Token() string {
	return s.token
}

// newGHTestClient creates a *GitHubClient with the given token and base URL for use in pkg/gh tests.
func newGHTestClient(t *testing.T, baseURL, token string) *GitHubClient {
	t.Helper()
	gc := github.NewClient(&http.Client{Transport: stubTokenRoundTripper{token: token}})
	parsed, err := url.Parse(baseURL)
	require.NoError(t, err)
	if parsed.Path == "" || parsed.Path[len(parsed.Path)-1] != '/' {
		parsed.Path += "/"
	}
	gc.BaseURL = parsed
	g, err := client.NewClient(gc)
	require.NoError(t, err)
	return g
}

func TestFindRepository(t *testing.T) {
	repo1 := &github.Repository{ID: github.Ptr(int64(1))}
	repo2 := &github.Repository{ID: github.Ptr(int64(2))}
	repos := []*github.Repository{repo1, repo2}

	target := &github.Repository{ID: github.Ptr(int64(1))}
	result := FindRepository(target, repos)

	if result == nil || *result.ID != *target.ID {
		t.Errorf("Expected repository with ID %d, got nil or different ID", *target.ID)
	}

	nonExistent := &github.Repository{ID: github.Ptr(int64(3))}
	result = FindRepository(nonExistent, repos)

	if result != nil {
		t.Errorf("Expected nil for non-existent repository, got %v", result)
	}
}

func TestFilterRepositoriesByNames(t *testing.T) {
	repo1 := &github.Repository{FullName: github.Ptr("owner/repo1")}
	repo2 := &github.Repository{FullName: github.Ptr("owner/repo2")}
	repos := []*github.Repository{repo1, repo2}

	names := []string{"repo1"}
	owner := "owner"
	filtered := FilterRepositoriesByNames(repos, names, owner)

	if len(filtered) != 1 || *filtered[0].FullName != "owner/repo1" {
		t.Errorf("Expected filtered repository to be 'owner/repo1', got %v", filtered)
	}

	names = []string{"repo3"}
	filtered = FilterRepositoriesByNames(repos, names, owner)

	if len(filtered) != 0 {
		t.Errorf("Expected no repositories to be filtered, got %v", filtered)
	}
}

func TestFilterTeamByNames(t *testing.T) {
	team1 := &github.Team{Slug: github.Ptr("team1")}
	team2 := &github.Team{Slug: github.Ptr("team2")}
	teams := []*github.Team{team1, team2}

	slugs := []string{"team1"}
	filtered := FilterTeamByNames(teams, slugs)

	if len(filtered) != 1 || *filtered[0].Slug != "team1" {
		t.Errorf("Expected filtered team to be 'team1', got %v", filtered)
	}

	slugs = []string{"team3"}
	filtered = FilterTeamByNames(teams, slugs)

	if len(filtered) != 0 {
		t.Errorf("Expected no teams to be filtered, got %v", filtered)
	}
}

// --- GitCmdEnv ---

func TestGitCmdEnv_StripsExistingGitConfigVars(t *testing.T) {
	t.Setenv("GIT_CONFIG_COUNT", "2")
	t.Setenv("GIT_CONFIG_KEY_0", "http.extraHeader")
	t.Setenv("GIT_CONFIG_KEY_1", "http.extraHeader2")
	t.Setenv("GIT_CONFIG_VALUE_0", "Authorization: Bearer old-token")
	t.Setenv("GIT_CONFIG_VALUE_1", "Authorization: Bearer old-token2")

	g := newGHTestClient(t, "https://api.github.com/", "new-token")
	env := GitCmdEnv(g, "https://github.com/owner/repo")

	// The old values must not appear; GitAuthEnvs replaces them with a single new entry.
	assert.NotContains(t, env, "GIT_CONFIG_COUNT=2")
	assert.NotContains(t, env, "GIT_CONFIG_KEY_0=http.extraHeader")
	assert.NotContains(t, env, "GIT_CONFIG_KEY_1=http.extraHeader2")
	assert.NotContains(t, env, "GIT_CONFIG_VALUE_0=Authorization: Bearer old-token")
	assert.NotContains(t, env, "GIT_CONFIG_VALUE_1=Authorization: Bearer old-token2")
	// GIT_CONFIG_KEY_1 / VALUE_1 must be absent entirely (GitAuthEnvs only uses index 0).
	for _, kv := range env {
		assert.False(t, strings.HasPrefix(kv, "GIT_CONFIG_KEY_1="), "GIT_CONFIG_KEY_1 should be stripped")
		assert.False(t, strings.HasPrefix(kv, "GIT_CONFIG_VALUE_1="), "GIT_CONFIG_VALUE_1 should be stripped")
	}
}

func TestGitCmdEnv_SetsGitTerminalPrompt(t *testing.T) {
	g := newGHTestClient(t, "https://api.github.com/", "test-token")
	env := GitCmdEnv(g, "https://github.com/owner/repo")

	assert.Contains(t, env, "GIT_TERMINAL_PROMPT=0")
}

func TestGitCmdEnv_AppendsAuthEnvs(t *testing.T) {
	g := newGHTestClient(t, "https://api.github.com/", "test-token")
	env := GitCmdEnv(g, "https://github.com/owner/repo")

	// GitAuthEnvs produces GIT_CONFIG_COUNT + KEY + VALUE entries.
	var configCount, configKey, configValue string
	for _, kv := range env {
		switch {
		case strings.HasPrefix(kv, "GIT_CONFIG_COUNT="):
			configCount = kv
		case strings.HasPrefix(kv, "GIT_CONFIG_KEY_0="):
			configKey = kv
		case strings.HasPrefix(kv, "GIT_CONFIG_VALUE_0="):
			configValue = kv
		}
	}
	assert.Equal(t, "GIT_CONFIG_COUNT=1", configCount)
	assert.Contains(t, configKey, "github.com")
	assert.Contains(t, configValue, "Authorization:")
}

func TestGitCmdEnv_NoAuthEnvsWhenTokenEmpty(t *testing.T) {
	// A client with no token should produce no GIT_CONFIG_* entries.
	g := newGHTestClient(t, "https://api.github.com/", "")
	env := GitCmdEnv(g, "https://github.com/owner/repo")

	for _, kv := range env {
		key, _, _ := strings.Cut(kv, "=")
		assert.NotEqual(t, "GIT_CONFIG_COUNT", key)
		assert.False(t, strings.HasPrefix(key, "GIT_CONFIG_KEY_"))
		assert.False(t, strings.HasPrefix(key, "GIT_CONFIG_VALUE_"))
	}
}

func TestGitCmdEnv_PreservesOtherEnvVars(t *testing.T) {
	const sentinelKey = "GIT_CMD_ENV_TEST_SENTINEL"
	t.Setenv(sentinelKey, "present")

	g := newGHTestClient(t, "https://api.github.com/", "test-token")
	env := GitCmdEnv(g, "https://github.com/owner/repo")

	found := false
	for _, kv := range env {
		if kv == sentinelKey+"=present" {
			found = true
			break
		}
	}
	assert.True(t, found, "unrelated env var should be preserved")
}

func TestGitCmdEnv_ContainsCurrentEnv(t *testing.T) {
	// Verify that os.Environ() entries (minus GIT_CONFIG_*) are included.
	g := newGHTestClient(t, "https://api.github.com/", "test-token")
	env := GitCmdEnv(g, "https://github.com/owner/repo")

	envSet := make(map[string]struct{}, len(env))
	for _, kv := range env {
		envSet[kv] = struct{}{}
	}

	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if key == "GIT_CONFIG_COUNT" || strings.HasPrefix(key, "GIT_CONFIG_KEY_") || strings.HasPrefix(key, "GIT_CONFIG_VALUE_") {
			continue
		}
		_, ok := envSet[kv]
		assert.True(t, ok, "expected %q to be present in GitCmdEnv output", kv)
	}
}
