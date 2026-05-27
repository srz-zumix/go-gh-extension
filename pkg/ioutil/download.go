package ioutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/httputil"
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
// prepending a FNV-1a 32-bit hash of the URL to reduce (not eliminate) the
// chance of name collisions when multiple assets share the same base filename.
// The provided filename is sanitized to strip path separators (both '/' and '\'),
// ".." components, characters reserved on Windows (: * ? " < > |), and Windows
// reserved device names (CON, PRN, AUX, NUL, COM1..9, LPT1..9) so the result is
// a safe, flat filename across platforms.
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
	// Windows does not allow filenames that end with a dot or space.
	safe = strings.TrimRight(safe, ". ")
	if safe == "" {
		safe = "_"
	}
	// Escape Windows reserved device names even when an extension is present.
	safe = escapeWindowsDeviceName(safe)
	return fmt.Sprintf("%x_%s", fnv32(rawURL), safe)
}

func escapeWindowsDeviceName(name string) string {
	base := name
	if i := strings.IndexRune(base, '.'); i >= 0 {
		base = base[:i]
	}
	switch strings.ToUpper(base) {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return "_" + name
	default:
		return name
	}
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
// the destination. If destPath already exists its permissions are preserved;
// otherwise 0644 is used.
//
// The request host is inferred from rawURL and used to build a host-aware
// client so redirects to other hosts strip GitHub-specific auth headers.
func DownloadFile(ctx context.Context, client *http.Client, rawURL, destPath string) error {
	ghHost := ""
	if ghHost == "" {
		u, err := url.Parse(rawURL)
		if err == nil {
			ghHost = u.Hostname()
		}
	}
	if ghHost == "" {
		return downloadFile(ctx, client, rawURL, destPath)
	}
	return downloadFile(ctx, httputil.NewHostAwareClient(client, ghHost), rawURL, destPath)
}

func downloadFile(ctx context.Context, client *http.Client, rawURL, destPath string) error {
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
		// Drain a bounded amount of the response body so the HTTP transport
		// can reuse the connection when possible.
		const maxErrorBodyDrain int64 = 4 << 10
		_, _ = io.CopyN(io.Discard, resp.Body, maxErrorBodyDrain)
		return fmt.Errorf("unexpected http status %s for %s", resp.Status, rawURL)
	}

	n, err := WriteFileAtomicFrom(destPath, resp.Body, 0644)
	if err != nil {
		return err
	}

	logger.Debug("file saved", "dest", destPath, "bytes", n)
	return nil
}
