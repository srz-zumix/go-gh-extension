package render

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
)

type danglingCommitFieldGetter func(c *gh.DanglingCommit) string
type danglingCommitFieldGetters struct {
	Func map[string]danglingCommitFieldGetter
}

func newDanglingCommitFieldGetters() *danglingCommitFieldGetters {
	return &danglingCommitFieldGetters{
		Func: map[string]danglingCommitFieldGetter{
			"SHA": func(c *gh.DanglingCommit) string {
				return c.SHA
			},
			"PR_NUMBER": func(c *gh.DanglingCommit) string {
				return fmt.Sprintf("%d", c.PRNumber)
			},
			"PR_URL": func(c *gh.DanglingCommit) string {
				return c.PRURL
			},
			"SIZE": func(c *gh.DanglingCommit) string {
				return humanize.Bytes(uint64(c.TotalBlobSize))
			},
			"MESSAGE": func(c *gh.DanglingCommit) string {
				return firstLineOf(c.Message)
			},
		},
	}
}

func (g *danglingCommitFieldGetters) getField(c *gh.DanglingCommit, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(c)
	}
	return ""
}

// RenderDanglingCommits renders a table of dangling commits with the specified headers.
// When an exporter is configured (e.g. --format json), the raw slice is exported instead.
func (r *Renderer) RenderDanglingCommits(commits []*gh.DanglingCommit, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(commits)
	}

	if len(commits) == 0 {
		r.writeLine("No dangling commits found.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"SHA", "PR_NUMBER", "PR_URL", "SIZE", "MESSAGE"}
	}

	getter := newDanglingCommitFieldGetters()
	table := r.newTableWriter(headers)
	for _, c := range commits {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.getField(c, header)
		}
		table.Append(row)
	}
	return table.Render()
}

type danglingBlobFieldGetter func(b *gh.DanglingBlob) string
type danglingBlobFieldGetters struct {
	Func map[string]danglingBlobFieldGetter
}

func newDanglingBlobFieldGetters() *danglingBlobFieldGetters {
	return &danglingBlobFieldGetters{
		Func: map[string]danglingBlobFieldGetter{
			"SHA": func(b *gh.DanglingBlob) string {
				return b.SHA
			},
			"PATH": func(b *gh.DanglingBlob) string {
				return b.Path
			},
			"SIZE": func(b *gh.DanglingBlob) string {
				return humanize.Bytes(uint64(b.Size))
			},
			"COMMIT_SHA": func(b *gh.DanglingBlob) string {
				return b.CommitSHA
			},
			"PR_NUMBER": func(b *gh.DanglingBlob) string {
				return fmt.Sprintf("%d", b.PRNumber)
			},
			"PR_URL": func(b *gh.DanglingBlob) string {
				return b.PRURL
			},
		},
	}
}

func (g *danglingBlobFieldGetters) getField(b *gh.DanglingBlob, field string) string {
	field = strings.ToUpper(field)
	if getter, ok := g.Func[field]; ok {
		return getter(b)
	}
	return ""
}

// RenderDanglingBlobs renders a table of dangling blobs with the specified headers.
// When an exporter is configured (e.g. --format json), the raw slice is exported instead.
func (r *Renderer) RenderDanglingBlobs(blobs []*gh.DanglingBlob, headers []string) error {
	if r.exporter != nil {
		return r.RenderExportedData(blobs)
	}

	if len(blobs) == 0 {
		r.writeLine("No dangling blobs found.")
		return nil
	}

	if len(headers) == 0 {
		headers = []string{"SHA", "PATH", "SIZE", "COMMIT_SHA", "PR_NUMBER", "PR_URL"}
	}

	getter := newDanglingBlobFieldGetters()
	table := r.newTableWriter(headers)
	for _, b := range blobs {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = getter.getField(b, header)
		}
		table.Append(row)
	}
	return table.Render()
}

// firstLineOf returns the first line of a potentially multi-line string.
func firstLineOf(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	return s
}
