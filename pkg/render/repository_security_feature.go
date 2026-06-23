package render

import (
	"fmt"

	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

// RenderRepositorySecurityFeatureStatus renders the status of a repository security feature.
func (r *Renderer) RenderRepositorySecurityFeatureStatus(status *gh.RepositorySecurityFeatureStatus) error {
	if r.exporter != nil {
		return r.RenderExportedData(status)
	}
	if status == nil {
		return nil
	}

	const labelWidth = 10
	labelFmt := fmt.Sprintf("%%-%ds %%s", labelWidth)

	r.writeLine(fmt.Sprintf(labelFmt, "Feature:", status.Feature))
	r.writeLine(fmt.Sprintf(labelFmt, "Enabled:", ToString(status.Enabled)))
	if status.Paused != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Paused:", ToString(*status.Paused)))
	}
	return nil
}
