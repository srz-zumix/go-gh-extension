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

func RepositoryFromTeamSlugs(owner string, teamSlug string) (repository.Repository, string, error) {
	repo, team := TeamSlugWithHostOwner(teamSlug)

	if repo.Owner == "" {
		repository, err := Repository(RepositoryOwner(owner))
		if err != nil {
			return repository, team, err
		}
		repo.Host = repository.Host
		repo.Owner = repository.Owner
	}
	return repo, team, nil
}
