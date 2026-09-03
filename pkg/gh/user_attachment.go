package gh

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
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

// UserAttachmentUpload describes a single asset upload.
type UserAttachmentUpload = client.UserAttachmentUpload

// UserAttachmentRateLimitError reports a response that hit a GitHub rate limit
// and is worth retrying with backoff.
type UserAttachmentRateLimitError = client.UserAttachmentRateLimitError

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

// UploadUserAttachment posts one asset to GitHub's REST user-attachments upload
// endpoint and returns the new asset URL. It rejects hosts/repository ids the
// endpoint cannot serve before delegating the request to the client layer.
func UploadUserAttachment(ctx context.Context, g *GitHubClient, up UserAttachmentUpload) (string, error) {
	if err := CheckUserAttachmentUploadSupported(up.Host, up.RepositoryID); err != nil {
		return "", err
	}
	return g.UploadUserAttachment(ctx, up)
}
