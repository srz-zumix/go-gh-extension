package parser

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
)

func TestTeamSlugWithHostOwner(t *testing.T) {
	tests := []struct {
		name         string
		teamSlug     string
		wantRepo     repository.Repository
		wantTeamSlug string
	}{
		{
			name:         "team slug only",
			teamSlug:     "team-name",
			wantRepo:     repository.Repository{},
			wantTeamSlug: "team-name",
		},
		{
			name:     "owner/team format",
			teamSlug: "my-org/team-name",
			wantRepo: repository.Repository{
				Owner: "my-org",
			},
			wantTeamSlug: "team-name",
		},
		{
			name:     "host/owner/team format",
			teamSlug: "github.com/my-org/team-name",
			wantRepo: repository.Repository{
				Host:  "github.com",
				Owner: "my-org",
			},
			wantTeamSlug: "team-name",
		},
		{
			name:     "GitHub Enterprise host/owner/team format",
			teamSlug: "github.example.com/my-org/team-name",
			wantRepo: repository.Repository{
				Host:  "github.example.com",
				Owner: "my-org",
			},
			wantTeamSlug: "team-name",
		},
		{
			name:     "owner with hyphen",
			teamSlug: "my-org-name/team-name",
			wantRepo: repository.Repository{
				Owner: "my-org-name",
			},
			wantTeamSlug: "team-name",
		},
		{
			name:     "team with hyphen",
			teamSlug: "my-org/my-team-name",
			wantRepo: repository.Repository{
				Owner: "my-org",
			},
			wantTeamSlug: "my-team-name",
		},
		{
			name:     "team with multiple slashes in name (only first 3 parts matter)",
			teamSlug: "github.com/my-org/team-name/extra",
			wantRepo: repository.Repository{
				Host:  "github.com",
				Owner: "my-org",
			},
			wantTeamSlug: "team-name/extra",
		},
		{
			name:         "empty string",
			teamSlug:     "",
			wantRepo:     repository.Repository{},
			wantTeamSlug: "",
		},
		{
			name:     "single slash only",
			teamSlug: "/",
			wantRepo: repository.Repository{
				Owner: "",
			},
			wantTeamSlug: "",
		},
		{
			name:     "double slash only",
			teamSlug: "//",
			wantRepo: repository.Repository{
				Host:  "",
				Owner: "",
			},
			wantTeamSlug: "",
		},
		{
			name:     "owner with trailing slash",
			teamSlug: "my-org/",
			wantRepo: repository.Repository{
				Owner: "my-org",
			},
			wantTeamSlug: "",
		},
		{
			name:     "host/owner with trailing slash",
			teamSlug: "github.com/my-org/",
			wantRepo: repository.Repository{
				Host:  "github.com",
				Owner: "my-org",
			},
			wantTeamSlug: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRepo, gotTeamSlug := TeamSlugWithHostOwner(tt.teamSlug)

			if gotRepo.Host != tt.wantRepo.Host {
				t.Errorf("TeamSlugWithHostOwner() Host = %v, want %v", gotRepo.Host, tt.wantRepo.Host)
			}
			if gotRepo.Owner != tt.wantRepo.Owner {
				t.Errorf("TeamSlugWithHostOwner() Owner = %v, want %v", gotRepo.Owner, tt.wantRepo.Owner)
			}
			if gotTeamSlug != tt.wantTeamSlug {
				t.Errorf("TeamSlugWithHostOwner() teamSlug = %v, want %v", gotTeamSlug, tt.wantTeamSlug)
			}
		})
	}
}

func TestRepositoryFromTeamSlugs(t *testing.T) {
	tests := []struct {
		name         string
		owner        string
		teamSlug     string
		wantRepo     repository.Repository
		wantTeamSlug string
		wantErr      bool
	}{
		{
			name:     "team slug with owner",
			owner:    "",
			teamSlug: "my-org/team-name",
			wantRepo: repository.Repository{
				Owner: "my-org",
			},
			wantTeamSlug: "team-name",
		},
		{
			name:     "team slug with host/owner",
			owner:    "",
			teamSlug: "github.com/my-org/team-name",
			wantRepo: repository.Repository{
				Host:  "github.com",
				Owner: "my-org",
			},
			wantTeamSlug: "team-name",
		},
		{
			name:     "team slug only with owner parameter",
			owner:    "my-org",
			teamSlug: "team-name",
			wantRepo: repository.Repository{
				Host:  "github.com",
				Owner: "my-org",
			},
			wantTeamSlug: "team-name",
		},
		{
			name:     "team slug with owner/repo format in owner parameter",
			owner:    "my-org/my-repo",
			teamSlug: "team-name",
			wantErr:  true,
		},
		{
			name:     "team slug overrides owner parameter",
			owner:    "other-org",
			teamSlug: "my-org/team-name",
			wantRepo: repository.Repository{
				Owner: "my-org",
			},
			wantTeamSlug: "team-name",
		},
		{
			name:     "empty team slug with owner parameter",
			owner:    "my-org",
			teamSlug: "",
			wantRepo: repository.Repository{
				Host:  "github.com",
				Owner: "my-org",
			},
			wantTeamSlug: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRepo, gotTeamSlug, err := RepositoryFromTeamSlugs(tt.owner, tt.teamSlug)
			if (err != nil) != tt.wantErr {
				t.Errorf("RepositoryFromTeamSlugs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Only compare Host and Owner as Name is not set in team context
			if gotRepo.Host != tt.wantRepo.Host {
				t.Errorf("RepositoryFromTeamSlugs() Host = %v, want %v", gotRepo.Host, tt.wantRepo.Host)
			}
			if gotRepo.Owner != tt.wantRepo.Owner {
				t.Errorf("RepositoryFromTeamSlugs() Owner = %v, want %v", gotRepo.Owner, tt.wantRepo.Owner)
			}
			if gotTeamSlug != tt.wantTeamSlug {
				t.Errorf("RepositoryFromTeamSlugs() teamSlug = %v, want %v", gotTeamSlug, tt.wantTeamSlug)
			}
		})
	}
}
