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

	// Use a sentinel directory as a sibling of destDir so that a successful escape
	// (writing to "<parent>/etc/passwd") would land in an inspectable location.
	parent := t.TempDir()
	destDir := filepath.Join(parent, "dest")
	if err := os.Mkdir(destDir, 0755); err != nil {
		t.Fatalf("failed to create destDir: %v", err)
	}

	err := ExtractTarGzSubdir(bytes.NewReader(archive), "ext", destDir)
	// path.Clean neutralizes ".." within the rooted archive path, so this entry
	// resolves outside "ext" and is simply skipped, leaving no files to extract.
	if err == nil {
		t.Fatal("expected error when traversal entry is the only candidate, got nil")
	}

	// Regardless of the error, nothing may have been written outside destDir.
	if _, statErr := os.Stat(filepath.Join(parent, "etc", "passwd")); !os.IsNotExist(statErr) {
		t.Fatalf("path traversal entry escaped destDir, stat err = %v", statErr)
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

func TestSafeJoin(t *testing.T) {
	sep := string(filepath.Separator)
	cases := []struct {
		name    string
		destDir string
		rel     string
		wantErr bool
	}{
		{name: "child file", destDir: filepath.Join("tmp", "dest"), rel: "a/b.txt", wantErr: false},
		{name: "destination itself", destDir: filepath.Join("tmp", "dest"), rel: ".", wantErr: false},
		{name: "parent escape", destDir: filepath.Join("tmp", "dest"), rel: "../escape", wantErr: true},
		{name: "root destination child", destDir: sep, rel: "file", wantErr: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := safeJoin(tc.destDir, tc.rel)
			if tc.wantErr && err == nil {
				t.Fatalf("safeJoin(%q, %q) = nil error, want error", tc.destDir, tc.rel)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("safeJoin(%q, %q) error = %v, want nil", tc.destDir, tc.rel, err)
			}
		})
	}
}

func TestExtractTarGzSubdir_SizeLimitCountsSkippedEntries(t *testing.T) {
	// A large file placed OUTSIDE the requested subdir must still count toward the
	// decompressed-size limit, because tar.Reader.Next drains (decompresses) its payload.
	orig := maxArchiveTotalSize
	maxArchiveTotalSize = 1 << 10 // 1 KiB
	defer func() { maxArchiveTotalSize = orig }()

	big := make([]byte, 8<<10) // 8 KiB, well above the lowered limit
	archive := buildTarGz(t, []tarEntry{
		{name: "owner-repo-abc123/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/big-outside.bin", content: string(big)},
		{name: "owner-repo-abc123/ext/", typeflag: tar.TypeDir},
		{name: "owner-repo-abc123/ext/small.txt", content: "small"},
	})

	err := ExtractTarGzSubdir(bytes.NewReader(archive), "ext", t.TempDir())
	if err == nil {
		t.Fatal("expected error when a skipped entry exceeds the size limit, got nil")
	}
}
