package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// GetUser retrieves a user by their username.
func (g *GitHubClient) GetUser(ctx context.Context, username string) (*github.User, error) {
	user, _, err := g.client.Users.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (g *GitHubClient) GetUserByID(ctx context.Context, id int64) (*github.User, error) {
	user, _, err := g.client.Users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (g *GitHubClient) GetUserHovercard(ctx context.Context, username string, subjectType, subjectId string) (*github.Hovercard, error) {
	var opts *github.HovercardOptions
	if subjectType != "" || subjectId != "" {
		opts = &github.HovercardOptions{
			SubjectType: subjectType,
			SubjectID:   subjectId,
		}
	}
	user, _, err := g.client.Users.GetHovercard(ctx, username, opts)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ListFollowers lists the users following user.
func (g *GitHubClient) ListFollowers(ctx context.Context, user string) ([]*github.User, error) {
	opts := &github.ListOptions{}
	return paginate(ctx, opts, 0, func(ctx context.Context) ([]*github.User, *github.Response, error) {
		return g.client.Users.ListFollowers(ctx, user, opts)
	})
}

// ListFollowing lists the users that user is following.
func (g *GitHubClient) ListFollowing(ctx context.Context, user string) ([]*github.User, error) {
	opts := &github.ListOptions{}
	return paginate(ctx, opts, 0, func(ctx context.Context) ([]*github.User, *github.Response, error) {
		return g.client.Users.ListFollowing(ctx, user, opts)
	})
}

// EditUser updates the authenticated user's profile.
func (g *GitHubClient) EditUser(ctx context.Context, body *github.User) (*github.User, error) {
	user, _, err := g.client.Users.Edit(ctx, body)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ListSocialAccounts lists all social accounts for the authenticated user.
func (g *GitHubClient) ListSocialAccounts(ctx context.Context) ([]*github.SocialAccount, error) {
	opts := &github.ListOptions{}
	return paginate(ctx, opts, 0, func(ctx context.Context) ([]*github.SocialAccount, *github.Response, error) {
		return g.client.Users.ListSocialAccounts(ctx, opts)
	})
}

// AddSocialAccounts adds social accounts for the authenticated user.
func (g *GitHubClient) AddSocialAccounts(ctx context.Context, accountURLs []string) ([]*github.SocialAccount, error) {
	accounts, _, err := g.client.Users.AddSocialAccounts(ctx, accountURLs)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}
