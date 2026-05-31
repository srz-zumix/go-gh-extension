package render

import (
	"fmt"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// RenderCodeQualitySetup renders a code quality setup configuration.
func (r *Renderer) RenderCodeQualitySetup(setup *client.CodeQualitySetup) error {
	if r.exporter != nil {
		return r.RenderExportedData(setup)
	}
	if setup == nil {
		return nil
	}

	labelFmt := "%-15s %s"
	runnerLabel := ""
	if setup.RunnerLabel != nil {
		runnerLabel = *setup.RunnerLabel
	}

	r.writeLine(fmt.Sprintf(labelFmt, "State:", setup.State))
	r.writeLine(fmt.Sprintf(labelFmt, "Languages:", strings.Join(setup.Languages, ", ")))
	r.writeLine(fmt.Sprintf(labelFmt, "Runner Type:", setup.RunnerType))
	if runnerLabel != "" {
		r.writeLine(fmt.Sprintf(labelFmt, "Runner Label:", runnerLabel))
	}
	r.writeLine(fmt.Sprintf(labelFmt, "Schedule:", setup.Schedule))
	r.writeLine(fmt.Sprintf(labelFmt, "Updated At:", setup.UpdatedAt))
	return nil
}
