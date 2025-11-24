package ioutil

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
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
