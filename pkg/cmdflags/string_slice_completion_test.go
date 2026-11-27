package cmdflags

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testEnumOptions = []string{"alpha", "beta", "gamma"}

func completeStringSliceEnum(t *testing.T, define func(*cobra.Command, *[]string), args ...string) []string {
	t.Helper()
	root := &cobra.Command{Use: "root"}
	var values []string
	sub := &cobra.Command{Use: "sub", Run: func(*cobra.Command, []string) {}}
	define(sub, &values)
	root.AddCommand(sub)

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(append([]string{cobra.ShellCompRequestCmd, "sub"}, args...))
	require.NoError(t, root.Execute())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.NotEmpty(t, lines)
	directive := lines[len(lines)-1]
	assert.Equal(t, ":6", directive)
	var completions []string
	for _, line := range lines[:len(lines)-1] {
		if line != "" && !strings.HasPrefix(line, "Completion ended") {
			completions = append(completions, line)
		}
	}
	return completions
}

func TestStringSliceEnumCompletion(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "all options", args: []string{"--item", ""}, want: []string{"alpha", "beta", "gamma"}},
		{name: "prefix filter", args: []string{"--item", "b"}, want: []string{"beta"}},
		{name: "keeps comma prefix and excludes selected", args: []string{"--item", "alpha,"}, want: []string{"alpha,beta", "alpha,gamma"}},
		{name: "excludes repeated flag values", args: []string{"--item", "beta", "--item", ""}, want: []string{"alpha", "gamma"}},
		{name: "equals syntax", args: []string{"--item=alpha,g"}, want: []string{"alpha,gamma"}},
		{name: "no candidates", args: []string{"--item", "alpha,beta,gamma,"}, want: nil},
	}
	defines := map[string]func(*cobra.Command, *[]string){
		"RegisterStringSliceEnumCompletion": func(cmd *cobra.Command, p *[]string) {
			cmd.Flags().StringSliceVar(p, "item", nil, "items")
			require.NoError(t, RegisterStringSliceEnumCompletion(cmd, "item", testEnumOptions))
		},
		"StringSliceEnumVar": func(cmd *cobra.Command, p *[]string) {
			StringSliceEnumVar(cmd, p, "item", nil, testEnumOptions, "items")
		},
	}
	for defineName, define := range defines {
		for _, tt := range tests {
			t.Run(defineName+"/"+tt.name, func(t *testing.T) {
				assert.Equal(t, tt.want, completeStringSliceEnum(t, define, tt.args...))
			})
		}
	}
}

func TestStringSliceEnumVar_Parse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    []string
		wantErr string
	}{
		{name: "default", args: nil, want: []string{"beta"}},
		{name: "comma separated and repeated", args: []string{"--item", "alpha,gamma", "--item", "beta"}, want: []string{"alpha", "gamma", "beta"}},
		{name: "invalid value", args: []string{"--item", "alpha,delta"}, wantErr: `invalid value "delta"`},
		{name: "case sensitive", args: []string{"--item", "ALPHA"}, wantErr: `invalid value "ALPHA"`},
		{name: "empty", args: []string{"--item", ""}, wantErr: "empty string is not allowed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "sub"}
			var values []string
			StringSliceEnumVarP(cmd, &values, "item", "i", []string{"beta"}, testEnumOptions, "items")
			err := cmd.ParseFlags(tt.args)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, values)
		})
	}
}

func TestStringSliceEnumVar_InvalidDefaultPanics(t *testing.T) {
	cmd := &cobra.Command{Use: "sub"}
	var values []string
	assert.Panics(t, func() {
		StringSliceEnumVar(cmd, &values, "item", []string{"delta"}, testEnumOptions, "items")
	})
}

func TestRegisterStringSliceEnumCompletion_UnknownFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "sub"}
	assert.Error(t, RegisterStringSliceEnumCompletion(cmd, "missing", testEnumOptions))
}
