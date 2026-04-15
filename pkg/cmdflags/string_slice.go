package cmdflags

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// nonEmptyStringSliceValue implements pflag.Value for a []string flag
// that accepts comma-separated values and rejects empty string entries.
type nonEmptyStringSliceValue struct {
	value   *[]string
	changed bool
}

func newNonEmptyStringSliceValue(val []string, p *[]string) *nonEmptyStringSliceValue {
	if err := validateNonEmptyStringArrayValues(val); err != nil {
		panic(fmt.Sprintf("invalid default value for non-empty string slice flag: %v", err))
	}

	s := new(nonEmptyStringSliceValue)
	s.value = p
	*s.value = append([]string(nil), val...)
	return s
}

func (s *nonEmptyStringSliceValue) Set(val string) error {
	r := csv.NewReader(strings.NewReader(val))
	parts, err := r.Read()
	if err != nil {
		return fmt.Errorf("failed to parse value %q: %w", val, err)
	}

	if err := validateNonEmptyStringArrayValues(parts); err != nil {
		return err
	}

	if !s.changed {
		*s.value = parts
		s.changed = true
	} else {
		*s.value = append(*s.value, parts...)
	}
	return nil
}

func (s *nonEmptyStringSliceValue) Type() string {
	return "stringSlice"
}

func (s *nonEmptyStringSliceValue) String() string {
	out := make([]string, len(*s.value))
	for i, v := range *s.value {
		out[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(out, ",") + "]"
}

// Ensure nonEmptyStringSliceValue implements pflag.Value.
var _ pflag.Value = (*nonEmptyStringSliceValue)(nil)

// NonEmptyStringSliceVar defines a string slice flag that accepts comma-separated values
// and rejects empty string entries.
// Multiple values can be supplied by comma-separated list (e.g. --flag a,b) or by repeating the flag.
func NonEmptyStringSliceVar(cmd *cobra.Command, p *[]string, name string, value []string, usage string) {
	cmd.Flags().VarP(newNonEmptyStringSliceValue(value, p), name, "", usage)
}

// NonEmptyStringSliceVarP is like NonEmptyStringSliceVar but accepts a shorthand letter.
func NonEmptyStringSliceVarP(cmd *cobra.Command, p *[]string, name, shorthand string, value []string, usage string) {
	cmd.Flags().VarP(newNonEmptyStringSliceValue(value, p), name, shorthand, usage)
}
