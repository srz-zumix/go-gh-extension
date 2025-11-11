package render

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/google/go-github/v73/github"
	"github.com/olekukonko/tablewriter"
	"github.com/shurcooL/githubv4"
)

type Renderer struct {
	IO       *iostreams.IOStreams
	Color    bool
	exporter cmdutil.Exporter
}

var defaultTimeFormat = "2006-01-02 15:04:05"

var TimeFormat = defaultTimeFormat

type StringRenderer struct {
	Renderer Renderer
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
}

const (
	ColorFlagAlways string = "always"
	ColorFlagNever  string = "never"
	ColorFlagAuto   string = "auto"
)

var ColorFlags = []string{
	ColorFlagAlways,
	ColorFlagNever,
	ColorFlagAuto,
}

func NewRenderer(ex cmdutil.Exporter) *Renderer {
	return &Renderer{
		IO:       iostreams.System(),
		exporter: ex,
	}
}

func NewStringRenderer(ex cmdutil.Exporter) *StringRenderer {
	io, _, out, errOut := iostreams.Test()
	return &StringRenderer{
		Renderer: Renderer{
			IO:       io,
			exporter: ex,
		},
		Stdout: out,
		Stderr: errOut,
	}
}

func (r *Renderer) SetColor(colorFlag string) {
	switch colorFlag {
	case ColorFlagAlways:
		r.Color = true
	case ColorFlagNever:
		r.Color = false
	default:
		r.Color = r.IO.ColorEnabled()
	}
}

func (r *Renderer) WriteLine(line string) {
	if r.exporter != nil {
		return
	}
	r.writeLine(line)
}

func (r *Renderer) writeLine(line string) {
	_, err := fmt.Fprintln(r.IO.Out, line)
	if err != nil {
		r.WriteError(err)
	}
}

func (r *Renderer) WriteError(err error) {
	fmt.Fprintf(r.IO.ErrOut, "%v\n", err) // nolint
}

func ToString(v any) string {
	if str, ok := v.(*string); ok {
		if str == nil {
			return ""
		}
		return *str
	} else if b, ok := v.(*bool); ok {
		if b == nil {
			return ""
		}
		return toString(*b)
	} else if i, ok := v.(*int); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*int64); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*int32); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*int8); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*int16); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*uint); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*uint64); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*uint32); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*uint8); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*uint16); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*float64); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if i, ok := v.(*float32); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if e, ok := v.(*error); ok {
		if e == nil {
			return ""
		}
		return toString(*e)
	} else if t, ok := v.(*time.Time); ok {
		if t == nil {
			return ""
		}
		return toString(*t)
	} else if t, ok := v.(*github.Timestamp); ok {
		if t == nil {
			return ""
		}
		return toString(*t)
	} else if p, ok := v.(*github.EmptyRuleParameters); ok {
		if p == nil {
			return "DISABLED"
		}
		return "ENABLED"
	} else if str, ok := v.(*githubv4.String); ok {
		if str == nil {
			return ""
		}
		return toString(*str)
	} else if s, ok := v.(*githubv4.Base64String); ok {
		if s == nil {
			return ""
		}
		return toString(*s)
	} else if b, ok := v.(*githubv4.Boolean); ok {
		if b == nil {
			return ""
		}
		return toString(*b)
	} else if id, ok := v.(*githubv4.ID); ok {
		if id == nil {
			return ""
		}
		return toString(*id)
	} else if i, ok := v.(*githubv4.Int); ok {
		if i == nil {
			return ""
		}
		return toString(*i)
	} else if f, ok := v.(*githubv4.Float); ok {
		if f == nil {
			return ""
		}
		return toString(*f)
	} else if uri, ok := v.(*githubv4.URI); ok {
		if uri == nil {
			return ""
		}
		return toString(*uri)
	} else if t, ok := v.(*githubv4.DateTime); ok {
		if t == nil {
			return ""
		}
		return toString(*t)
	} else if a, ok := v.(*any); ok {
		if a == nil {
			return ""
		}
		return toString(*a)
	}
	return toString(v)
}

func toString(v any) string {
	if str, ok := v.(string); ok {
		return str
	} else if b, ok := v.(bool); ok {
		if b {
			return "YES"
		}
		return "NO"
	} else if i, ok := v.(int); ok {
		return fmt.Sprintf("%d", i)
	} else if i, ok := v.(int64); ok {
		return fmt.Sprintf("%d", i)
	} else if i, ok := v.(int32); ok {
		return fmt.Sprintf("%d", i)
	} else if i, ok := v.(int8); ok {
		return fmt.Sprintf("%d", i)
	} else if i, ok := v.(int16); ok {
		return fmt.Sprintf("%d", i)
	} else if i, ok := v.(uint); ok {
		return fmt.Sprintf("%d", i)
	} else if i, ok := v.(uint64); ok {
		return fmt.Sprintf("%d", i)
	} else if i, ok := v.(uint32); ok {
		return fmt.Sprintf("%d", i)
	} else if i, ok := v.(uint8); ok {
		return fmt.Sprintf("%d", i)
	} else if i, ok := v.(uint16); ok {
		return fmt.Sprintf("%d", i)
	} else if f, ok := v.(float64); ok {
		return fmt.Sprintf("%f", f)
	} else if f, ok := v.(float32); ok {
		return fmt.Sprintf("%f", f)
	} else if i, ok := v.(error); ok {
		return i.Error()
	} else if t, ok := v.(time.Time); ok {
		return t.Format(TimeFormat)
	} else if t, ok := v.(github.Timestamp); ok {
		return t.Format(TimeFormat)
	} else if str, ok := v.(githubv4.String); ok {
		return string(str)
	} else if s, ok := v.(githubv4.Base64String); ok {
		return string(s)
	} else if b, ok := v.(githubv4.Boolean); ok {
		if b {
			return "YES"
		}
		return "NO"
	} else if id, ok := v.(githubv4.ID); ok {
		return fmt.Sprintf("%v", id)
	} else if i, ok := v.(githubv4.Int); ok {
		return fmt.Sprintf("%d", i)
	} else if f, ok := v.(githubv4.Float); ok {
		return fmt.Sprintf("%f", f)
	} else if uri, ok := v.(githubv4.URI); ok {
		return uri.String()
	} else if t, ok := v.(githubv4.DateTime); ok {
		return t.Format(TimeFormat)
	}
	return ""
}

func ToRGB(c string) (int, int, int, error) {
	if len(c) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid color code: %s", c)
	}
	r, err := strconv.ParseInt(c[:2], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	g, err := strconv.ParseInt(c[2:4], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	b, err := strconv.ParseInt(c[4:6], 16, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	return int(r), int(g), int(b), nil
}

func (r *Renderer) newTableWriter(header []string) *tablewriter.Table {
	table := tablewriter.NewWriter(r.IO.Out)
	table.SetHeader(header)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	return table
}
