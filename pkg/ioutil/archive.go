package ioutil

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
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
)

// maxArchiveTotalSize caps the total decompressed size (bytes) read when extracting an
// archive, to guard against decompression-bomb style inputs. It is a var (not a const)
// so tests can temporarily lower it. Every byte decompressed from the tar stream is
// counted, including the payloads of entries that are skipped because they fall outside
// the requested subdir.
var maxArchiveTotalSize int64 = 200 << 20 // 200 MiB

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
//
// destDir is expected to be a freshly created, empty directory that the caller controls;
// the zip-slip check is lexical and does not resolve symlinks, so a pre-existing symlink
// at destDir or in any directory component beneath it (e.g. "destDir/sub" -> "/etc")
// could still redirect writes outside destDir. Callers must not pass a directory that may
// contain attacker-controlled symlinks.
func ExtractTarGzSubdir(r io.Reader, subdir string, destDir string) (err error) {
	// Reject subdir paths that are absolute or escape the archive root before doing any
	// work; otherwise inputs like "../" or "/" would collapse to "" during normalization
	// and be treated as "extract the whole archive", silently broadening the scope.
	if path.IsAbs(subdir) {
		return fmt.Errorf("subdir %q must be a relative path", subdir)
	}
	if cleaned := path.Clean(subdir); cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("subdir %q escapes the archive root", subdir)
	}

	gzr, gzErr := gzip.NewReader(r)
	if gzErr != nil {
		return fmt.Errorf("failed to decompress archive: %w", gzErr)
	}
	defer func() {
		closeErr := gzr.Close()
		if closeErr == nil {
			return
		}
		wrapped := fmt.Errorf("failed to close gzip reader: %w", closeErr)
		switch {
		case err == nil:
			err = wrapped
		case errors.Is(err, closeErr):
			// The primary error already wraps this decompression error; don't duplicate it.
		default:
			err = errors.Join(err, wrapped)
		}
	}()

	subdir = strings.Trim(path.Clean("/"+subdir), "/")

	// Bound the total number of decompressed bytes read from the archive, including the
	// payloads of entries that are skipped (drained by tar.Reader.Next) because they fall
	// outside subdir or are non-regular. This prevents a decompression bomb placed outside
	// subdir from being fully decompressed without counting toward the limit.
	tr := tar.NewReader(&limitedReader{r: gzr, limit: maxArchiveTotalSize})
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

// limitedReader wraps an io.Reader and returns an error once the cumulative number of
// bytes read exceeds limit. Unlike io.LimitedReader, it reports an explicit error rather
// than a silent EOF, so it can enforce a hard cap on the total decompressed size of an
// archive (including the payloads of skipped entries drained by tar.Reader.Next).
type limitedReader struct {
	r     io.Reader
	n     int64
	limit int64
}

func (l *limitedReader) Read(p []byte) (int, error) {
	remaining := l.limit - l.n
	if remaining < 0 {
		return 0, fmt.Errorf("archive exceeds maximum extracted size (limit %d bytes)", l.limit)
	}
	// Read at most one byte beyond the remaining budget so crossing the limit is
	// always detected, even when the underlying reader reports io.EOF in the same
	// call. This bounds the total bytes read from the source to limit+1.
	if int64(len(p)) > remaining+1 {
		p = p[:remaining+1]
	}
	n, err := l.r.Read(p)
	l.n += int64(n)
	if l.n > l.limit {
		return n, fmt.Errorf("archive exceeds maximum extracted size (limit %d bytes)", l.limit)
	}
	return n, err
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
		// destAbs may be a filesystem root (e.g. "/"), for which destAbs+Separator
		// is "//" and the prefix check above would reject valid children. Fall back
		// to filepath.Rel, which reports containment correctly for root destinations.
		relToDest, relErr := filepath.Rel(destAbs, joinedAbs)
		if relErr != nil || relToDest == ".." || strings.HasPrefix(relToDest, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("archive entry %q escapes destination directory", rel)
		}
	}
	return joinedAbs, nil
}
