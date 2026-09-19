package copilotext

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve(t *testing.T) {
	ext := Extension{
		Name: "my-extension",
		URL:  "https://github.com/owner/repo/tree/v1.2.3/.github/extensions/my-extension",
	}

	src, err := resolve(ext, "")
	if err != nil {
		t.Fatalf("resolve() error = %v", err)
	}
	if src.Repo.Owner != "owner" || src.Repo.Name != "repo" {
		t.Fatalf("resolve() repo = %+v, want owner/repo", src.Repo)
	}
	if src.Ref != "v1.2.3" {
		t.Fatalf("resolve() ref = %q, want %q", src.Ref, "v1.2.3")
	}
	if src.Path != ".github/extensions/my-extension" {
		t.Fatalf("resolve() path = %q, want %q", src.Path, ".github/extensions/my-extension")
	}

	src, err = resolve(ext, "main")
	if err != nil {
		t.Fatalf("resolve() with override error = %v", err)
	}
	if src.Ref != "main" {
		t.Fatalf("resolve() with override ref = %q, want %q", src.Ref, "main")
	}
}

func TestResolveInvalidURL(t *testing.T) {
	ext := Extension{Name: "bad", URL: "https://github.com/owner/repo"}
	if _, err := resolve(ext, ""); err == nil {
		t.Fatal("resolve() with non-tree URL: expected error, got nil")
	}
}

