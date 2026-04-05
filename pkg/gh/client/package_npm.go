package client

// npm registry URL helpers and HTTP functions.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const npmDefaultHost = DefaultHost

// NPMRegistryBase returns the npm registry base URL for the given GitHub host.
// For github.com, it returns "https://npm.pkg.github.com/".
// For GitHub Enterprise Server with subdomain isolation enabled, it returns "https://npm.<host>/".
func NPMRegistryBase(host string) string {
	if host == "" || host == npmDefaultHost {
		return "https://npm.pkg.github.com/"
	}
	return fmt.Sprintf("https://npm.%s/", host)
}

// NPMPackageURL returns the URL to download or access an npm package.
// If packageName is not already scoped (e.g., @org/pkg) and owner is non-empty,
// it is automatically scoped as @owner/packageName as required by the GitHub npm registry.
func NPMPackageURL(host, owner, packageName string) string {
	base := strings.TrimRight(NPMRegistryBase(host), "/")
	if owner != "" && !strings.HasPrefix(packageName, "@") {
		packageName = "@" + owner + "/" + packageName
	}
	return base + "/" + packageName
}

// rewritePackageJSON updates the "repository" field in package.json to point to the new repo URL.
func rewritePackageJSON(content []byte, repoURL string) ([]byte, error) {
	var pkg map[string]any
	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	// Update repository field
	if repoURL != "" {
		pkg["repository"] = map[string]any{
			"type": "git",
			"url":  repoURL,
		}
	}

	// Marshal back to JSON with proper indentation
	rewritten, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package.json: %w", err)
	}
	return rewritten, nil
}

// npmBearerAuthTransport adds an Authorization: Bearer <token> header to requests
// targeting the npm registry host. Requests to other hosts (e.g., pre-signed S3/CDN
// URLs in the dist.tarball field) are forwarded without the Authorization header,
// because S3 pre-signed URLs embed auth in query parameters and reject extra headers.
// Both github.com (npm.pkg.github.com) and GHES (npm.<host>) npm registries require
// Bearer format, unlike the GitHub REST API which uses "token X" format.
type npmBearerAuthTransport struct {
	token            string
	registryHostname string
	registryPort     string
	transport        http.RoundTripper
}

func (t *npmBearerAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	if t.token != "" && r2.URL.Hostname() == t.registryHostname && r2.URL.Port() == t.registryPort {
		r2.Header.Set("Authorization", "Bearer "+t.token)
	}
	tr := t.transport
	if tr == nil {
		tr = http.DefaultTransport
	}
	return tr.RoundTrip(r2)
}

// npmHTTPClient returns an http.Client configured for the GitHub npm registry.
// Bearer auth is only added for requests to the registry host; other hosts
// (e.g., pre-signed S3/CDN tarball URLs) receive no Authorization header.
func (g *GitHubClient) npmHTTPClient() *http.Client {
	registryBase := NPMRegistryBase(g.Host())
	registryBase = strings.TrimRight(registryBase, "/")
	var registryHostname, registryPort string
	if u, err := url.Parse(registryBase); err == nil {
		registryHostname = u.Hostname()
		registryPort = u.Port()
	}
	return &http.Client{
		Transport: &npmBearerAuthTransport{
			token:            g.bearerToken(),
			registryHostname: registryHostname,
			registryPort:     registryPort,
			transport:        g.rawHTTPTransport(),
		},
	}
}

// npmMetadataHTTPClient returns an http.Client for metadata requests only.
// Redirects are disabled so authentication failures surface as clear 3xx errors
// rather than silently following through to a login page.
func (g *GitHubClient) npmMetadataHTTPClient() *http.Client {
	c := g.npmHTTPClient()
	c.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return c
}

// npmVersionDist holds the dist information for a specific package version.
type npmVersionDist struct {
	Tarball string `json:"tarball"`
}

// npmVersionMeta holds the metadata for a specific package version.
type npmVersionMeta struct {
	Dist npmVersionDist `json:"dist"`
}

// npmMetadata holds the package version metadata information from the npm registry.
// Compatible with both application/json and application/vnd.npm.install-v1+json responses.
type npmMetadata struct {
	Versions map[string]npmVersionMeta `json:"versions"`
}

