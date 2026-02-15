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

// OverrideFormatFlagOptions replaces the Value of the "format" flag (created by cmdutil.AddFormatFlags)
// with a new enumValue that accepts additional options beyond "json".
// This must be called AFTER cmdutil.AddFormatFlags.
func OverrideFormatFlagOptions(cmd *cobra.Command, defaultValue string, options []string) {
	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		return
	}
	flag.Value = &formatEnumValue{value: defaultValue, options: options}
	flag.DefValue = defaultValue
	flag.Usage = fmt.Sprintf("Output format: {%s}", strings.Join(options, "|"))
	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return options, cobra.ShellCompDirectiveNoFileComp
	})
}

// SetupFormatFlagWithNonJSONFormats configures the format flag to accept additional non-JSON formats
// and sets up PreRunE to validate --jq and --template flags are only used with JSON format.
// The "json" format is automatically added to the options list.
// This must be called AFTER cmdutil.AddFormatFlags.
func SetupFormatFlagWithNonJSONFormats(cmd *cobra.Command, exportTarget *cmdutil.Exporter, exportFormat *string, defaultValue string, options []string) {
	// Always include "json" in the options
	allOptions := append([]string{"json"}, options...)

	// Override the format flag options
	OverrideFormatFlagOptions(cmd, defaultValue, allOptions)

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
}
