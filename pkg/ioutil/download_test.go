package ioutil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeFilename_EscapeWindowsReservedDeviceNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantSuffix string
	}{
		{name: "con", input: "CON", wantSuffix: "__CON"},
		{name: "nul txt", input: "NUL.txt", wantSuffix: "__NUL.txt"},
		{name: "com1", input: "COM1", wantSuffix: "__COM1"},
		{name: "lpt9", input: "LPT9.log", wantSuffix: "__LPT9.log"},
		{name: "normal", input: "report.txt", wantSuffix: "_report.txt"},
	}

	rawURL := "https://example.com/path/image.png"
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := SafeFilename(rawURL, tt.input)
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Fatalf("SafeFilename() = %q, want suffix %q", got, tt.wantSuffix)
			}
		})
	}
}

func TestSafeFilename_UnsafeInputFallsBackToURL(t *testing.T) {
	t.Parallel()

	rawURL := "https://example.com/path/image.png"
	urlSuffix := "_image.png"

	unsafeInputs := []struct {
		name  string
		input string
	}{
		{name: "root slash", input: "/"},
		{name: "double dot", input: ".."},
		{name: "backslash root", input: `\`},
	}

	for _, tt := range unsafeInputs {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := SafeFilename(rawURL, tt.input)
			if !strings.HasSuffix(got, urlSuffix) {
				t.Fatalf("SafeFilename(%q) = %q, want suffix %q (should fall back to URL-derived name)", tt.input, got, urlSuffix)
			}
		})
	}
}

func TestSafeFilename_AlwaysFlatFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawURL   string
		filename string
	}{
		{
			name:     "backslash in filename",
			rawURL:   "https://example.com/image.png",
			filename: `sub\file.png`,
		},
		{
			name:     "slash in URL-derived fallback",
			rawURL:   "https://example.com/sub/image.png",
			filename: "/",
		},
		{
			name:     "backslash in malformed URL used as fallback",
			rawURL:   "https://example.com/a\\b\\file.png",
			filename: "..",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := SafeFilename(tt.rawURL, tt.filename)
			// Strip the leading hash prefix (e.g. "abc123_") before checking.
			if i := strings.IndexByte(got, '_'); i >= 0 {
				got = got[i+1:]
			}
			if strings.ContainsAny(got, "/\\") {
				t.Fatalf("SafeFilename() = %q contains a path separator", got)
			}
		})
	}
}

func TestGetFilename_StripsQueryAndFragment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{name: "simple path", url: "https://example.com/image.png", want: "image.png"},
		{name: "with query", url: "https://example.com/image.png?token=secret", want: "image.png"},
		{name: "with fragment", url: "https://example.com/image.png#section", want: "image.png"},
		{name: "with both", url: "https://example.com/image.png?token=secret#section", want: "image.png"},
		{name: "deep path with query", url: "https://example.com/path/to/file.txt?jwt=abc123", want: "file.txt"},
		{name: "trailing slash", url: "https://example.com/", want: ""},
		{name: "bare host", url: "https://example.com", want: ""},
		{name: "trailing slash with query", url: "https://example.com/?token=secret", want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := GetFilename(tt.url)
			if got != tt.want {
				t.Fatalf("GetFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

type authInjectingTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authInjectingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	if t.token != "" {
		r2.Header.Set("Authorization", "token "+t.token)
	}
	return t.base.RoundTrip(r2)
}

func (t *authInjectingTransport) RawTransport() http.RoundTripper {
	return t.base
}

func TestDownloadFile_StripsAuthOnCrossHostRedirect(t *testing.T) {
	t.Parallel()

	var leakedAuthorization string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leakedAuthorization = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, "payload")
	}))
	defer target.Close()
	targetURL, err := url.Parse(target.URL)
	if err != nil {
		t.Fatalf("failed to parse target URL: %v", err)
	}
	crossHostTargetURL := targetURL.String()
	crossHostTargetURL = strings.Replace(crossHostTargetURL, "127.0.0.1", "localhost", 1)

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, crossHostTargetURL+"/file.bin", http.StatusFound)
	}))
	defer origin.Close()

	client := &http.Client{
		Transport: &authInjectingTransport{
			base:  http.DefaultTransport,
			token: "secret",
		},
	}

	dest := filepath.Join(t.TempDir(), "download.bin")
	if err := DownloadFile(context.Background(), client, origin.URL+"/asset", dest); err != nil {
		t.Fatalf("DownloadFile returned error: %v", err)
	}

	if leakedAuthorization != "" {
		t.Fatalf("authorization header leaked to redirect target: %q", leakedAuthorization)
	}
}

func TestDownloadFile_PreservesAuthOnSameHost(t *testing.T) {
	t.Parallel()

	var gotAuthorization string
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, "payload")
	}))
	defer origin.Close()

	client := &http.Client{
		Transport: &authInjectingTransport{
			base:  http.DefaultTransport,
			token: "secret",
		},
	}

	dest := filepath.Join(t.TempDir(), "download.bin")
	if err := DownloadFile(context.Background(), client, origin.URL+"/asset", dest); err != nil {
		t.Fatalf("DownloadFile returned error: %v", err)
	}

	if gotAuthorization != "token secret" {
		t.Fatalf("authorization header = %q, want %q", gotAuthorization, "token secret")
	}
}
