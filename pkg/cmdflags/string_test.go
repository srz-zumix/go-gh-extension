package cmdflags

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNonEmptyStringFlags(t *testing.T) {
	registrations := map[string]func(*cobra.Command, *string, string){
		"long": func(cmd *cobra.Command, value *string, def string) {
			NonEmptyStringVar(cmd, value, "prefix", def, "Prefix")
		},
		"shorthand": func(cmd *cobra.Command, value *string, def string) {
			NonEmptyStringVarP(cmd, value, "prefix", "p", def, "Prefix")
		},
	}
	for name, register := range registrations {
		t.Run(name, func(t *testing.T) {
			flag := "--prefix"
			if name == "shorthand" {
				flag = "-p"
			}
			cases := []struct {
				name    string
				def     string
				args    []string
				want    string
				wantErr bool
			}{
				{"default", "cordoned-", nil, "cordoned-", false},
				{"unset", "", nil, "", false},
				{"explicit", "cordoned-", []string{flag, "custom-"}, "custom-", false},
				{"empty", "cordoned-", []string{flag, ""}, "cordoned-", true},
				{"empty equals", "cordoned-", []string{"--prefix="}, "cordoned-", true},
				{"empty without default", "", []string{flag, ""}, "", true},
				{"repeated", "cordoned-", []string{flag, "first-", flag, "last-"}, "last-", false},
				{"invalid repeated", "cordoned-", []string{flag, "first-", flag, ""}, "first-", true},
				{"whitespace", "cordoned-", []string{flag, " "}, " ", false},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					cmd := &cobra.Command{Use: "test"}
					var value string
					register(cmd, &value, tc.def)
					if value != tc.def || cmd.Flags().Lookup("prefix").DefValue != tc.def {
						t.Fatalf("default value = %q, want %q", value, tc.def)
					}
					err := cmd.ParseFlags(tc.args)
					if (err != nil) != tc.wantErr {
						t.Fatalf("ParseFlags() error = %v, wantErr %v", err, tc.wantErr)
					}
					if err != nil && !strings.Contains(err.Error(), "empty string is not allowed") {
						t.Fatalf("unexpected error: %v", err)
					}
					got, err := cmd.Flags().GetString("prefix")
					if err != nil || got != tc.want || value != tc.want {
						t.Fatalf("GetString() = %q, %v; value = %q, want %q", got, err, value, tc.want)
					}
				})
			}
		})
	}
}

func TestNonEmptyStringFlagSet(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var value string
	NonEmptyStringVar(cmd, &value, "prefix", "cordoned-", "Prefix")
	if err := cmd.Flags().Set("prefix", "custom-"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("prefix", ""); err == nil {
		t.Fatal("Set() accepted an empty string")
	}
	if value != "custom-" {
		t.Fatalf("value = %q, want custom-", value)
	}
}
