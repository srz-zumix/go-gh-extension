package gh

import (
	"context"

	"github.com/shurcooL/githubv4"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// UpdateProjectV2FieldName renames a custom field in a GitHub Project v2.
func UpdateProjectV2FieldName(ctx context.Context, g *GitHubClient, fieldID string, name string) error {
	n := githubv4.String(name)
	return g.UpdateProjectV2Field(ctx, client.UpdateProjectV2FieldInput{
		FieldID: githubv4.ID(fieldID),
		Name:    &n,
	})
}

// UpdateProjectV2FieldSingleSelectOptions replaces the options of a SINGLE_SELECT field.
// Options with a non-empty ID are updated; options without an ID are added.
func UpdateProjectV2FieldSingleSelectOptions(ctx context.Context, g *GitHubClient, fieldID string, options []ProjectV2SingleSelectOption) error {
	opts := make([]client.ProjectV2SingleSelectFieldOptionInput, len(options))
	for i, o := range options {
		opt := client.ProjectV2SingleSelectFieldOptionInput{
			Name:        githubv4.String(o.Name),
			Color:       githubv4.String(o.Color),
			Description: githubv4.String(o.Description),
		}
		if o.ID != "" {
			id := githubv4.String(o.ID)
			opt.ID = &id
		}
		opts[i] = opt
	}
	return g.UpdateProjectV2Field(ctx, client.UpdateProjectV2FieldInput{
		FieldID:             githubv4.ID(fieldID),
		SingleSelectOptions: opts,
	})
}

// DeleteProjectV2Field deletes a custom field from a GitHub Project v2.
func DeleteProjectV2Field(ctx context.Context, g *GitHubClient, fieldID string) error {
	return g.DeleteProjectV2Field(ctx, client.DeleteProjectV2FieldInput{
		FieldID: githubv4.ID(fieldID),
	})
}
