package cmdflags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// nonEmptyStringArrayValue implements pflag.Value for a []string flag that rejects empty string entries.
type nonEmptyStringArrayValue struct {
	value   *[]string
	changed bool
}

func newNonEmptyStringArrayValue(p *[]string) *nonEmptyStringArrayValue {
	*p = nil
	return &nonEmptyStringArrayValue{value: p}
}

func (s *nonEmptyStringArrayValue) Set(val string) error {
	if val == "" {
		return fmt.Errorf("empty string is not allowed")
	}
	if !s.changed {
		*s.value = []string{val}
		s.changed = true
	} else {
		*s.value = append(*s.value, val)
	}
	return nil
}

func (s *nonEmptyStringArrayValue) Type() string {
	return "stringArray"
}

func (s *nonEmptyStringArrayValue) String() string {
	out := make([]string, len(*s.value))
	for i, v := range *s.value {
		out[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(out, ",") + "]"
}

// Ensure nonEmptyStringArrayValue implements pflag.Value.
var _ pflag.Value = (*nonEmptyStringArrayValue)(nil)

// NonEmptyStringArrayVar defines a string array flag that rejects empty string entries.
// Multiple values can be supplied by repeating the flag (e.g. --flag a --flag b).
func NonEmptyStringArrayVar(cmd *cobra.Command, p *[]string, name string, usage string) {
	cmd.Flags().VarP(newNonEmptyStringArrayValue(p), name, "", usage)
}

// NonEmptyStringArrayVarP is like NonEmptyStringArrayVar but accepts a shorthand letter.
func NonEmptyStringArrayVarP(cmd *cobra.Command, p *[]string, name, shorthand string, usage string) {
	cmd.Flags().VarP(newNonEmptyStringArrayValue(p), name, shorthand, usage)
}
