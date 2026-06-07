package render

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v84/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// DependencyChangeFieldGetter defines a function to get a field value from a DependencyChange
type DependencyChangeFieldGetter func(c *client.DependencyChange) string

// DependencyChangeFieldGetters manages getter functions for each header
type DependencyChangeFieldGetters struct {
	Func map[string]DependencyChangeFieldGetter
}

func newDependencyChangeFieldGetters() *DependencyChangeFieldGetters {
	return &DependencyChangeFieldGetters{
		Func: map[string]DependencyChangeFieldGetter{
			"CHANGETYPE": func(c *client.DependencyChange) string { return c.ChangeType },
			"MANIFEST":   func(c *client.DependencyChange) string { return c.Manifest },
			"ECOSYSTEM":  func(c *client.DependencyChange) string { return c.Ecosystem },
			"NAME":       func(c *client.DependencyChange) string { return c.Name },
			"VERSION":    func(c *client.DependencyChange) string { return c.Version },
			"PACKAGEURL": func(c *client.DependencyChange) string { return c.PackageURL },
			"LICENSE":    func(c *client.DependencyChange) string { return c.License },
			"SCOPE":      func(c *client.DependencyChange) string { return c.Scope },
			"VULNERABILITIES": func(c *client.DependencyChange) string {
				if len(c.Vulnerabilities) == 0 {
					return ""
				}
				ids := make([]string, len(c.Vulnerabilities))
				for i, v := range c.Vulnerabilities {
					ids[i] = v.AdvisoryGHSAID
				}
				return strings.Join(ids, ", ")
			},
		},
	}
}

func (g *DependencyChangeFieldGetters) getField(c *client.DependencyChange, field string) string {
	key := strings.ToUpper(strings.ReplaceAll(field, "_", ""))
	if getter, ok := g.Func[key]; ok {
		return getter(c)
	}
	return ""
}

// RenderDependencyChanges renders a list of dependency changes as a table.
func (r *Renderer) RenderDependencyChanges(changes []*client.DependencyChange, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(changes)
	}
	if len(changes) == 0 {
		return nil
	}
	if len(headers) == 0 {
		headers = []string{"ChangeType", "Manifest", "Ecosystem", "Name", "Version", "Scope"}
	}
	getter := newDependencyChangeFieldGetters()
	table := r.newTableWriter(headers)
	for _, c := range changes {
		row := make([]string, len(headers))
		for i, h := range headers {
			row[i] = getter.getField(c, h)
		}
		table.Append(row)
	}
	return table.Render()
}

// RenderDependencyGraphSnapshotResult renders the result of creating a dependency graph snapshot.
func (r *Renderer) RenderDependencyGraphSnapshotResult(result *github.DependencyGraphSnapshotCreationData) error {
	if r.exporter != nil {
		return r.RenderExportedData(result)
	}
	if result == nil {
		return nil
	}
	createdAt := ""
	if result.CreatedAt != nil {
		createdAt = result.CreatedAt.Format(TimeFormat)
	}
	message := ""
	if result.Message != nil {
		message = *result.Message
	}
	res := ""
	if result.Result != nil {
		res = *result.Result
	}
	fmt.Fprintf(r.IO.Out, "ID:         %d\n", result.ID)
	fmt.Fprintf(r.IO.Out, "CreatedAt:  %s\n", createdAt)
	fmt.Fprintf(r.IO.Out, "Result:     %s\n", res)
	fmt.Fprintf(r.IO.Out, "Message:    %s\n", message)
	return nil
}
