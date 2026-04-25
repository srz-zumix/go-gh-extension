package render

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v84/github"
)

// CodeScanningAlertFieldGetter defines a function to get a field value from a code scanning Alert.
type CodeScanningAlertFieldGetter func(alert *github.Alert) string

// CodeScanningAlertFieldGetters holds field getters for code scanning Alert table rendering.
type CodeScanningAlertFieldGetters struct {
	Func map[string]CodeScanningAlertFieldGetter
}

// NewCodeScanningAlertFieldGetters creates field getters for code scanning Alert table rendering.
func NewCodeScanningAlertFieldGetters() *CodeScanningAlertFieldGetters {
	return &CodeScanningAlertFieldGetters{
		Func: map[string]CodeScanningAlertFieldGetter{
			"NUMBER": func(alert *github.Alert) string {
				return ToString(alert.Number)
			},
			"STATE": func(alert *github.Alert) string {
				return ToString(alert.State)
			},
			"RULE": func(alert *github.Alert) string {
				if alert.Rule != nil {
					return ToString(alert.Rule.ID)
				}
				return ""
			},
			"SEVERITY": func(alert *github.Alert) string {
				if alert.Rule != nil {
					return ToString(alert.Rule.Severity)
				}
				return ""
			},
			"DESCRIPTION": func(alert *github.Alert) string {
				if alert.Rule != nil {
					return ToString(alert.Rule.Description)
				}
				return ""
			},
			"TOOL": func(alert *github.Alert) string {
				if alert.Tool != nil {
					return ToString(alert.Tool.Name)
				}
				return ""
			},
			"URL": func(alert *github.Alert) string {
				return ToString(alert.HTMLURL)
			},
			"CREATED": func(alert *github.Alert) string {
				if alert.CreatedAt != nil {
					return ToString(alert.CreatedAt)
				}
				return ""
			},
			"UPDATED": func(alert *github.Alert) string {
				if alert.UpdatedAt != nil {
					return ToString(alert.UpdatedAt)
				}
				return ""
			},
			"REPOSITORY": func(alert *github.Alert) string {
				if alert.Repository != nil {
					return ToString(alert.Repository.FullName)
				}
				return ""
			},
		},
	}
}

// GetField returns the value of the specified field for the given alert.
func (g *CodeScanningAlertFieldGetters) GetField(alert *github.Alert, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(alert)
	}
	return ""
}

// RenderCodeScanningAlerts renders a list of code scanning alerts as a table.
func (r *Renderer) RenderCodeScanningAlerts(alerts []*github.Alert, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(alerts)
	}
	if len(alerts) == 0 {
		r.writeLine("No code scanning alerts found.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"Number", "State", "Severity", "Rule", "Tool", "Description"}
	}
	getter := NewCodeScanningAlertFieldGetters()
	table := r.newTableWriter(headers)
	for _, alert := range alerts {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(alert, header)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderCodeScanningAlert renders a single code scanning alert in detail format.
func (r *Renderer) RenderCodeScanningAlert(alert *github.Alert) error {
	if r.exporter != nil {
		return r.RenderExportedData(alert)
	}
	if alert == nil {
		return nil
	}

	// label width: longest label is "Sec Severity:" (13 chars)
	const labelFmt = "%-13s %s"

	r.writeLine(fmt.Sprintf(labelFmt, "Number:", ToString(alert.Number)))
	r.writeLine(fmt.Sprintf(labelFmt, "State:", ToString(alert.State)))

	if alert.Rule != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Rule:", ToString(alert.Rule.ID)))
		r.writeLine(fmt.Sprintf(labelFmt, "Severity:", ToString(alert.Rule.Severity)))
		if alert.Rule.SecuritySeverityLevel != nil {
			r.writeLine(fmt.Sprintf(labelFmt, "Sec Severity:", ToString(alert.Rule.SecuritySeverityLevel)))
		}
		r.writeLine(fmt.Sprintf(labelFmt, "Description:", ToString(alert.Rule.Description)))
	}

	if alert.Tool != nil {
		toolStr := ToString(alert.Tool.Name)
		if alert.Tool.Version != nil {
			toolStr = fmt.Sprintf("%s (%s)", toolStr, ToString(alert.Tool.Version))
		}
		r.writeLine(fmt.Sprintf(labelFmt, "Tool:", toolStr))
	}

	if alert.MostRecentInstance != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Ref:", ToString(alert.MostRecentInstance.Ref)))
		if alert.MostRecentInstance.Location != nil {
			loc := alert.MostRecentInstance.Location
			r.writeLine(fmt.Sprintf(labelFmt, "Location:", fmt.Sprintf("%s:%d", ToString(loc.Path), loc.StartLine)))
		}
	}

	r.writeLine(fmt.Sprintf(labelFmt, "URL:", ToString(alert.HTMLURL)))
	r.writeLine(fmt.Sprintf(labelFmt, "Created:", ToString(alert.CreatedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "Updated:", ToString(alert.UpdatedAt)))

	if alert.DismissedAt != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Dismissed:", ToString(alert.DismissedAt)))
		r.writeLine(fmt.Sprintf(labelFmt, "Reason:", ToString(alert.DismissedReason)))
		if alert.DismissedComment != nil {
			r.writeLine(fmt.Sprintf(labelFmt, "Comment:", ToString(alert.DismissedComment)))
		}
	}

	if alert.FixedAt != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Fixed:", ToString(alert.FixedAt)))
	}

	return nil
}

