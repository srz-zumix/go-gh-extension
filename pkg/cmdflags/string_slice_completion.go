package cmdflags

import (
	"fmt"
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

// RegisterStringSliceEnumCompletion registers StringSliceEnumCompletion for the named flag.
func RegisterStringSliceEnumCompletion(cmd *cobra.Command, name string, options []string) error {
	if err := cmd.RegisterFlagCompletionFunc(name, StringSliceEnumCompletion(name, options)); err != nil {
		return fmt.Errorf("failed to register completion for flag %q: %w", name, err)
	}
	return nil
}
