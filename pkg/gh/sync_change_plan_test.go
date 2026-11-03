package gh

import (
	"context"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
)

func TestPlanSyncRepoUserPermissionsIsReadOnly(t *testing.T) {
	sourceURL := fixtureAPI + "repos/source-org/source-repo/collaborators?affiliation=direct&per_page=100"
	destinationURL := fixtureAPI + "repos/destination-org/destination-repo/collaborators?affiliation=direct&per_page=100"
	client, transport := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + sourceURL: ok(`[
			{"login":"sample-user","permissions":{"pull":true,"push":true}}
		]`),
		"GET " + destinationURL: ok(`[
			{"login":"sample-user","permissions":{"pull":true}},
			{"login":"retired-user","permissions":{"admin":true}}
		]`),
	})
	dst := repository.Repository{
		Host:  "github.com",
		Owner: "destination-org",
		Name:  "destination-repo",
	}

	changes, err := PlanSyncRepoUserPermissions(
		context.Background(),
		client,
		repository.Repository{Owner: "source-org", Name: "source-repo"},
		client,
		dst,
	)
	if err != nil {
		t.Fatalf("PlanSyncRepoUserPermissions() error = %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("got %d changes, want 2: %#v", len(changes), changes)
	}
	if changes[0].Action != SyncChangeUpdate || changes[0].Target != "sample-user" || changes[0].Permission != "push" || changes[0].Current != "pull" {
		t.Errorf("unexpected update plan: %#v", changes[0])
	}
	if changes[0].Destination != "github.com/destination-org/destination-repo" {
		t.Errorf("destination = %q, want host-qualified repository", changes[0].Destination)
	}
	if changes[1].Action != SyncChangeRemove || changes[1].Target != "retired-user" || changes[1].Current != "admin" {
		t.Errorf("unexpected removal plan: %#v", changes[1])
	}
	for _, request := range transport.requests {
		if request.Method != "GET" {
			t.Errorf("plan issued a write request: %s %s", request.Method, request.URL)
		}
	}
}
