package render

import (
	"maps"
	"slices"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/unity"
)

// UnityPackageFields lists all available field names for UnityPackage table rendering.
var UnityPackageFields = []string{"Name", "Version", "SHA", "Path", "URL"}

// UnityPackageFieldGetter defines a function to get a field value from a unity.UnityPackage.
type UnityPackageFieldGetter func(pkg *unity.UnityPackage) string

// UnityPackageFieldGetters holds field getters for UnityPackage table rendering.
type UnityPackageFieldGetters struct {
	Func map[string]UnityPackageFieldGetter
}

// NewUnityPackageFieldGetters creates field getters for UnityPackage table rendering.
func NewUnityPackageFieldGetters() *UnityPackageFieldGetters {
	return &UnityPackageFieldGetters{
		Func: map[string]UnityPackageFieldGetter{
			"NAME":    func(pkg *unity.UnityPackage) string { return pkg.Name },
			"VERSION": func(pkg *unity.UnityPackage) string { return pkg.Version },
			"SHA":     func(pkg *unity.UnityPackage) string { return pkg.SHA },
			"PATH":    func(pkg *unity.UnityPackage) string { return pkg.Path },
			"URL":     func(pkg *unity.UnityPackage) string { return pkg.URL },
		},
	}
}

// GetField returns the value of the specified field for the given package.
func (g *UnityPackageFieldGetters) GetField(pkg *unity.UnityPackage, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(pkg)
	}
	return ""
}

// Fields returns a sorted list of available field names.
func (g *UnityPackageFieldGetters) Fields() []string {
	return slices.Sorted(maps.Keys(g.Func))
}

// RenderUnityPackages renders Unity manifest dependencies as a table.
// headers specifies which fields to display; defaults to all fields when empty.
func (r *Renderer) RenderUnityPackages(packages []unity.UnityPackage, headers []string) {
	if r.exporter != nil {
		r.RenderExportedData(packages)
		return
	}
	if len(packages) == 0 {
		return
	}
	if len(headers) == 0 {
		headers = UnityPackageFields
	}
	getter := NewUnityPackageFieldGetters()
	table := r.newTableWriter(headers)
	for i := range packages {
		row := make([]string, len(headers))
		for j, header := range headers {
			row[j] = getter.GetField(&packages[i], header)
		}
		table.Append(row)
	}
	table.Render()
}
