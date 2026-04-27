package cmdflags

import (
	"fmt"

	"github.com/spf13/cobra"
)

// safeOptionsValue implements pflag.Value for a free-form options string that
// must not contain newlines or control characters (codepoints < U+0020).
// This prevents a pasted or shell-expanded value from injecting extra lines
// into a generated shell script.
type safeOptionsValue struct {
	value *string
}

func (v *safeOptionsValue) Set(s string) error {
	for i, r := range s {
		if r < 0x20 {
			return fmt.Errorf("character at byte offset %d (U+%04X) is not allowed; value must not contain newlines or control characters", i, r)
		}
	}
	*v.value = s
	return nil
}

func (v *safeOptionsValue) String() string {
	if v.value == nil {
		return ""
	}
	return *v.value
}

func (v *safeOptionsValue) Type() string {
	return "string"
}

// SafeOptionsVar registers a string flag on cmd whose value is rejected when it
// contains newlines or other ASCII control characters (U+0000–U+001F).
// Quote the value when it contains spaces, e.g. --flag '--opt value'.
func SafeOptionsVar(cmd *cobra.Command, p *string, name, value, usage string) {
	v := &safeOptionsValue{value: p}
	if err := v.Set(value); err != nil {
		panic(fmt.Errorf("invalid default value for flag %q: %w", name, err))
	}
	cmd.Flags().Var(v, name, usage)
}
