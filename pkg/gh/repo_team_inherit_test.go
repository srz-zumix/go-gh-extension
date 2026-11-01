package gh

import (
	"context"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTeamReposWithoutInheritUsesDirectPermission(t *testing.T) {
	g, tr := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "orgs/example-org/teams/child-team/repos?per_page=100": ok(`[
			{"id":1,"name":"shared-repo","full_name":"example-org/shared-repo","owner":{"login":"example-org"},"permissions":{"admin":true,"maintain":true,"push":true,"triage":true,"pull":true}},
			{"id":2,"name":"inherited-repo","full_name":"example-org/inherited-repo","owner":{"login":"example-org"},"permissions":{"admin":false,"maintain":false,"push":false,"triage":false,"pull":true}}
		]`),
		"GET " + fixtureAPI + "orgs/example-org/teams/child-team": ok(`{"slug":"child-team","parent":{"slug":"parent-team"}}`),
		"GET " + fixtureAPI + "orgs/example-org/teams/parent-team/repos?per_page=100": ok(`[
			{"id":1,"name":"shared-repo","full_name":"example-org/shared-repo","owner":{"login":"example-org"},"permissions":{"admin":true,"maintain":true,"push":true,"triage":true,"pull":true}},
			{"id":2,"name":"inherited-repo","full_name":"example-org/inherited-repo","owner":{"login":"example-org"},"permissions":{"admin":false,"maintain":false,"push":false,"triage":false,"pull":true}}
		]`),
		"GET " + fixtureAPI + "repos/example-org/shared-repo/teams?per_page=100": ok(`[
			{"slug":"parent-team","permission":"admin","permissions":{"admin":true,"maintain":true,"push":true,"triage":true,"pull":true}},
			{"slug":"child-team","permission":"pull","permissions":{"admin":false,"maintain":false,"push":false,"triage":false,"pull":true}}
		]`),
		"GET " + fixtureAPI + "repos/example-org/inherited-repo/teams?per_page=100": ok(`[
			{"slug":"parent-team","permission":"pull","permissions":{"admin":false,"maintain":false,"push":false,"triage":false,"pull":true}}
		]`),
	})

	repos, err := ListTeamRepos(context.Background(), g, repository.Repository{Owner: "example-org"}, "child-team", nil, false)
	if err != nil {
		for _, r := range tr.requests {
			t.Log(r.Method, r.URL)
		}
	}
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "shared-repo", repos[0].GetName())
	assert.Equal(t, "pull", GetRepositoryPermissions(repos[0]))
}
