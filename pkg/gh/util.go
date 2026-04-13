package gh

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"slices"

	"github.com/google/go-github/v84/github"
)

// GitCmdEnv returns a git command environment that disables prompts and injects
// bearer-token credentials for rawURL via GIT_CONFIG_* variables.
// Any existing GIT_CONFIG_COUNT/KEY/VALUE entries in the parent environment are
// stripped to avoid leaving git with duplicate or conflicting config variables.
func GitCmdEnv(g *GitHubClient, rawURL string) []string {
	base := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if key == "GIT_CONFIG_COUNT" || strings.HasPrefix(key, "GIT_CONFIG_KEY_") || strings.HasPrefix(key, "GIT_CONFIG_VALUE_") {
			continue
		}
		base = append(base, kv)
	}
	base = append(base, "GIT_TERMINAL_PROMPT=0")
	return append(base, g.GitAuthEnvs(rawURL)...)
}

// FindRepository searches for a repository in a list by matching IDs and returns it, or nil if not found.
func FindRepository(target *github.Repository, repos []*github.Repository) *github.Repository {
	for _, r := range repos {
		if *r.ID == *target.ID {
			return r
		}
	}
	return nil
}

// FilterRepositoriesByNames filters a list of repositories by their full names (owner/repo).
// If the names do not include the owner, the owner is prepended.
func FilterRepositoriesByNames(repos []*github.Repository, names []string, owner string) []*github.Repository {
	nameSet := make(map[string]struct{})
	for _, name := range names {
		if !strings.Contains(name, "/") {
			name = owner + "/" + name
		}
		nameSet[name] = struct{}{}
	}

	var filteredRepos []*github.Repository
	for _, repo := range repos {
		repoFullName := repo.GetFullName() // owner/repo
		if _, exists := nameSet[repoFullName]; exists {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	return filteredRepos
}

// FilterTeamByNames filters a list of teams by their slugs.
func FilterTeamByNames(teams []*github.Team, slugs []string) []*github.Team {
	var filteredTeams []*github.Team
	for _, team := range teams {
		teamSlug := team.GetSlug()
		if slices.Contains(slugs, teamSlug) {
			filteredTeams = append(filteredTeams, team)
		}
	}
	return filteredTeams
}

// IsHTTPNotFound returns true if err is a GitHub API 404 Not Found response.
func IsHTTPNotFound(err error) bool {
	var errResp *github.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusNotFound
}

// IsHTTPForbidden returns true if err is a GitHub API 403 Forbidden response.
func IsHTTPForbidden(err error) bool {
	var errResp *github.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusForbidden
}

// IsHTTPBadRequest returns true if err is a GitHub API 400 Bad Request response.
func IsHTTPBadRequest(err error) bool {
	var errResp *github.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusBadRequest
}

// IsHTTPUnprocessableEntity returns true if err is a GitHub API 422 Unprocessable Entity response.
func IsHTTPUnprocessableEntity(err error) bool {
	var errResp *github.ErrorResponse
	return errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusUnprocessableEntity
}
