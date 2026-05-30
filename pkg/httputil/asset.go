// Package httputil provides utilities for HTTP metadata fetching.
package httputil

import (
	"context"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// githubExactHeaders lists the HTTP request headers that are specific to the
// GitHub API and must be stripped by exact (case-insensitive) name match when
// forwarding requests to third-party storage backends.
var githubExactHeaders = []string{
	"Authorization",
}

// githubHeaderPrefixes lists the HTTP request header name prefixes that are
// specific to the GitHub API. Any header whose name starts with one of these
// prefixes (case-insensitive) must not be forwarded to third-party storage
// backends (e.g. Azure Blob Storage on GHES, AWS S3 on github.com) during
// redirects.
var githubHeaderPrefixes = []string{
	"X-Github-",
	"X-Hub-",
}

// headerStrippingTransport wraps a base RoundTripper and removes GitHub
// API-specific headers before forwarding the request. This preserves the
// connection settings (proxy, TLS config, timeouts, custom dialer) of the
// base transport while preventing credential leakage to third-party hosts.
type headerStrippingTransport struct {
	base http.RoundTripper
}

type transportUnwrapper interface {
	Unwrap() http.RoundTripper
}

type rawTransportGetter interface {
	RawTransport() http.RoundTripper
}

func unwrapTransport(tr http.RoundTripper) http.RoundTripper {
	for {
		u, ok := tr.(transportUnwrapper)
		if !ok {
			return tr
		}
		next := u.Unwrap()
		if next == nil || next == tr {
			return tr
		}
		tr = next
	}
}

func crossHostBaseTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		return http.DefaultTransport
	}

	inner := unwrapTransport(base)
	if rtg, ok := inner.(rawTransportGetter); ok {
		if raw := rtg.RawTransport(); raw != nil {
			return raw
		}
	}

	return inner
}

func (t *headerStrippingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	for key := range r2.Header {
		if shouldStripHeader(key) {
			delete(r2.Header, key)
		}
	}
	return t.base.RoundTrip(r2)
}

