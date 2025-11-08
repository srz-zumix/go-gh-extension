package client

import (
	"context"

	"github.com/google/go-github/v73/github"
	"github.com/shurcooL/githubv4"
)

func (g *GitHubClient) GetLabel(ctx context.Context, owner, repo, name string) (*github.Label, error) {
	label, _, err := g.client.Issues.GetLabel(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}
	return label, nil
}

func (g *GitHubClient) CreateLabel(ctx context.Context, owner, repo string, label *github.Label) (*github.Label, error) {
	label, _, err := g.client.Issues.CreateLabel(ctx, owner, repo, label)
	if err != nil {
		return nil, err
	}
	return label, nil
}

func (g *GitHubClient) DeleteLabel(ctx context.Context, owner, repo, name string) error {
	_, err := g.client.Issues.DeleteLabel(ctx, owner, repo, name)
	if err != nil {
		return err
	}
	return nil
}

func (g *GitHubClient) EditLabel(ctx context.Context, owner, repo, name string, label *github.Label) (*github.Label, error) {
	label, _, err := g.client.Issues.EditLabel(ctx, owner, repo, name, label)
	if err != nil {
		return nil, err
	}
	return label, nil
}

func (g *GitHubClient) ListLabels(ctx context.Context, owner, repo string) ([]*github.Label, error) {
	var allLabels []*github.Label
	opt := &github.ListOptions{PerPage: 50}

	for {
		labels, resp, err := g.client.Issues.ListLabels(ctx, owner, repo, opt)
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

// GetRepositoryLabelID retrieves the ID of a label in a repository
func (g *GitHubClient) GetRepositoryLabelID(ctx context.Context, owner string, repo string, labelName string) (string, error) {
	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return "", err
	}

	var query struct {
		Repository struct {
			Label struct {
				ID githubv4.String
			} `graphql:"label(name: $labelName)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner":     githubv4.String(owner),
		"repo":      githubv4.String(repo),
		"labelName": githubv4.String(labelName),
	}

	if err := graphql.Query(ctx, &query, variables); err != nil {
		return "", err
	}

	return string(query.Repository.Label.ID), nil
}

// GetRepositoryLabelIDs retrieves the IDs of multiple labels in a repository with a single query
func (g *GitHubClient) GetRepositoryLabelIDs(ctx context.Context, owner string, repo string, labelNames []string) (map[string]string, error) {
	if len(labelNames) == 0 {
		return map[string]string{}, nil
	}

	graphql, err := g.GetOrCreateGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query struct {
		Repository struct {
			Labels struct {
				Nodes []struct {
					ID   githubv4.String
					Name githubv4.String
				}
			} `graphql:"labels(first: 100)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"repo":  githubv4.String(repo),
	}

	if err := graphql.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	// Create a map of label names to label IDs
	labelMap := make(map[string]string)
	for _, label := range query.Repository.Labels.Nodes {
		labelMap[string(label.Name)] = string(label.ID)
	}

	// Filter to only requested labels
	result := make(map[string]string)
	for _, name := range labelNames {
		if id, ok := labelMap[name]; ok {
			result[name] = id
		}
	}

	return result, nil
}
