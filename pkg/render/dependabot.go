package render

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// DependabotAlertFieldGetter defines a function to get a field value from a DependabotAlert.
type DependabotAlertFieldGetter func(alert *github.DependabotAlert) string

// DependabotAlertFieldGetters holds field getters for DependabotAlert table rendering.
type DependabotAlertFieldGetters struct {
	Func map[string]DependabotAlertFieldGetter
}

// NewDependabotAlertFieldGetters creates field getters for DependabotAlert table rendering.
func NewDependabotAlertFieldGetters() *DependabotAlertFieldGetters {
	return &DependabotAlertFieldGetters{
		Func: map[string]DependabotAlertFieldGetter{
			"NUMBER": func(alert *github.DependabotAlert) string {
				return ToString(alert.Number)
			},
			"STATE": func(alert *github.DependabotAlert) string {
				return ToString(alert.State)
			},
			"PACKAGE": func(alert *github.DependabotAlert) string {
				if alert.Dependency != nil && alert.Dependency.Package != nil {
					return ToString(alert.Dependency.Package.Name)
				}
				return ""
			},
			"ECOSYSTEM": func(alert *github.DependabotAlert) string {
				if alert.Dependency != nil && alert.Dependency.Package != nil {
					return ToString(alert.Dependency.Package.Ecosystem)
				}
				return ""
			},
			"SCOPE": func(alert *github.DependabotAlert) string {
				if alert.Dependency != nil {
					return ToString(alert.Dependency.Scope)
				}
				return ""
			},
			"SEVERITY": func(alert *github.DependabotAlert) string {
				if alert.SecurityAdvisory != nil {
					return ToString(alert.SecurityAdvisory.Severity)
				}
				return ""
			},
			"GHSA": func(alert *github.DependabotAlert) string {
				if alert.SecurityAdvisory != nil {
					return ToString(alert.SecurityAdvisory.GHSAID)
				}
				return ""
			},
			"CVE": func(alert *github.DependabotAlert) string {
				if alert.SecurityAdvisory != nil {
					return ToString(alert.SecurityAdvisory.CVEID)
				}
				return ""
			},
			"SUMMARY": func(alert *github.DependabotAlert) string {
				if alert.SecurityAdvisory != nil {
					return ToString(alert.SecurityAdvisory.Summary)
				}
				return ""
			},
			"MANIFEST": func(alert *github.DependabotAlert) string {
				if alert.Dependency != nil {
					return ToString(alert.Dependency.ManifestPath)
				}
				return ""
			},
			"URL": func(alert *github.DependabotAlert) string {
				return ToString(alert.HTMLURL)
			},
			"CREATED": func(alert *github.DependabotAlert) string {
				if alert.CreatedAt != nil {
					return alert.CreatedAt.Format("2006-01-02")
				}
				return ""
			},
			"UPDATED": func(alert *github.DependabotAlert) string {
				if alert.UpdatedAt != nil {
					return alert.UpdatedAt.Format("2006-01-02")
				}
				return ""
			},
			"REPOSITORY": func(alert *github.DependabotAlert) string {
				if alert.Repository != nil {
					return ToString(alert.Repository.FullName)
				}
				return ""
			},
		},
	}
}

// GetField returns the value of the specified field for the given alert.
func (g *DependabotAlertFieldGetters) GetField(alert *github.DependabotAlert, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(alert)
	}
	return ""
}

// RenderDependabotAlerts renders a list of Dependabot alerts as a table.
func (r *Renderer) RenderDependabotAlerts(alerts []*github.DependabotAlert, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(alerts)
	}
	if len(alerts) == 0 {
		r.writeLine("No Dependabot alerts found.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"Number", "State", "Package", "Ecosystem", "Severity", "Summary"}
	}
	getter := NewDependabotAlertFieldGetters()
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

// RenderDependabotAlert renders a single Dependabot alert in detail format.
func (r *Renderer) RenderDependabotAlert(alert *github.DependabotAlert) error {
	if r.exporter != nil {
		return r.RenderExportedData(alert)
	}
	if alert == nil {
		return nil
	}

	// label width: longest label is "Vulnerable" (10 chars) + ":"
	const labelFmt = "%-12s %s"

	r.writeLine(fmt.Sprintf(labelFmt, "Number:", ToString(alert.Number)))
	r.writeLine(fmt.Sprintf(labelFmt, "State:", ToString(alert.State)))

	if alert.Dependency != nil {
		if alert.Dependency.Package != nil {
			r.writeLine(fmt.Sprintf(labelFmt, "Package:", fmt.Sprintf("%s (%s)", ToString(alert.Dependency.Package.Name), ToString(alert.Dependency.Package.Ecosystem))))
		}
		r.writeLine(fmt.Sprintf(labelFmt, "Manifest:", ToString(alert.Dependency.ManifestPath)))
		r.writeLine(fmt.Sprintf(labelFmt, "Scope:", ToString(alert.Dependency.Scope)))
	}

	if alert.SecurityAdvisory != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Severity:", ToString(alert.SecurityAdvisory.Severity)))
		r.writeLine(fmt.Sprintf(labelFmt, "GHSA ID:", ToString(alert.SecurityAdvisory.GHSAID)))
		r.writeLine(fmt.Sprintf(labelFmt, "CVE ID:", ToString(alert.SecurityAdvisory.CVEID)))
		r.writeLine(fmt.Sprintf(labelFmt, "Summary:", ToString(alert.SecurityAdvisory.Summary)))
		if alert.SecurityAdvisory.CVSS != nil && alert.SecurityAdvisory.CVSS.Score != nil {
			r.writeLine(fmt.Sprintf(labelFmt, "CVSS:", fmt.Sprintf("%.1f (%s)", *alert.SecurityAdvisory.CVSS.Score, ToString(alert.SecurityAdvisory.CVSS.VectorString))))
		}
	}

	if alert.SecurityVulnerability != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Vulnerable:", ToString(alert.SecurityVulnerability.VulnerableVersionRange)))
		if alert.SecurityVulnerability.FirstPatchedVersion != nil {
			r.writeLine(fmt.Sprintf(labelFmt, "Patched:", ToString(alert.SecurityVulnerability.FirstPatchedVersion.Identifier)))
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

// RenderDependabotRepositoryAccess renders the Dependabot repository access information.
func (r *Renderer) RenderDependabotRepositoryAccess(access *client.DependabotRepositoryAccess) error {
	if r.exporter != nil {
		return r.RenderExportedData(access)
	}
	if access == nil {
		return nil
	}

	r.writeLine(fmt.Sprintf("Default Level: %s", access.DefaultLevel))
	r.writeLine("")

	if len(access.AccessibleRepositories) == 0 {
		r.writeLine("No accessible repositories.")
		return nil
	}

	headers := []string{"Repository", "Visibility", "Description"}
	getter := NewRepositoryFieldGetters()
	table := r.newTableWriter(headers)
	for _, repo := range access.AccessibleRepositories {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(repo, header)
		}
		table.Append(row)
	}
	return table.Render()
}