// shouldStripHeader reports whether the header with the given name should be
// removed before forwarding a request to a non-GitHub host. Headers in
// githubExactHeaders are matched case-insensitively by full name; headers in
// githubHeaderPrefixes are matched case-insensitively by prefix.
func shouldStripHeader(key string) bool {
	for _, exact := range githubExactHeaders {
		if strings.EqualFold(key, exact) {
			return true
		}
	}
	keyLower := strings.ToLower(key)
	for _, prefix := range githubHeaderPrefixes {
		if strings.HasPrefix(keyLower, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

// crossHostTransport returns a RoundTripper that strips GitHub-specific request
// headers from base while keeping all connection settings intact (proxy, TLS,
// timeouts, custom dialer). Use this instead of http.DefaultTransport when
// forwarding redirected requests to non-GitHub hosts.
func crossHostTransport(base http.RoundTripper) http.RoundTripper {
	return &headerStrippingTransport{base: crossHostBaseTransport(base)}
}

// AssetMeta holds HTTP metadata for a single asset URL.
type AssetMeta struct {
	// Size is the content length in bytes, or -1 when unknown.
	Size int64
	// Filename is the original upload name extracted from the Content-Disposition
	// response header, or empty when not available.
	Filename string
	// ContentType is the MIME type from the Content-Type response header.
	ContentType string
	// ExtHint is a file extension inferred from the redirect URL path when
	// Content-Disposition is not available (e.g. ".mov" from an S3 path).
	ExtHint string
}

// cdCapture is an http.RoundTripper wrapper that extracts the original filename
// from GitHub asset redirect responses.
//
// GitHub encodes the filename in multiple ways depending on the storage backend:
//  1. A Content-Disposition header directly on the response.
//  2. A query parameter in the 302 Location URL encoding the desired
//     Content-Disposition for the storage backend to mirror:
//     - AWS S3 / GitHub.com CDN:  "response-content-disposition"
//     - Azure Blob Storage (GHES): "rscd"
type cdCapture struct {
	base        http.RoundTripper
	ghHost      string // use base transport for this host; other hosts use crossHostTransport(base) to strip GitHub-specific headers while preserving base transport settings
	Filename    string
	ExtHint     string // file extension extracted from redirect URL path (e.g. ".mov")
	ContentType string // captured from the final response
}

// cdParamNames lists the query parameter names used by various storage backends
// to encode the desired Content-Disposition of a pre-signed redirect URL.
var cdParamNames = []string{
	"response-content-disposition", // AWS S3, GitHub.com CDN
	"rscd",                         // Azure Blob Storage (used by GHES)
}

func (c *cdCapture) RoundTrip(req *http.Request) (*http.Response, error) {
	// For cross-host requests (e.g. media CDN or Azure Blob Storage on GHES),
	// strip GitHub API-specific headers while preserving the base transport's
	// connection settings (proxy, TLS config, timeouts, custom dialer).
	transport := c.base
	if c.ghHost != "" && req.URL.Hostname() != c.ghHost {
		transport = crossHostTransport(c.base)
	}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Capture Content-Type from every hop; the final value wins.
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		c.ContentType = ct
	}

	if c.Filename != "" {
		return resp, nil
	}

	// Check Content-Disposition on this response first.
	if name := FilenameFromContentDisposition(resp.Header.Get("Content-Disposition")); name != "" {
		logger.Debug("filename from Content-Disposition header", "url", req.URL.Host+req.URL.EscapedPath(), "filename", name)
		c.Filename = name
		return resp, nil
	}

	// For redirect responses, parse the Location URL for content-disposition
	// query parameters used by various cloud storage backends.
	if loc := resp.Header.Get("Location"); loc != "" {
		logger.Debug("redirect hop", "from", req.URL.Host+req.URL.EscapedPath(), "status", resp.StatusCode)
		if u, parseErr := url.Parse(loc); parseErr == nil {
			q := u.Query()
			for _, key := range cdParamNames {
				if cd := q.Get(key); cd != "" {
					if name := FilenameFromContentDisposition(cd); name != "" {
						logger.Debug("filename from redirect param", "key", key, "filename", name)
						c.Filename = name
						return resp, nil
					}
				}
			}
			// Fall back to extracting the extension from the redirect URL path.
			// e.g. S3 paths like /owner/123-<uuid>.mov expose the extension even
			// when no response-content-disposition param is present.
			if c.ExtHint == "" {
				if base := path.Base(u.Path); base != "." && base != "" {
					if ext := path.Ext(base); ext != "" && ext != base {
						logger.Debug("extension hint from redirect path", "ext", ext, "path", u.Path)
						c.ExtHint = ext
					}
				}
			}
		}
	}
	return resp, nil
}

// FetchAssetMeta performs an HTTP HEAD request and extracts the content length and
// original filename from the response.
//
// GitHub returns the upload filename via Content-Disposition, either directly on
// the response or encoded as a query parameter in the 302 Location URL. A custom
// RoundTripper captures data from every hop in the redirect chain.
//
// Some GHES media servers return 404 for HEAD but respond correctly to GET.
// The fallback sequence is:
//  1. HEAD request
//  2. GET with "Range: bytes=0-0" (minimal download)
//  3. Full GET without a Range header (for servers that reject partial requests)
//
// Size is set to -1 when Content-Length is absent or the request fails.
// Filename is empty when Content-Disposition is absent in all hops.
// ghHost is used to select the correct transport per hop: authenticated for the
// GitHub host, and a transport derived from the same base transport for
// CDN/other hosts so proxy, TLS, and dialer settings are preserved while
// GitHub-specific headers are stripped.
func FetchAssetMeta(ctx context.Context, client *http.Client, assetURL, ghHost string) AssetMeta {
	if client == nil {
		client = http.DefaultClient
	}
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	meta := fetchAssetMetaWithMethod(ctx, client, transport, ghHost, http.MethodHead, false, assetURL)

	// Some servers (e.g. GHES media backends) return 404 for HEAD requests.
	// Fall back to a range-GET first, then a full GET as last resort.
	if meta.Size == -1 && meta.Filename == "" {
		logger.Debug("HEAD returned no useful data, trying range GET")
		meta = fetchAssetMetaWithMethod(ctx, client, transport, ghHost, http.MethodGet, true, assetURL)
	}
	if meta.Size == -1 && meta.Filename == "" {
		logger.Debug("range GET returned no useful data, trying full GET")
		meta = fetchAssetMetaWithMethod(ctx, client, transport, ghHost, http.MethodGet, false, assetURL)
	}
	}

	return meta
}

func fetchAssetMetaWithMethod(ctx context.Context, client *http.Client, transport http.RoundTripper, ghHost, method string, withRange bool, assetURL string) AssetMeta {
	cdc := &cdCapture{base: transport, ghHost: ghHost}
	c := &http.Client{
		Transport:     cdc,
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
		Timeout:       client.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, method, assetURL, nil)
	if err != nil {
		return AssetMeta{Size: -1}
	}
	if method == http.MethodGet && withRange {
		req.Header.Set("Range", "bytes=0-0")
	}
	resp, err := c.Do(req)
	if err != nil {
		logger.Debug("request failed", "method", method, "url", assetURL, "error", err)
		return AssetMeta{Size: -1}
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Debug("failed to close response body", "url", assetURL, "error", closeErr)
		}
	}()
	// For Range GET drain a small portion so the connection can be reused.
	if method == http.MethodGet && withRange {
		if _, drainErr := io.CopyN(io.Discard, resp.Body, 512); drainErr != nil && drainErr != io.EOF {
			logger.Debug("failed to drain response body", "url", assetURL, "error", drainErr)
		}
	}

	logger.Debug("response",
		"method", method,
		"url", req.URL.Host+req.URL.EscapedPath(),
		"status", resp.StatusCode,
		"content-length", resp.ContentLength,
		"content-range", resp.Header.Get("Content-Range"),
		"content-type", resp.Header.Get("Content-Type"),
		"content-disposition", resp.Header.Get("Content-Disposition"),
		"captured-filename", cdc.Filename,
	)

	// 4xx/5xx with no useful data.
	if resp.StatusCode >= 400 && cdc.Filename == "" {
		cd := resp.Header.Get("Content-Disposition")
		if FilenameFromContentDisposition(cd) == "" {
			return AssetMeta{Size: -1}
		}
	}

	size := resp.ContentLength
	if size < 0 {
		size = -1
	}
	// For 206 Partial Content, Content-Length is the range size (e.g. 1 byte for bytes=0-0).
	// Extract the actual total size from Content-Range: bytes 0-0/TOTAL.
	if resp.StatusCode == http.StatusPartialContent {
		if total := ParseTotalFromContentRange(resp.Header.Get("Content-Range")); total >= 0 {
			size = total
		}
	}

	filename := cdc.Filename
	if filename == "" {
		filename = FilenameFromContentDisposition(resp.Header.Get("Content-Disposition"))
	}
	return AssetMeta{Size: size, Filename: filename, ContentType: cdc.ContentType, ExtHint: cdc.ExtHint}
}

