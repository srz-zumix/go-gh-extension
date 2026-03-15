package render

import "fmt"

func (r *Renderer) renderExportedData(data any) error {
	if r.exporter == nil {
		return fmt.Errorf("no exporter available")
	}
	return r.exporter.Write(r.IO, data)
}

func (r *Renderer) RenderExportedData(data any) error {
	return r.renderExportedData(data)
}
