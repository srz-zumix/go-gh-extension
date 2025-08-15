package client

import (
	"context"

	"github.com/google/go-github/v73/github"
)

func (g *GitHubClient) GetMilestone(ctx context.Context, owner, repo string, number int) (*github.Milestone, error) {
	milestone, _, err := g.client.Issues.GetMilestone(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return milestone, nil
}

func (g *GitHubClient) CreateMilestone(ctx context.Context, owner, repo string, milestone *github.Milestone) (*github.Milestone, error) {
	milestone, _, err := g.client.Issues.CreateMilestone(ctx, owner, repo, milestone)
	if err != nil {
		return nil, err
	}
	return milestone, nil
}

func (g *GitHubClient) ListMilestones(ctx context.Context, owner, repo, state, sort, direction string) ([]*github.Milestone, error) {
	var allMilestones []*github.Milestone
	if state == "" {
		state = "open"
	}
	if sort == "" {
		sort = "due_on"
	}
	if direction == "" {
		direction = "asc"
	}
	opt := &github.MilestoneListOptions{
		State:     state,
		Sort:      sort,
		Direction: direction,
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}

	for {
		milestones, resp, err := g.client.Issues.ListMilestones(ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allMilestones = append(allMilestones, milestones...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allMilestones, nil
}
func (g *GitHubClient) EditMilestone(ctx context.Context, owner, repo string, number int, milestone *github.Milestone) (*github.Milestone, error) {
	milestone, _, err := g.client.Issues.EditMilestone(ctx, owner, repo, number, milestone)
	if err != nil {
		return nil, err
	}
	return milestone, nil
}

func (g *GitHubClient) DeleteMilestone(ctx context.Context, owner, repo string, number int) error {
	_, err := g.client.Issues.DeleteMilestone(ctx, owner, repo, number)
	if err != nil {
		return err
	}
	return nil
}

func (g *GitHubClient) ListLabelsForMilestone(ctx context.Context, owner, repo string, number int) ([]*github.Label, error) {
	var allLabels []*github.Label
	opt := &github.ListOptions{PerPage: 50}

	for {
		labels, resp, err := g.client.Issues.ListLabelsForMilestone(ctx, owner, repo, number, opt)
		if err != nil {
			return nil, err
		}
		allLabels = append(allLabels, labels...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allLabels, nil
}
