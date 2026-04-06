package parser

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

func TeamSlugWithHostOwner(teamSlug string) (repository.Repository, string) {
	teamURL, err := ParseTeamURL(teamSlug)
	if err == nil && teamURL != nil {
		return repository.Repository{Host: teamURL.Host, Owner: teamURL.Org}, teamURL.TeamSlug
	}
	parts := strings.SplitN(teamSlug, "/", 3)
	if len(parts) == 2 {
		return repository.Repository{Owner: parts[0]}, parts[1]
	}
	if len(parts) == 3 {
		return repository.Repository{Host: parts[0], Owner: parts[1]}, parts[2]
	}
	return repository.Repository{}, teamSlug
}

// Deprecated: Use RepositoryWithTeamSlugs with RepositoryOwner(owner) instead.
func RepositoryFromTeamSlugs(owner string, teamSlug string) (repository.Repository, string, error) {
	return RepositoryWithTeamSlugs(teamSlug, RepositoryOwner(owner))
}

// RepositoryWithTeamSlugs resolves a team slug against RepositoryOptions instead of a plain
// owner string. The owner (and host) are taken from the resolved repository when the team slug
// does not already carry host/owner information.
func RepositoryWithTeamSlugs(teamSlug string, opts ...RepositoryOption) (repository.Repository, string, error) {
	repo, team := TeamSlugWithHostOwner(teamSlug)

	if repo.Owner == "" {
		r, err := Repository(opts...)
		if err != nil {
			return repo, team, err
		}
		repo.Host = r.Host
		repo.Owner = r.Owner
	}
	return repo, team, nil
}
