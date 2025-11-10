package render

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v73/github"
)

type repositoryRulesetFieldGetter func(ruleset *github.RepositoryRuleset) string
type repositoryRulesetFieldGetters struct {
	Func map[string]repositoryRulesetFieldGetter
}

func NewRepositoryRulesetFieldGetters() *repositoryRulesetFieldGetters {
	return &repositoryRulesetFieldGetters{
		Func: map[string]repositoryRulesetFieldGetter{
			"ID": func(ruleset *github.RepositoryRuleset) string {
				return ToString(ruleset.ID)
			},
			"NAME": func(ruleset *github.RepositoryRuleset) string {
				return ruleset.Name
			},
			"TARGET": func(ruleset *github.RepositoryRuleset) string {
				return ToString((*string)(ruleset.Target))
			},
			"ENFORCEMENT": func(ruleset *github.RepositoryRuleset) string {
				return string(ruleset.Enforcement)
			},
			"SOURCE_TYPE": func(ruleset *github.RepositoryRuleset) string {
				return ToString((*string)(ruleset.SourceType))
			},
			"SOURCE": func(ruleset *github.RepositoryRuleset) string {
				return ruleset.Source
			},
			"CURRENTUSER_CAN_BYPASS": func(ruleset *github.RepositoryRuleset) string {
				return ToString((*string)(ruleset.CurrentUserCanBypass))
			},
			"NODE_ID": func(ruleset *github.RepositoryRuleset) string {
				return ToString(ruleset.NodeID)
			},
			"CREATED_AT": func(ruleset *github.RepositoryRuleset) string {
				if ruleset.CreatedAt != nil {
					return ruleset.CreatedAt.Format("2006-01-02 15:04:05")
				}
				return ""
			},
			"UPDATED_AT": func(ruleset *github.RepositoryRuleset) string {
				if ruleset.UpdatedAt != nil {
					return ruleset.UpdatedAt.Format("2006-01-02 15:04:05")
				}
				return ""
			},
			"BYPASS_ACTORS": func(ruleset *github.RepositoryRuleset) string {
				actors := []string{}
				for _, actor := range ruleset.BypassActors {
					actors = append(actors, fmt.Sprintf("ID: %d, Type: %s, Mode: %s",
						actor.ActorID, ToString((*string)(actor.ActorType)), ToString((*string)(actor.BypassMode))))
				}
				return strings.Join(actors, "/ ")
			},
		},
	}
}

func (u *repositoryRulesetFieldGetters) GetField(ruleset *github.RepositoryRuleset, field string) string {
	if getter, ok := u.Func[field]; ok {
		return getter(ruleset)
	}
	return ""
}

// RenderRepositoryRulesets renders repository rulesets in a table format with specified headers
func (r *Renderer) RenderRepositoryRulesets(rulesets []*github.RepositoryRuleset, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(rulesets)
		return
	}

	getter := NewRepositoryRulesetFieldGetters()
	table := r.newTableWriter(headers)

	for _, ruleset := range rulesets {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(ruleset, header)
		}
		table.Append(row)
	}

	table.Render()
}

// RenderRepositoryRulesetsDefault renders repository rulesets with default headers
func (r *Renderer) RenderRepositoryRulesetsDefault(rulesets []*github.RepositoryRuleset) {
	headers := []string{"ID", "NAME", "TARGET", "ENFORCEMENT", "SOURCE"}
	r.RenderRepositoryRulesets(rulesets, headers)
}

