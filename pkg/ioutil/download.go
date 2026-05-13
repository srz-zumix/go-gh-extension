package ioutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

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

// SafeFilename returns a filesystem-safe name for a URL-keyed asset by prepending
// a FNV-1a hash of the URL to avoid collisions when multiple assets share the
// same base filename.
func SafeFilename(rawURL, filename string) string {
	return fmt.Sprintf("%x_%s", fnv32(rawURL), filename)
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

// DownloadFile performs an HTTP GET using client and saves the response body to destPath.
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
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, rawURL)
	}

	out, err := os.Create(destPath) //nolint:gosec
	if err != nil {
		return err
	}
	defer out.Close() //nolint:errcheck

	n, err := io.Copy(out, resp.Body)
	if err == nil {
		logger.Debug("file saved", "dest", destPath, "bytes", n)
	}
	return err
}
