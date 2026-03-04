package render

import (
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

func getName(item any) string {
	return gh.GetObjectName(item)
}
func getNames(items any) []string {
	return gh.GetObjectNames(items)
}

func (r *Renderer) RenderNames(items any) {
	names := getNames(items)
	if r.exporter != nil {
		r.RenderExportedData(names)
		return
	}

	if names == nil {
		return
	}

	r.writeLine(strings.Join(names, "\n"))
}

// RenderNamesWithSeparator renders the names joined by the specified separator
func (r *Renderer) RenderNamesWithSeparator(items any, sep string) {
	names := getNames(items)
	if r.exporter != nil {
		r.RenderExportedData(names)
		return
	}

	if names == nil {
		return
	}

	r.writeLine(strings.Join(names, sep))
}

// RenderVersionedNames renders versioned names (name@ref) for ActionReference slices.
// For other types, falls back to RenderNames.
func (r *Renderer) RenderVersionedNames(items any) {
	switch v := items.(type) {
	case []parser.ActionReference:
		names := make([]string, len(v))
		for i, ref := range v {
			names[i] = ref.VersionedName()
		}
		if r.exporter != nil {
			r.RenderExportedData(names)
			return
		}
		r.writeLine(strings.Join(names, "\n"))
	default:
		r.RenderNames(items)
	}
}
