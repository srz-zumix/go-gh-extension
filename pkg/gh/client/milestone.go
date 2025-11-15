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

func (g *GitHubClient) ListMilestones(ctx context.Context, owner, repo string, opts *github.MilestoneListOptions) ([]*github.Milestone, error) {
	var allMilestones []*github.Milestone
	if opts == nil {
		opts = &github.MilestoneListOptions{
			ListOptions: github.ListOptions{PerPage: defaultPerPage},
		}
	} else if opts.PerPage == 0 {
		opts.PerPage = defaultPerPage
	}

	for {
		milestones, resp, err := g.client.Issues.ListMilestones(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allMilestones = append(allMilestones, milestones...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
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
	opt := &github.ListOptions{PerPage: defaultPerPage}

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
