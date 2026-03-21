package render

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v84/github"
)

type environmentFieldGetter func(env *github.Environment) string
type environmentFieldGetters struct {
	Func map[string]environmentFieldGetter
}

func newEnvironmentFieldGetters() *environmentFieldGetters {
	return &environmentFieldGetters{
		Func: map[string]environmentFieldGetter{
			"NAME": func(env *github.Environment) string {
				return ToString(env.Name)
			},
			"CREATED_AT": func(env *github.Environment) string {
				return ToString(env.CreatedAt)
			},
			"UPDATED_AT": func(env *github.Environment) string {
				return ToString(env.UpdatedAt)
			},
		},
	}
}

func (g *environmentFieldGetters) GetField(env *github.Environment, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(env)
	}
	return ""
}

// RenderEnvironments renders a table of environments with the specified headers.
func (r *Renderer) RenderEnvironments(envs []*github.Environment, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(envs)
	}

	if len(envs) == 0 {
		r.writeLine("No environments.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"NAME", "CREATED_AT", "UPDATED_AT"}
	}

	getter := newEnvironmentFieldGetters()
	table := r.newTableWriter(headers)

	for _, env := range envs {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(env, header)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderEnvironment prints the details of an environment in key:value format.
// policies should contain the custom deployment branch policies when
// DeploymentBranchPolicy.CustomBranchPolicies is true; pass nil otherwise.
func (r *Renderer) RenderEnvironment(env *github.Environment, policies []*github.DeploymentBranchPolicy) error {
	if r.exporter != nil {
		return r.RenderExportedData(env)
	}

	r.writeLine(fmt.Sprintf("Name:                %s", ToString(env.Name)))
	r.writeLine(fmt.Sprintf("URL:                 %s", ToString(env.HTMLURL)))
	r.writeLine(fmt.Sprintf("Created at:          %s", ToString(env.CreatedAt)))
	r.writeLine(fmt.Sprintf("Updated at:          %s", ToString(env.UpdatedAt)))
	r.writeLine(fmt.Sprintf("Can admins bypass:   %s", ToString(env.CanAdminsBypass)))

	// Deployment branch policy
	branchPolicyStr := "none"
	if env.DeploymentBranchPolicy != nil {
		switch {
		case env.DeploymentBranchPolicy.ProtectedBranches != nil && *env.DeploymentBranchPolicy.ProtectedBranches:
			branchPolicyStr = "protected_branches"
		case env.DeploymentBranchPolicy.CustomBranchPolicies != nil && *env.DeploymentBranchPolicy.CustomBranchPolicies:
			branchPolicyStr = "custom"
		}
	}
	r.writeLine(fmt.Sprintf("Branch policy:       %s", branchPolicyStr))
	for _, p := range policies {
		r.writeLine(fmt.Sprintf("  - %s (%s)", ToString(p.Name), ToString(p.Type)))
	}

	// Protection rules
	var waitTimer int
	var reviewers []string
	preventSelfReview := false
	for _, rule := range env.ProtectionRules {
		if rule.Type == nil {
			continue
		}
		switch *rule.Type {
		case "wait_timer":
			if rule.WaitTimer != nil {
				waitTimer = *rule.WaitTimer
			}
		case "required_reviewers":
			if rule.PreventSelfReview != nil {
				preventSelfReview = *rule.PreventSelfReview
			}
			for _, rev := range rule.Reviewers {
				if rev.Type == nil {
					continue
				}
				var name string
				switch v := rev.Reviewer.(type) {
				case *github.User:
					name = fmt.Sprintf("%s: %s", *rev.Type, ToString(v.Login))
				case *github.Team:
					name = fmt.Sprintf("%s: %s", *rev.Type, ToString(v.Slug))
				default:
					name = *rev.Type
				}
				reviewers = append(reviewers, name)
			}
		}
	}
	r.writeLine(fmt.Sprintf("Wait timer:          %d minutes", waitTimer))
	r.writeLine(fmt.Sprintf("Prevent self-review: %v", preventSelfReview))
	if len(reviewers) == 0 {
		r.writeLine("Reviewers:           none")
	} else {
		r.writeLine("Reviewers:")
		for _, rev := range reviewers {
			r.writeLine(fmt.Sprintf("  - %s", rev))
		}
	}
	return nil
}
