package gh

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestUserAttachmentSupported(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		size         int64
		wantOK       bool
		wantMIMEType string
	}{
		{"png under limit", "shot.png", 1024, true, "image/png"},
		{"jpeg alias", "shot.jpeg", 1024, true, "image/jpeg"},
		{"png at limit", "shot.png", MaxUserAttachmentImageBytes, true, "image/png"},
		{"png over limit", "shot.png", MaxUserAttachmentImageBytes + 1, false, ""},
		{"mp4 under limit", "clip.mp4", MaxUserAttachmentImageBytes + 1, true, "video/mp4"},
		{"mp4 over limit", "clip.mp4", MaxUserAttachmentVideoBytes + 1, false, ""},
		{"mov supported", "clip.mov", 1024, true, "video/quicktime"},
		{"unknown size skips size check", "shot.png", -1, true, "image/png"},
		{"unsupported extension bmp", "shot.bmp", 1024, false, ""},
		{"unsupported extension pdf", "doc.pdf", 1024, false, ""},
		{"case insensitive extension", "SHOT.PNG", 1024, true, "image/png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentType, ok := UserAttachmentSupported(tt.filename, tt.size)
			if ok != tt.wantOK {
				t.Fatalf("UserAttachmentSupported(%q, %d) ok = %v, want %v", tt.filename, tt.size, ok, tt.wantOK)
			}
			if ok && contentType != tt.wantMIMEType {
				t.Fatalf("UserAttachmentSupported(%q, %d) contentType = %q, want %q", tt.filename, tt.size, contentType, tt.wantMIMEType)
			}
		})
	}
}

func TestCheckUserAttachmentUploadSupported(t *testing.T) {
	if err := CheckUserAttachmentUploadSupported("github.com", 1234); err != nil {
		t.Fatalf("CheckUserAttachmentUploadSupported() error = %v, want nil", err)
	}
	if err := CheckUserAttachmentUploadSupported("ghe.example.com", 1234); err == nil {
		t.Fatal("CheckUserAttachmentUploadSupported() error = nil, want an error for a GHES host")
	}
	if err := CheckUserAttachmentUploadSupported("github.com", 0); err == nil {
		t.Fatal("CheckUserAttachmentUploadSupported() error = nil, want an error for a zero repository id")
	}
}

// redirectingClient returns an http.Client whose transport rewrites every
// request's host/scheme to target srv, so tests exercise the real
// "uploads.<host>" URL-building logic while hitting a local test server.
func redirectingClient(t *testing.T, srv *httptest.Server) *http.Client {
	t.Helper()
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	base := srv.Client()
	base.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		return http.DefaultTransport.RoundTrip(req)
	})
	return base
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testUpload(body string) UserAttachmentUpload {
	return UserAttachmentUpload{
		Host:         "github.com",
		RepositoryID: 1234,
		Name:         "shot.png",
		ContentType:  "image/png",
		Size:         int64(len(body)),
		Open:         func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(body)), nil },
	}
}

func TestUploadUserAttachment_Success(t *testing.T) {
	const body = "the bytes"
	var gotQuery url.Values
	var gotHeader http.Header
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		gotHeader = r.Header.Clone()
		b, _ := io.ReadAll(r.Body)
		gotBody = b
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://example.com/user-attachments/assets/1"}`))
	}))
	defer srv.Close()

	assetURL, err := UploadUserAttachment(context.Background(), redirectingClient(t, srv), testUpload(body))
	if err != nil {
		t.Fatalf("UploadUserAttachment() error = %v", err)
	}
	if assetURL != "https://example.com/user-attachments/assets/1" {
		t.Fatalf("UploadUserAttachment() = %q, want asset url", assetURL)
	}
	if got := gotQuery.Get("name"); got != "shot.png" {
		t.Errorf("query name = %q, want shot.png", got)
	}
	if got := gotQuery.Get("content_type"); got != "image/png" {
		t.Errorf("query content_type = %q, want image/png", got)
	}
	if got := gotQuery.Get("repository_id"); got != "1234" {
		t.Errorf("query repository_id = %q, want 1234", got)
	}
	if got := gotHeader.Get("Content-Type"); got != "application/octet-stream" {
		t.Errorf("Content-Type header = %q, want application/octet-stream", got)
	}
	if got := gotHeader.Get("Accept"); got != "application/vnd.github+json" {
		t.Errorf("Accept header = %q, want application/vnd.github+json", got)
	}
	if string(gotBody) != body {
		t.Errorf("request body = %q, want %q", gotBody, body)
	}
}

func TestUploadUserAttachment_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	_, err := UploadUserAttachment(context.Background(), redirectingClient(t, srv), testUpload("the bytes"))
	if err == nil {
		t.Fatal("UploadUserAttachment() error = nil, want an error")
	}
	if got := err.Error(); got != "could not upload asset: attaching files requires write access to the repository" {
		t.Fatalf("UploadUserAttachment() error = %q, want the write-access message", got)
	}
}

func TestUploadUserAttachment_ValidationFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"Validation Failed"}`))
	}))
	defer srv.Close()

	_, err := UploadUserAttachment(context.Background(), redirectingClient(t, srv), testUpload("the bytes"))
	if err == nil {
		t.Fatal("UploadUserAttachment() error = nil, want an error")
	}
	if !strings.Contains(err.Error(), "422") || !strings.Contains(err.Error(), "Validation Failed") {
		t.Fatalf("UploadUserAttachment() error = %q, want the status and server message", err.Error())
	}
}

func TestUploadUserAttachment_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"Too Many Requests"}`))
	}))
	defer srv.Close()

	_, err := UploadUserAttachment(context.Background(), redirectingClient(t, srv), testUpload("the bytes"))
	var rl *UserAttachmentRateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("UploadUserAttachment() error = %v, want *UserAttachmentRateLimitError", err)
	}
	if rl.Status != http.StatusTooManyRequests {
		t.Errorf("Status = %d, want 429", rl.Status)
	}
	if got := rl.Header.Get("Retry-After"); got != "120" {
		t.Errorf("Retry-After header = %q, want 120", got)
	}
}

func TestUploadUserAttachment_NoAssetURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	_, err := UploadUserAttachment(context.Background(), redirectingClient(t, srv), testUpload("the bytes"))
	if err == nil {
		t.Fatal("UploadUserAttachment() error = nil, want an error")
	}
}