func TestConfigFindAndSelectExtensions(t *testing.T) {
	cfg := Config{
		Extensions: []Extension{
			{Name: "a", URL: "https://github.com/owner/repo/tree/v1/a"},
			{Name: "b", URL: "https://github.com/owner/repo/tree/v1/b"},
		},
	}

	if _, err := cfg.find("a"); err != nil {
		t.Fatalf("find(a) error = %v", err)
	}
	if _, err := cfg.find("missing"); err == nil {
		t.Fatal("find(missing): expected error, got nil")
	}

	all, err := cfg.selectExtensions(nil)
	if err != nil {
		t.Fatalf("selectExtensions(nil) error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("selectExtensions(nil) len = %d, want 2", len(all))
	}

	one, err := cfg.selectExtensions([]string{"b"})
	if err != nil {
		t.Fatalf("selectExtensions([b]) error = %v", err)
	}
	if len(one) != 1 || one[0].Name != "b" {
		t.Fatalf("selectExtensions([b]) = %+v, want [b]", one)
	}

	if _, err := cfg.selectExtensions([]string{"missing"}); err == nil {
		t.Fatal("selectExtensions([missing]): expected error, got nil")
	}
}

func TestExtensionsRootPrefixOverridesScope(t *testing.T) {
	dir, err := extensionsRoot(context.Background(), ScopeUser, "/custom/prefix")
	if err != nil {
		t.Fatalf("extensionsRoot() error = %v", err)
	}
	if dir != "/custom/prefix" {
		t.Fatalf("extensionsRoot() = %q, want %q", dir, "/custom/prefix")
	}
}

func TestExtensionsRootUserScopeUsesCopilotHome(t *testing.T) {
	t.Setenv("COPILOT_HOME", "/tmp/copilot-home")
	dir, err := extensionsRoot(context.Background(), ScopeUser, "")
	if err != nil {
		t.Fatalf("extensionsRoot() error = %v", err)
	}
	want := filepath.Join("/tmp/copilot-home", "extensions")
	if dir != want {
		t.Fatalf("extensionsRoot() = %q, want %q", dir, want)
	}
}

func TestExtensionsRootUserScopeFallsBackToHomeDir(t *testing.T) {
	t.Setenv("COPILOT_HOME", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home directory: %v", err)
	}
	dir, err := extensionsRoot(context.Background(), ScopeUser, "")
	if err != nil {
		t.Fatalf("extensionsRoot() error = %v", err)
	}
	want := filepath.Join(home, ".copilot", "extensions")
	if dir != want {
		t.Fatalf("extensionsRoot() = %q, want %q", dir, want)
	}
}

func TestMetadataRoundTrip(t *testing.T) {
	dir := t.TempDir()

	if m, err := readMetadata(dir); err != nil || m != nil {
		t.Fatalf("readMetadata() on dir with no metadata = %+v, %v, want nil, nil", m, err)
	}

	want := metadata{Tool: "gh-team-kit", ToolVersion: "1.0.0", Host: "github.com", Owner: "owner", Repo: "repo", Ref: "main", CommitSHA: "abc123"}
	if err := writeMetadata(dir, want); err != nil {
		t.Fatalf("writeMetadata() error = %v", err)
	}

	got, err := readMetadata(dir)
	if err != nil {
		t.Fatalf("readMetadata() error = %v", err)
	}
	if got == nil || got.CommitSHA != want.CommitSHA || got.Ref != want.Ref {
		t.Fatalf("readMetadata() = %+v, want %+v", got, want)
	}
}

func TestReadMetadataCorruptFileIsUnmanaged(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(metadataPath(dir), []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to write corrupt metadata file: %v", err)
	}
	m, err := readMetadata(dir)
	if err != nil {
		t.Fatalf("readMetadata() error = %v, want nil", err)
	}
	if m != nil {
		t.Fatalf("readMetadata() = %+v, want nil for corrupt file", m)
	}
}

func TestRequireOverwritable(t *testing.T) {
	base := t.TempDir()

	missing := filepath.Join(base, "missing")
	if err := requireOverwritable(missing, false); err != nil {
		t.Fatalf("requireOverwritable() on missing dir error = %v, want nil", err)
	}

	unmanaged := filepath.Join(base, "unmanaged")
	if err := os.Mkdir(unmanaged, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := requireOverwritable(unmanaged, false); err == nil {
		t.Fatal("requireOverwritable() on unmanaged dir: expected error, got nil")
	}
	if err := requireOverwritable(unmanaged, true); err != nil {
		t.Fatalf("requireOverwritable() with force error = %v, want nil", err)
	}

	managed := filepath.Join(base, "managed")
	if err := os.Mkdir(managed, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := writeMetadata(managed, metadata{Tool: "gh-team-kit"}); err != nil {
		t.Fatalf("writeMetadata() error = %v", err)
	}
	if err := requireOverwritable(managed, false); err != nil {
		t.Fatalf("requireOverwritable() on managed dir error = %v, want nil", err)
	}
}

func TestGetStatusNotInstalled(t *testing.T) {
	cfg := Config{Extensions: []Extension{{Name: "my-extension", URL: "https://github.com/owner/repo/tree/v1/dir"}}}
	prefix := t.TempDir()

	status, err := GetStatus(context.Background(), cfg, "my-extension", ScopeUser, prefix)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Installed {
		t.Fatalf("GetStatus() = %+v, want Installed=false", status)
	}
}

func TestGetStatusManaged(t *testing.T) {
	cfg := Config{Extensions: []Extension{{Name: "my-extension", URL: "https://github.com/owner/repo/tree/v1/dir"}}}
	prefix := t.TempDir()
	dir := filepath.Join(prefix, "my-extension")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := writeMetadata(dir, metadata{Ref: "v1", CommitSHA: "abc"}); err != nil {
		t.Fatalf("writeMetadata() error = %v", err)
	}

	status, err := GetStatus(context.Background(), cfg, "my-extension", ScopeUser, prefix)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if !status.Installed || !status.Managed || status.Ref != "v1" || status.CommitSHA != "abc" {
		t.Fatalf("GetStatus() = %+v, want installed and managed with ref=v1 commit=abc", status)
	}
}

func TestUpdateNotInstalledFailsBeforeResolvingRef(t *testing.T) {
	cfg := Config{Extensions: []Extension{{Name: "my-extension", URL: "https://github.com/owner/repo/tree/v1/dir"}}}
	prefix := t.TempDir()
	dir := filepath.Join(prefix, "my-extension")

	// Neither update nor update --force may bootstrap a not-installed extension; both
	// must fail before any network call (which is why resolve/GetCommitSHA1 are never
	// reached and no HTTP mocking is required here).
	for _, force := range []bool{false, true} {
		opts := InstallOptions{Scope: ScopeUser, Prefix: prefix, Force: force}
		_, err := Update(context.Background(), cfg, "my-extension", opts)
		if err == nil {
			t.Fatalf("Update(force=%v) on not-installed extension: expected error, got nil", force)
		}
		if !strings.Contains(err.Error(), "not installed") {
			t.Fatalf("Update(force=%v) error = %v, want it to mention \"not installed\"", force, err)
		}
		if _, statErr := os.Stat(dir); !os.IsNotExist(statErr) {
			t.Fatalf("Update(force=%v) created or left %q; stat err = %v", force, dir, statErr)
		}
	}
}

func TestUpdateUnmanagedWithoutForceRefuses(t *testing.T) {
	cfg := Config{Extensions: []Extension{{Name: "my-extension", URL: "https://github.com/owner/repo/tree/v1/dir"}}}
	prefix := t.TempDir()
	dir := filepath.Join(prefix, "my-extension")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	_, err := Update(context.Background(), cfg, "my-extension", InstallOptions{Scope: ScopeUser, Prefix: prefix})
	if err == nil {
		t.Fatal("Update() on unmanaged dir without force: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not managed by this command") {
		t.Fatalf("Update() error = %v, want it to mention \"not managed by this command\"", err)
	}
}

func TestUninstallRefusesUnmanagedWithoutForce(t *testing.T) {
	cfg := Config{Extensions: []Extension{{Name: "my-extension", URL: "https://github.com/owner/repo/tree/v1/dir"}}}
	prefix := t.TempDir()
	dir := filepath.Join(prefix, "my-extension")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	if err := Uninstall(context.Background(), cfg, "my-extension", ScopeUser, prefix, false, false); err == nil {
		t.Fatal("Uninstall() on unmanaged dir without force: expected error, got nil")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("Uninstall() should not have removed the directory: %v", err)
	}

	if err := Uninstall(context.Background(), cfg, "my-extension", ScopeUser, prefix, false, true); err != nil {
		t.Fatalf("Uninstall() with force error = %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("Uninstall() with force should have removed the directory, stat err = %v", err)
	}
}

func TestUninstallDryRunDoesNotRemove(t *testing.T) {
	cfg := Config{Extensions: []Extension{{Name: "my-extension", URL: "https://github.com/owner/repo/tree/v1/dir"}}}
	prefix := t.TempDir()
	dir := filepath.Join(prefix, "my-extension")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := writeMetadata(dir, metadata{Ref: "v1", CommitSHA: "abc"}); err != nil {
		t.Fatalf("writeMetadata() error = %v", err)
	}

	if err := Uninstall(context.Background(), cfg, "my-extension", ScopeUser, prefix, true, false); err != nil {
		t.Fatalf("Uninstall() dry-run error = %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("Uninstall() dry-run should not have removed the directory: %v", err)
	}
}

func TestUninstallNotInstalled(t *testing.T) {
	cfg := Config{Extensions: []Extension{{Name: "my-extension", URL: "https://github.com/owner/repo/tree/v1/dir"}}}
	prefix := t.TempDir()

	if err := Uninstall(context.Background(), cfg, "my-extension", ScopeUser, prefix, false, false); err == nil {
		t.Fatal("Uninstall() on non-installed extension: expected error, got nil")
	}
}

func TestMetadataJSONFieldNames(t *testing.T) {
	dir := t.TempDir()
	if err := writeMetadata(dir, metadata{Tool: "gh-team-kit", Ref: "main", CommitSHA: "abc123"}); err != nil {
		t.Fatalf("writeMetadata() error = %v", err)
	}
	data, err := os.ReadFile(metadataPath(dir))
	if err != nil {
		t.Fatalf("failed to read metadata file: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal metadata file: %v", err)
	}
	for _, key := range []string{"tool", "ref", "commit_sha"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("metadata JSON missing key %q, got %v", key, raw)
		}
	}
}
