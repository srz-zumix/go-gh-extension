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
				if len(names) > 1 {
					return names[0]
				}
				return ""
			},
			"NAME": func(pkg *github.RepoDependencies) string {
				name := ToString(pkg.Name)
				names := strings.Split(name, ":")
				if len(names) > 1 {
					return names[1]
				}
				return name
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


func (r *Renderer) RenderRepositoryDependencies(deps []*github.RepoDependencies, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(deps)
		return
	}
	if deps == nil {
		return
	}

	if len(headers) == 0 {
		headers = []string{"Package", "Name", "Version"}
	}
	getter := NewSBOMPackageFieldGetters()
	table := r.newTableWriter(headers)
	for _, pkg := range deps {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(pkg, header)
		}
		table.Append(row)
	}
	table.Render()
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
	r.RenderRepositoryDependencies(sbom.SBOM.Packages, headers)
}

// RenderSBOMPackagesDefault renders SBOM packages with default headers
func (r *Renderer) RenderSBOMPackagesDefault(sbom *github.SBOM) {
	r.RenderSBOMPackages(sbom, nil)
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

type SBOMInfoFieldGetter func(pkg *github.SBOMInfo) string
type SBOMInfoFieldGetters struct {
	Func map[string]SBOMInfoFieldGetter
}

func NewSBOMInfoFieldGetters() *SBOMInfoFieldGetters {
	return &SBOMInfoFieldGetters{
		Func: map[string]SBOMInfoFieldGetter{
			"NAME": func(info *github.SBOMInfo) string {
				return ToString(info.Name)
			},
			"SPDXVERSION": func(info *github.SBOMInfo) string {
				return ToString(info.SPDXVersion)
			},
			"SPDXID": func(info *github.SBOMInfo) string {
				return ToString(info.SPDXID)
			},
			// Add more fields as needed
		},
	}
}

func (g *SBOMInfoFieldGetters) GetField(info *github.SBOMInfo, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(info)
	}
	return ""
}

func (r *Renderer) RenderSBOMInfo(sbomInfo *github.SBOMInfo, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(sbomInfo)
		return
	}

	if sbomInfo == nil {
		return
	}

	if len(headers) == 0 {
		headers = []string{"Name", "SPDXVersion", "SPDXID"}
	}
	getter := NewSBOMInfoFieldGetters()
	table := r.newTableWriter(headers)
	row := make([]string, len(headers))
	for i, header := range headers {
		row[i] = getter.GetField(sbomInfo, header)
	}
	table.Append(row)
	table.Render()
}

// RenderMultipleSBOMPackages merges packages from multiple SBOMs and renders them as a single table
func (r *Renderer) RenderMultipleSBOMPackages(sboms []*github.SBOM, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(sboms)
		return
	}
	for _, sbom := range sboms {
		if sbom == nil || sbom.SBOM == nil {
			// skip SBOM entries that do not contain valid SBOM metadata
			continue
		}
		r.writeLine(ToString(sbom.SBOM.Name))
		r.RenderSBOMPackages(sbom, headers)
	}
}
