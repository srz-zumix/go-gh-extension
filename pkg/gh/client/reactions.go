package client

import (
	"context"

	"github.com/google/go-github/v90/github"
)

// CreatePullRequestCommentReaction creates a reaction for a pull request review comment.
// The content should have one of the following values: "+1", "-1", "laugh",
// "confused", "heart", "hooray", "rocket", or "eyes".
func (g *GitHubClient) CreatePullRequestCommentReaction(ctx context.Context, owner string, repo string, commentID int64, content string) (*github.Reaction, error) {
	reaction, _, err := g.client.Reactions.CreatePullRequestCommentReaction(ctx, owner, repo, commentID, content)
	if err != nil {
		return nil, err
	}
	return reaction, nil
}
