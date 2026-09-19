package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// ListEventsPerformedByUser lists public (and, when authenticated as user, private)
// events performed by user, most recent first.
// GitHub only returns events from the last 90 days, up to a maximum of 300 events.
func (g *GitHubClient) ListEventsPerformedByUser(ctx context.Context, user string, publicOnly bool) ([]*github.Event, error) {
	opts := &github.ListOptions{}
	return paginate(ctx, opts, 0, func(ctx context.Context) ([]*github.Event, *github.Response, error) {
		return g.client.Activity.ListEventsPerformedByUser(ctx, user, publicOnly, opts)
	})
}

// ListNotifications lists notification threads for the authenticated user.
// The GitHub API does not support listing notifications for another user.
func (g *GitHubClient) ListNotifications(ctx context.Context, opts *github.NotificationListOptions) ([]*github.Notification, error) {
	if opts == nil {
		opts = &github.NotificationListOptions{}
	}
	return paginate(ctx, &opts.ListOptions, 0, func(ctx context.Context) ([]*github.Notification, *github.Response, error) {
		return g.client.Activity.ListNotifications(ctx, opts)
	})
}

// ListStarred lists the repositories starred by user, including the time they were starred.
func (g *GitHubClient) ListStarred(ctx context.Context, user string) ([]*github.StarredRepository, error) {
	opts := &github.ActivityListStarredOptions{}
	return paginate(ctx, &opts.ListOptions, 0, func(ctx context.Context) ([]*github.StarredRepository, *github.Response, error) {
		return g.client.Activity.ListStarred(ctx, user, opts)
	})
}

// ListWatched lists the repositories watched by user.
func (g *GitHubClient) ListWatched(ctx context.Context, user string) ([]*github.Repository, error) {
	opts := &github.ListOptions{}
	return paginate(ctx, opts, 0, func(ctx context.Context) ([]*github.Repository, *github.Response, error) {
		return g.client.Activity.ListWatched(ctx, user, opts)
	})
}
