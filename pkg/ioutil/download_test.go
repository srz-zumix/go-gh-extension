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

func TestDownloadFile_InferHostFromRawURL(t *testing.T) {
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
