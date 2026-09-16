package ioutil

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	// maxArchiveEntries caps the number of entries processed from an archive to bound
	// resource usage when extracting untrusted tarballs.
	maxArchiveEntries = 20000
	// maxArchiveTotalSize caps the total decompressed size (bytes) written when extracting
	// an archive, to guard against decompression-bomb style inputs.
	maxArchiveTotalSize = 200 << 20 // 200 MiB
)

// DownloadZipArchive downloads the workflow run logs archive and returns a zip.Reader for accessing the contents.
func DownloadZipArchive(ctx context.Context, logURL string) (*zip.Reader, int64, error) {
	// Download the zip file
	req, err := http.NewRequestWithContext(ctx, "GET", logURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download log archive: %w", err)
	}
	defer resp.Body.Close() // nolint

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("failed to download log archive: status code %d", resp.StatusCode)
	}

	// Read the entire zip file into memory
	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read log archive: %w", err)
	}

	// Create a zip reader from the in-memory data
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create zip reader: %w", err)
	}

	return zipReader, int64(len(zipData)), nil
}

// ExtractTarGzSubdir extracts subdir from a gzip-compressed tar archive (as returned by
// the GitHub repository archive API) into destDir, which must already exist. The
// archive's single top-level directory (e.g. "owner-repo-sha1234/") is stripped
// automatically, then subdir is stripped so its contents land directly under destDir.
// Only regular files and directories are extracted: symlinks, hardlinks, devices, and
// other special entries are skipped. Entries are also rejected if resolving them would
// escape destDir (zip-slip), and the total entry count and decompressed size are capped
// to guard against decompression bombs.
func ExtractTarGzSubdir(r io.Reader, subdir string, destDir string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to decompress archive: %w", err)
	}
	defer gzr.Close() //nolint:errcheck

	subdir = strings.Trim(path.Clean("/"+subdir), "/")

	tr := tar.NewReader(gzr)
	var totalSize int64
	entries := 0
	extracted := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read archive entry: %w", err)
		}
		entries++
		if entries > maxArchiveEntries {
			return fmt.Errorf("archive has too many entries (limit %d)", maxArchiveEntries)
		}

		rel, ok := stripArchiveRoot(hdr.Name, subdir)
		if !ok {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if rel == "" {
				continue
			}
			destPath, err := safeJoin(destDir, rel)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %q: %w", rel, err)
			}
		case tar.TypeReg:
			if rel == "" {
				return fmt.Errorf("archive entry %q is a file where a directory was expected", hdr.Name)
			}
			destPath, err := safeJoin(destDir, rel)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %q: %w", rel, err)
			}
			totalSize += hdr.Size
			if totalSize > maxArchiveTotalSize {
				return fmt.Errorf("archive exceeds maximum extracted size (limit %d bytes)", maxArchiveTotalSize)
			}
			if _, err := WriteFileAtomicFrom(destPath, io.LimitReader(tr, hdr.Size), 0644); err != nil {
				return fmt.Errorf("failed to write file %q: %w", rel, err)
			}
			extracted++
		default:
			// Skip symlinks, hardlinks, devices, and other non-regular entries.
			continue
		}
	}
	if extracted == 0 {
		return fmt.Errorf("no files found under %q in archive", subdir)
	}
	return nil
}

// stripArchiveRoot removes the archive's single top-level directory component from name,
// then removes the subdir prefix. Returns ok=false if name is not under subdir.
func stripArchiveRoot(name string, subdir string) (rel string, ok bool) {
	cleaned := path.Clean("/" + filepath.ToSlash(name))
	parts := strings.Split(strings.Trim(cleaned, "/"), "/")
	if len(parts) <= 1 {
		// Only the top-level directory itself, with no further path.
		return "", false
	}
	rest := strings.Join(parts[1:], "/")
	if subdir == "" {
		return rest, true
	}
	if rest == subdir {
		return "", true
	}
	if rest, ok := strings.CutPrefix(rest, subdir+"/"); ok {
		return rest, true
	}
	return "", false
}

// safeJoin joins destDir with the cleaned, slash-separated relative path rel and returns
// an error if the result would resolve outside destDir (zip-slip protection).
func safeJoin(destDir, rel string) (string, error) {
	joined := filepath.Join(destDir, filepath.FromSlash(rel))
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve destination directory: %w", err)
	}
	joinedAbs, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %q: %w", rel, err)
	}
	if joinedAbs != destAbs && !strings.HasPrefix(joinedAbs, destAbs+string(filepath.Separator)) {
		return "", fmt.Errorf("archive entry %q escapes destination directory", rel)
	}
	return joinedAbs, nil
}
