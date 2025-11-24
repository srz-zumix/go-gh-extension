package render

import (
	"github.com/fatih/color"
	"github.com/google/go-github/v79/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

// checkRunFieldGetter is a function type that extracts a specific field value from a gh.CheckRun.
type checkRunFieldGetter func(checkRun *gh.CheckRun) string

// checkRunFieldGetters holds a map of field names to their corresponding getter functions for gh.CheckRun.
type checkRunFieldGetters struct {
	Func map[string]checkRunFieldGetter
}

// NewCheckRunFieldGetters returns field getter functions for github.CheckRun
func NewCheckRunFieldGetters(enableColor bool) *checkRunFieldGetters {
	return &checkRunFieldGetters{
		Func: map[string]checkRunFieldGetter{
			"NAME": func(checkRun *gh.CheckRun) string {
				if enableColor && checkRun.IsRequired != nil && *checkRun.IsRequired {
					return color.YellowString(ToString(checkRun.Name))
				}
				return ToString(checkRun.Name)
			},
			"FULL_NAME": func(checkRun *gh.CheckRun) string {
				if enableColor && checkRun.IsRequired != nil && *checkRun.IsRequired {
					return color.YellowString(checkRun.GetFullName())
				}
				return checkRun.GetFullName()
			},
			"WORKFLOW_NAME": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.WorkflowName)
			},
			"STATUS": func(checkRun *gh.CheckRun) string {
				status := checkRun.GetStatus()
				if status == "" {
					return "unknown"
				}
				return status
			},
			"CONCLUSION": func(checkRun *gh.CheckRun) string {
				conclusion := checkRun.GetConclusion()
				if conclusion == "" {
					return "pending"
				}
				return conclusion
			},
			"RUN_ID": func(checkRun *gh.CheckRun) string {
				runID := gh.ExtractRunIDFromCheckRun(checkRun)
				return runID
			},
			"JOB_ID": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.ID)
			},
			"HTML_URL": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.HTMLURL)
			},
			"EXTERNAL_ID": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.ExternalID)
			},
			"STARTED_AT": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.StartedAt)
			},
			"COMPLETED_AT": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.CompletedAt)
			},
			"DETAILS_URL": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.DetailsURL)
			},
			"URL": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.URL)
			},
			"ELAPSED": func(checkRun *gh.CheckRun) string {
				elapsed := checkRun.GetCompletedAt().Sub(checkRun.GetStartedAt().Time)
				return ToString(elapsed)
			},
			"REQUIRED": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.IsRequired)
			},
			"S": func(checkRun *gh.CheckRun) string {
				return GetStatusIconColored(checkRun.GetStatus(), enableColor)
			},
			"C": func(checkRun *gh.CheckRun) string {
				return GetConclusionIconColored(checkRun.GetConclusion(), enableColor)
			},
			"_": func(checkRun *gh.CheckRun) string {
				return GetConclusionIconColored(checkRun.GetConclusion(), enableColor)
			},
			"R": func(checkRun *gh.CheckRun) string {
				if checkRun.IsRequired != nil && *checkRun.IsRequired {
					return "R"
				}
				return " "
			},
			"SUITE": func(checkRun *gh.CheckRun) string {
				return ToString(checkRun.GetCheckSuite().ID)
			},
		},
	}
}

// GetField returns the value of the specified field for the check run
func (g *checkRunFieldGetters) GetField(checkRun *gh.CheckRun, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(checkRun)
	}
	return ""
}

