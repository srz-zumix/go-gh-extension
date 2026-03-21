package client

import (
	"context"

	"github.com/google/go-github/v84/github"
)

// ListGists lists gists for the specified user. Pass an empty string for the
// authenticated user's gists.
func (g *GitHubClient) ListGists(ctx context.Context, username string) ([]*github.Gist, error) {
	allGists := []*github.Gist{}
	opts := &github.GistListOptions{ListOptions: github.ListOptions{PerPage: defaultPerPage}}
	for {
		gists, resp, err := g.client.Gists.List(ctx, username, opts)
		if err != nil {
			return nil, err
		}
		allGists = append(allGists, gists...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allGists, nil
}

// GetGist retrieves a single gist by its ID.
func (g *GitHubClient) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	gist, _, err := g.client.Gists.Get(ctx, gistID)
	if err != nil {
		return nil, err
	}
	return gist, nil
}

// CreateGist creates a new gist.
func (g *GitHubClient) CreateGist(ctx context.Context, gist *github.Gist) (*github.Gist, error) {
	created, _, err := g.client.Gists.Create(ctx, gist)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// DeleteGist deletes a gist by its ID.
func (g *GitHubClient) DeleteGist(ctx context.Context, gistID string) error {
	_, err := g.client.Gists.Delete(ctx, gistID)
	return err
}
