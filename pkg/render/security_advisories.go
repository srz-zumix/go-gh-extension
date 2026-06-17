package render

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v88/github"
)

// RenderRepositorySecurityAdvisories renders a list of repository security advisories as a table.
func (r *Renderer) RenderRepositorySecurityAdvisories(advisories []*github.SecurityAdvisory) error {
	if r.exporter != nil {
		return r.RenderExportedData(advisories)
	}
	if len(advisories) == 0 {
		r.writeLine("No repository security advisories")
		return nil
	}
	headers := []string{"GHSA ID", "CVE ID", "State", "Severity", "Summary", "Published At"}
	table := r.newTableWriter(headers)
	for _, a := range advisories {
		table.Append([]string{
			ToString(a.GHSAID),
			ToString(a.CVEID),
			ToString(a.State),
			ToString(a.Severity),
			ToString(a.Summary),
			ToString(a.PublishedAt),
		})
	}
	return table.Render()
}

// RenderRepositorySecurityAdvisory renders a single repository security advisory.
func (r *Renderer) RenderRepositorySecurityAdvisory(advisory *github.SecurityAdvisory) error {
	if r.exporter != nil {
		return r.RenderExportedData(advisory)
	}
	if advisory == nil {
		return nil
	}

	labelFmt := "%-13s %s"

	r.writeLine(fmt.Sprintf(labelFmt, "GHSA ID:", ToString(advisory.GHSAID)))
	r.writeLine(fmt.Sprintf(labelFmt, "CVE ID:", ToString(advisory.CVEID)))
	r.writeLine(fmt.Sprintf(labelFmt, "State:", ToString(advisory.State)))
	r.writeLine(fmt.Sprintf(labelFmt, "Severity:", ToString(advisory.Severity)))
	r.writeLine(fmt.Sprintf(labelFmt, "Summary:", ToString(advisory.Summary)))
	r.writeLine(fmt.Sprintf(labelFmt, "Published At:", ToString(advisory.PublishedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "Created At:", ToString(advisory.CreatedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "Updated At:", ToString(advisory.UpdatedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "URL:", ToString(advisory.HTMLURL)))

	if len(advisory.CWEIDs) > 0 {
		r.writeLine(fmt.Sprintf(labelFmt, "CWEs:", strings.Join(advisory.CWEIDs, ", ")))
	}

	if len(advisory.Vulnerabilities) > 0 {
		r.writeLine("")
		r.writeLine("Vulnerabilities:")
		headers := []string{"Ecosystem", "Package", "Vulnerable Range", "Patched Versions"}
		table := r.newTableWriter(headers)
		for _, v := range advisory.Vulnerabilities {
			ecosystem := ""
			name := ""
			if v.Package != nil {
				ecosystem = ToString(v.Package.Ecosystem)
				name = ToString(v.Package.Name)
			}
			table.Append([]string{
				ecosystem,
				name,
				ToString(v.VulnerableVersionRange),
				ToString(v.PatchedVersions),
			})
		}
		if err := table.Render(); err != nil {
			return err
		}
	}

	if len(advisory.Credits) > 0 {
		r.writeLine("")
		r.writeLine("Credits:")
		headers := []string{"Login", "Type"}
		table := r.newTableWriter(headers)
		for _, c := range advisory.Credits {
			table.Append([]string{
				ToString(c.Login),
				ToString(c.Type),
			})
		}
		if err := table.Render(); err != nil {
			return err
		}
	}

	return nil
}

// RenderGlobalSecurityAdvisories renders a list of global security advisories as a table.
func (r *Renderer) RenderGlobalSecurityAdvisories(advisories []*github.GlobalSecurityAdvisory) error {
	if r.exporter != nil {
		return r.RenderExportedData(advisories)
	}
	if len(advisories) == 0 {
		r.writeLine("No global security advisories")
		return nil
	}
	headers := []string{"GHSA ID", "CVE ID", "Type", "Severity", "Summary", "Published At"}
	table := r.newTableWriter(headers)
	for _, a := range advisories {
		table.Append([]string{
			ToString(a.GHSAID),
			ToString(a.CVEID),
			ToString(a.Type),
			ToString(a.Severity),
			ToString(a.Summary),
			ToString(a.PublishedAt),
		})
	}
	return table.Render()
}

// RenderGlobalSecurityAdvisory renders a single global security advisory.
func (r *Renderer) RenderGlobalSecurityAdvisory(advisory *github.GlobalSecurityAdvisory) error {
	if r.exporter != nil {
		return r.RenderExportedData(advisory)
	}
	if advisory == nil {
		return nil
	}

	labelFmt := "%-20s %s"

	r.writeLine(fmt.Sprintf(labelFmt, "GHSA ID:", ToString(advisory.GHSAID)))
	r.writeLine(fmt.Sprintf(labelFmt, "CVE ID:", ToString(advisory.CVEID)))
	r.writeLine(fmt.Sprintf(labelFmt, "Type:", ToString(advisory.Type)))
	r.writeLine(fmt.Sprintf(labelFmt, "Severity:", ToString(advisory.Severity)))
	r.writeLine(fmt.Sprintf(labelFmt, "Summary:", ToString(advisory.Summary)))
	r.writeLine(fmt.Sprintf(labelFmt, "Published At:", ToString(advisory.PublishedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "GitHub Reviewed At:", ToString(advisory.GithubReviewedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "NVD Published At:", ToString(advisory.NVDPublishedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "Updated At:", ToString(advisory.UpdatedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "URL:", ToString(advisory.HTMLURL)))

	if advisory.SourceCodeLocation != nil && *advisory.SourceCodeLocation != "" {
		r.writeLine(fmt.Sprintf(labelFmt, "Source Code:", ToString(advisory.SourceCodeLocation)))
	}

	if advisory.RepositoryAdvisoryURL != nil && *advisory.RepositoryAdvisoryURL != "" {
		r.writeLine(fmt.Sprintf(labelFmt, "Repository Advisory:", ToString(advisory.RepositoryAdvisoryURL)))
	}

	cwes := make([]string, 0)
	for _, cwe := range advisory.CWEs {
		if cwe != nil && cwe.CWEID != nil {
			cwes = append(cwes, *cwe.CWEID)
		}
	}
	if len(cwes) > 0 {
		r.writeLine(fmt.Sprintf(labelFmt, "CWEs:", strings.Join(cwes, ", ")))
	}

	if len(advisory.Vulnerabilities) > 0 {
		r.writeLine("")
		r.writeLine("Vulnerabilities:")
		headers := []string{"Ecosystem", "Package", "Vulnerable Range", "Patched Version"}
		table := r.newTableWriter(headers)
		for _, v := range advisory.Vulnerabilities {
			ecosystem := ""
			name := ""
			if v.Package != nil {
				ecosystem = ToString(v.Package.Ecosystem)
				name = ToString(v.Package.Name)
			}
			table.Append([]string{
				ecosystem,
				name,
				ToString(v.VulnerableVersionRange),
				ToString(v.FirstPatchedVersion),
			})
		}
		if err := table.Render(); err != nil {
			return err
		}
	}

	if len(advisory.Credits) > 0 {
		r.writeLine("")
		r.writeLine("Credits:")
		headers := []string{"Login", "Type"}
		table := r.newTableWriter(headers)
		for _, c := range advisory.Credits {
			login := ""
			if c.User != nil {
				login = ToString(c.User.Login)
			}
			table.Append([]string{
				login,
				ToString(c.Type),
			})
		}
		if err := table.Render(); err != nil {
			return err
		}
	}

	return nil
}

