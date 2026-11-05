package gh

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionsPolicyWrappers(t *testing.T) {
	target := repository.Repository{Owner: "fixture-org", Name: "fixture-repo"}
	responses := map[string]fixtureResponse{
		"GET " + fixtureAPI + "repos/fixture-org/fixture-repo/actions/policies?has_parents=true&page=1&per_page=100": ok(`{"total_count":1,"policies":[{"id":11}]}`),
		"GET " + fixtureAPI + "orgs/fixture-org/actions/policies?has_parents=true&page=1&per_page=100":               ok(`{"total_count":1,"policies":[{"id":12}]}`),
		"GET " + fixtureAPI + "repos/fixture-org/fixture-repo/actions/policies/11":                                   ok(`{"id":11}`),
		"GET " + fixtureAPI + "orgs/fixture-org/actions/policies/12":                                                 ok(`{"id":12}`),
	}
	g, transport := newFixtureClient(t, responses)

	repoPolicies, err := ListRepositoryActionsPolicies(t.Context(), g, target)
	require.NoError(t, err)
	require.Len(t, repoPolicies, 1)
	assert.JSONEq(t, `{"id":11}`, string(repoPolicies[0]))

	orgPolicies, err := ListOrgActionsPolicies(t.Context(), g, target)
	require.NoError(t, err)
	require.Len(t, orgPolicies, 1)
	assert.JSONEq(t, `{"id":12}`, string(orgPolicies[0]))

	repoPolicy, err := GetRepositoryActionsPolicy(t.Context(), g, target, 11)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":11}`, string(repoPolicy))

	orgPolicy, err := GetOrgActionsPolicy(t.Context(), g, target, 12)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":12}`, string(orgPolicy))

	require.Len(t, transport.requests, 4)
	for _, request := range transport.requests {
		assert.Equal(t, "GET", request.Method)
		assert.Equal(t, "2026-03-10", request.Version)
	}
}
