package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// MilestoneParameter represents parameters for creating or editing a milestone
type MilestoneParameter struct {
	Title 	 	*string
	State	 	*string
	Description *string
	DueOn    	*github.Timestamp
}

// ToGitHubMilestone converts MilestoneParameter to github.Milestone
func (mp *MilestoneParameter) ToGitHubMilestone() *github.Milestone {
	if mp == nil {
		return &github.Milestone{}
	}
	return &github.Milestone{
		Title:       mp.Title,
		State:       mp.State,
		Description: mp.Description,
		DueOn:       mp.DueOn,
	}
}

// MilestoneListOptions represents options for listing milestones
type MilestoneListOptions struct {
	// State filters milestones based on their state. Possible values are:
	// open, closed, all. Default is "open".
	State		string

	// Sort specifies how to sort milestones. Possible values are: due_on, completeness.
	// Default value is "due_on".
	Sort		string

	// Direction in which to sort milestones. Possible values are: asc, desc.
	// Default is "asc".
	Direction	string
}

func (opts *MilestoneListOptions) ToGitHubOptions() *github.MilestoneListOptions {
	if opts == nil {
		return &github.MilestoneListOptions{}
	}
	return &github.MilestoneListOptions{
		State:     opts.State,
		Sort:      opts.Sort,
		Direction: opts.Direction,
	}
}

func GetMilestoneNumber(milestone any) (int, error) {
	switch t := milestone.(type) {
	case string:
		return parser.GetMilestoneNumberFromString(t)
	case int:
		return t, nil
	case *github.Milestone:
		return t.GetNumber(), nil
	default:
		return 0, fmt.Errorf("unsupported milestone type: %T", milestone)
	}
}

// ListMilestones lists all milestones in a repository
func ListMilestones(ctx context.Context, g *GitHubClient, repo repository.Repository, options *MilestoneListOptions) ([]*github.Milestone, error) {
	opts := options.ToGitHubOptions()
	return g.ListMilestones(ctx, repo.Owner, repo.Name, opts)
}

// GetMilestone gets a single milestone by number for a repository (wrapper)
func GetMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, milestone any) (*github.Milestone, error) {
	number, err := GetMilestoneNumber(milestone)
	if err != nil {
		return nil, err
	}
	return g.GetMilestone(ctx, repo.Owner, repo.Name, number)
}

// CreateMilestone creates a new milestone for a repository (wrapper)
func CreateMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, milestone *github.Milestone) (*github.Milestone, error) {
	return g.CreateMilestone(ctx, repo.Owner, repo.Name, milestone)
}

// EditMilestone edits an existing milestone for a repository (wrapper)
func EditMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, milestone any, edit *github.Milestone) (*github.Milestone, error) {
	number, err := GetMilestoneNumber(milestone)
	if err != nil {
		return nil, err
	}
	return g.EditMilestone(ctx, repo.Owner, repo.Name, number, edit)
}

// DeleteMilestone deletes a milestone by number for a repository (wrapper)
func DeleteMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, milestone any) error {
	number, err := GetMilestoneNumber(milestone)
	if err != nil {
		return err
	}
	return g.DeleteMilestone(ctx, repo.Owner, repo.Name, number)
}

func ListLabelsForMilestone(ctx context.Context, g *GitHubClient, repo repository.Repository, milestone any) ([]*github.Label, error) {
	number, err := GetMilestoneNumber(milestone)
	if err != nil {
		return nil, err
	}
	return g.ListLabelsForMilestone(ctx, repo.Owner, repo.Name, number)
}
