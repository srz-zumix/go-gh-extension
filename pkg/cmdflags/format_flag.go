package cmdflags

import (
	"fmt"
	"strings"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// formatEnumValue implements pflag.Value for enum-constrained string flags.
type formatEnumValue struct {
	value   string
	options []string
}

func (e *formatEnumValue) Set(value string) error {
	for _, opt := range e.options {
		if opt == value {
			e.value = value
			return nil
		}
	}
	return fmt.Errorf("valid values are {%s}", strings.Join(e.options, "|"))
}

func (e *formatEnumValue) String() string {
	return e.value
}

func (e *formatEnumValue) Type() string {
	return "string"
}

// OverrideFormatFlagOptions replaces the allowed values, default, and usage of the
// "format" flag that was previously registered by cmdutil.AddFormatFlags.
// Returns an error only when shell-completion registration fails for an unexpected reason.
// This must be called AFTER cmdutil.AddFormatFlags.
func OverrideFormatFlagOptions(cmd *cobra.Command, defaultValue string, options []string) error {
	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		return fmt.Errorf("OverrideFormatFlagOptions: \"format\" flag not found; call cmdutil.AddFormatFlags before this function")
	}
	flag.Value = &formatEnumValue{value: defaultValue, options: options}
	flag.DefValue = defaultValue
	flag.Usage = fmt.Sprintf("Output format: {%s}", strings.Join(options, "|"))
	// RegisterFlagCompletionFunc returns an error when the flag already has a registered
	// completion (cmdutil.AddFormatFlags → StringEnumFlag registers ["json"] first).
	// Only that specific case is handled via go:linkname override; any other error surfaces.
	compFunc := func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return options, cobra.ShellCompDirectiveNoFileComp
	}
	if err := cmd.RegisterFlagCompletionFunc("format", compFunc); err != nil {
		if !isAlreadyRegisteredError(err, "format") {
			return err
		}
		if err := overrideFlagCompletion(flag, compFunc); err != nil {
			return err
		}
	}
	return nil
}

// SetupFormatFlagWithNonJSONFormats configures the format flag to accept additional non-JSON formats
// and sets up PreRunE to validate --jq and --template flags are only used with JSON format.
// The "json" format is automatically added to the options list.
// This must be called AFTER cmdutil.AddFormatFlags.
func SetupFormatFlagWithNonJSONFormats(cmd *cobra.Command, exportTarget *cmdutil.Exporter, exportFormat *string, defaultValue string, options []string) error {
	// Always include "json" in the options
	allOptions := append([]string{"json"}, options...)

	// Override the format flag options
	if err := OverrideFormatFlagOptions(cmd, defaultValue, allOptions); err != nil {
		return err
	}

	// Wrap the existing PreRunE (set by AddFormatFlags)
	oldPreRun := cmd.PreRunE
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		// Run the original PreRunE from AddFormatFlags
		if oldPreRun != nil {
			if err := oldPreRun(c, args); err != nil {
				return err
			}
		}

		// Check if format is not the JSON format
		formatFlag := c.Flags().Lookup("format")
		if formatFlag == nil {
			return nil
		}

		format := formatFlag.Value.String()
		if format != "json" {
			jqFlag := c.Flags().Lookup("jq")
			if jqFlag != nil && jqFlag.Changed {
				return fmt.Errorf("cannot use `--jq` without specifying `--format json`")
			}
			templateFlag := c.Flags().Lookup("template")
			if templateFlag != nil && templateFlag.Changed {
				return fmt.Errorf("cannot use `--template` without specifying `--format json`")
			}
			// Clear the exporter and set the export format for non-JSON formats
			*exportTarget = nil
			*exportFormat = format
		}

		return nil
	}
	return nil
}

// AddFormatFlags is a helper that combines cmdutil.AddFormatFlags with SetupFormatFlagWithNonJSONFormats for convenience. See those functions for details.
func AddFormatFlags(cmd *cobra.Command, exporter *cmdutil.Exporter, exportFormat *string, defaultValue string, additionalFormats []string) error {
	cmdutil.AddFormatFlags(cmd, exporter)
	return SetupFormatFlagWithNonJSONFormats(cmd, exporter, exportFormat, defaultValue, additionalFormats)
}