// ExtFromContentType returns the preferred file extension for a given MIME type,
// e.g. "video/quicktime" → ".mov". Returns "" when the type is unknown.
func ExtFromContentType(contentType string) string {
	// Strip parameters (e.g. charset) before lookup.
	mediaType, _, _ := mime.ParseMediaType(contentType)
	// Prefer well-known mappings over the OS MIME database for portability.
	switch mediaType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "video/webm":
		return ".webm"
	case "video/x-msvideo":
		return ".avi"
	}
	// Fall back to the stdlib MIME database.
	exts, err := mime.ExtensionsByType(mediaType)
	if err != nil || len(exts) == 0 {
		return ""
	}
	return exts[0]
}

// AssetTypeFromContentType classifies a media type from a MIME Content-Type string.
// Returns "image", "video", or "other".
func AssetTypeFromContentType(contentType string) string {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	switch {
	case strings.HasPrefix(mediaType, "image/"):
		return "image"
	case strings.HasPrefix(mediaType, "video/"):
		return "video"
	}
	return "other"
}

// ParseTotalFromContentRange extracts the total byte count from a Content-Range header.
// For example "bytes 0-0/12345" returns 12345. Returns -1 when the value cannot be parsed.
func ParseTotalFromContentRange(cr string) int64 {
	// Format: bytes <start>-<end>/<total>  or  bytes */<total>
	if i := strings.LastIndex(cr, "/"); i >= 0 {
		total, err := strconv.ParseInt(strings.TrimSpace(cr[i+1:]), 10, 64)
		if err == nil && total >= 0 {
			return total
		}
	}
	return -1
}

// FilenameFromContentDisposition parses the filename from a Content-Disposition header value.
// It handles both the plain ASCII form (filename="foo.png") and the RFC 5987 extended
// form (filename*=UTF-8''foo%20bar.png).
func FilenameFromContentDisposition(header string) string {
	if header == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(header)
	if err != nil {
		return ""
	}
	// mime.ParseMediaType handles filename* (RFC 5987) automatically via the Go stdlib.
	if name, ok := params["filename"]; ok && name != "" {
		return name
	}
	return ""
}

// hostSwitchTransport is an http.RoundTripper that uses base for requests to
// ghHost and strips GitHub-specific headers for all other hosts while preserving
// the base transport's connection settings (proxy, TLS config, timeouts, custom
// dialer). This prevents credential leakage to CDN / storage backends
// (e.g. Azure Blob Storage on GHES) during redirects.
type hostSwitchTransport struct {
	base   http.RoundTripper
	ghHost string
}

func (t *hostSwitchTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.ghHost != "" && req.URL.Hostname() != t.ghHost {
		return crossHostTransport(t.base).RoundTrip(req)
	}
	return t.base.RoundTrip(req)
}

// NewHostAwareClient returns an *http.Client that uses the authenticated transport
// for requests to ghHost and, for all other hosts (e.g. CDN, Azure Blob Storage
// on GHES), derives a transport from the client's base transport while stripping
// GitHub-specific headers. This preserves the base transport's connection
// settings, such as proxy, TLS config, and custom dialer, when following
// redirects across host boundaries to third-party storage backends.
func NewHostAwareClient(client *http.Client, ghHost string) *http.Client {
	if client == nil {
		client = http.DefaultClient
	}
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	return &http.Client{
		Transport:     &hostSwitchTransport{base: base, ghHost: ghHost},
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
		Timeout:       client.Timeout,
	}
}
