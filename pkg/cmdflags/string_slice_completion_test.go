package cmdflags

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func completeStringSliceEnum(t *testing.T, args ...string) []string {
	t.Helper()
	root := &cobra.Command{Use: "root"}
	var values []string
	sub := &cobra.Command{Use: "sub", Run: func(*cobra.Command, []string) {}}
	sub.Flags().StringSliceVar(&values, "item", nil, "items")
	require.NoError(t, RegisterStringSliceEnumCompletion(sub, "item", []string{"alpha", "beta", "gamma"}))
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, completeStringSliceEnum(t, tt.args...))
		})
	}
}

func TestRegisterStringSliceEnumCompletion_UnknownFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "sub"}
	assert.Error(t, RegisterStringSliceEnumCompletion(cmd, "missing", []string{"a"}))
}
