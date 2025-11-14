package render

import (
	"fmt"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

type ruleSuiteFieldGetter func(ruleSuite *client.RuleSuite) string
type ruleSuiteFieldGetters struct {
	Func map[string]ruleSuiteFieldGetter
}

func NewRuleSuiteFieldGetters() *ruleSuiteFieldGetters {
	return &ruleSuiteFieldGetters{
		Func: map[string]ruleSuiteFieldGetter{
			"ID": func(ruleSuite *client.RuleSuite) string {
				return ToString(ruleSuite.ID)
			},
			"ACTOR_NAME": func(ruleSuite *client.RuleSuite) string {
				return ToString(ruleSuite.ActorName)
			},
			"REF": func(ruleSuite *client.RuleSuite) string {
				if ruleSuite.Ref != nil && strings.HasPrefix(*ruleSuite.Ref, "refs/heads/") {
					return ToString((*ruleSuite.Ref)[len("refs/heads/"):])
				}
				return ToString(ruleSuite.Ref)
			},
			"RESULT": func(ruleSuite *client.RuleSuite) string {
				return ToString(ruleSuite.Result)
			},
			"EVALUATION_RESULT": func(ruleSuite *client.RuleSuite) string {
				return ToString(ruleSuite.EvaluationResult)
			},
			"EVALUATED": func(ruleSuite *client.RuleSuite) string {
				return ToString(ruleSuite.EvaluationResult)
			},
			"PUSHED_AT": func(ruleSuite *client.RuleSuite) string {
				return ToString(ruleSuite.PushedAt)
			},
			"BEFORE_SHA": func(ruleSuite *client.RuleSuite) string {
				return ToString(ruleSuite.BeforeSHA)
			},
			"AFTER_SHA": func(ruleSuite *client.RuleSuite) string {
				return ToString(ruleSuite.AfterSHA)
			},
		},
	}
}

func (u *ruleSuiteFieldGetters) GetField(ruleSuite *client.RuleSuite, field string) string {
	if getter, ok := u.Func[field]; ok {
		return getter(ruleSuite)
	}
	return ""
}

// RenderRuleSuites renders rule suites in a table format with specified headers
func (r *Renderer) RenderRuleSuites(ruleSuites []*client.RuleSuite, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(ruleSuites)
		return
	}

	if len(ruleSuites) == 0 {
		r.writeLine("No rule suites.")
		return
	}

	getter := NewRuleSuiteFieldGetters()
	table := r.newTableWriter(headers)

	for _, ruleSuite := range ruleSuites {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(ruleSuite, header)
		}
		table.Append(row)
	}

	table.Render()
}

// RenderRuleSuitesDefault renders rule suites with default headers
func (r *Renderer) RenderRuleSuitesDefault(ruleSuites []*client.RuleSuite) {
	headers := []string{"ID", "ACTOR_NAME", "REF", "RESULT", "EVALUATED", "PUSHED_AT"}
	r.RenderRuleSuites(ruleSuites, headers)
}

// RenderRuleSuiteDetail renders detailed information about a single rule suite
func (r *Renderer) RenderRuleSuiteDetail(ruleSuite *client.RuleSuite) {
	if r.exporter != nil {
		r.RenderExportedData(ruleSuite)
		return
	}

	r.writeLine(fmt.Sprintf("Rule Suite ID: %s", ToString(ruleSuite.ID)))
	r.writeLine(fmt.Sprintf("Actor: %s (ID: %s)", ToString(ruleSuite.ActorName), ToString(ruleSuite.ActorID)))
	r.writeLine(fmt.Sprintf("Repository: %s (ID: %s)", ToString(ruleSuite.RepositoryName), ToString(ruleSuite.RepositoryID)))
	r.writeLine(fmt.Sprintf("Ref: %s", ToString(ruleSuite.Ref)))
	r.writeLine(fmt.Sprintf("Before SHA: %s", ToString(ruleSuite.BeforeSHA)))
	r.writeLine(fmt.Sprintf("After SHA: %s", ToString(ruleSuite.AfterSHA)))
	r.writeLine(fmt.Sprintf("Result: %s", ToString(ruleSuite.Result)))
	r.writeLine(fmt.Sprintf("Evaluation Result: %s", ToString(ruleSuite.EvaluationResult)))
	if ruleSuite.PushedAt != nil {
		r.writeLine(fmt.Sprintf("Pushed At: %s", ruleSuite.PushedAt.String()))
	}

	if len(ruleSuite.RuleEvaluations) > 0 {
		r.writeLine("\nRule Evaluations:")
		for i, evaluation := range ruleSuite.RuleEvaluations {
			r.writeLine(fmt.Sprintf("  [%d] Rule Type: %s", i+1, ToString(evaluation.RuleType)))
			r.writeLine(fmt.Sprintf("      Result: %s", ToString(evaluation.Result)))
			r.writeLine(fmt.Sprintf("      Enforcement Mode: %s", ToString(evaluation.EnforcementMode)))
			if evaluation.RuleSource != nil {
				r.writeLine(fmt.Sprintf("      Source: %s (ID: %s, Name: %s)",
					ToString(evaluation.RuleSource.Type),
					ToString(evaluation.RuleSource.ID),
					ToString(evaluation.RuleSource.Name)))
			}
			if evaluation.Details != nil && *evaluation.Details != "" {
				r.writeLine(fmt.Sprintf("      Details: %s", *evaluation.Details))
			}
		}
	}
}
