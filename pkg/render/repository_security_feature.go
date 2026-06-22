package render

import (
	"fmt"
	"strconv"

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
	r.writeLine(fmt.Sprintf(labelFmt, "Enabled:", strconv.FormatBool(status.Enabled)))
	if status.Paused != nil {
		r.writeLine(fmt.Sprintf(labelFmt, "Paused:", strconv.FormatBool(*status.Paused)))
	}
	return nil
}
