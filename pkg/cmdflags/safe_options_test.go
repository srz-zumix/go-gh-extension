package cmdflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeOptionsValue_AcceptsNormalString(t *testing.T) {
	var p string
	v := &safeOptionsValue{value: &p}
	require.NoError(t, v.Set("--flag value"))
	assert.Equal(t, "--flag value", p)
}

func TestSafeOptionsValue_AcceptsEmpty(t *testing.T) {
	var p string
	v := &safeOptionsValue{value: &p}
	require.NoError(t, v.Set(""))
	assert.Equal(t, "", p)
}

func TestSafeOptionsValue_RejectsNewline(t *testing.T) {
	var p string
	v := &safeOptionsValue{value: &p}
	err := v.Set("--flag\ninjected")
	assert.Error(t, err)
}

func TestSafeOptionsValue_RejectsCarriageReturn(t *testing.T) {
	var p string
	v := &safeOptionsValue{value: &p}
	err := v.Set("--flag\rinjected")
	assert.Error(t, err)
}

func TestSafeOptionsValue_RejectsControlCharacter(t *testing.T) {
	var p string
	v := &safeOptionsValue{value: &p}
	err := v.Set("--flag\x01value")
	assert.Error(t, err)
}

func TestSafeOptionsValue_Type(t *testing.T) {
	var p string
	v := &safeOptionsValue{value: &p}
	assert.Equal(t, "string", v.Type())
}

func TestSafeOptionsVar_RegistersFlag(t *testing.T) {
	var p string
	cmd := &cobra.Command{Use: "test"}
	SafeOptionsVar(cmd, &p, "extra-opts", "", "usage")

	flag := cmd.Flags().Lookup("extra-opts")
	require.NotNil(t, flag)

	// Ensure the registered default metadata stays consistent with the value
	// string representation used by help text rendering.
	assert.Equal(t, flag.DefValue, flag.Value.String())
	require.NoError(t, cmd.Flags().Set("extra-opts", "--timeout 5m"))
	assert.Equal(t, "--timeout 5m", p)
}

func TestSafeOptionsVar_RejectsNewlineViaFlag(t *testing.T) {
	var p string
	cmd := &cobra.Command{Use: "test"}
	SafeOptionsVar(cmd, &p, "extra-opts", "", "usage")

	err := cmd.Flags().Set("extra-opts", "--flag\ninjected")
	assert.Error(t, err)
}
