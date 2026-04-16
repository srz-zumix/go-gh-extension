package cmdflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- nonEmptyStringArrayValue unit tests ---

func TestNonEmptyStringArrayValue_EmptyStringReturnsError(t *testing.T) {
	var p []string
	v := newNonEmptyStringArrayValue(nil, &p)
	err := v.Set("")
	assert.Error(t, err, "Set(\"\") must return an error")
}

func TestNonEmptyStringArrayValue_FirstSetOverridesDefault(t *testing.T) {
	var p []string
	v := newNonEmptyStringArrayValue([]string{"default"}, &p)

	// Default must be visible before any Set call.
	assert.Equal(t, []string{"default"}, p)

	require.NoError(t, v.Set("first"))
	// After the first Set the default is replaced, not appended.
	assert.Equal(t, []string{"first"}, p)
}

func TestNonEmptyStringArrayValue_SubsequentSetAppends(t *testing.T) {
	var p []string
	v := newNonEmptyStringArrayValue(nil, &p)

	require.NoError(t, v.Set("a"))
	require.NoError(t, v.Set("b"))
	require.NoError(t, v.Set("c"))
	assert.Equal(t, []string{"a", "b", "c"}, p)
}

func TestNonEmptyStringArrayValue_Type(t *testing.T) {
	var p []string
	v := newNonEmptyStringArrayValue(nil, &p)
	assert.Equal(t, "stringArray", v.Type())
}

func TestNonEmptyStringArrayValue_StringEmpty(t *testing.T) {
	var p []string
	v := newNonEmptyStringArrayValue(nil, &p)
	assert.Equal(t, "[]", v.String())
}

func TestNonEmptyStringArrayValue_StringWithValues(t *testing.T) {
	var p []string
	v := newNonEmptyStringArrayValue([]string{"foo", "bar baz"}, &p)
	// Each element must be double-quoted; elements separated by commas.
	assert.Equal(t, `["foo","bar baz"]`, v.String())
}

// --- NonEmptyStringArrayVar / NonEmptyStringArrayVarP integration tests ---

func TestNonEmptyStringArrayVar_RejectsEmptyString(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringArrayVar(cmd, &vals, "items", nil, "test flag")

	err := cmd.Flags().Set("items", "")
	assert.Error(t, err)
}

func TestNonEmptyStringArrayVar_DefaultAndParse(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringArrayVar(cmd, &vals, "items", []string{"default"}, "test flag")

	// Before parsing, default value must be reflected.
	assert.Equal(t, []string{"default"}, vals)

	// After parsing with explicit values, default must be replaced then appended.
	require.NoError(t, cmd.ParseFlags([]string{"--items", "x", "--items", "y"}))
	assert.Equal(t, []string{"x", "y"}, vals)
}

func TestNonEmptyStringArrayVarP_ShorthandWorks(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringArrayVarP(cmd, &vals, "items", "i", nil, "test flag")

	require.NoError(t, cmd.ParseFlags([]string{"-i", "alpha", "-i", "beta"}))
	assert.Equal(t, []string{"alpha", "beta"}, vals)
}

func TestNonEmptyStringArrayVarP_RejectsEmptyString(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringArrayVarP(cmd, &vals, "items", "i", nil, "test flag")

	err := cmd.Flags().Set("items", "")
	assert.Error(t, err)
}

func TestNonEmptyStringArrayVar_DefValueMatchesStringOutput(t *testing.T) {
	// The flag's DefValue (used in help text) must match String() on the default value.
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringArrayVar(cmd, &vals, "items", []string{"a", "b"}, "test flag")

	flag := cmd.Flags().Lookup("items")
	require.NotNil(t, flag)
	assert.Equal(t, flag.Value.String(), flag.DefValue)
}

// --- nonEmptyStringSliceValue unit tests ---

func TestNonEmptyStringSliceValue_EmptyStringReturnsError(t *testing.T) {
	var p []string
	v := newNonEmptyStringSliceValue(nil, &p)
	err := v.Set("")
	assert.Error(t, err, "Set(\"\") must return an error")
}

