package gh

import (
	"context"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v93/github"
)

func TestPlanTeamRepoPermissionChanges(t *testing.T) {
	repo := func(name, permission string) *github.Repository {
		return &github.Repository{
			Name:        github.Ptr(name),
			RoleName:    github.Ptr(permission),
			Permissions: CreateRepositoryPermissions([]string{permission}),
		}
	}

	tests := []struct {
		name                 string
		source               []*github.Repository
		destinationDirect    []*github.Repository
		destinationEffective []*github.Repository
		force                bool
		sync                 bool
		wantGrant            int
		wantRevoke           int
		wantWarnings         int
	}{
		{
			name:      "grant missing repository",
			source:    []*github.Repository{repo("alpha", "push")},
			wantGrant: 1,
		},
		{
			name:                 "same direct permission is unchanged",
			source:               []*github.Repository{repo("alpha", "push")},
			destinationDirect:    []*github.Repository{repo("alpha", "push")},
			destinationEffective: []*github.Repository{repo("alpha", "push")},
		},
		{
			name:                 "same inherited permission is not materialized",
			source:               []*github.Repository{repo("alpha", "push")},
			destinationEffective: []*github.Repository{repo("alpha", "push")},
		},
		{
			name:                 "sync adds the source direct permission even when inherited permission matches",
			source:               []*github.Repository{repo("alpha", "push")},
			destinationEffective: []*github.Repository{repo("alpha", "push")},
			sync:                 true,
			wantGrant:            1,
		},
		{
			name:                 "different effective permission warns by default",
			source:               []*github.Repository{repo("alpha", "push")},
			destinationEffective: []*github.Repository{repo("alpha", "pull")},
			wantWarnings:         1,
		},
		{
			name:                 "force updates different effective permission",
			source:               []*github.Repository{repo("alpha", "push")},
			destinationEffective: []*github.Repository{repo("alpha", "pull")},
			force:                true,
			wantGrant:            1,
		},
		{
			name:                 "force updates a different direct permission",
			source:               []*github.Repository{repo("alpha", "push")},
			destinationDirect:    []*github.Repository{repo("alpha", "pull")},
			destinationEffective: []*github.Repository{repo("alpha", "pull")},
			force:                true,
			wantGrant:            1,
		},
		{
			name:                 "sync updates a differing direct permission even if inheritance matches",
			source:               []*github.Repository{repo("alpha", "push")},
			destinationDirect:    []*github.Repository{repo("alpha", "pull")},
			destinationEffective: []*github.Repository{repo("alpha", "push")},
			sync:                 true,
			wantGrant:            1,
		},
		{
			name:                 "sync updates different permission and removes extra direct repository",
			source:               []*github.Repository{repo("alpha", "push")},
			destinationDirect:    []*github.Repository{repo("alpha", "pull"), repo("beta", "push")},
			destinationEffective: []*github.Repository{repo("alpha", "pull"), repo("beta", "push")},
			sync:                 true,
			wantGrant:            1,
			wantRevoke:           1,
		},
		{
			name:                 "sync preserves and warns about extra inherited repository",
			source:               []*github.Repository{},
			destinationEffective: []*github.Repository{repo("beta", "push")},
			sync:                 true,
			wantWarnings:         1,
		},
		{
			name:                 "sync warns when inherited permission differs from source",
			source:               []*github.Repository{repo("alpha", "pull")},
			destinationEffective: []*github.Repository{repo("alpha", "push")},
			sync:                 true,
			wantGrant:            1,
			wantWarnings:         1,
		},
		{
			name:                 "custom roles are compared by name",
			source:               []*github.Repository{repo("alpha", "custom-read")},
			destinationDirect:    []*github.Repository{repo("alpha", "custom-write")},
			destinationEffective: []*github.Repository{repo("alpha", "custom-write")},
			wantWarnings:         1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := PlanTeamRepoPermissionChanges(tt.source, tt.destinationDirect, tt.destinationEffective, tt.force, tt.sync)
			if len(plan.Grant) != tt.wantGrant || len(plan.Revoke) != tt.wantRevoke || len(plan.Warnings) != tt.wantWarnings {
				t.Fatalf("got grant=%d revoke=%d warnings=%d, want grant=%d revoke=%d warnings=%d",
					len(plan.Grant), len(plan.Revoke), len(plan.Warnings), tt.wantGrant, tt.wantRevoke, tt.wantWarnings)
			}
			if tt.name == "custom roles are compared by name" && !strings.Contains(plan.Warnings[0], "different direct permission") {
				t.Fatalf("unexpected custom-role warning: %q", plan.Warnings[0])
			}
		})
	}
}

func TestCompareRepositoryDetectsCustomRoleDifferences(t *testing.T) {
	left := &github.Repository{RoleName: github.Ptr("custom-read"), Permissions: CreateRepositoryPermissions([]string{"pull"})}
	right := &github.Repository{RoleName: github.Ptr("custom-write"), Permissions: CreateRepositoryPermissions([]string{"pull"})}
	if CompareRepository(left, right) == nil {
		t.Fatal("expected different custom roles to produce a diff")
	}
}

