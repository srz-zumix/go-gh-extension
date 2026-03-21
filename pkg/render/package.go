package render

import (
	"maps"
	"slices"
	"strings"

	"github.com/google/go-github/v79/github"
)

// PackageFieldGetter defines a function to get a field value from a github.Package
type PackageFieldGetter func(pkg *github.Package) string

// PackageFieldGetters holds field getters for Package table rendering.
type PackageFieldGetters struct {
	Func map[string]PackageFieldGetter
}

// NewPackageFieldGetters creates field getters for Package table rendering
func NewPackageFieldGetters() *PackageFieldGetters {
	return &PackageFieldGetters{
		Func: map[string]PackageFieldGetter{
			"ID": func(pkg *github.Package) string {
				return ToString(pkg.ID)
			},
			"NAME": func(pkg *github.Package) string {
				return ToString(pkg.Name)
			},
			"TYPE": func(pkg *github.Package) string {
				return ToString(pkg.PackageType)
			},
			"VISIBILITY": func(pkg *github.Package) string {
				return ToString(pkg.Visibility)
			},
			"OWNER": func(pkg *github.Package) string {
				if pkg.Owner == nil {
					return ""
				}
				return ToString(pkg.Owner.Login)
			},
			"VERSIONS": func(pkg *github.Package) string {
				return ToString(pkg.VersionCount)
			},
			"URL": func(pkg *github.Package) string {
				return ToString(pkg.HTMLURL)
			},
			"CREATED_AT": func(pkg *github.Package) string {
				return ToString(pkg.CreatedAt)
			},
			"UPDATED_AT": func(pkg *github.Package) string {
				return ToString(pkg.UpdatedAt)
			},
			"REPOSITORY": func(pkg *github.Package) string {
				if pkg.Repository == nil {
					return ""
				}
				return ToString(pkg.Repository.FullName)
			},
		},
	}
}

// GetField returns the value of the specified field for the given package.
func (g *PackageFieldGetters) GetField(pkg *github.Package, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(pkg)
	}
	return ""
}

// Fields returns a sorted list of available field names.
func (g *PackageFieldGetters) Fields() []string {
	return slices.Sorted(maps.Keys(g.Func))
}

// RenderPackages renders a list of packages as a table using the given headers.
func (r *Renderer) RenderPackages(packages []*github.Package, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(packages)
		return
	}

	if len(packages) == 0 {
		r.writeLine("No packages.")
		return
	}

	if len(headers) == 0 {
		headers = []string{"NAME", "TYPE", "VISIBILITY", "VERSIONS"}
	}

	getter := NewPackageFieldGetters()
	table := r.newTableWriter(headers)

	for _, pkg := range packages {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(pkg, header)
		}
		table.Append(row)
	}
	table.Render()
}

// RenderPackage renders a single package as a key-value table.
func (r *Renderer) RenderPackage(pkg *github.Package) {
	if r.exporter != nil {
		r.RenderExportedData(pkg)
		return
	}

	if pkg == nil {
		r.writeLine("No package found.")
		return
	}

	table := r.newTableWriter([]string{"FIELD", "VALUE"})
	getter := NewPackageFieldGetters()
	for _, field := range getter.Fields() {
		value := getter.GetField(pkg, field)
		if value != "" {
			table.Append([]string{field, value})
		}
	}
	table.Render()
}

// PackageVersionFieldGetter defines a function to get a field value from a github.PackageVersion
type PackageVersionFieldGetter func(v *github.PackageVersion) string

// PackageVersionFieldGetters holds field getters for PackageVersion table rendering.
type PackageVersionFieldGetters struct {
	Func map[string]PackageVersionFieldGetter
}

// NewPackageVersionFieldGetters creates field getters for PackageVersion table rendering
func NewPackageVersionFieldGetters() *PackageVersionFieldGetters {
	return &PackageVersionFieldGetters{
		Func: map[string]PackageVersionFieldGetter{
			"ID": func(v *github.PackageVersion) string {
				return ToString(v.ID)
			},
			"NAME": func(v *github.PackageVersion) string {
				return ToString(v.Name)
			},
			"URL": func(v *github.PackageVersion) string {
				if v.HTMLURL != nil {
					return ToString(v.HTMLURL)
				}
				return ToString(v.URL)
			},
			"CREATED_AT": func(v *github.PackageVersion) string {
				return ToString(v.CreatedAt)
			},
			"UPDATED_AT": func(v *github.PackageVersion) string {
				return ToString(v.UpdatedAt)
			},
			"DESCRIPTION": func(v *github.PackageVersion) string {
				return ToString(v.Description)
			},
			"LICENSE": func(v *github.PackageVersion) string {
				return ToString(v.License)
			},
		},
	}
}

// GetField returns the value of the specified field for the given package version.
func (g *PackageVersionFieldGetters) GetField(v *github.PackageVersion, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(v)
	}
	return ""
}

// Fields returns a sorted list of available field names.
func (g *PackageVersionFieldGetters) Fields() []string {
	return slices.Sorted(maps.Keys(g.Func))
}

// RenderPackageVersions renders a list of package versions as a table using the given headers.
func (r *Renderer) RenderPackageVersions(versions []*github.PackageVersion, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(versions)
	}

	if len(versions) == 0 {
		r.writeLine("No package versions.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"ID", "NAME", "CREATED_AT"}
	}

	getter := NewPackageVersionFieldGetters()
	table := r.newTableWriter(headers)

	for _, v := range versions {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.GetField(v, header)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderPackageVersion renders a single package version as a key-value table.
func (r *Renderer) RenderPackageVersion(v *github.PackageVersion) error {
	if r.exporter != nil {
		return r.RenderExportedData(v)
	}

	if v == nil {
		r.writeLine("No package version found.")
		return nil
	}

	table := r.newTableWriter([]string{"FIELD", "VALUE"})
	getter := NewPackageVersionFieldGetters()
	for _, field := range getter.Fields() {
		value := getter.GetField(v, field)
		if value != "" {
			table.Append([]string{field, value})
		}
	}
	return table.Render()
}