func TestNonEmptyStringSliceValue_EmptyEntryInCommaListReturnsError(t *testing.T) {
	var p []string
	v := newNonEmptyStringSliceValue(nil, &p)
	err := v.Set("a,,b")
	assert.Error(t, err, "Set(\"a,,b\") must return an error")
}

func TestNonEmptyStringSliceValue_CommaSeparatedSplits(t *testing.T) {
	var p []string
	v := newNonEmptyStringSliceValue(nil, &p)
	require.NoError(t, v.Set("a,b,c"))
	assert.Equal(t, []string{"a", "b", "c"}, p)
}

func TestNonEmptyStringSliceValue_FirstSetOverridesDefault(t *testing.T) {
	var p []string
	v := newNonEmptyStringSliceValue([]string{"default"}, &p)

	// Default must be visible before any Set call.
	assert.Equal(t, []string{"default"}, p)

	require.NoError(t, v.Set("first"))
	// After the first Set the default is replaced, not appended.
	assert.Equal(t, []string{"first"}, p)
}

func TestNonEmptyStringSliceValue_SubsequentSetAppends(t *testing.T) {
	var p []string
	v := newNonEmptyStringSliceValue(nil, &p)

	require.NoError(t, v.Set("a"))
	require.NoError(t, v.Set("b"))
	require.NoError(t, v.Set("c"))
	assert.Equal(t, []string{"a", "b", "c"}, p)
}

func TestNonEmptyStringSliceValue_Type(t *testing.T) {
	var p []string
	v := newNonEmptyStringSliceValue(nil, &p)
	assert.Equal(t, "stringSlice", v.Type())
}

func TestNonEmptyStringSliceValue_StringEmpty(t *testing.T) {
	var p []string
	v := newNonEmptyStringSliceValue(nil, &p)
	assert.Equal(t, "[]", v.String())
}

func TestNonEmptyStringSliceValue_StringWithValues(t *testing.T) {
	var p []string
	v := newNonEmptyStringSliceValue([]string{"foo", "bar baz"}, &p)
	assert.Equal(t, `["foo","bar baz"]`, v.String())
}

func TestNonEmptyStringSliceValue_InvalidDefaultPanics(t *testing.T) {
	var p []string
	assert.Panics(t, func() {
		newNonEmptyStringSliceValue([]string{""}, &p)
	})
}

// --- NonEmptyStringSliceVar / NonEmptyStringSliceVarP integration tests ---

func TestNonEmptyStringSliceVar_RejectsEmptyString(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringSliceVar(cmd, &vals, "items", nil, "test flag")

	err := cmd.Flags().Set("items", "")
	assert.Error(t, err)
}

func TestNonEmptyStringSliceVar_DefaultAndParse(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{"default"}, "test flag")

	// Before parsing, default value must be reflected.
	assert.Equal(t, []string{"default"}, vals)

	// After parsing with explicit values, default must be replaced then appended.
	require.NoError(t, cmd.ParseFlags([]string{"--items", "x", "--items", "y"}))
	assert.Equal(t, []string{"x", "y"}, vals)
}

func TestNonEmptyStringSliceVar_CommaSeparated(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringSliceVar(cmd, &vals, "items", nil, "test flag")

	require.NoError(t, cmd.ParseFlags([]string{"--items", "a,b,c"}))
	assert.Equal(t, []string{"a", "b", "c"}, vals)
}

func TestNonEmptyStringSliceVar_MixedCommaAndRepeated(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringSliceVar(cmd, &vals, "items", nil, "test flag")

	require.NoError(t, cmd.ParseFlags([]string{"--items", "a,b", "--items", "c"}))
	assert.Equal(t, []string{"a", "b", "c"}, vals)
}

func TestNonEmptyStringSliceVarP_ShorthandWorks(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringSliceVarP(cmd, &vals, "items", "i", nil, "test flag")

	require.NoError(t, cmd.ParseFlags([]string{"-i", "alpha", "-i", "beta"}))
	assert.Equal(t, []string{"alpha", "beta"}, vals)
}

func TestNonEmptyStringSliceVarP_RejectsEmptyString(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var vals []string
	NonEmptyStringSliceVarP(cmd, &vals, "items", "i", nil, "test flag")

	err := cmd.Flags().Set("items", "")
	assert.Error(t, err)
}
