package httputil

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type captureRoundTripper struct {
	lastAuthorization string
	lastXGitHubFoo    string
}

func (c *captureRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	c.lastAuthorization = r.Header.Get("Authorization")
	c.lastXGitHubFoo = r.Header.Get("X-GitHub-Foo")
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    r,
	}, nil
}

type authInjectingRoundTripper struct {
	raw   http.RoundTripper
	token string
}

func (t *authInjectingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	if t.token != "" {
		r2.Header.Set("Authorization", "token "+t.token)
	}
	return t.raw.RoundTrip(r2)
}

func (t *authInjectingRoundTripper) RawTransport() http.RoundTripper {
	return t.raw
}

type wrapperRoundTripper struct {
	inner http.RoundTripper
}

func (w *wrapperRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return w.inner.RoundTrip(r)
}

func (w *wrapperRoundTripper) Unwrap() http.RoundTripper {
	return w.inner
}

func TestCrossHostTransport_StripsGitHubHeaders(t *testing.T) {
	capture := &captureRoundTripper{}
	transport := crossHostTransport(capture)

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "token test")
	req.Header.Set("X-GitHub-Foo", "bar")

	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip returned error: %v", err)
	}

	if capture.lastAuthorization != "" {
		t.Fatalf("Authorization header leaked: %q", capture.lastAuthorization)
	}
	if capture.lastXGitHubFoo != "" {
		t.Fatalf("X-GitHub-* header leaked: %q", capture.lastXGitHubFoo)
	}
}

func TestCrossHostTransport_UsesRawTransportWhenAvailable(t *testing.T) {
	capture := &captureRoundTripper{}
	auth := &authInjectingRoundTripper{raw: capture, token: "secret"}
	transport := crossHostTransport(auth)

	req, err := http.NewRequest(http.MethodGet, "https://objects.example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip returned error: %v", err)
	}

	if capture.lastAuthorization != "" {
		t.Fatalf("Authorization header leaked from auth wrapper: %q", capture.lastAuthorization)
	}
}

func TestCrossHostTransport_UsesRawTransportThroughUnwrap(t *testing.T) {
	capture := &captureRoundTripper{}
	auth := &authInjectingRoundTripper{raw: capture, token: "secret"}
	wrapped := &wrapperRoundTripper{inner: auth}
	transport := crossHostTransport(wrapped)

	req, err := http.NewRequest(http.MethodGet, "https://objects.example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip returned error: %v", err)
	}

	if capture.lastAuthorization != "" {
		t.Fatalf("Authorization header leaked from wrapped auth transport: %q", capture.lastAuthorization)
	}
}
