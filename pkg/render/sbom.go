package render

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v79/github"
)

// SBOMPackageFieldGetter defines a function to get a field value from github.SBOMPackage
// This struct manages getter functions for each header
type SBOMPackageFieldGetter func(pkg *github.RepoDependencies) string
type SBOMPackageFieldGetters struct {
	Func map[string]SBOMPackageFieldGetter
}

func NewSBOMPackageFieldGetters() *SBOMPackageFieldGetters {
	return &SBOMPackageFieldGetters{
		Func: map[string]SBOMPackageFieldGetter{
			"PACKAGE": func(pkg *github.RepoDependencies) string {
				names := strings.Split(ToString(pkg.Name), ":")
				if len(names) > 0 {
					return names[0]
				}
				return ""
			},
			"NAME": func(pkg *github.RepoDependencies) string {
				names := strings.Split(ToString(pkg.Name), ":")
				if len(names) > 1 {
					return names[1]
				}
				return ""
			},
			"VERSION": func(pkg *github.RepoDependencies) string {
				return ToString(pkg.VersionInfo)
			},
			"ID": func(pkg *github.RepoDependencies) string {
				return ToString(pkg.SPDXID)
			},
			"DOWNLOADLOCATION": func(pkg *github.RepoDependencies) string {
				return ToString(pkg.DownloadLocation)
			},
			// Add more fields as needed
		},
	}
}

func (g *SBOMPackageFieldGetters) GetField(pkg *github.RepoDependencies, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(pkg)
	}
	return ""
}

// RenderSBOMPackagesWithHeaders renders SBOM packages with custom headers
func (r *Renderer) RenderSBOMPackages(sbom *github.SBOM, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(sbom)
		return
	}
	if sbom == nil {
		return
	}
	getter := NewSBOMPackageFieldGetters()
	table := r.newTableWriter(headers)
	for _, pkg := range sbom.SBOM.Packages {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(pkg, header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderSBOMPackagesDefault renders SBOM packages with default headers
func (r *Renderer) RenderSBOMPackagesDefault(sbom *github.SBOM) {
	headers := []string{"Package", "Name", "Version"}
	r.RenderSBOMPackages(sbom, headers)
}

func (r *Renderer) RenderSBOMPackagesVersionList(sbom *github.SBOM) {
	if r.exporter != nil {
		r.RenderExportedData(sbom)
		return
	}

	if sbom == nil {
		return
	}
	packages := []string{}
	for _, pkg := range sbom.SBOM.Packages {
		names := strings.Split(*pkg.Name, ":")
		packages = append(packages, fmt.Sprintf("%s@%s", names[1], *pkg.VersionInfo))
	}
	r.writeLine(strings.Join(packages, "\n"))
}