// RenderCheckRuns renders a table of check runs with the specified headers
func (r *Renderer) RenderCheckRuns(checkRuns []*gh.CheckRun, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(checkRuns)
		return
	}

	if len(checkRuns) == 0 {
		r.writeLine("No check runs.")
		return
	}

	getter := NewCheckRunFieldGetters(r.Color)
	table := r.newTableWriter(headers)
	table.SetAutoWrapText(false)

	for _, checkRun := range checkRuns {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(checkRun, header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderCheckRunsDefault renders check runs with default columns
func (r *Renderer) RenderCheckRunsDefault(checkRuns []*gh.CheckRun) {
	headers := []string{"_", "NAME", "ELAPSED", "DETAILS_URL"}
	r.RenderCheckRuns(checkRuns, headers)
}

func (r *Renderer) RenderCheckRunsDetails(checkRuns []*gh.CheckRun) {
	headers := []string{"_", "WORKFLOW_NAME", "NAME", "STATUS", "RUN_ID", "JOB_ID", "ELAPSED", "DETAILS_URL"}
	r.RenderCheckRuns(checkRuns, headers)
}

// checkSuiteFieldGetter is a function type that extracts a string field from a github.CheckSuite.
type checkSuiteFieldGetter func(checkSuite *github.CheckSuite) string

// checkSuiteFieldGetters holds a map of field names to their corresponding getter functions for github.CheckSuite.
type checkSuiteFieldGetters struct {
	Func map[string]checkSuiteFieldGetter
}

// NewCheckSuiteFieldGetters returns field getter functions for github.CheckSuite
func NewCheckSuiteFieldGetters(enableColor bool) *checkSuiteFieldGetters {
	return &checkSuiteFieldGetters{
		Func: map[string]checkSuiteFieldGetter{
			"ID": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.ID)
			},
			"NODE_ID": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.NodeID)
			},
			"HEAD_BRANCH": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.HeadBranch)
			},
			"HEAD_SHA": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.HeadSHA)
			},
			"STATUS": func(checkSuite *github.CheckSuite) string {
				status := checkSuite.GetStatus()
				if status == "" {
					return "unknown"
				}
				return status
			},
			"CONCLUSION": func(checkSuite *github.CheckSuite) string {
				conclusion := checkSuite.GetConclusion()
				if conclusion == "" {
					return "pending"
				}
				return conclusion
			},
			"URL": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.URL)
			},
			"BEFORE_SHA": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.BeforeSHA)
			},
			"AFTER_SHA": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.AfterSHA)
			},
			"CREATED_AT": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.CreatedAt)
			},
			"UPDATED_AT": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.UpdatedAt)
			},
			"APP": func(checkSuite *github.CheckSuite) string {
				if checkSuite.App != nil {
					return ToString(checkSuite.App.Name)
				}
				return ""
			},
			"LATEST_CHECK_RUNS_COUNT": func(checkSuite *github.CheckSuite) string {
				return ToString(checkSuite.LatestCheckRunsCount)
			},
			"S": func(checkSuite *github.CheckSuite) string {
				return GetStatusIconColored(checkSuite.GetStatus(), enableColor)
			},
			"C": func(checkSuite *github.CheckSuite) string {
				return GetConclusionIconColored(checkSuite.GetConclusion(), enableColor)
			},
			"_": func(checkSuite *github.CheckSuite) string {
				return GetConclusionIconColored(checkSuite.GetConclusion(), enableColor)
			},
		},
	}
}

// GetField returns the value of the specified field for the check suite
func (g *checkSuiteFieldGetters) GetField(checkSuite *github.CheckSuite, field string) string {
	if getter, ok := g.Func[field]; ok {
		return getter(checkSuite)
	}
	return ""
}

// RenderCheckSuites renders a table of check suites with the specified headers
func (r *Renderer) RenderCheckSuites(checkSuites []*github.CheckSuite, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(checkSuites)
		return
	}

	if len(checkSuites) == 0 {
		r.writeLine("No check suites.")
		return
	}

	getter := NewCheckSuiteFieldGetters(r.Color)
	table := r.newTableWriter(headers)
	table.SetAutoWrapText(false)

	for _, checkSuite := range checkSuites {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(checkSuite, header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderCheckSuitesDefault renders check suites with default columns
func (r *Renderer) RenderCheckSuitesDefault(checkSuites []*github.CheckSuite) {
	headers := []string{"_", "ID", "HEAD_BRANCH", "STATUS", "APP"}
	r.RenderCheckSuites(checkSuites, headers)
}