func TestTeamRepoPermissionWarnings(t *testing.T) {
	repo := func(name, permission string) *github.Repository {
		return &github.Repository{Name: github.Ptr(name), RoleName: github.Ptr(permission)}
	}

	tests := []struct {
		name        string
		source      []*github.Repository
		effective   []*github.Repository
		skipped     map[string]struct{}
		wantWarning []string
	}{
		{
			name:        "repository absent from source remains accessible",
			effective:   []*github.Repository{repo("alpha", "push")},
			wantWarning: []string{"repository alpha remains accessible after synchronization"},
		},
		{
			name:      "equivalent permissions and case-insensitive names match",
			source:    []*github.Repository{repo("alpha", "write")},
			effective: []*github.Repository{repo("ALPHA", "push")},
		},
		{
			name:        "remaining permission difference is reported",
			source:      []*github.Repository{repo("alpha", "read")},
			effective:   []*github.Repository{repo("alpha", "push")},
			wantWarning: []string{"repository alpha has effective permission push after synchronization; expected pull"},
		},
		{
			name:        "custom permission difference is reported",
			source:      []*github.Repository{repo("alpha", "custom-read")},
			effective:   []*github.Repository{repo("alpha", "custom-write")},
			wantWarning: []string{"repository alpha has effective permission custom-write after synchronization; expected custom-read"},
		},
		{
			name:        "source repository missing from effective access is reported",
			source:      []*github.Repository{repo("alpha", "pull")},
			wantWarning: []string{"repository alpha is missing from destination effective access after synchronization"},
		},
		{
			name:        "repository skipped because it is absent from destination is not reported twice",
			source:      []*github.Repository{repo("alpha", "pull")},
			skipped:     map[string]struct{}{"alpha": {}},
			wantWarning: nil,
		},
		{
			name:        "duplicate effective repositories produce one warning",
			effective:   []*github.Repository{repo("alpha", "push"), repo("ALPHA", "push")},
			wantWarning: []string{"repository alpha remains accessible after synchronization"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := teamRepoPermissionWarnings(tt.source, tt.effective, tt.skipped)
			if len(warnings) != len(tt.wantWarning) {
				t.Fatalf("got warnings %q, want %q", warnings, tt.wantWarning)
			}
			for i := range warnings {
				if warnings[i] != tt.wantWarning[i] {
					t.Fatalf("got warnings %q, want %q", warnings, tt.wantWarning)
				}
			}
		})
	}
}

func TestSyncTeamRepoPermissionsWarnsWhenRevokedRepositoryRemainsAccessible(t *testing.T) {
	g, transport := newFixtureClient(t, map[string]fixtureResponse{
		"GET " + fixtureAPI + "orgs/source-org/teams/source-team/repos?per_page=100": ok(`[]`),
		"GET " + fixtureAPI + "orgs/source-org/teams/source-team":                    ok(`{"slug":"source-team"}`),
		"GET " + fixtureAPI + "orgs/destination-org/teams/destination-team/repos?per_page=100": ok(`[
			{"id":1,"name":"alpha","role_name":"push","owner":{"login":"destination-org"},"permissions":{"push":true,"pull":true}}
		]`),
		"GET " + fixtureAPI + "orgs/destination-org/teams/destination-team": ok(`{"slug":"destination-team","parent":{"slug":"parent-team"}}`),
		"GET " + fixtureAPI + "orgs/destination-org/teams/parent-team/repos?per_page=100": ok(`[
			{"id":1,"name":"alpha","role_name":"push","owner":{"login":"destination-org"},"permissions":{"push":true,"pull":true}}
		]`),
		"GET " + fixtureAPI + "repos/destination-org/alpha/teams?per_page=100": ok(`[
			{"slug":"parent-team","permission":"push"},
			{"slug":"destination-team","permission":"push"}
		]`),
		"DELETE " + fixtureAPI + "orgs/destination-org/teams/destination-team/repos/destination-org/alpha": {
			Status: 204,
		},
	})

	warnings, err := SyncTeamRepoPermissions(
		context.Background(),
		g, repository.Repository{Owner: "source-org"}, "source-team",
		g, repository.Repository{Owner: "destination-org"}, "destination-team",
	)
	if err != nil {
		t.Fatalf("SyncTeamRepoPermissions() error = %v", err)
	}
	want := []string{"repository alpha remains accessible after synchronization"}
	if len(warnings) != len(want) || warnings[0] != want[0] {
		t.Fatalf("SyncTeamRepoPermissions() warnings = %q, want %q", warnings, want)
	}
	foundRevoke := false
	for _, request := range transport.requests {
		if request.Method == "DELETE" && request.URL == fixtureAPI+"orgs/destination-org/teams/destination-team/repos/destination-org/alpha" {
			foundRevoke = true
			break
		}
	}
	if !foundRevoke {
		t.Fatal("SyncTeamRepoPermissions() did not revoke the destination team's direct permission")
	}
}
