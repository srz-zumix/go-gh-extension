package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// BuildAssetURLPatterns returns compiled regular expressions matching GitHub-hosted
// asset URLs for the given host.
//
// For github.com the patterns also cover the separate CDN hostnames
// (user-images.githubusercontent.com and private-user-images.githubusercontent.com).
// For GitHub Enterprise Server instances only the GHES host itself is used, because
// assets are served from the same hostname rather than a separate CDN.
func BuildAssetURLPatterns(host string) []*regexp.Regexp {
	escapedHost := regexp.QuoteMeta(host)
	patterns := []*regexp.Regexp{
		// https://<host>/user-attachments/assets/...
		regexp.MustCompile(`https://` + escapedHost + `/user-attachments/assets/[^\s)<>"]+`),
		// https://<host>/<owner>/<repo>/assets/...
		regexp.MustCompile(`https://` + escapedHost + `/[^/\s]+/[^/\s]+/assets/[^\s)<>"]+`),
	}
	// GitHub.com additionally serves images from a separate CDN hostname.
	if host == "github.com" {
		patterns = append(patterns,
			regexp.MustCompile(`https://user-images\.githubusercontent\.com/[^\s)<>"]+`),
			regexp.MustCompile(`https://private-user-images\.githubusercontent\.com/[^\s)<>"]+`),
		)
	}
	return patterns
}

// UploadAsset uploads a local file to the GitHub repository as an issue asset and
// returns the new CDN URL.
//
// It uses POST /repos/{owner}/{repo}/issues/assets (multipart/form-data).
// The response JSON is expected to contain { "url": "..." }.
func UploadAsset(ctx context.Context, httpClient *http.Client, host, owner, repo, localPath string) (string, error) {
	f, err := os.Open(localPath) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to open local file %q: %w", localPath, err)
	}
	defer f.Close() //nolint:errcheck

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(localPath))
	if err != nil {
		return "", fmt.Errorf("failed to create multipart field: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", fmt.Errorf("failed to write file to multipart: %w", err)
	}
	writer.Close() //nolint:errcheck

	apiURL := fmt.Sprintf("https://api.%s/repos/%s/%s/issues/assets", host, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, &body)
	if err != nil {
		return "", fmt.Errorf("failed to build upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse upload response: %w", err)
	}
	if result.URL == "" {
		return "", fmt.Errorf("upload response did not contain a URL")
	}
	return result.URL, nil
}
