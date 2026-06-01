package gitutil
package gitutil

import (
	"testing"
)

func TestValidateBranchName(t *testing.T) {
	t.Run("valid names", func(t *testing.T) {
		valid := []string{
			"main",
			"feature/my-branch",
			"release/1.2.3",
			"fix_issue-42",
			"UPPER/lower/Mixed",
			"a",
			"a/b/c",
			"branch.name",
		}
		for _, name := range valid {
			if err := ValidateBranchName(name); err != nil {
				t.Errorf("ValidateBranchName(%q) returned unexpected error: %v", name, err)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		if err := ValidateBranchName(""); err == nil {
			t.Error("expected error for empty branch name")
		}
	})

	t.Run("starts with slash", func(t *testing.T) {
		if err := ValidateBranchName("/branch"); err == nil {
			t.Error("expected error for branch starting with /")
		}
	})

	t.Run("ends with slash", func(t *testing.T) {
		if err := ValidateBranchName("branch/"); err == nil {
			t.Error("expected error for branch ending with /")
		}
	})

	t.Run("ends with dot", func(t *testing.T) {
		if err := ValidateBranchName("branch."); err == nil {
			t.Error("expected error for branch ending with .")
		}
	})

	t.Run("double slash", func(t *testing.T) {
		if err := ValidateBranchName("feat//branch"); err == nil {
			t.Error("expected error for branch containing //")
		}
	})

	t.Run("double dot", func(t *testing.T) {
		if err := ValidateBranchName("feat..branch"); err == nil {
			t.Error("expected error for branch containing ..")
		}
	})

	t.Run("component starts with dot", func(t *testing.T) {
		if err := ValidateBranchName("feat/.hidden"); err == nil {
			t.Error("expected error for path component starting with .")
		}
	})

	t.Run("component ends with .lock", func(t *testing.T) {
		if err := ValidateBranchName("feat/branch.lock"); err == nil {
			t.Error("expected error for path component ending with .lock")
		}
	})

	t.Run("shell metacharacter dollar", func(t *testing.T) {
		if err := ValidateBranchName("branch-$HOME"); err == nil {
			t.Error("expected error for branch containing $")
		}
	})

	t.Run("shell metacharacter backtick", func(t *testing.T) {
		if err := ValidateBranchName("branch-`id`"); err == nil {
			t.Error("expected error for branch containing backtick")
		}
	})

	t.Run("shell metacharacter semicolon", func(t *testing.T) {
		if err := ValidateBranchName("branch;evil"); err == nil {
			t.Error("expected error for branch containing ;")
		}
	})

	t.Run("space", func(t *testing.T) {
		if err := ValidateBranchName("branch name"); err == nil {
			t.Error("expected error for branch containing space")
		}
	})

	t.Run("tilde", func(t *testing.T) {
		if err := ValidateBranchName("branch~1"); err == nil {
			t.Error("expected error for branch containing ~")
		}
	})

	t.Run("caret", func(t *testing.T) {
		if err := ValidateBranchName("branch^0"); err == nil {
			t.Error("expected error for branch containing ^")
		}
	})

	t.Run("colon", func(t *testing.T) {
		if err := ValidateBranchName("branch:name"); err == nil {
			t.Error("expected error for branch containing :")
		}
	})
}