// RenderRepositoryRulesetDetail renders detailed information about a single repository ruleset
func (r *Renderer) RenderRepositoryRuleset(ruleset *github.RepositoryRuleset, showConditionsAndRules bool) {
	if r.exporter != nil {
		r.RenderExportedData(ruleset)
		return
	}
	{
		headers := []string{"ID", "NAME", "TARGET", "ENFORCEMENT", "SOURCE_TYPE", "SOURCE", "CURRENTUSER_CAN_BYPASS", "NODE_ID", "CREATED_AT", "UPDATED_AT"}
		getter := NewRepositoryRulesetFieldGetters()
		table := r.newTableWriter([]string{"FIELD", "VALUE"})

		for _, header := range headers {
			row := []string{header, getter.GetField(ruleset, header)}
			table.Append(row)
		}

		table.Render()
	}

	if showConditionsAndRules {
		r.writeLine("Bypass Actors:")
		table := r.newTableWriter([]string{"ACTOR_ID", "ACTOR_TYPE", "BYPASS_MODE"})
		for _, actor := range ruleset.BypassActors {
			row := []string{
				ToString(actor.ActorID),
				ToString((*string)(actor.ActorType)),
				ToString((*string)(actor.BypassMode)),
			}
			table.Append(row)
		}
		table.Render()
	}

	if ruleset.Conditions != nil && showConditionsAndRules {
		r.writeLine("Targets:")
		table := r.newTableWriter([]string{"CONDITION", "FIELD", "VALUE"})

		if ruleset.Conditions.RefName != nil {
			condition := "Ref Name"
			table.Append([]string{condition, "Include", strings.Join(ruleset.Conditions.RefName.Include, ", ")})
			table.Append([]string{condition, "Exclude", strings.Join(ruleset.Conditions.RefName.Exclude, ", ")})
		}

		if ruleset.Conditions.RepositoryID != nil {
			condition := "Repository ID"
			ids := []string{}
			for _, id := range ruleset.Conditions.RepositoryID.RepositoryIDs {
				ids = append(ids, ToString(id))
			}
			table.Append([]string{condition, "ID", strings.Join(ids, ", ")})
		}

		if ruleset.Conditions.RepositoryName != nil {
			condition := "Repository Name"
			table.Append([]string{condition, "Include", strings.Join(ruleset.Conditions.RepositoryName.Include, ", ")})
			table.Append([]string{condition, "Exclude", strings.Join(ruleset.Conditions.RepositoryName.Exclude, ", ")})
			table.Append([]string{condition, "Protected", ToString(ruleset.Conditions.RepositoryName.Protected)})
		}

		if ruleset.Conditions.RepositoryProperty != nil {
			condition := "Repository Property"
			var indcludes []string
			for _, include := range ruleset.Conditions.RepositoryProperty.Include {
				indcludes = append(indcludes, include.Name+": "+strings.Join(include.PropertyValues, ", ")+": from "+ToString(include.Source))
			}
			var excludes []string
			for _, exclude := range ruleset.Conditions.RepositoryProperty.Exclude {
				excludes = append(excludes, exclude.Name+": "+strings.Join(exclude.PropertyValues, ", ")+": from "+ToString(exclude.Source))
			}
			table.Append([]string{condition, "Include", strings.Join(indcludes, "\n")})
			table.Append([]string{condition, "Exclude", strings.Join(excludes, "\n")})
		}

		if ruleset.Conditions.OrganizationID != nil {
			condition := "Organization ID"
			ids := []string{}
			for _, id := range ruleset.Conditions.OrganizationID.OrganizationIDs {
				ids = append(ids, ToString(id))
			}
			table.Append([]string{condition, "ID", strings.Join(ids, ", ")})
		}

		if ruleset.Conditions.OrganizationName != nil {
			condition := "Organization Name"
			table.Append([]string{condition, "Include", strings.Join(ruleset.Conditions.OrganizationName.Include, ", ")})
			table.Append([]string{condition, "Exclude", strings.Join(ruleset.Conditions.OrganizationName.Exclude, ", ")})
		}
		table.Render()
	}

	if ruleset.Rules != nil && showConditionsAndRules {
		r.writeLine("Rules:")
		table := r.newTableWriter([]string{"FIELD", "VALUE"})
		table.Append([]string{"Restrict creations", ToString(ruleset.Rules.Creation)})
		if ruleset.Rules.Update != nil {
			table.Append([]string{"Restrict updates", "ENABLED"})
			table.Append([]string{"  - Allows fetch and merge", ToString(ruleset.Rules.Update.UpdateAllowsFetchAndMerge)})
		} else {
			table.Append([]string{"Restrict updates", "DISABLED"})
		}
		table.Append([]string{"Restrict deletions", ToString(ruleset.Rules.Deletion)})
		table.Append([]string{"Require linear history", ToString(ruleset.Rules.RequiredLinearHistory)})
		if ruleset.Rules.RequiredDeployments != nil {
			table.Append([]string{"Require deployments to succeed", "ENABLED"})
			table.Append([]string{"  - Environments", strings.Join(ruleset.Rules.RequiredDeployments.RequiredDeploymentEnvironments, ", ")})
		} else {
			table.Append([]string{"Require deployments to succeed", "DISABLED"})
		}
		table.Append([]string{"Require signed commits", ToString(ruleset.Rules.RequiredSignatures)})
		if ruleset.Rules.PullRequest != nil {
			table.Append([]string{"Require a pull request before merging", "ENABLED"})
			table.Append([]string{"  - Required approvals", ToString(ruleset.Rules.PullRequest.RequiredApprovingReviewCount)})
			table.Append([]string{"  - Dismiss stale pull request approvals when new commits are pushed", ToString(ruleset.Rules.PullRequest.DismissStaleReviewsOnPush)})
			// Require review from specific teams
			table.Append([]string{"  - Require review from Code Owners", ToString(ruleset.Rules.PullRequest.RequireCodeOwnerReview)})
			table.Append([]string{"  - Require approval of the most recent reviewable push", ToString(ruleset.Rules.PullRequest.RequireLastPushApproval)})
			table.Append([]string{"  - Require conversation resolution before merging", ToString(ruleset.Rules.PullRequest.RequiredReviewThreadResolution)})
			table.Append([]string{"  - Automatically request Copilot code review", ToString(ruleset.Rules.PullRequest.AutomaticCopilotCodeReviewEnabled)})
			allowedMergeMethod := []string{}
			for _, method := range ruleset.Rules.PullRequest.AllowedMergeMethods {
				allowedMergeMethod = append(allowedMergeMethod, string(method))
			}
			table.Append([]string{"  - Allowed merge methods", strings.Join(allowedMergeMethod, ", ")})
		} else {
			table.Append([]string{"Require a pull request before merging", "DISABLED"})
		}
		if ruleset.Rules.RequiredStatusChecks != nil {
			table.Append([]string{"Require status checks to pass", "ENABLED"})
			table.Append([]string{"  - Require branches to be up to date before merging", ToString(ruleset.Rules.RequiredStatusChecks.StrictRequiredStatusChecksPolicy)})
			table.Append([]string{"  - Do not require status checks on creation", ToString(ruleset.Rules.RequiredStatusChecks.DoNotEnforceOnCreate)})
			for _, check := range ruleset.Rules.RequiredStatusChecks.RequiredStatusChecks {
				table.Append([]string{"  - Require status check: " + check.Context, ToString(check.IntegrationID)})
			}
		} else {
			table.Append([]string{"Require status checks to pass", "DISABLED"})
		}
		table.Append([]string{"Block force pushes", ToString(ruleset.Rules.NonFastForward)})
		if ruleset.Rules.CodeScanning != nil {
			table.Append([]string{"Require code scanning results to be up to date", "ENABLED"})
			for _, tools := range ruleset.Rules.CodeScanning.CodeScanningTools {
				table.Append([]string{"  - " + tools.Tool, fmt.Sprintf("AlertsThreshold: %s, SecurityAlertsThreshold: %s", tools.AlertsThreshold, tools.SecurityAlertsThreshold)})
			}
		}

		table.Render()
	}
}
