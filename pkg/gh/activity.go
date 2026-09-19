package gh

import (
	"context"
	"time"

	"github.com/google/go-github/v90/github"
)

// ListUserEvents lists events performed by username, most recent first.
// GitHub only returns events from the last 90 days, up to a maximum of 300 events.
// Private events are only included when username is the authenticated user.
func ListUserEvents(ctx context.Context, g *GitHubClient, username string, publicOnly bool) ([]*github.Event, error) {
	return g.ListEventsPerformedByUser(ctx, username, publicOnly)
}

// ListUserNotifications lists notification threads for the authenticated user.
// The GitHub API does not support listing notifications for another user.
func ListUserNotifications(ctx context.Context, g *GitHubClient, since *time.Time) ([]*github.Notification, error) {
	opts := &github.NotificationListOptions{All: true}
	if since != nil {
		opts.Since = *since
	}
	return g.ListNotifications(ctx, opts)
}

// ListUserStarredRepositories lists the repositories starred by username.
func ListUserStarredRepositories(ctx context.Context, g *GitHubClient, username string) ([]*github.StarredRepository, error) {
	return g.ListStarred(ctx, username)
}

// ListUserWatchedRepositories lists the repositories watched by username.
func ListUserWatchedRepositories(ctx context.Context, g *GitHubClient, username string) ([]*github.Repository, error) {
	return g.ListWatched(ctx, username)
}

// ListUserFollowers lists the users following username.
func ListUserFollowers(ctx context.Context, g *GitHubClient, username string) ([]*github.User, error) {
	return g.ListFollowers(ctx, username)
}

// ListUserFollowing lists the users that username is following.
func ListUserFollowing(ctx context.Context, g *GitHubClient, username string) ([]*github.User, error) {
	return g.ListFollowing(ctx, username)
}

// ListUserOrganizations lists the organizations username is a member of.
func ListUserOrganizations(ctx context.Context, g *GitHubClient, username string) ([]*github.Organization, error) {
	return g.ListOrganizationsForUser(ctx, username)
}
