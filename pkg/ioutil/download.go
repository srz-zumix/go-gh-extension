package ioutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// GetFilename extracts the file name from a URL.
// It strips query-string parameters (e.g. JWT tokens on private images).
func GetFilename(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return path.Base(rawURL)
	}
	return path.Base(u.Path)
}

// SafeFilename returns a filesystem-safe flat filename for a URL-keyed asset by
// prepending a FNV-1a hash of the URL to avoid collisions when multiple assets
// share the same base filename.
// The provided filename is sanitized to strip path separators (both '/' and '\'),
// ".." components, and characters reserved on Windows (: * ? " < > |) so the
// result is a safe, flat filename across platforms.
func SafeFilename(rawURL, filename string) string {
	// Normalize backslashes to forward slashes so filepath.Base and path.Base
	// both treat them as separators, then take the base component.
	normalized := strings.ReplaceAll(filename, `\`, "/")
	safe := path.Base(filepath.Base(normalized))
	// path.Base returns "." for empty or separator-only input; fall back to the
	// URL-derived name in that case.
	if safe == "." || safe == "" {
		safe = GetFilename(rawURL)
	}
	// Remove characters that are reserved or problematic on Windows and common
	// filesystems (: * ? " < > | and control characters).
	safe = strings.Map(func(r rune) rune {
		switch r {
		case ':', '*', '?', '"', '<', '>', '|':
			return '_'
		}
		if r < 0x20 {
			return '_'
		}
		return r
	}, safe)
	return fmt.Sprintf("%x_%s", fnv32(rawURL), safe)
}

// fnv32 computes a FNV-1a 32-bit hash of a string for filename deduplication.
func fnv32(s string) uint32 {
	const (
		offset32 uint32 = 2166136261
		prime32  uint32 = 16777619
	)
	h := offset32
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return h
}

// DownloadFile performs an HTTP GET using client and saves the response body to
// destPath. It writes to a sibling temp file first and renames to destPath only
// on success, so a failed download never leaves a partial or corrupted file at
// the destination. The temp file is removed on any failure.
func DownloadFile(ctx context.Context, client *http.Client, rawURL, destPath string) error {
	logger.Debug("downloading file", "url", rawURL, "dest", destPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Debug("failed to close response body", "url", rawURL, "error", closeErr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected http status %s for %s", resp.Status, rawURL)
	}

	// Write to a temp file in the same directory so the final rename is atomic
	// (same filesystem) and destPath is never left in a partial state.
	tmp, err := os.CreateTemp(filepath.Dir(destPath), ".dl-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	n, copyErr := io.Copy(tmp, resp.Body)
	closeErr := tmp.Close()

	if copyErr != nil {
		os.Remove(tmpName) //nolint:errcheck
		return fmt.Errorf("failed to write download: %w", copyErr)
	}
	if closeErr != nil {
		os.Remove(tmpName) //nolint:errcheck
		return fmt.Errorf("failed to close temp file: %w", closeErr)
	}

	if err := ReplaceFile(tmpName, destPath); err != nil {
		os.Remove(tmpName) //nolint:errcheck
		return fmt.Errorf("failed to move download to destination: %w", err)
	}

	logger.Debug("file saved", "dest", destPath, "bytes", n)
	return nil
}
