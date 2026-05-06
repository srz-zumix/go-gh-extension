package render

import (
	"fmt"

	"github.com/google/go-github/v84/github"
)

// RenderSecretScanningPatternConfigs renders secret scanning pattern configurations as a table.
func (r *Renderer) RenderSecretScanningPatternConfigs(configs *github.SecretScanningPatternConfigs) error {
	if r.exporter != nil {
		return r.RenderExportedData(configs)
	}
	if configs == nil {
		return nil
	}

	wroteSection := false

	if len(configs.ProviderPatternOverrides) > 0 {
		r.writeLine(fmt.Sprintf("Pattern Config Version: %s", ToString(configs.PatternConfigVersion)))
		r.writeLine("")
		r.writeLine("Provider Patterns:")
		headers := []string{"Token Type", "Slug", "Display Name", "Setting", "Default Setting", "Enterprise Setting"}
		table := r.newTableWriter(headers)
		for _, p := range configs.ProviderPatternOverrides {
			table.Append([]string{
				ToString(p.TokenType),
				ToString(p.Slug),
				ToString(p.DisplayName),
				ToString(p.Setting),
				ToString(p.DefaultSetting),
				ToString(p.EnterpriseSetting),
			})
		}
		if err := table.Render(); err != nil {
			return err
		}
		wroteSection = true
	}

	if len(configs.CustomPatternOverrides) > 0 {
		if wroteSection {
			r.writeLine("")
		}
		r.writeLine("Custom Patterns:")
		headers := []string{"Token Type", "Version", "Slug", "Display Name", "Setting", "Default Setting"}
		table := r.newTableWriter(headers)
		for _, p := range configs.CustomPatternOverrides {
			table.Append([]string{
				ToString(p.TokenType),
				ToString(p.CustomPatternVersion),
				ToString(p.Slug),
				ToString(p.DisplayName),
				ToString(p.Setting),
				ToString(p.DefaultSetting),
			})
		}
		if err := table.Render(); err != nil {
			return err
		}
	}

	return nil
}

// RenderSecretScanningAlerts renders a list of secret scanning alerts as a table.
func (r *Renderer) RenderSecretScanningAlerts(alerts []*github.SecretScanningAlert) error {
	if r.exporter != nil {
		return r.RenderExportedData(alerts)
	}
	if len(alerts) == 0 {
		fmt.Println("No secret scanning alerts")
		return nil
	}
	headers := []string{"Number", "State", "Secret Type", "Validity", "Created At", "URL"}
	table := r.newTableWriter(headers)
	for _, a := range alerts {
		table.Append([]string{
			ToString(a.Number),
			ToString(a.State),
			ToString(a.SecretType),
			ToString(a.Validity),
			ToString(a.CreatedAt),
			ToString(a.HTMLURL),
		})
	}
	return table.Render()
}

// RenderSecretScanningAlert renders a single secret scanning alert.
func (r *Renderer) RenderSecretScanningAlert(alert *github.SecretScanningAlert) error {
	if r.exporter != nil {
		return r.RenderExportedData(alert)
	}
	if alert == nil {
		return nil
	}
	r.writeLine(fmt.Sprintf("Number:      %s", ToString(alert.Number)))
	r.writeLine(fmt.Sprintf("State:       %s", ToString(alert.State)))
	r.writeLine(fmt.Sprintf("Secret Type: %s", ToString(alert.SecretType)))
	r.writeLine(fmt.Sprintf("Validity:    %s", ToString(alert.Validity)))
	r.writeLine(fmt.Sprintf("Resolution:  %s", ToString(alert.Resolution)))
	r.writeLine(fmt.Sprintf("Created At:  %s", ToString(alert.CreatedAt)))
	r.writeLine(fmt.Sprintf("Updated At:  %s", ToString(alert.UpdatedAt)))
	r.writeLine(fmt.Sprintf("URL:         %s", ToString(alert.HTMLURL)))
	return nil
}

// RenderSecretScanningAlertLocations renders a list of secret scanning alert locations as a table.
func (r *Renderer) RenderSecretScanningAlertLocations(locations []*github.SecretScanningAlertLocation) error {
	if r.exporter != nil {
		return r.RenderExportedData(locations)
	}
	if len(locations) == 0 {
		r.writeLine("No secret scanning alert locations")
		return nil
	}
	headers := []string{"Type", "Path", "Start Line", "End Line", "Commit SHA"}
	table := r.newTableWriter(headers)
	for _, l := range locations {
		path := ""
		startLine := ""
		endLine := ""
		commitSHA := ""
		if l.Details != nil {
			path = ToString(l.Details.Path)
			startLine = ToString(l.Details.Startline)
			endLine = ToString(l.Details.EndLine)
			commitSHA = ToString(l.Details.CommitSHA)
		}
		table.Append([]string{
			ToString(l.Type),
			path,
			startLine,
			endLine,
			commitSHA,
		})
	}
	return table.Render()
}

// RenderSecretScanningScanHistory renders the secret scanning scan history.
func (r *Renderer) RenderSecretScanningScanHistory(history *github.SecretScanningScanHistory) error {
	if r.exporter != nil {
		return r.RenderExportedData(history)
	}
	if history == nil {
		return nil
	}
	headers := []string{"Category", "Type", "Status", "Started At", "Completed At"}
	table := r.newTableWriter(headers)
	for _, s := range history.IncrementalScans {
		table.Append([]string{"incremental", s.Type, s.Status, ToString(s.StartedAt), ToString(s.CompletedAt)})
	}
	for _, s := range history.BackfillScans {
		table.Append([]string{"backfill", s.Type, s.Status, ToString(s.StartedAt), ToString(s.CompletedAt)})
	}
	for _, s := range history.PatternUpdateScans {
		table.Append([]string{"pattern_update", s.Type, s.Status, ToString(s.StartedAt), ToString(s.CompletedAt)})
	}
	for _, s := range history.CustomPatternBackfillScans {
		table.Append([]string{"custom_pattern_backfill", s.Type, s.Status, ToString(s.StartedAt), ToString(s.CompletedAt)})
	}
	return table.Render()
}


