package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// GetBranchProtection retrieves the branch protection settings for the given branch.
func GetBranchProtection(ctx context.Context, g *GitHubClient, repo repository.Repository, branch string) (*github.Protection, error) {
	return g.GetBranchProtection(ctx, repo.Owner, repo.Name, branch)
}

// RemoveBranchProtection removes the branch protection settings for the given branch.
func RemoveBranchProtection(ctx context.Context, g *GitHubClient, repo repository.Repository, branch string) error {
	return g.RemoveBranchProtection(ctx, repo.Owner, repo.Name, branch)
}

// ConvertBranchProtectionToRuleset converts a branch protection rule to a repository ruleset.
// The generated ruleset targets the given branch name and includes rules equivalent to the
// branch protection settings. Some unsupported or partially unsupported fields (such as certain
// restrictions) are logged as warnings.
func ConvertBranchProtectionToRuleset(branch string, protection *github.Protection) *github.RepositoryRuleset {
	target := github.RulesetTargetBranch
	enforcement := github.RulesetEnforcementActive

	refInclude := fmt.Sprintf("refs/heads/%s", branch)
	conditions := &github.RepositoryRulesetConditions{
		RefName: &github.RepositoryRulesetRefConditionParameters{
			Include: []string{refInclude},
			Exclude: []string{},
		},
	}

	rules := &github.RepositoryRulesetRules{}

	// treat nil protection as empty protection to avoid panic on field access
	if protection == nil {
		protection = &github.Protection{}
	}

	// required_linear_history
	if protection.RequireLinearHistory != nil && protection.RequireLinearHistory.Enabled {
		rules.RequiredLinearHistory = &github.EmptyRuleParameters{}
	}

	// deletion: allow_deletions=false → add deletion rule (prevent deletion)
	if protection.AllowDeletions != nil && !protection.AllowDeletions.Enabled {
		rules.Deletion = &github.EmptyRuleParameters{}
	}

	// non_fast_forward: allow_force_pushes=false → add non_fast_forward rule
	if protection.AllowForcePushes != nil && !protection.AllowForcePushes.Enabled {
		rules.NonFastForward = &github.EmptyRuleParameters{}
	}

	// creation: block_creations=true → add creation rule (prevent branch creation matching the pattern)
	if protection.BlockCreations != nil && protection.BlockCreations.GetEnabled() {
		rules.Creation = &github.EmptyRuleParameters{}
	}

	// required_signatures
	if protection.RequiredSignatures != nil && protection.RequiredSignatures.GetEnabled() {
		rules.RequiredSignatures = &github.EmptyRuleParameters{}
	}

	// required_status_checks
	if protection.RequiredStatusChecks != nil {
		var statusChecks []*github.RuleStatusCheck
		if protection.RequiredStatusChecks.Checks != nil {
			for _, check := range *protection.RequiredStatusChecks.Checks {
				sc := &github.RuleStatusCheck{
					Context: check.Context,
				}
				if check.AppID != nil && *check.AppID > 0 {
					sc.IntegrationID = check.AppID
				}
				statusChecks = append(statusChecks, sc)
			}
		} else if protection.RequiredStatusChecks.Contexts != nil {
			// Legacy contexts field
			for _, ctx := range *protection.RequiredStatusChecks.Contexts {
				statusChecks = append(statusChecks, &github.RuleStatusCheck{Context: ctx})
			}
		}
		if len(statusChecks) > 0 {
			rules.RequiredStatusChecks = &github.RequiredStatusChecksRuleParameters{
				RequiredStatusChecks:             statusChecks,
				StrictRequiredStatusChecksPolicy: protection.RequiredStatusChecks.Strict,
			}
		} else if protection.RequiredStatusChecks.Strict {
			// GitHub rulesets may reject a required-status-checks rule without any checks configured
			logger.Warn("required status checks 'strict' is enabled but no checks or contexts are configured; skipping required-status-checks rule")
		}
	}

	// pull_request
	if protection.RequiredPullRequestReviews != nil || protection.RequiredConversationResolution != nil {
		prParams := &github.PullRequestRuleParameters{}
		hasPRRequirements := false

		if protection.RequiredPullRequestReviews != nil {
			pr := protection.RequiredPullRequestReviews
			prParams.DismissStaleReviewsOnPush = pr.DismissStaleReviews
			prParams.RequireCodeOwnerReview = pr.RequireCodeOwnerReviews
			prParams.RequiredApprovingReviewCount = pr.RequiredApprovingReviewCount
			prParams.RequireLastPushApproval = pr.RequireLastPushApproval
			hasPRRequirements = true
		}

		// required_conversation_resolution maps to RequiredReviewThreadResolution
		if protection.RequiredConversationResolution != nil && protection.RequiredConversationResolution.Enabled {
			prParams.RequiredReviewThreadResolution = true
			hasPRRequirements = true
		}

		if hasPRRequirements {
			rules.PullRequest = prParams
		}
	}

	// bypass actors from EnforceAdmins and Restrictions
	var bypassActors []*github.BypassActor

	// enforce_admins=false → allow repository admins to bypass
	if protection.EnforceAdmins != nil && !protection.EnforceAdmins.Enabled {
		bypassMode := github.BypassModeAlways
		actorType := github.BypassActorTypeRepositoryRole
		// Repository Role ID 5 = admin
		adminRoleID := int64(5)
		bypassActors = append(bypassActors, &github.BypassActor{
			ActorID:    &adminRoleID,
			ActorType:  &actorType,
			BypassMode: &bypassMode,
		})
	}

	// Restrictions define who can push. In rulesets there is no direct push-access
	// restriction equivalent; log a warning so the user is aware.
	if protection.Restrictions != nil {
		if len(protection.Restrictions.Users) > 0 {
			logger.Warn("Branch protection restriction (users) cannot be directly converted to a ruleset; push-access restrictions are not supported in rulesets",
				"branch", branch,
				"users", GetObjectNames(protection.Restrictions.Users))
		}
		if len(protection.Restrictions.Teams) > 0 {
			logger.Warn("Branch protection restriction (teams) cannot be directly converted to a ruleset; push-access restrictions are not supported in rulesets",
				"branch", branch,
				"teams", GetObjectNames(protection.Restrictions.Teams))
		}
		if len(protection.Restrictions.Apps) > 0 {
			logger.Warn("Branch protection restriction (apps) cannot be directly converted to a ruleset; push-access restrictions are not supported in rulesets",
				"branch", branch,
				"apps", GetObjectNames(protection.Restrictions.Apps))
		}
	}

	ruleset := &github.RepositoryRuleset{
		Name:         branch,
		Target:       &target,
		Enforcement:  enforcement,
		Conditions:   conditions,
		Rules:        rules,
		BypassActors: bypassActors,
	}

	return ruleset
}
