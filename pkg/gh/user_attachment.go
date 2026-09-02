package gh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/auth"
)

// userAttachmentContentTypes maps the file extensions GitHub's REST
// user-attachments upload endpoint accepts to the content type it expects.
var userAttachmentContentTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".mp4":  "video/mp4",
	".mov":  "video/quicktime",
	".webm": "video/webm",
}

var userAttachmentVideoExtensions = map[string]bool{".mp4": true, ".mov": true, ".webm": true}

const (
	// MaxUserAttachmentImageBytes and MaxUserAttachmentVideoBytes match the
	// limits GitHub's REST user-attachments upload endpoint enforces.
	MaxUserAttachmentImageBytes int64 = 10 * 1024 * 1024
	MaxUserAttachmentVideoBytes int64 = 100 * 1024 * 1024
)

// ErrUserAttachmentUnsupported marks a file the user-attachments upload
// endpoint cannot accept (unsupported extension or over its size limit).
var ErrUserAttachmentUnsupported = errors.New("file type or size not supported by the user-attachments upload endpoint")

// UserAttachmentSupported reports whether the user-attachments upload
// endpoint accepts a file by name and size, and returns the content type to
// send when it does. size < 0 skips the size check (used when the file size
// is not yet known).
func UserAttachmentSupported(filename string, size int64) (string, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	contentType, ok := userAttachmentContentTypes[ext]
	if !ok {
		return "", false
	}
	limit := MaxUserAttachmentImageBytes
	if userAttachmentVideoExtensions[ext] {
		limit = MaxUserAttachmentVideoBytes
	}
	if size >= 0 && size > limit {
		return "", false
	}
	return contentType, true
}

// CheckUserAttachmentUploadSupported reports whether host and repositoryID can
// be used with UploadUserAttachment. GitHub Enterprise Server does not serve
// the upload endpoint, and the endpoint requires the repository's numeric id.
func CheckUserAttachmentUploadSupported(host string, repositoryID int64) error {
	if auth.IsEnterprise(host) {
		return fmt.Errorf("uploading via the REST API is not supported on GitHub Enterprise Server (%s)", host)
	}
	if repositoryID <= 0 {
		return errors.New("could not determine the repository id required to upload assets")
	}
	return nil
}

// NewUserAttachmentHTTPClient returns a shallow copy of g's authenticated
// http.Client with timeout applied, so a caller can allow longer than the
// default for potentially large uploads without mutating the shared client.
func NewUserAttachmentHTTPClient(g *GitHubClient, timeout time.Duration) *http.Client {
	base := *g.GetClient().Client()
	base.Timeout = timeout
	return &base
}

// UserAttachmentUpload describes a single asset upload.
type UserAttachmentUpload struct {
	// Host is the GitHub host (e.g. "github.com"); the request is sent to
	// "uploads.<Host>".
	Host string
	// RepositoryID is the numeric REST id of the repository to attach to.
	RepositoryID int64
	// Name is the filename recorded for the asset.
	Name string
	// ContentType is the type the endpoint expects for Name's extension, as
	// returned by UserAttachmentSupported.
	ContentType string
	// Size is the number of bytes Open yields.
	Size int64
	// Open opens the asset bytes; it may be called again to replay the body on
	// a redirect.
	Open func() (io.ReadCloser, error)
}

// UserAttachmentRateLimitError reports a response that hit a GitHub rate limit
// and is worth retrying with backoff. Header carries the response headers so a
// caller can honor Retry-After / X-RateLimit-Reset.
type UserAttachmentRateLimitError struct {
	Status int
	Header http.Header
	Body   string
}

func (e *UserAttachmentRateLimitError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("upload endpoint rate limited (status %d): %s", e.Status, e.Body)
	}
	return fmt.Sprintf("upload endpoint rate limited (status %d)", e.Status)
}

// UploadUserAttachment posts one asset to GitHub's REST user-attachments
// upload endpoint (the same endpoint the gh CLI's --attach flag and GitHub's
// own web/mobile clients use) and returns the new asset URL. It makes a single
// attempt: a rate-limited response is returned as *UserAttachmentRateLimitError
// so the caller decides the retry policy.
func UploadUserAttachment(ctx context.Context, httpClient *http.Client, up UserAttachmentUpload) (string, error) {
	reqURL := &url.URL{
		Scheme: "https",
		Host:   "uploads." + up.Host,
		Path:   "/user-attachments/assets",
		RawQuery: url.Values{
			"name":          {up.Name},
			"content_type":  {up.ContentType},
			"repository_id": {strconv.FormatInt(up.RepositoryID, 10)},
		}.Encode(),
	}

	f, err := up.Open()
	if err != nil {
		return "", err
	}
	defer f.Close() //nolint:errcheck

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), f)
	if err != nil {
		return "", fmt.Errorf("create upload request: %w", err)
	}
	req.ContentLength = up.Size
	req.GetBody = func() (io.ReadCloser, error) { return up.Open() }
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if isUserAttachmentRateLimitStatus(resp.StatusCode) {
			return "", &UserAttachmentRateLimitError{
				Status: resp.StatusCode,
				Header: resp.Header,
				Body:   strings.TrimSpace(string(body)),
			}
		}
		return "", newUserAttachmentUploadStatusError(resp.StatusCode, string(body))
	}

	var asset struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&asset); err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}
	if asset.URL == "" {
		return "", errors.New("upload response had no asset url")
	}
	return asset.URL, nil
}

// newUserAttachmentUploadStatusError formats an upload failure by status
// code: the endpoint answers 404 rather than 403 when the token cannot write,
// so the status code alone would misname the problem.
func newUserAttachmentUploadStatusError(status int, body string) error {
	switch status {
	case http.StatusNotFound:
		return errors.New("could not upload asset: attaching files requires write access to the repository")
	default:
		body = strings.TrimSpace(body)
		if body == "" {
			return fmt.Errorf("upload asset returned status %d", status)
		}
		return fmt.Errorf("upload asset returned status %d: %s", status, body)
	}
}

// isUserAttachmentRateLimitStatus reports whether an HTTP status indicates a
// GitHub rate limit worth retrying with backoff.
func isUserAttachmentRateLimitStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusForbidden
}