// RenderCodeQLDatabases renders a list of CodeQL databases as a table.
func (r *Renderer) RenderCodeQLDatabases(databases []*github.CodeQLDatabase) error {
	if r.exporter != nil {
		return r.RenderExportedData(databases)
	}
	if len(databases) == 0 {
		r.writeLine("No CodeQL databases found.")
		return nil
	}

	headers := []string{"ID", "Language", "Size", "Created", "Updated"}
	table := r.newTableWriter(headers)
	for _, db := range databases {
		row := []string{
			ToString(db.ID),
			ToString(db.Language),
			ToString(db.Size),
			ToString(db.CreatedAt),
			ToString(db.UpdatedAt),
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderCodeQLDatabase renders a single CodeQL database in detail format.
func (r *Renderer) RenderCodeQLDatabase(db *github.CodeQLDatabase) error {
	if r.exporter != nil {
		return r.RenderExportedData(db)
	}
	if db == nil {
		return nil
	}

	const labelFmt = "%-13s %s"

	r.writeLine(fmt.Sprintf(labelFmt, "ID:", ToString(db.ID)))
	r.writeLine(fmt.Sprintf(labelFmt, "Language:", ToString(db.Language)))
	r.writeLine(fmt.Sprintf(labelFmt, "Name:", ToString(db.Name)))
	r.writeLine(fmt.Sprintf(labelFmt, "Content Type:", ToString(db.ContentType)))
	r.writeLine(fmt.Sprintf(labelFmt, "Size:", ToString(db.Size)))
	if db.Uploader != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Uploader:", ToString(db.Uploader.Login)))
	}
	r.writeLine(fmt.Sprintf(labelFmt, "Created:", ToString(db.CreatedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "Updated:", ToString(db.UpdatedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "URL:", ToString(db.URL)))

	return nil
}

// RenderCodeScanningAnalyses renders a list of code scanning analyses as a table.
func (r *Renderer) RenderCodeScanningAnalyses(analyses []*github.ScanningAnalysis) error {
	if r.exporter != nil {
		return r.RenderExportedData(analyses)
	}
	if len(analyses) == 0 {
		r.writeLine("No code scanning analyses found.")
		return nil
	}

	headers := []string{"ID", "Ref", "Tool", "Results", "Rules", "Created"}
	table := r.newTableWriter(headers)
	for _, a := range analyses {
		toolName := ""
		if a.Tool != nil {
			toolName = ToString(a.Tool.Name)
		}
		row := []string{
			ToString(a.ID),
			ToString(a.Ref),
			toolName,
			ToString(a.ResultsCount),
			ToString(a.RulesCount),
			ToString(a.CreatedAt),
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderCodeScanningAnalysis renders a single code scanning analysis in detail format.
func (r *Renderer) RenderCodeScanningAnalysis(analysis *github.ScanningAnalysis) error {
	if r.exporter != nil {
		return r.RenderExportedData(analysis)
	}
	if analysis == nil {
		return nil
	}

	const labelFmt = "%-14s %s"

	r.writeLine(fmt.Sprintf(labelFmt, "ID:", ToString(analysis.ID)))
	r.writeLine(fmt.Sprintf(labelFmt, "Ref:", ToString(analysis.Ref)))
	r.writeLine(fmt.Sprintf(labelFmt, "Commit SHA:", ToString(analysis.CommitSHA)))
	r.writeLine(fmt.Sprintf(labelFmt, "Analysis Key:", ToString(analysis.AnalysisKey)))
	r.writeLine(fmt.Sprintf(labelFmt, "Category:", ToString(analysis.Category)))
	if analysis.Tool != nil {
		toolStr := ToString(analysis.Tool.Name)
		if analysis.Tool.Version != nil {
			toolStr = fmt.Sprintf("%s (%s)", toolStr, ToString(analysis.Tool.Version))
		}
		r.writeLine(fmt.Sprintf(labelFmt, "Tool:", toolStr))
	}
	r.writeLine(fmt.Sprintf(labelFmt, "Results:", ToString(analysis.ResultsCount)))
	r.writeLine(fmt.Sprintf(labelFmt, "Rules:", ToString(analysis.RulesCount)))
	r.writeLine(fmt.Sprintf(labelFmt, "SARIF ID:", ToString(analysis.SarifID)))
	r.writeLine(fmt.Sprintf(labelFmt, "Created:", ToString(analysis.CreatedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "Deletable:", ToString(analysis.Deletable)))
	if analysis.Warning != nil && *analysis.Warning != "" {
		r.writeLine(fmt.Sprintf(labelFmt, "Warning:", ToString(analysis.Warning)))
	}
	r.writeLine(fmt.Sprintf(labelFmt, "URL:", ToString(analysis.URL)))

	return nil
}

// RenderSarifUpload renders a single SARIF upload status in detail format.
func (r *Renderer) RenderSarifUpload(upload *github.SARIFUpload) error {
	if r.exporter != nil {
		return r.RenderExportedData(upload)
	}
	if upload == nil {
		return nil
	}

	const labelFmt = "%-14s %s"

	r.writeLine(fmt.Sprintf(labelFmt, "Status:", ToString(upload.ProcessingStatus)))
	r.writeLine(fmt.Sprintf(labelFmt, "Analyses URL:", ToString(upload.AnalysesURL)))

	return nil
}

// RenderSarifID renders the result of a SARIF upload.
func (r *Renderer) RenderSarifID(sarifID *github.SarifID) error {
	if r.exporter != nil {
		return r.RenderExportedData(sarifID)
	}
	if sarifID == nil {
		return nil
	}

	const labelFmt = "%-14s %s"

	r.writeLine(fmt.Sprintf(labelFmt, "SARIF ID:", ToString(sarifID.ID)))
	r.writeLine(fmt.Sprintf(labelFmt, "URL:", ToString(sarifID.URL)))

	return nil
}
