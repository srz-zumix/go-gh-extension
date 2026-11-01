package gh

import (
	"testing"

	"github.com/google/go-github/v93/github"
)

func TestWithTeamDirectPermission(t *testing.T) {
	inherited := func() *github.Repository {
		return &github.Repository{
			Name:        github.Ptr("repo"),
			RoleName:    github.Ptr("admin"),
			Permissions: CreateRepositoryPermissions([]string{"pull", "triage", "push", "maintain", "admin"}),
		}
	}

	tests := []struct {
		name         string
		team         *github.Team
		wantRoleName string
		want         github.RepositoryPermissions
		wantAdmin    bool
	}{
		{
			name:         "read alias without permissions map",
			team:         &github.Team{Permission: github.Ptr("read")},
			wantRoleName: "read",
			want:         github.RepositoryPermissions{Pull: github.Ptr(true)},
		},
		{
			name:         "write alias without permissions map",
			team:         &github.Team{Permission: github.Ptr("write")},
			wantRoleName: "write",
			want:         github.RepositoryPermissions{Pull: github.Ptr(true), Triage: github.Ptr(true), Push: github.Ptr(true)},
		},
		{
			name:         "maintain without permissions map",
			team:         &github.Team{Permission: github.Ptr("maintain")},
			wantRoleName: "maintain",
			want:         github.RepositoryPermissions{Pull: github.Ptr(true), Triage: github.Ptr(true), Push: github.Ptr(true), Maintain: github.Ptr(true)},
		},
		{
			name: "direct pull clears inherited admin",
			team: &github.Team{
				Permission:  github.Ptr("pull"),
				Permissions: map[string]bool{"pull": true, "triage": false, "push": false, "maintain": false, "admin": false},
			},
			wantRoleName: "pull",
			want:         github.RepositoryPermissions{Pull: github.Ptr(true)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := withTeamDirectPermission(inherited(), tt.team)
			if got := r.GetRoleName(); got != tt.wantRoleName {
				t.Errorf("RoleName = %q, want %q", got, tt.wantRoleName)
			}
			p := r.GetPermissions()
			checks := []struct {
				name      string
				got, want bool
			}{
				{"admin", p.GetAdmin(), tt.want.GetAdmin()},
				{"maintain", p.GetMaintain(), tt.want.GetMaintain()},
				{"push", p.GetPush(), tt.want.GetPush()},
				{"triage", p.GetTriage(), tt.want.GetTriage()},
				{"pull", p.GetPull(), tt.want.GetPull()},
			}
			for _, c := range checks {
				if c.got != c.want {
					t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
				}
			}
			if len(FilterByRepositoryPermissions([]*github.Repository{r}, []string{"admin"})) != 0 {
				t.Errorf("expected repository to be excluded by admin role filter")
			}
		})
	}
}
