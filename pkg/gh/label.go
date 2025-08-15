package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v73/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

func GetLabel(ctx context.Context, g *client.GitHubClient, repo repository.Repository, name string) (*github.Label, error) {
	label, err := g.GetLabel(ctx, repo.Owner, repo.Name, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get label %s: %w", name, err)
	}
	return label, nil
}

func CreateLabel(ctx context.Context, g *client.GitHubClient, repo repository.Repository, name, description, color *string) (*github.Label, error) {
	label := &github.Label{
		Name:        name,
		Description: description,
		Color:       color,
	}
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

func DeleteUnusedLabel(ctx context.Context, g *client.GitHubClient, repo repository.Repository, name string) (*github.Label, error) {
	issues, err := g.SearchIssues(ctx, fmt.Sprintf("repo:%s/%s label:%s", repo.Owner, repo.Name, name))
	if err != nil {
		return nil, fmt.Errorf("failed to search issues with label %s: %w", name, err)
	}
	if len(issues) > 0 {
		return GetLabel(ctx, g, repo, name)
	}
	return nil, DeleteLabel(ctx, g, repo, name)
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

// SyncLabels synchronizes all labels from the src repository to the dst repository.
// Existing labels are updated, new labels are created, and labels that exist only in dst are not deleted.
func SyncLabels(ctx context.Context, g *client.GitHubClient, src, dst repository.Repository, force bool) error {
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
		if dstLabel, exists := dstLabelMap[*label.Name]; exists {
			// update if any field differs
			if !equalLabel(label, dstLabel) {
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
	// Delete labels from dst that do not exist in src
	srcLabelMap := make(map[string]struct{})
	for _, label := range srcLabels {
		if label.Name != nil {
			srcLabelMap[*label.Name] = struct{}{}
		}
	}
	for _, l := range dstLabels {
		if l.Name != nil {
			if _, exists := srcLabelMap[*l.Name]; !exists {
				if force {
					if err := g.DeleteLabel(ctx, dst.Owner, dst.Name, *l.Name); err != nil {
						return fmt.Errorf("failed to delete label %s: %w", *l.Name, err)
					}
				} else {
					if _, err := DeleteUnusedLabel(ctx, g, dst, *l.Name); err != nil {
						return fmt.Errorf("failed to delete unused label %s: %w", *l.Name, err)
					}
				}
			}
		}
	}
	return nil
}

func equalLabel(a, b *github.Label) bool {
	if a.Name == nil || b.Name == nil {
		return false
	}
	if *a.Name != *b.Name {
		return false
	}
	if (a.Color == nil) != (b.Color == nil) {
		return false
	}
	if a.Color != nil && b.Color != nil && *a.Color != *b.Color {
		return false
	}
	if (a.Description == nil) != (b.Description == nil) {
		return false
	}
	if a.Description != nil && b.Description != nil && *a.Description != *b.Description {
		return false
	}
	return true
}
