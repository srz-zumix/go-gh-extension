package gh

import (
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
)

func TestGetObjectName(t *testing.T) {
	tests := []struct {
		name     string
		item     any
		expected string
	}{
		{
			name:     "Repository",
			item:     &github.Repository{FullName: github.Ptr("owner/repo")},
			expected: "owner/repo",
		},
		{
			name:     "Team",
			item:     &github.Team{Slug: github.Ptr("team-slug")},
			expected: "team-slug",
		},
		{
			name:     "User",
			item:     &github.User{Login: github.Ptr("username")},
			expected: "username",
		},
		{
			name:     "CustomOrgRoles",
			item:     &github.CustomOrgRoles{Name: github.Ptr("role-name")},
			expected: "role-name",
		},
		{
			name:     "RepositoryPermissionLevel",
			item:     &github.RepositoryPermissionLevel{User: &github.User{Login: github.Ptr("user-login")}},
			expected: "user-login",
		},
		{
			name: "RepositoryPermissionLevel (custom type)",
			item: &RepositoryPermissionLevel{Repository: repository.Repository{Owner: "owner", Name: "repo"}},
			expected: "owner/repo",
		},
		{
			name:     "Label",
			item:     &github.Label{Name: github.Ptr("bug")},
			expected: "bug",
		},
		{
			name:     "RepoDependencies",
			item:     &github.RepoDependencies{Name: github.Ptr("npm:lodash")},
			expected: "lodash",
		},
		{
			name:     "SBOM",
			item:     &github.SBOM{SBOM: &github.SBOMInfo{Name: github.Ptr("sbom-name")}},
			expected: "sbom-name",
		},
		{
			name:     "SBOMInfo",
			item:     &github.SBOMInfo{Name: github.Ptr("sbom-info-name")},
			expected: "sbom-info-name",
		},
		{
			name:     "Unknown type",
			item:     "unknown",
			expected: "",
		},
		{
			name:     "Nil",
			item:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetObjectName(tt.item); got != tt.expected {
				t.Errorf("GetObjectName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetObjectNames(t *testing.T) {
	tests := []struct {
		name     string
		items    any
		expected []string
	}{
		{
			name: "Repositories",
			items: []*github.Repository{
				{FullName: github.Ptr("owner/repo1")},
				{FullName: github.Ptr("owner/repo2")},
			},
			expected: []string{"owner/repo1", "owner/repo2"},
		},
		{
			name: "Teams",
			items: []*github.Team{
				{Slug: github.Ptr("team1")},
				{Slug: github.Ptr("team2")},
			},
			expected: []string{"team1", "team2"},
		},
		{
			name: "Users",
			items: []*github.User{
				{Login: github.Ptr("user1")},
				{Login: github.Ptr("user2")},
			},
			expected: []string{"user1", "user2"},
		},
		{
			name: "CustomOrgRoles",
			items: []*github.CustomOrgRoles{
				{Name: github.Ptr("role1")},
				{Name: github.Ptr("role2")},
			},
			expected: []string{"role1", "role2"},
		},
		{
			name: "Labels",
			items: []*github.Label{
				{Name: github.Ptr("bug")},
				{Name: github.Ptr("feature")},
			},
			expected: []string{"bug", "feature"},
		},
		{
			name: "RepoDependencies",
			items: []*github.RepoDependencies{
				{Name: github.Ptr("npm:lodash")},
				{Name: github.Ptr("npm:react")},
			},
			expected: []string{"lodash", "react"},
		},
		{
			name:     "Empty slice",
			items:    []*github.Repository{},
			expected: []string{},
		},
		{
			name:     "Unknown type",
			items:    []string{"unknown"},
			expected: nil,
		},
		{
			name:     "Nil",
			items:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetObjectNames(tt.items)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("GetObjectNames() = %v, want nil", got)
				}
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("GetObjectNames() length = %d, want %d", len(got), len(tt.expected))
				return
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("GetObjectNames()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}
