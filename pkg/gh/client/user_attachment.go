package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// UserAttachmentUpload describes a single asset upload.
type UserAttachmentUpload struct {
	// Host is the GitHub host (e.g. "github.com"); the request is sent to
	// "uploads.<Host>".
	Host string
	// RepositoryID is the numeric REST id of the repository to attach to.
	RepositoryID int64
	// Name is the filename recorded for the asset.
	Name string
	// ContentType is the type the endpoint expects for Name's extension.
	ContentType string
	// Size is the number of bytes Open yields.
	Size int64
	// Open opens the asset bytes; it may be called again to replay the body on
	// a redirect.
	Open func() (io.ReadCloser, error)
	// Timeout bounds this upload. Zero keeps the client's configured timeout,
	// which is usually too short for a large asset.
	Timeout time.Duration
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
func (g *GitHubClient) UploadUserAttachment(ctx context.Context, up UserAttachmentUpload) (string, error) {
	if g == nil {
		return "", errors.New("github client must not be nil")
	}
	if up.Open == nil {
		return "", errors.New("upload body opener (Open) must not be nil")
	}
	if up.Host == "" {
		return "", errors.New("upload host must not be empty")
	}
	if up.Name == "" {
		return "", errors.New("upload name must not be empty")
	}
	if up.ContentType == "" {
		return "", errors.New("upload content type must not be empty")
	}
	if up.Size < 0 {
		return "", errors.New("upload size must not be negative")
	}
	if up.RepositoryID <= 0 {
		return "", errors.New("upload repository id must be positive")
	}

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
		return "", fmt.Errorf("open upload body: %w", err)
	}
	if f == nil {
		return "", errors.New("open upload body returned a nil reader")
	}
	defer func() { _ = f.Close() }()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), f)
	if err != nil {
		return "", fmt.Errorf("create upload request: %w", err)
	}
	req.ContentLength = up.Size
	req.GetBody = func() (io.ReadCloser, error) {
		r, err := up.Open()
		if err != nil {
			return nil, err
		}
		if r == nil {
			return nil, errors.New("open upload body returned a nil reader")
		}
		return r, nil
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/vnd.github+json")

	// Client() hands back a copy, so overriding Timeout here leaves the shared
	// client untouched while still reusing its authenticated transport.
	httpClient := g.GetClient().Client()
	if up.Timeout > 0 {
		httpClient.Timeout = up.Timeout
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if isUserAttachmentRateLimit(resp) {
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

// isUserAttachmentRateLimit reports whether resp indicates a GitHub rate limit
// worth retrying with backoff. 429 always qualifies. A 403 only qualifies when
// its headers identify a rate limit (Retry-After for a secondary limit, or
// X-RateLimit-Remaining: 0 for a primary limit); other 403s (SSO enforcement,
// insufficient scopes, policy blocks, etc.) are not retryable.
func isUserAttachmentRateLimit(resp *http.Response) bool {
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if resp.StatusCode == http.StatusForbidden {
		if resp.Header.Get("Retry-After") != "" {
			return true
		}
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return true
		}
	}
	return false
}
