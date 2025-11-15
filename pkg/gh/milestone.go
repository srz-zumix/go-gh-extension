package gh

import (
	"context"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
)

type MilestoneListOptions struct {
	// State filters milestones based on their state. Possible values are:
	// open, closed, all. Default is "open".
	State string

	// Sort specifies how to sort milestones. Possible values are: due_on, completeness.
	// Default value is "due_on".
	Sort string

	// Direction in which to sort milestones. Possible values are: asc, desc.
	// Default is "asc".
	Direction string
}

// ListMilestones lists all milestones in a repository
func ListMilestones(ctx context.Context, g *GitHubClient, repo repository.Repository, options *MilestoneListOptions) ([]*github.Milestone, error) {
	opts := &github.MilestoneListOptions{}
	if options != nil {
		opts.State = options.State
		opts.Sort = options.Sort
		opts.Direction = options.Direction
	}
	return g.ListMilestones(ctx, repo.Owner, repo.Name, opts)
}

// GetMilestone gets a single milestone by number for a repository (wrapper)
func GetMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, number int) (*github.Milestone, error) {
	return g.GetMilestone(ctx, repo.Owner, repo.Name, number)
}

// CreateMilestone creates a new milestone for a repository (wrapper)
func CreateMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, milestone *github.Milestone) (*github.Milestone, error) {
	return g.CreateMilestone(ctx, repo.Owner, repo.Name, milestone)
}

// EditMilestone edits an existing milestone for a repository (wrapper)
func EditMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, number int, milestone *github.Milestone) (*github.Milestone, error) {
	return g.EditMilestone(ctx, repo.Owner, repo.Name, number, milestone)
}

// DeleteMilestone deletes a milestone by number for a repository (wrapper)
func DeleteMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, number int) error {
	return g.DeleteMilestone(ctx, repo.Owner, repo.Name, number)
}

func ListLabelsForMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, number int) ([]*github.Label, error) {
	return g.ListLabelsForMilestone(ctx, repo.Owner, repo.Name, number)
}
