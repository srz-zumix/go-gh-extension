package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// ListMilestones lists all milestones for a repository (wrapper)
func ListMilestones(ctx context.Context, g *client.GitHubClient, repo repository.Repository, state, sort, direction string) ([]*github.Milestone, error) {
	return g.ListMilestones(ctx, repo.Owner, repo.Name, state, sort, direction)
}

// GetMilestone gets a single milestone by number for a repository (wrapper)
func GetMilestone(ctx context.Context, g *client.GitHubClient, repo repository.Repository, number int) (*github.Milestone, error) {
	return g.GetMilestone(ctx, repo.Owner, repo.Name, number)
}

// CreateMilestone creates a new milestone for a repository (wrapper)
func CreateMilestone(ctx context.Context, g *client.GitHubClient, repo repository.Repository, milestone *github.Milestone) (*github.Milestone, error) {
	return g.CreateMilestone(ctx, repo.Owner, repo.Name, milestone)
}

// EditMilestone edits an existing milestone for a repository (wrapper)
func EditMilestone(ctx context.Context, g *client.GitHubClient, repo repository.Repository, number int, milestone *github.Milestone) (*github.Milestone, error) {
	return g.EditMilestone(ctx, repo.Owner, repo.Name, number, milestone)
}

// DeleteMilestone deletes a milestone by number for a repository (wrapper)
func DeleteMilestone(ctx context.Context, g *client.GitHubClient, repo repository.Repository, number int) error {
	return g.DeleteMilestone(ctx, repo.Owner, repo.Name, number)
}

func ListLabelsForMilestone(ctx context.Context, g *client.GitHubClient, repo repository.Repository, number int) ([]*github.Label, error) {
	return g.ListLabelsForMilestone(ctx, repo.Owner, repo.Name, number)
}
