package cmdflags

import (
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func TestOverrideFormatFlagOptions(t *testing.T) {
	t.Run("override format flag after AddFormatFlags", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
		}
		var exporter cmdutil.Exporter

		// AddFormatFlags creates a format flag with only "json" option
		cmdutil.AddFormatFlags(cmd, &exporter)

		// Override to support both json and mermaid
		OverrideFormatFlagOptions(cmd, "mermaid", []string{"json", "mermaid"})

		flag := cmd.Flags().Lookup("format")
		if flag == nil {
			t.Fatal("format flag not found")
		}

		// Check default value is "mermaid"
		if flag.DefValue != "mermaid" {
			t.Errorf("DefValue = %v, want %v", flag.DefValue, "mermaid")
		}

		// Check usage contains both options
		expectedUsage := "Output format: {json|mermaid}"
		if flag.Usage != expectedUsage {
			t.Errorf("Usage = %v, want %v", flag.Usage, expectedUsage)
		}

		// Check the value is set to default
		if flag.Value.String() != "mermaid" {
			t.Errorf("Value.String() = %v, want %v", flag.Value.String(), "mermaid")
		}
	})

	t.Run("format flag not exists", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
		}

		// Should not panic when format flag doesn't exist
		OverrideFormatFlagOptions(cmd, "mermaid", []string{"json", "mermaid"})

		flag := cmd.Flags().Lookup("format")
		if flag != nil {
			t.Errorf("format flag should not exist, got %v", flag)
		}
	})
}

func TestFormatEnumValue_Set(t *testing.T) {
	tests := []struct {
		name      string
		options   []string
		setValue  string
		wantError bool
	}{
		{
			name:      "valid json",
			options:   []string{"json", "mermaid"},
			setValue:  "json",
			wantError: false,
		},
		{
			name:      "valid mermaid",
			options:   []string{"json", "mermaid"},
			setValue:  "mermaid",
			wantError: false,
		},
		{
			name:      "invalid value",
			options:   []string{"json", "mermaid"},
			setValue:  "invalid",
			wantError: true,
		},
		{
			name:      "empty string",
			options:   []string{"json", "mermaid"},
			setValue:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &formatEnumValue{
				value:   "",
				options: tt.options,
			}

			err := v.Set(tt.setValue)
			if tt.wantError {
				if err == nil {
					t.Errorf("Set(%v) expected error, got nil", tt.setValue)
				}
			} else {
				if err != nil {
					t.Errorf("Set(%v) unexpected error: %v", tt.setValue, err)
				}
				if v.String() != tt.setValue {
					t.Errorf("String() = %v, want %v", v.String(), tt.setValue)
				}
			}
		})
	}
}

func TestFormatEnumValue_String(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "json",
			value:    "json",
			expected: "json",
		},
		{
			name:     "mermaid",
			value:    "mermaid",
			expected: "mermaid",
		},
		{
			name:     "empty",
			value:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &formatEnumValue{
				value:   tt.value,
				options: []string{"json", "mermaid"},
			}
			if got := v.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatEnumValue_Type(t *testing.T) {
	v := &formatEnumValue{
		value:   "json",
		options: []string{"json", "mermaid"},
	}
	if got := v.Type(); got != "string" {
		t.Errorf("Type() = %v, want %v", got, "string")
	}
}

func TestSetupFormatFlagWithNonJSONFormats(t *testing.T) {
	t.Run("setup with mermaid default", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		var exporter cmdutil.Exporter
		var format string

		// AddFormatFlags first
		cmdutil.AddFormatFlags(cmd, &exporter)
		// Then setup with non-JSON formats
		SetupFormatFlagWithNonJSONFormats(cmd, &exporter, &format, "mermaid", []string{"mermaid"})

		flag := cmd.Flags().Lookup("format")
		if flag == nil {
			t.Fatal("format flag not found")
		}

		// Check default value is "mermaid"
		if flag.DefValue != "mermaid" {
			t.Errorf("DefValue = %v, want %v", flag.DefValue, "mermaid")
		}

		// Check usage contains both options
		expectedUsage := "Output format: {json|mermaid}"
		if flag.Usage != expectedUsage {
			t.Errorf("Usage = %v, want %v", flag.Usage, expectedUsage)
		}
	})

	t.Run("exporter cleared for non-json format", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		var exporter cmdutil.Exporter
		var format string

		cmdutil.AddFormatFlags(cmd, &exporter)
		SetupFormatFlagWithNonJSONFormats(cmd, &exporter, &format, "mermaid", []string{"mermaid"})

		// Execute PreRunE with format=mermaid (default)
		if err := cmd.PreRunE(cmd, []string{}); err != nil {
			t.Fatalf("PreRunE failed: %v", err)
		}

		// Exporter should be nil for mermaid format
		if exporter != nil {
			t.Errorf("exporter should be nil for mermaid format, got %v", exporter)
		}
	})

	t.Run("jq flag error with non-json format", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		var exporter cmdutil.Exporter
		var format string

		cmdutil.AddFormatFlags(cmd, &exporter)
		SetupFormatFlagWithNonJSONFormats(cmd, &exporter, &format, "mermaid", []string{"mermaid"})

		// Set --jq flag
		if err := cmd.Flags().Set("jq", ".test"); err != nil {
			t.Fatalf("failed to set jq flag: %v", err)
		}

		// Execute PreRunE should fail
		err := cmd.PreRunE(cmd, []string{})
		if err == nil {
			t.Fatal("PreRunE should fail when --jq is used with non-JSON format")
		}
		expectedError := "cannot use `--jq` without specifying `--format json`"
		if err.Error() != expectedError {
			t.Errorf("error = %v, want %v", err.Error(), expectedError)
		}
	})

	t.Run("template flag error with non-json format", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		var exporter cmdutil.Exporter
		var format string

		cmdutil.AddFormatFlags(cmd, &exporter)
		SetupFormatFlagWithNonJSONFormats(cmd, &exporter, &format, "mermaid", []string{"mermaid"})

		// Set --template flag
		if err := cmd.Flags().Set("template", "{{.}}"); err != nil {
			t.Fatalf("failed to set template flag: %v", err)
		}

		// Execute PreRunE should fail
		err := cmd.PreRunE(cmd, []string{})
		if err == nil {
			t.Fatal("PreRunE should fail when --template is used with non-JSON format")
		}
		expectedError := "cannot use `--template` without specifying `--format json`"
		if err.Error() != expectedError {
			t.Errorf("error = %v, want %v", err.Error(), expectedError)
		}
	})

	t.Run("json format with jq allowed", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		var exporter cmdutil.Exporter
		var format string

		cmdutil.AddFormatFlags(cmd, &exporter)
		SetupFormatFlagWithNonJSONFormats(cmd, &exporter, &format, "mermaid", []string{"mermaid"})

		// Set format to json and jq
		if err := cmd.Flags().Set("format", "json"); err != nil {
			t.Fatalf("failed to set format flag: %v", err)
		}
		if err := cmd.Flags().Set("jq", ".test"); err != nil {
			t.Fatalf("failed to set jq flag: %v", err)
		}

		// Execute PreRunE should succeed
		if err := cmd.PreRunE(cmd, []string{}); err != nil {
			t.Errorf("PreRunE should not fail when --jq is used with --format json: %v", err)
		}

		// Exporter should not be nil for json format
		if exporter == nil {
			t.Error("exporter should not be nil for json format")
		}
	})
}
