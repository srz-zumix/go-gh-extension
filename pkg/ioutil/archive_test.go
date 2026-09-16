package ioutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// tarEntry describes one entry to write into a test tar.gz archive.
type tarEntry struct {
	name     string
	typeflag byte
	content  string
	linkname string
}

func buildTarGz(t *testing.T, entries []tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for _, e := range entries {
		typeflag := e.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		hdr := &tar.Header{
			Name:     e.name,
			Typeflag: typeflag,
			Mode:     0644,
			Size:     int64(len(e.content)),
			Linkname: e.linkname,
		}
		if typeflag == tar.TypeDir {
			hdr.Mode = 0755
			hdr.Size = 0
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write tar header for %q: %v", e.name, err)
		}
		if typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(e.content)); err != nil {
				t.Fatalf("failed to write tar content for %q: %v", e.name, err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	return buf.Bytes()
}

func TestExtractTarGzSubdir_ExtractsFilesUnderSubdir(t *testing.T) {
	archive := buildTarGz(t, []tarEntry{
		{name: "owner-repo-abc123/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/README.md", content: "root readme"},
		{name: "owner-repo-abc123/.github/extensions/pr-graph-dashboard/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/.github/extensions/pr-graph-dashboard/extension.mjs", content: "export default {}"},
		{name: "owner-repo-abc123/.github/extensions/pr-graph-dashboard/ui/index.html", content: "<html></html>"},
	})

	destDir := t.TempDir()
	if err := ExtractTarGzSubdir(bytes.NewReader(archive), ".github/extensions/pr-graph-dashboard", destDir); err != nil {
		t.Fatalf("ExtractTarGzSubdir() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "extension.mjs"))
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(got) != "export default {}" {
		t.Errorf("extension.mjs content = %q, want %q", got, "export default {}")
	}

	got, err = os.ReadFile(filepath.Join(destDir, "ui", "index.html"))
	if err != nil {
		t.Fatalf("failed to read extracted nested file: %v", err)
	}
	if string(got) != "<html></html>" {
		t.Errorf("ui/index.html content = %q, want %q", got, "<html></html>")
	}

	if _, err := os.Stat(filepath.Join(destDir, "README.md")); !os.IsNotExist(err) {
		t.Errorf("expected README.md outside subdir to be excluded, stat err = %v", err)
	}
}

func TestExtractTarGzSubdir_RejectsPathTraversal(t *testing.T) {
	archive := buildTarGz(t, []tarEntry{
		{name: "owner-repo-abc123/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/ext/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/ext/../../etc/passwd", content: "malicious"},
	})

	destDir := t.TempDir()
	err := ExtractTarGzSubdir(bytes.NewReader(archive), "ext", destDir)
	// path.Clean neutralizes ".." within the rooted archive path, so this entry
	// resolves outside "ext" and is simply skipped rather than extracted.
	if err == nil {
		if _, statErr := os.Stat(filepath.Join(destDir, "..", "etc", "passwd")); statErr == nil {
			t.Fatalf("path traversal entry was extracted outside destDir")
		}
	}
}

func TestExtractTarGzSubdir_SkipsSymlinks(t *testing.T) {
	archive := buildTarGz(t, []tarEntry{
		{name: "owner-repo-abc123/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/ext/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/ext/evil-link", typeflag: tar.TypeSymlink, linkname: "/etc/passwd"},
		{name: "owner-repo-abc123/ext/real-file.txt", content: "ok"},
	})

	destDir := t.TempDir()
	if err := ExtractTarGzSubdir(bytes.NewReader(archive), "ext", destDir); err != nil {
		t.Fatalf("ExtractTarGzSubdir() error = %v", err)
	}

	if _, err := os.Lstat(filepath.Join(destDir, "evil-link")); !os.IsNotExist(err) {
		t.Errorf("expected symlink entry to be skipped, lstat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "real-file.txt")); err != nil {
		t.Errorf("expected regular file to be extracted, stat err = %v", err)
	}
}

func TestExtractTarGzSubdir_NoFilesUnderSubdirReturnsError(t *testing.T) {
	archive := buildTarGz(t, []tarEntry{
		{name: "owner-repo-abc123/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/other/file.txt", content: "not in subdir"},
	})

	destDir := t.TempDir()
	err := ExtractTarGzSubdir(bytes.NewReader(archive), "missing-subdir", destDir)
	if err == nil {
		t.Fatal("expected error when subdir has no files, got nil")
	}
}

func TestExtractTarGzSubdir_EmptySubdirExtractsWholeArchive(t *testing.T) {
	archive := buildTarGz(t, []tarEntry{
		{name: "owner-repo-abc123/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/a.txt", content: "a"},
		{name: "owner-repo-abc123/dir/b.txt", content: "b"},
	})

	destDir := t.TempDir()
	if err := ExtractTarGzSubdir(bytes.NewReader(archive), "", destDir); err != nil {
		t.Fatalf("ExtractTarGzSubdir() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "a.txt")); err != nil {
		t.Errorf("expected a.txt to be extracted, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "dir", "b.txt")); err != nil {
		t.Errorf("expected dir/b.txt to be extracted, stat err = %v", err)
	}
}

func TestExtractTarGzSubdir_InvalidGzip(t *testing.T) {
	err := ExtractTarGzSubdir(bytes.NewReader([]byte("not gzip data")), "ext", t.TempDir())
	if err == nil {
		t.Fatal("expected error for invalid gzip data, got nil")
	}
}
