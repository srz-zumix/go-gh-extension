package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v84/github"
)

// ListTagProtections retrieves all tag protection settings for the repository.
func ListTagProtections(ctx context.Context, g *GitHubClient, repo repository.Repository) ([]*github.TagProtection, error) {
	return g.ListTagProtection(ctx, repo.Owner, repo.Name)
}

// GetTagProtection retrieves a tag protection setting by pattern.
func GetTagProtection(ctx context.Context, g *GitHubClient, repo repository.Repository, pattern string) (*github.TagProtection, error) {
	tagProtections, err := ListTagProtections(ctx, g, repo)
	if err != nil {
		return nil, err
	}

	for _, tagProtection := range tagProtections {
		if tagProtection != nil && tagProtection.Pattern != nil && *tagProtection.Pattern == pattern {
			return tagProtection, nil
		}
	}

	return nil, fmt.Errorf("tag protection pattern %q not found", pattern)
}

// RemoveTagProtection removes a tag protection setting by ID.
func RemoveTagProtection(ctx context.Context, g *GitHubClient, repo repository.Repository, tagProtectionID int64) error {
	return g.DeleteTagProtection(ctx, repo.Owner, repo.Name, tagProtectionID)
}

// ConvertTagProtectionToRuleset converts a tag protection pattern to a repository ruleset.
func ConvertTagProtectionToRuleset(pattern string) *github.RepositoryRuleset {
	target := github.RulesetTargetTag
	enforcement := github.RulesetEnforcementActive

	refInclude := fmt.Sprintf("refs/tags/%s", pattern)
	conditions := &github.RepositoryRulesetConditions{
		RefName: &github.RepositoryRulesetRefConditionParameters{
			Include: []string{refInclude},
			Exclude: []string{},
		},
	}

	rules := &github.RepositoryRulesetRules{
		Update:   &github.UpdateRuleParameters{},
		Deletion: &github.EmptyRuleParameters{},
	}

	ruleset := &github.RepositoryRuleset{
		Name:        pattern,
		Target:      &target,
		Enforcement: enforcement,
		Conditions:  conditions,
		Rules:       rules,
	}

	return ruleset
}
