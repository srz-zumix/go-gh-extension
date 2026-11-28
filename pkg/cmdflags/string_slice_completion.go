package cmdflags

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// StringSliceEnumCompletion returns a completion function for a comma-separated
// string slice flag whose values are restricted to options.
// It completes the last comma-separated element while keeping the preceding
// elements as a prefix, and excludes values that were already selected either
// earlier in the same argument or by previous occurrences of the flag.
// The flag value must implement pflag.SliceValue (e.g. StringSliceVar) for
// previous occurrences to be excluded.
func StringSliceEnumCompletion(name string, options []string) cobra.CompletionFunc {
	return func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		prefix := ""
		if idx := strings.LastIndexByte(toComplete, ','); idx >= 0 {
			prefix, toComplete = toComplete[:idx+1], toComplete[idx+1:]
		}

		selected := map[string]bool{}
		if flag := cmd.Flags().Lookup(name); flag != nil && flag.Changed {
			if sv, ok := flag.Value.(pflag.SliceValue); ok {
				for _, v := range sv.GetSlice() {
					selected[v] = true
				}
			}
		}
		for _, v := range strings.Split(prefix, ",") {
			selected[v] = true
		}

		var completions []string
		for _, option := range options {
			if !selected[option] && strings.HasPrefix(option, toComplete) {
				completions = append(completions, prefix+option)
			}
		}
		return completions, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	}
}

// stringSliceEnumValue implements pflag.SliceValue for a comma-separated
// string slice flag whose values are restricted to options.
type stringSliceEnumValue struct {
	value   *[]string
	options []string
	changed bool
}

func (s *stringSliceEnumValue) validate(vals []string) error {
	for _, v := range vals {
		if !slices.Contains(s.options, v) {
			return fmt.Errorf("invalid value %q; valid values: %s", v, strings.Join(s.options, ", "))
		}
	}
	return nil
}

func (s *stringSliceEnumValue) Set(val string) error {
	parts, err := readCSVValues(val)
	if err != nil {
		return err
	}
	if err := s.validate(parts); err != nil {
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

func (s *stringSliceEnumValue) Type() string {
	return "stringSlice"
}

func (s *stringSliceEnumValue) String() string {
	// Use the same bracketed CSV format as pflag's stringSlice so that
	// FlagSet.GetStringSlice can round-trip the value.
	return "[" + writeCSVValues(*s.value) + "]"
}

// writeCSVValues encodes vals as a single CSV record without a trailing newline.
func writeCSVValues(vals []string) string {
	if len(vals) == 0 {
		return ""
	}
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	_ = w.Write(vals)
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n")
}

func (s *stringSliceEnumValue) Append(val string) error {
	if err := s.validate([]string{val}); err != nil {
		return err
	}
	*s.value = append(*s.value, val)
	return nil
}

func (s *stringSliceEnumValue) Replace(vals []string) error {
	if err := s.validate(vals); err != nil {
		return err
	}
	*s.value = append([]string(nil), vals...)
	return nil
}

func (s *stringSliceEnumValue) GetSlice() []string {
	return append([]string(nil), *s.value...)
}

// Ensure stringSliceEnumValue implements pflag.SliceValue.
var _ pflag.SliceValue = (*stringSliceEnumValue)(nil)

// StringSliceEnumVar defines a comma-separated string slice flag whose values are
// restricted to options (case-sensitive) and registers StringSliceEnumCompletion for it.
// Multiple values can be supplied by comma-separated list (e.g. --flag a,b) or by repeating the flag.
// It panics if defaultValues contains a value not included in options.
func StringSliceEnumVar(cmd *cobra.Command, p *[]string, name string, defaultValues, options []string, usage string) *pflag.Flag {
	return StringSliceEnumVarP(cmd, p, name, "", defaultValues, options, usage)
}

// StringSliceEnumVarP is like StringSliceEnumVar but accepts a shorthand letter.
func StringSliceEnumVarP(cmd *cobra.Command, p *[]string, name, shorthand string, defaultValues, options []string, usage string) *pflag.Flag {
	val := &stringSliceEnumValue{value: p, options: append([]string(nil), options...)}
	if err := val.validate(defaultValues); err != nil {
		panic(fmt.Sprintf("invalid default value for enum string slice flag %q: %v", name, err))
	}
	*p = append([]string(nil), defaultValues...)
	f := cmd.Flags().VarPF(val, name, shorthand, usage)
	if len(defaultValues) == 0 {
		// Let pflag treat an empty default as zero and omit it from help.
		f.DefValue = ""
	}
	if err := RegisterStringSliceEnumCompletion(cmd, name, val.options); err != nil {
		panic(err)
	}
	return f
}

// RegisterStringSliceEnumCompletion registers StringSliceEnumCompletion for the named flag.
func RegisterStringSliceEnumCompletion(cmd *cobra.Command, name string, options []string) error {
	if err := cmd.RegisterFlagCompletionFunc(name, StringSliceEnumCompletion(name, options)); err != nil {
		return fmt.Errorf("failed to register completion for flag %q: %w", name, err)
	}
	return nil
}
