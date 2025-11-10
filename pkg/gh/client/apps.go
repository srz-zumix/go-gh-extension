package client

import (
	"context"

	"github.com/google/go-github/v73/github"
)

// ListAppInstallations retrieves all app installations for the authenticated app.
func (g *GitHubClient) ListAppInstallations(ctx context.Context) ([]*github.Installation, error) {
	var allInstallations []*github.Installation
	opt := &github.ListOptions{PerPage: defaultPerPage}

	for {
		installations, resp, err := g.client.Apps.ListInstallations(ctx, opt)
		if err != nil {
			return nil, err
		}
		allInstallations = append(allInstallations, installations...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allInstallations, nil
}

func (g *GitHubClient) GetAppInstallation(ctx context.Context, installationID int64) (*github.Installation, error) {
	installation, _, err := g.client.Apps.GetInstallation(ctx, installationID)
	if err != nil {
		return nil, err
	}
	return installation, nil
}
