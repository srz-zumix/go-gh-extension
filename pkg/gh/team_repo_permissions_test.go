package gh

import (
	"strings"
	"testing"

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
