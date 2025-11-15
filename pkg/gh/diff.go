package gh

import (
	"fmt"

	"github.com/google/go-github/v79/github"
)

// RepositoryPermissionsDiff represents the diff between two repositories.
type RepositoryPermissionsDiff struct {
	Left  *github.Repository
	Right *github.Repository
}

func (d *RepositoryPermissionsDiff) GetName() string {
	if d.Left != nil {
		return *d.Left.Name
	}
	if d.Right != nil {
		return *d.Right.Name
	}
	return ""
}

func (d *RepositoryPermissionsDiff) GetFullName() string {
	if d.Left != nil {
		return *d.Left.FullName
	}
	if d.Right != nil {
		return *d.Right.FullName
	}
	return ""
}

type RepositoryPermissionsDiffs []RepositoryPermissionsDiff

func (d RepositoryPermissionsDiffs) Left() []*github.Repository {
	var repos []*github.Repository
	for _, diff := range d {
		if diff.Left != nil {
			repos = append(repos, diff.Left)
		}
	}
	return repos
}

func (d RepositoryPermissionsDiffs) Right() []*github.Repository {
	var repos []*github.Repository
	for _, diff := range d {
		if diff.Right != nil {
			repos = append(repos, diff.Right)
		}
	}
	return repos
}

func CompareRepository(left, right *github.Repository) *RepositoryPermissionsDiff {
	if GetRepositoryPermissions(left) == GetRepositoryPermissions(right) {
		return nil
	}
	diff := RepositoryPermissionsDiff{
		Left:  left,
		Right: right,
	}
	return &diff
}

func CompareRepositories(left, right []*github.Repository) RepositoryPermissionsDiffs {
	var diffs RepositoryPermissionsDiffs
	rightMap := make(map[string]*github.Repository)

	// Map right repositories by their name
	for _, r := range right {
		if r.Name != nil {
			rightMap[*r.Name] = r
		}
	}

	// Compare repositories in left with rightMap
	for _, l := range left {
		if l.Name == nil {
			continue
		}
		r := rightMap[*l.Name]
		diff := CompareRepository(l, r)
		if diff != nil {
			diffs = append(diffs, *diff)
		}
		delete(rightMap, *l.Name) // Remove matched repository from rightMap
	}

	// Add remaining repositories in rightMap as differences
	for _, r := range rightMap {
		diffs = append(diffs, RepositoryPermissionsDiff{
			Left:  nil,
			Right: r,
		})
	}

	return diffs
}

// TeamPermissionsDiff represents the differences in permissions for a team between two repositories.
type TeamPermissionsDiff struct {
	Left  *github.Team
	Right *github.Team
}

// TeamPermissionsDiffs represents a collection of TeamPermissionsDiff.
type TeamPermissionsDiffs []TeamPermissionsDiff

func (d *TeamPermissionsDiff) GetSlug() string {
	if d.Left != nil {
		return *d.Left.Slug
	}
	if d.Right != nil {
		return *d.Right.Slug
	}
	return ""
}

// ComparePermissions compares the permissions of two teams and returns the difference as a TeamPermissionsDiff.
// If the left and right teams are different, it returns an error.
// If the permissions match, it returns nil.
func ComparePermissions(leftTeam, rightTeam *github.Team) (*TeamPermissionsDiff, error) {
	if leftTeam != nil && rightTeam != nil && leftTeam.GetSlug() != rightTeam.GetSlug() {
		return nil, fmt.Errorf("team mismatch: left team is %s, right team is %s", leftTeam.GetSlug(), rightTeam.GetSlug())
	}

	var leftPerm, rightPerm string

	if leftTeam != nil && leftTeam.Permission != nil {
		leftPerm = *leftTeam.Permission
	}

	if rightTeam != nil && rightTeam.Permission != nil {
		rightPerm = *rightTeam.Permission
	}

	if leftPerm == rightPerm {
		return nil, nil
	}

	return &TeamPermissionsDiff{
		Left:  leftTeam,
		Right: rightTeam,
	}, nil
}

// CompareTeamsPermissions compares the permissions of two slices of teams and returns the differences as a slice of TeamPermissionsDiff.
func CompareTeamsPermissions(leftTeams, rightTeams []*github.Team) (TeamPermissionsDiffs, error) {
	diffs := TeamPermissionsDiffs{}

	// Create a map for quick lookup of right teams by team slug
	rightMap := make(map[string]*github.Team)
	for _, team := range rightTeams {
		if team.Slug != nil {
			rightMap[*team.Slug] = team
		}
	}

	// Compare teams in leftTeams with rightMap
	for _, leftTeam := range leftTeams {
		if leftTeam.Slug == nil {
			continue
		}
		rightTeam := rightMap[*leftTeam.Slug]
		diff, err := ComparePermissions(leftTeam, rightTeam)
		if err != nil {
			return nil, fmt.Errorf("error comparing permissions for team %s: %w", *leftTeam.Slug, err)
		}
		if diff != nil {
			diffs = append(diffs, *diff)
		}
		delete(rightMap, *leftTeam.Slug) // Remove matched team from rightMap
	}

	// Add remaining teams in rightMap as differences
	for _, rightTeam := range rightMap {
		diffs = append(diffs, TeamPermissionsDiff{
			Left:  nil,
			Right: rightTeam,
		})
	}

	return diffs, nil
}
