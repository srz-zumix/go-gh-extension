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
