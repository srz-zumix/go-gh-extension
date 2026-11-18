package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// OutsideCollaboratorFilterList is the list of valid filter values for listing outside collaborators.
var OutsideCollaboratorFilterList = []string{
	"all",
	"2fa_disabled",
	"2fa_insecure",
}

// ListOutsideCollaborators lists outside collaborators of the organization specified by repo.Owner.
func ListOutsideCollaborators(ctx context.Context, g *GitHubClient, repo repository.Repository, filter string) ([]*GitHubUser, error) {
	return g.ListOutsideCollaborators(ctx, repo.Owner, filter)
}

// RemoveOutsideCollaborator removes an outside collaborator from all repositories of the organization specified by repo.Owner.
func RemoveOutsideCollaborator(ctx context.Context, g *GitHubClient, repo repository.Repository, username string) error {
	return g.RemoveOutsideCollaborator(ctx, repo.Owner, username)
}

// ConvertMemberToOutsideCollaborator converts a member of the organization specified by repo.Owner to an outside collaborator.
func ConvertMemberToOutsideCollaborator(ctx context.Context, g *GitHubClient, repo repository.Repository, username string) error {
	return g.ConvertMemberToOutsideCollaborator(ctx, repo.Owner, username)
}
