package gh

import (
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterRepositoriesByNamesIgnoresCase(t *testing.T) {
	repos := []*github.Repository{
		{Name: github.Ptr("Example-Repo"), FullName: github.Ptr("Example-Org/Example-Repo")},
		{Name: github.Ptr("other-repo"), FullName: github.Ptr("Example-Org/other-repo")},
	}
	got := FilterRepositoriesByNames(repos, []string{"example-repo"}, "example-org")
	require.Len(t, got, 1)
	assert.Equal(t, "Example-Repo", got[0].GetName())

	got = FilterRepositoriesByNames(repos, []string{"EXAMPLE-ORG/OTHER-REPO"}, "example-org")
	require.Len(t, got, 1)
	assert.Equal(t, "other-repo", got[0].GetName())
}

func TestFilterTeamByNamesIgnoresCase(t *testing.T) {
	teams := []*github.Team{{Slug: github.Ptr("team-a")}, {Slug: github.Ptr("team-b")}}
	got := FilterTeamByNames(teams, []string{"Team-A"})
	require.Len(t, got, 1)
	assert.Equal(t, "team-a", got[0].GetSlug())
}

func TestCompareRepositoriesIgnoresNameCase(t *testing.T) {
	perms := CreateRepositoryPermissions([]string{"pull"})
	left := []*github.Repository{{Name: github.Ptr("Example-Repo"), FullName: github.Ptr("org-a/Example-Repo"), Permissions: perms}}
	right := []*github.Repository{{Name: github.Ptr("example-repo"), FullName: github.Ptr("org-b/example-repo"), Permissions: perms}}
	assert.Empty(t, CompareRepositories(left, right))
}

func TestExcludeUsersIgnoresCase(t *testing.T) {
	users := []*GitHubUser{{Login: github.Ptr("Example-User")}, {Login: github.Ptr("another-user")}}
	got := ExcludeUsers(users, []string{"example-user"})
	require.Len(t, got, 1)
	assert.Equal(t, "another-user", got[0].GetLogin())
}
