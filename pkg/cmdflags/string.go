package cmdflags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// nonEmptyStringValue implements pflag.Value for a string flag that rejects an empty value.
type nonEmptyStringValue struct {
	value *string
}

func newNonEmptyStringValue(val string, p *string) *nonEmptyStringValue {
	*p = val
	return &nonEmptyStringValue{value: p}
}

func (s *nonEmptyStringValue) Set(val string) error {
	if err := validateNonEmptyStringValues([]string{val}); err != nil {
		return err
	}
	*s.value = val
	return nil
}

func (s *nonEmptyStringValue) Type() string {
	return "string"
}

func (s *nonEmptyStringValue) String() string {
	return *s.value
}

// Ensure nonEmptyStringValue implements pflag.Value.
var _ pflag.Value = (*nonEmptyStringValue)(nil)

// NonEmptyStringVar defines a string flag that rejects an empty value.
// An empty default value is allowed and means the flag is unset.
func NonEmptyStringVar(cmd *cobra.Command, p *string, name string, value string, usage string) {
	cmd.Flags().VarP(newNonEmptyStringValue(value, p), name, "", usage)
}

// NonEmptyStringVarP is like NonEmptyStringVar but accepts a shorthand letter.
func NonEmptyStringVarP(cmd *cobra.Command, p *string, name, shorthand string, value string, usage string) {
	cmd.Flags().VarP(newNonEmptyStringValue(value, p), name, shorthand, usage)
}
