package copilotext

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtensionCommandsDryRunFlags(t *testing.T) {
	tests := []struct {
		name string
		flag string
		want bool
	}{
		{name: "default"},
		{name: "long", flag: "--dryrun", want: true},
		{name: "short", flag: "-n", want: true},
		{name: "long false", flag: "--dryrun=false"},
		{name: "short false", flag: "-n=false"},
	}

	for _, action := range []string{"install", "update", "uninstall"} {
		t.Run(action, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					cmd := NewExtensionCmd(Config{})
					args := []string{action}
					if tt.flag != "" {
						args = append(args, tt.flag)
					}
					cmd.SetArgs(args)

					executed, err := cmd.ExecuteC()
					if err != nil {
						t.Fatalf("ExecuteC() error = %v", err)
					}
					got, err := executed.Flags().GetBool("dryrun")
					if err != nil {
						t.Fatalf("GetBool(dryrun) error = %v", err)
					}
					if got != tt.want {
						t.Errorf("dryrun = %v, want %v", got, tt.want)
					}
				})
			}

			t.Run("help", func(t *testing.T) {
				cmd := NewExtensionCmd(Config{})
				var output bytes.Buffer
				cmd.SetOut(&output)
				cmd.SetArgs([]string{action, "--help"})
				if err := cmd.Execute(); err != nil {
					t.Fatalf("Execute() error = %v", err)
				}
				if !strings.Contains(output.String(), "-n, --dryrun") {
					t.Errorf("help does not contain -n, --dryrun: %s", output.String())
				}
				if strings.Contains(output.String(), "--dry-run") {
					t.Errorf("help contains old --dry-run flag: %s", output.String())
				}
			})

			t.Run("old flag rejected", func(t *testing.T) {
				cmd := NewExtensionCmd(Config{})
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true
				cmd.SetArgs([]string{action, "--dry-run"})
				err := cmd.Execute()
				if err == nil || !strings.Contains(err.Error(), "unknown flag: --dry-run") {
					t.Errorf("Execute() error = %v, want unknown flag error", err)
				}
			})
		})
	}
}

func TestExtensionUninstallDryRunFlags(t *testing.T) {
	tests := []struct {
		name   string
		flag   string
		dryRun bool
	}{
		{name: "default"},
		{name: "long", flag: "--dryrun", dryRun: true},
		{name: "short", flag: "-n", dryRun: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix := t.TempDir()
			name := "example-extension"
			dir := filepath.Join(prefix, name)
			if err := os.Mkdir(dir, 0755); err != nil {
				t.Fatalf("Mkdir() error = %v", err)
			}
			if err := writeMetadata(dir, metadata{Ref: "v1", CommitSHA: "abc"}); err != nil {
				t.Fatalf("writeMetadata() error = %v", err)
			}
			marker := filepath.Join(dir, "marker.txt")
			if err := os.WriteFile(marker, []byte("original"), 0644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}

			cmd := NewExtensionCmd(Config{Extensions: []Extension{{Name: name}}})
			var output bytes.Buffer
			cmd.SetOut(&output)
			args := []string{"uninstall", "--prefix", prefix}
			if tt.flag != "" {
				args = append(args, tt.flag)
			}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			wantOutput := name + "\tuninstalled\n"
			if tt.dryRun {
				wantOutput = name + "\twould be uninstalled\n"
				content, err := os.ReadFile(marker)
				if err != nil || string(content) != "original" {
					t.Errorf("marker content = %q, error = %v, want unchanged", content, err)
				}
				if _, err := os.Stat(metadataPath(dir)); err != nil {
					t.Errorf("metadata should be preserved: %v", err)
				}
			} else if _, err := os.Stat(dir); !os.IsNotExist(err) {
				t.Errorf("extension directory should be removed, stat error = %v", err)
			}
			if output.String() != wantOutput {
				t.Errorf("output = %q, want %q", output.String(), wantOutput)
			}
		})
	}
}
