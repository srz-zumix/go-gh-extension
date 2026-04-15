package cmdflags

import (
	"testing"

	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	return &cobra.Command{Use: "test", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
}

// --- NonEmptyStringArrayVar tests ---

func TestNonEmptyStringArrayVar_SingleValue(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringArrayVar(cmd, &vals, "items", []string{}, "test flag")

	cmd.SetArgs([]string{"--items", "a"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 || vals[0] != "a" {
		t.Fatalf("expected [a], got %v", vals)
	}
}

func TestNonEmptyStringArrayVar_MultipleValues(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringArrayVar(cmd, &vals, "items", []string{}, "test flag")

	cmd.SetArgs([]string{"--items", "a", "--items", "b"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 2 || vals[0] != "a" || vals[1] != "b" {
		t.Fatalf("expected [a b], got %v", vals)
	}
}

func TestNonEmptyStringArrayVar_RejectsEmpty(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringArrayVar(cmd, &vals, "items", []string{}, "test flag")

	cmd.SetArgs([]string{"--items", ""})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}

func TestNonEmptyStringArrayVar_DefaultValue(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringArrayVar(cmd, &vals, "items", []string{"default"}, "test flag")

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 || vals[0] != "default" {
		t.Fatalf("expected [default], got %v", vals)
	}
}

func TestNonEmptyStringArrayVar_InvalidDefaultPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty default value")
		}
	}()

	var vals []string
	cmd := newTestCmd()
	NonEmptyStringArrayVar(cmd, &vals, "items", []string{""}, "test flag")
}

func TestNonEmptyStringArrayVar_CommaSeparatedNotSplit(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringArrayVar(cmd, &vals, "items", []string{}, "test flag")

	// StringArray does NOT split on comma; "a,b" is a single value
	cmd.SetArgs([]string{"--items", "a,b"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 || vals[0] != "a,b" {
		t.Fatalf("expected [a,b], got %v", vals)
	}
}

// --- NonEmptyStringSliceVar tests ---

func TestNonEmptyStringSliceVar_SingleValue(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{}, "test flag")

	cmd.SetArgs([]string{"--items", "a"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 || vals[0] != "a" {
		t.Fatalf("expected [a], got %v", vals)
	}
}

func TestNonEmptyStringSliceVar_CommaSeparated(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{}, "test flag")

	cmd.SetArgs([]string{"--items", "a,b,c"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 3 || vals[0] != "a" || vals[1] != "b" || vals[2] != "c" {
		t.Fatalf("expected [a b c], got %v", vals)
	}
}

func TestNonEmptyStringSliceVar_RepeatedFlag(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{}, "test flag")

	cmd.SetArgs([]string{"--items", "a", "--items", "b"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 2 || vals[0] != "a" || vals[1] != "b" {
		t.Fatalf("expected [a b], got %v", vals)
	}
}

func TestNonEmptyStringSliceVar_MixedCommaAndRepeated(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{}, "test flag")

	cmd.SetArgs([]string{"--items", "a,b", "--items", "c"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 3 || vals[0] != "a" || vals[1] != "b" || vals[2] != "c" {
		t.Fatalf("expected [a b c], got %v", vals)
	}
}

func TestNonEmptyStringSliceVar_RejectsEmpty(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{}, "test flag")

	cmd.SetArgs([]string{"--items", ""})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}

func TestNonEmptyStringSliceVar_RejectsEmptyInCommaList(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{}, "test flag")

	cmd.SetArgs([]string{"--items", "a,,b"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for empty entry in comma list, got nil")
	}
}

func TestNonEmptyStringSliceVar_DefaultValue(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{"default"}, "test flag")

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 || vals[0] != "default" {
		t.Fatalf("expected [default], got %v", vals)
	}
}

func TestNonEmptyStringSliceVar_InvalidDefaultPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty default value")
		}
	}()

	var vals []string
	cmd := newTestCmd()
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{""}, "test flag")
}

func TestNonEmptyStringSliceVar_OverridesDefault(t *testing.T) {
	var vals []string
	cmd := newTestCmd()
	NonEmptyStringSliceVar(cmd, &vals, "items", []string{"default"}, "test flag")

	cmd.SetArgs([]string{"--items", "x"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 1 || vals[0] != "x" {
		t.Fatalf("expected [x], got %v", vals)
	}
}
