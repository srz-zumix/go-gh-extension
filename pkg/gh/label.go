package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v71/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

func GetLabel(ctx context.Context, g *client.GitHubClient, repo repository.Repository, name string) (*github.Label, error) {
	label, err := g.GetLabel(ctx, repo.Owner, repo.Name, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get label %s: %w", name, err)
	}
	return label, nil
}

func CreateLabel(ctx context.Context, g *client.GitHubClient, repo repository.Repository, label *github.Label) (*github.Label, error) {
	createdLabel, err := g.CreateLabel(ctx, repo.Owner, repo.Name, label)
	if err != nil {
		return nil, fmt.Errorf("failed to create label %s: %w", *label.Name, err)
	}
	return createdLabel, nil
}

func DeleteLabel(ctx context.Context, g *client.GitHubClient, repo repository.Repository, name string) error {
	if err := g.DeleteLabel(ctx, repo.Owner, repo.Name, name); err != nil {
		return fmt.Errorf("failed to delete label %s: %w", name, err)
	}
	return nil
}

func EditLabel(ctx context.Context, g *client.GitHubClient, repo repository.Repository, name string, label *github.Label) (*github.Label, error) {
	editedLabel, err := g.EditLabel(ctx, repo.Owner, repo.Name, name, label)
	if err != nil {
		return nil, fmt.Errorf("failed to edit label %s: %w", name, err)
	}
	return editedLabel, nil
}

func ListLabels(ctx context.Context, g *client.GitHubClient, repo repository.Repository) ([]*github.Label, error) {
	labels, err := g.ListLabels(ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}
	return labels, nil
}

// CopyLabels copies all labels from src repository to dst repository.
// If force is true, existing labels in dst will be updated.
func CopyLabels(ctx context.Context, g *client.GitHubClient, src, dst repository.Repository, force bool) error {
	srcLabels, err := g.ListLabels(ctx, src.Owner, src.Name)
	if err != nil {
		return fmt.Errorf("failed to list labels from source: %w", err)
	}
	dstLabels, err := g.ListLabels(ctx, dst.Owner, dst.Name)
	if err != nil {
		return fmt.Errorf("failed to list labels from destination: %w", err)
	}
	dstLabelMap := make(map[string]*github.Label)
	for _, l := range dstLabels {
		if l.Name != nil {
			dstLabelMap[*l.Name] = l
		}
	}
	for _, label := range srcLabels {
		if label.Name == nil {
			continue
		}
		if _, exists := dstLabelMap[*label.Name]; exists {
			if force {
				_, err := g.EditLabel(ctx, dst.Owner, dst.Name, *label.Name, label)
				if err != nil {
					return fmt.Errorf("failed to update label %s: %w", *label.Name, err)
				}
			}
			continue
		}
		_, err := g.CreateLabel(ctx, dst.Owner, dst.Name, label)
		if err != nil {
			return fmt.Errorf("failed to create label %s: %w", *label.Name, err)
		}
	}
	return nil
}
