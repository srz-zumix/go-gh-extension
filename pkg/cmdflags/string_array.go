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

// validateNonEmptyStringArrayValues validates that all entries are non-empty.
func validateNonEmptyStringArrayValues(val []string) error {
	for _, v := range val {
		if v == "" {
			return fmt.Errorf("empty string is not allowed")
		}
	}
	return nil
}

func newNonEmptyStringArrayValue(val []string, p *[]string) *nonEmptyStringArrayValue {
	if err := validateNonEmptyStringArrayValues(val); err != nil {
		panic(fmt.Sprintf("invalid default value for non-empty string array flag: %v", err))
	}

	s := new(nonEmptyStringArrayValue)
	s.value = p
	*s.value = append([]string(nil), val...)
	return s
}

func (s *nonEmptyStringArrayValue) Set(val string) error {
	if err := validateNonEmptyStringArrayValues([]string{val}); err != nil {
		return err
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
func NonEmptyStringArrayVar(cmd *cobra.Command, p *[]string, name string, value []string, usage string) {
	cmd.Flags().VarP(newNonEmptyStringArrayValue(value, p), name, "", usage)
}

// NonEmptyStringArrayVarP is like NonEmptyStringArrayVar but accepts a shorthand letter.
func NonEmptyStringArrayVarP(cmd *cobra.Command, p *[]string, name, shorthand string, value []string, usage string) {
	cmd.Flags().VarP(newNonEmptyStringArrayValue(value, p), name, shorthand, usage)
}
