package cmdflags

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// validateNonEmptyStringValues validates that all entries are non-empty.
func validateNonEmptyStringValues(val []string) error {
	for _, v := range val {
		if v == "" {
			return fmt.Errorf("empty string is not allowed")
		}
	}
	return nil
}

// quotedSliceString returns a bracketed, comma-separated string of double-quoted values.
func quotedSliceString(vals []string) string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(out, ",") + "]"
}

// initNonEmptyStringSlice validates val and copies it into *p, panicking with typeName on invalid defaults.
func initNonEmptyStringSlice(val []string, p *[]string, typeName string) {
	if err := validateNonEmptyStringValues(val); err != nil {
		panic(fmt.Sprintf("invalid default value for non-empty %s flag: %v", typeName, err))
	}
	*p = append([]string(nil), val...)
}

// --- nonEmptyStringArrayValue ---

// nonEmptyStringArrayValue implements pflag.Value for a []string flag that rejects empty string entries.
type nonEmptyStringArrayValue struct {
	value   *[]string
	changed bool
}

func newNonEmptyStringArrayValue(val []string, p *[]string) *nonEmptyStringArrayValue {
	s := new(nonEmptyStringArrayValue)
	s.value = p
	initNonEmptyStringSlice(val, p, "string array")
	return s
}

func (s *nonEmptyStringArrayValue) Set(val string) error {
	if err := validateNonEmptyStringValues([]string{val}); err != nil {
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
	return quotedSliceString(*s.value)
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

// --- nonEmptyStringSliceValue ---

// nonEmptyStringSliceValue implements pflag.Value for a []string flag
// that accepts comma-separated values and rejects empty string entries.
type nonEmptyStringSliceValue struct {
	value   *[]string
	changed bool
}

func newNonEmptyStringSliceValue(val []string, p *[]string) *nonEmptyStringSliceValue {
	s := new(nonEmptyStringSliceValue)
	s.value = p
	initNonEmptyStringSlice(val, p, "string slice")
	return s
}

func (s *nonEmptyStringSliceValue) Set(val string) error {
	r := csv.NewReader(strings.NewReader(val))
	r.TrimLeadingSpace = true
	parts, err := r.Read()
	if err != nil {
		return fmt.Errorf("failed to parse value %q: %w", val, err)
	}

	if err := validateNonEmptyStringValues(parts); err != nil {
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
	return quotedSliceString(*s.value)
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
