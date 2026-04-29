package render

import (
	"fmt"

	"github.com/google/go-github/v84/github"
)

// RenderCodeSecurityConfigurations renders a list of code security configurations as a table.
func (r *Renderer) RenderCodeSecurityConfigurations(configs []*github.CodeSecurityConfiguration) error {
	if r.exporter != nil {
		return r.RenderExportedData(configs)
	}
	if len(configs) == 0 {
		r.writeLine("No code security configurations found.")
		return nil
	}
	headers := []string{"ID", "Name", "Target", "Enforcement", "Updated"}
	table := r.newTableWriter(headers)
	for _, c := range configs {
		row := []string{
			ToString(c.ID),
			c.Name,
			ToString(c.TargetType),
			ToString(c.Enforcement),
			ToString(c.UpdatedAt),
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderCodeSecurityConfiguration renders a single code security configuration in detail format.
func (r *Renderer) RenderCodeSecurityConfiguration(c *github.CodeSecurityConfiguration) error {
	if r.exporter != nil {
		return r.RenderExportedData(c)
	}
	if c == nil {
		return nil
	}

	// labelWidth keeps detail output aligned. It includes padding beyond the longest current label.
	const labelWidth = 37
	labelFmt := fmt.Sprintf("%%-%ds %%s", labelWidth)

	r.writeLine(fmt.Sprintf(labelFmt, "ID:", ToString(c.ID)))
	r.writeLine(fmt.Sprintf(labelFmt, "Name:", c.Name))
	r.writeLine(fmt.Sprintf(labelFmt, "Description:", c.Description))
	r.writeLine(fmt.Sprintf(labelFmt, "Target Type:", ToString(c.TargetType)))
	r.writeLine(fmt.Sprintf(labelFmt, "Enforcement:", ToString(c.Enforcement)))

	r.writeLine(fmt.Sprintf(labelFmt, "Advanced Security:", ToString(c.AdvancedSecurity)))
	if c.CodeSecurity != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Code Security:", ToString(c.CodeSecurity)))
	}
	if c.SecretProtection != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Secret Protection:", ToString(c.SecretProtection)))
	}

	r.writeLine(fmt.Sprintf(labelFmt, "Dependency Graph:", ToString(c.DependencyGraph)))
	r.writeLine(fmt.Sprintf(labelFmt, "Dep. Graph Autosubmit:", ToString(c.DependencyGraphAutosubmitAction)))
	r.writeLine(fmt.Sprintf(labelFmt, "Dependabot Alerts:", ToString(c.DependabotAlerts)))
	r.writeLine(fmt.Sprintf(labelFmt, "Dependabot Security Updates:", ToString(c.DependabotSecurityUpdates)))

	r.writeLine(fmt.Sprintf(labelFmt, "Code Scanning Default Setup:", ToString(c.CodeScanningDefaultSetup)))
	if c.CodeScanningDelegatedAlertDismissal != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Code Scanning Delegated Dismissal:", ToString(c.CodeScanningDelegatedAlertDismissal)))
	}

	r.writeLine(fmt.Sprintf(labelFmt, "Secret Scanning:", ToString(c.SecretScanning)))
	r.writeLine(fmt.Sprintf(labelFmt, "Secret Scanning Push Protection:", ToString(c.SecretScanningPushProtection)))
	if c.SecretScanningDelegatedBypass != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Secret Scanning Delegated Bypass:", ToString(c.SecretScanningDelegatedBypass)))
	}
	if c.SecretScanningValidityChecks != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Secret Scanning Validity Checks:", ToString(c.SecretScanningValidityChecks)))
	}
	if c.SecretScanningNonProviderPatterns != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Secret Scanning Non-Provider Patterns:", ToString(c.SecretScanningNonProviderPatterns)))
	}
	if c.SecretScanningGenericSecrets != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Secret Scanning Generic Secrets:", ToString(c.SecretScanningGenericSecrets)))
	}
	if c.SecretScanningDelegatedAlertDismissal != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Secret Scanning Delegated Dismissal:", ToString(c.SecretScanningDelegatedAlertDismissal)))
	}

	r.writeLine(fmt.Sprintf(labelFmt, "Private Vulnerability Reporting:", ToString(c.PrivateVulnerabilityReporting)))

	r.writeLine(fmt.Sprintf(labelFmt, "URL:", ToString(c.HTMLURL)))
	r.writeLine(fmt.Sprintf(labelFmt, "Created:", ToString(c.CreatedAt)))
	r.writeLine(fmt.Sprintf(labelFmt, "Updated:", ToString(c.UpdatedAt)))
	return nil
}

// RenderDefaultCodeSecurityConfigurations renders default configurations as a table.
func (r *Renderer) RenderDefaultCodeSecurityConfigurations(defaults []*github.CodeSecurityConfigurationWithDefaultForNewRepos) error {
	if r.exporter != nil {
		return r.RenderExportedData(defaults)
	}
	if len(defaults) == 0 {
		r.writeLine("No default code security configurations found.")
		return nil
	}
	headers := []string{"Default For", "ID", "Name", "Target", "Enforcement"}
	table := r.newTableWriter(headers)
	for _, d := range defaults {
		row := []string{ToString(d.DefaultForNewRepos), "", "", "", ""}
		if d.Configuration != nil {
			row[1] = ToString(d.Configuration.ID)
			row[2] = d.Configuration.Name
			row[3] = ToString(d.Configuration.TargetType)
			row[4] = ToString(d.Configuration.Enforcement)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderDefaultCodeSecurityConfiguration renders a single default configuration result in detail format.
func (r *Renderer) RenderDefaultCodeSecurityConfiguration(d *github.CodeSecurityConfigurationWithDefaultForNewRepos) error {
	if r.exporter != nil {
		return r.RenderExportedData(d)
	}
	if d == nil {
		return nil
	}
	const labelFmt = "%-15s %s"
	r.writeLine(fmt.Sprintf(labelFmt, "Default For:", ToString(d.DefaultForNewRepos)))
	if d.Configuration != nil {
		r.writeLine("")
		return r.RenderCodeSecurityConfiguration(d.Configuration)
	}
	return nil
}

// RenderCodeSecurityConfigurationRepositories renders repository attachments as a table.
func (r *Renderer) RenderCodeSecurityConfigurationRepositories(attachments []*github.RepositoryAttachment) error {
	if r.exporter != nil {
		return r.RenderExportedData(attachments)
	}
	if len(attachments) == 0 {
		r.writeLine("No repositories attached to this configuration.")
		return nil
	}
	headers := []string{"Status", "Repository", "ID"}
	table := r.newTableWriter(headers)
	for _, a := range attachments {
		row := []string{ToString(a.Status), "", ""}
		if a.Repository != nil {
			row[1] = ToString(a.Repository.FullName)
			row[2] = ToString(a.Repository.ID)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderRepositoryCodeSecurityConfiguration renders the configuration attached to a repository in detail format.
func (r *Renderer) RenderRepositoryCodeSecurityConfiguration(c *github.RepositoryCodeSecurityConfiguration) error {
	if r.exporter != nil {
		return r.RenderExportedData(c)
	}
	if c == nil {
		return nil
	}
	const labelFmt = "%-10s %s"
	r.writeLine(fmt.Sprintf(labelFmt, "State:", ToString(c.State)))
	if c.Configuration != nil {
		r.writeLine("")
		return r.RenderCodeSecurityConfiguration(c.Configuration)
	}
	return nil
}