// DownloadNPMPackage downloads an npm package tarball from the registry.
// Returns the tarball data as bytes.
// Fetches package metadata first to obtain the canonical tarball URL (dist.tarball),
// then downloads the tarball with Bearer authentication as required by the GitHub npm registry.
func (g *GitHubClient) DownloadNPMPackage(ctx context.Context, owner, packageName, version string) ([]byte, error) {
	// Step 1: Fetch package metadata to get the tarball URL.
	// Use a no-redirect client so auth failures (3xx→login page) are caught immediately.
	metaClient := g.npmMetadataHTTPClient()
	metaURL := NPMPackageURL(g.Host(), owner, packageName)
	metaReq, err := http.NewRequestWithContext(ctx, http.MethodGet, metaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata request: %w", err)
	}
	metaReq.Header.Set("Accept", "application/vnd.npm.install-v1+json; q=1.0, application/json; q=0.8, */*")

	metaResp, err := metaClient.Do(metaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get npm package metadata: %w", err)
	}
	defer func() {
		if closeErr := metaResp.Body.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	switch {
	case metaResp.StatusCode == http.StatusMovedPermanently ||
		metaResp.StatusCode == http.StatusFound ||
		metaResp.StatusCode == http.StatusSeeOther ||
		metaResp.StatusCode == http.StatusTemporaryRedirect ||
		metaResp.StatusCode == http.StatusPermanentRedirect:
		// Authentication failed: the registry redirected to a login page.
		// This typically means the token is missing, invalid, or lacks read:packages scope.
		return nil, fmt.Errorf("failed to get npm package metadata: authentication required (registry redirected to %s)", metaResp.Header.Get("Location"))
	case metaResp.StatusCode != http.StatusOK:
		body, _ := io.ReadAll(metaResp.Body)
		return nil, fmt.Errorf("failed to get npm package metadata: %d - %s", metaResp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Detect non-JSON responses as a fallback (e.g., HTML from intermediate proxies).
	// Accept any media type containing "json" to cover standard application/json,
	// vendor types such as application/vnd.npm.install-v1+json, and +json suffixed types.
	contentType := metaResp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "json") {
		return nil, fmt.Errorf("failed to get npm package metadata: unexpected content type '%s' - check that the token has read:packages scope and the package '%s' exists under owner '%s'", contentType, packageName, owner)
	}

	var meta npmMetadata
	if err := json.NewDecoder(metaResp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("failed to parse npm package metadata: %w", err)
	}

	versionMeta, ok := meta.Versions[version]
	if !ok {
		return nil, fmt.Errorf("version '%s' not found in npm package metadata", version)
	}
	tarballURL := versionMeta.Dist.Tarball
	if tarballURL == "" {
		return nil, fmt.Errorf("tarball URL is empty for version '%s'", version)
	}

	// Step 2: Download the tarball from the URL provided in the metadata.
	// Use a redirect-following client; the tarball URL may redirect to S3 or CDN.
	tarReq, err := http.NewRequestWithContext(ctx, http.MethodGet, tarballURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create tarball request: %w", err)
	}

	tarResp, err := g.npmHTTPClient().Do(tarReq)
	if err != nil {
		return nil, fmt.Errorf("failed to download npm tarball: %w", err)
	}
	defer tarResp.Body.Close() // nolint

	if tarResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(tarResp.Body)
		return nil, fmt.Errorf("failed to download npm tarball: %d - %s", tarResp.StatusCode, string(body))
	}

	return io.ReadAll(tarResp.Body)
}

// PushNPMPackage publishes an npm package tarball to the registry.
func (g *GitHubClient) PushNPMPackage(ctx context.Context, owner, packageName string, tarballData []byte) (retErr error) {
	pkgURL := NPMPackageURL(g.Host(), owner, packageName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, pkgURL, bytes.NewReader(tarballData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := g.npmHTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to push npm package: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to push npm package: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// RewriteNPMPackageJSON rewrites the package.json file inside an npm tarball (.tgz).
// It reads the tarball from src, rewrites package.json to update the repository URL,
// and returns the modified tarball data.
func RewriteNPMPackageJSON(src io.Reader, repoURL string) ([]byte, error) {
	// Decompress gzip
	gzr, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress tarball: %w", err)
	}
	defer func() {
		if err := gzr.Close(); err != nil {
			_ = err // gzip reader close error after full read is not critical
		}
	}()

	// Read tar archive
	tr := tar.NewReader(gzr)

	var outBuf bytes.Buffer
	gzw := gzip.NewWriter(&outBuf)
	tw := tar.NewWriter(gzw)
	// NOTE: tw and gzw must NOT be deferred.
	// Both Close() calls flush trailing bytes (tar end-of-archive blocks and gzip trailer/checksum)
	// into outBuf. If deferred, they would run after "return outBuf.Bytes()", so the returned
	// slice would be missing those bytes and the archive would be corrupt.
	// On early-error returns the writers are abandoned, but since outBuf is in-memory there
	// is no OS resource leak; the GC will reclaim them.

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Check if this is package.json
		if strings.HasSuffix(hdr.Name, "package.json") {
			body, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read package.json: %w", err)
			}

			// Rewrite package.json
			rewritten, err := rewritePackageJSON(body, repoURL)
			if err != nil {
				return nil, err
			}

			// Update header with new size
			hdr.Size = int64(len(rewritten))
			if err := tw.WriteHeader(hdr); err != nil {
				return nil, fmt.Errorf("failed to write tar header: %w", err)
			}
			if _, err := tw.Write(rewritten); err != nil {
				return nil, fmt.Errorf("failed to write package.json to tar: %w", err)
			}
		} else {
			// Copy entry as-is
			if err := tw.WriteHeader(hdr); err != nil {
				return nil, fmt.Errorf("failed to write tar header: %w", err)
			}
			if _, err := io.Copy(tw, tr); err != nil {
				return nil, fmt.Errorf("failed to copy tar entry: %w", err)
			}
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize tar: %w", err)
	}

	if err := gzw.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize gzip: %w", err)
	}

	return outBuf.Bytes(), nil
}
