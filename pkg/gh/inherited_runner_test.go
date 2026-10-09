package gh

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v90/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newInheritedTestClient(t *testing.T, routes map[string]string) (*GitHubClient, map[string]int) {
	t.Helper()
	calls := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls[req.URL.Path]++
		body, ok := routes[req.URL.Path]
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			_, _ = fmt.Fprint(w, `{"message":"Forbidden"}`)
			return
		}
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	base := server.URL + "/"
	gc, err := github.NewClient(github.WithURLs(&base, nil))
	require.NoError(t, err)
	g, err := client.NewClient(gc)
	require.NoError(t, err)
	return g, calls
}

const inheritedGroups = `{"total_count":2,"runner_groups":[{"id":1,"name":"Default","inherited":false},{"id":3,"name":"shared","inherited":true}]}`

func TestListOrgRunnersWithInherited(t *testing.T) {
	g, calls := newInheritedTestClient(t, map[string]string{
		"/orgs/octo/actions/runners":                 `{"total_count":2,"runners":[{"id":1,"name":"org","runner_group_id":1},{"id":7,"name":"dup","runner_group_id":3}]}`,
		"/orgs/octo/actions/runner-groups":           inheritedGroups,
		"/orgs/octo/actions/runner-groups/3/runners": `{"total_count":3,"runners":[{"id":1,"name":"ent","runner_group_id":3},{"id":7,"name":"dup","runner_group_id":3},{"id":9,"name":"ent2","runner_group_id":3}]}`,
	})

	runners, err := ListOrgRunnersWithInherited(t.Context(), g, repository.Repository{Owner: "octo"})
	require.NoError(t, err)

	names := make([]string, len(runners))
	for i, r := range runners {
		names[i] = r.GetName()
	}
	// The enterprise runner sharing ID 1 with an organization runner is kept; the duplicate of ID 7 is not.
	assert.Equal(t, []string{"org", "dup", "ent", "ent2"}, names)
	assert.Zero(t, calls["/orgs/octo/actions/runner-groups/1/runners"])
}

func TestListOrgHostedRunnersWithInherited(t *testing.T) {
	g, calls := newInheritedTestClient(t, map[string]string{
		"/orgs/octo/actions/hosted-runners":                 `{"total_count":1,"runners":[{"id":1,"name":"org-pool","runner_group_id":1}]}`,
		"/orgs/octo/actions/runner-groups":                  inheritedGroups,
		"/orgs/octo/actions/runner-groups/3/hosted-runners": `{"total_count":2,"runners":[{"id":1,"name":"ent-small","runner_group_id":3},{"id":2,"name":"ent-large","runner_group_id":3}]}`,
	})

	runners, err := ListOrgHostedRunnersWithInherited(t.Context(), g, repository.Repository{Owner: "octo"})
	require.NoError(t, err)

	names := make([]string, len(runners))
	for i, r := range runners {
		names[i] = r.GetName()
	}
	assert.Equal(t, []string{"org-pool", "ent-small", "ent-large"}, names)
	assert.Zero(t, calls["/orgs/octo/actions/runner-groups/1/hosted-runners"])
}

func TestListOrgRunnersWithInheritedErrors(t *testing.T) {
	for name, routes := range map[string]map[string]string{
		"runner groups": {
			"/orgs/octo/actions/runners":        `{"total_count":0,"runners":[]}`,
			"/orgs/octo/actions/hosted-runners": `{"total_count":0,"runners":[]}`,
		},
		"inherited group": {
			"/orgs/octo/actions/runners":        `{"total_count":0,"runners":[]}`,
			"/orgs/octo/actions/hosted-runners": `{"total_count":0,"runners":[]}`,
			"/orgs/octo/actions/runner-groups":  inheritedGroups,
		},
	} {
		t.Run(name, func(t *testing.T) {
			g, _ := newInheritedTestClient(t, routes)
			repo := repository.Repository{Owner: "octo"}

			runners, err := ListOrgRunnersWithInherited(t.Context(), g, repo)
			assert.Error(t, err)
			assert.Nil(t, runners)

			hosted, err := ListOrgHostedRunnersWithInherited(t.Context(), g, repo)
			assert.Error(t, err)
			assert.Nil(t, hosted)
		})
	}
}
