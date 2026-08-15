package render

import (
	"fmt"

	"github.com/google/go-github/v90/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// RenderSBOMReportGeneration renders the result of requesting an SBOM report generation.
func (r *Renderer) RenderSBOMReportGeneration(result *client.SBOMReportGeneration) error {
	if r.exporter != nil {
		return r.RenderExportedData(result)
	}
	if result == nil {
		return nil
	}
	r.WriteLine(fmt.Sprintf("SBOMURL: %s", result.SBOMURL))
	return nil
}

// RenderSBOMReport renders a fetched SBOM report, or a pending message if the
// report is still being generated.
func (r *Renderer) RenderSBOMReport(report *client.SBOMReport) error {
	if r.exporter != nil {
		return r.RenderExportedData(report)
	}
	if report == nil {
		return nil
	}
	if report.Pending {
		r.WriteLine("SBOM report is still being generated; try again later")
		return nil
	}
	return r.RenderSBOMPackages(&github.SBOM{SBOM: report.SBOM}, nil)
}
