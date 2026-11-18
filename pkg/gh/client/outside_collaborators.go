package client

import (
	"context"

	"github.com/google/go-github/v93/github"
)

// ListOutsideCollaborators retrieves all outside collaborators of the specified organization.
func (g *GitHubClient) ListOutsideCollaborators(ctx context.Context, org string, filter string) ([]*github.User, error) {
	opts := &github.ListOutsideCollaboratorsOptions{Filter: filter}
	return paginate(ctx, &opts.ListOptions, 0, func(ctx context.Context) ([]*github.User, *github.Response, error) {
		return g.client.Organizations.ListOutsideCollaborators(ctx, org, opts)
	})
}

// RemoveOutsideCollaborator removes an outside collaborator from all repositories of the specified organization.
func (g *GitHubClient) RemoveOutsideCollaborator(ctx context.Context, org string, username string) error {
	_, err := g.client.Organizations.RemoveOutsideCollaborator(ctx, org, username)
	return err
}

// ConvertMemberToOutsideCollaborator converts an organization member to an outside collaborator.
func (g *GitHubClient) ConvertMemberToOutsideCollaborator(ctx context.Context, org string, username string) error {
	_, err := g.client.Organizations.ConvertMemberToOutsideCollaborator(ctx, org, username)
	return err
}
