package client

// NuGet registry URL helpers and HTTP functions.
// URL construction uses the host derived from the GitHubClient's BaseURL.

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const nugetDefaultHost = DefaultHost

// repositoryElemRe matches the full <repository .../> or <repository ...>...</repository> element
// in a .nuspec file, including any child content between the tags.
// (?is) enables case-insensitive matching and makes '.' match newlines for multi-line elements.
var repositoryElemRe = regexp.MustCompile(`(?is)<repository\b[^>]*(?:/>|>[\s\S]*?</repository\s*>)`)

// NuGetRegistryBase returns the NuGet registry base URL for the given GitHub host and owner.
// For github.com, it returns "https://nuget.pkg.github.com/<owner>".
// For GitHub Enterprise Server, it returns "https://nuget.<host>/<owner>".
func NuGetRegistryBase(host, owner string) string {
	if host == "" || host == nugetDefaultHost {
		return fmt.Sprintf("https://nuget.pkg.github.com/%s", owner)
	}
	return fmt.Sprintf("https://nuget.%s/%s", host, owner)
}

// NuGetServiceIndexURL returns the NuGet V3 service index URL for the given host and owner.
func NuGetServiceIndexURL(host, owner string) string {
	return NuGetRegistryBase(host, owner) + "/index.json"
}

// NuGetDownloadURL returns the URL to download a .nupkg file from the GitHub NuGet registry.
// Package name is lowercased to comply with NuGet V3 API conventions.
func NuGetDownloadURL(host, owner, packageName, version string) string {
	base := NuGetRegistryBase(host, owner)
	lower := strings.ToLower(packageName)
	return fmt.Sprintf("%s/download/%s/%s/%s.%s.nupkg", base, lower, version, lower, version)
}

// NuGetPushURL returns the URL to push a .nupkg file to the GitHub NuGet registry.
func NuGetPushURL(host, owner string) string {
	return NuGetRegistryBase(host, owner)
}

// rewriteNuspecRepository updates or adds the <repository type="git" url="..." /> element
// in the given .nuspec XML bytes. GitHub Packages requires this element to associate
// the package with a repository on the same GitHub instance.
func rewriteNuspecRepository(nuspec []byte, repoURL string) []byte {
	newElem := fmt.Sprintf(`<repository type="git" url="%s" />`, repoURL)
	if repositoryElemRe.Match(nuspec) {
		return repositoryElemRe.ReplaceAll(nuspec, []byte(newElem))
	}
	// No existing <repository> element — insert before </metadata>
	return bytes.Replace(nuspec, []byte("</metadata>"), []byte("\t\t"+newElem+"\n\t</metadata>"), 1)
}

// RewriteNuPkgRepository rewrites the <repository> element in the .nuspec file inside
// a .nupkg (ZIP archive) to use the given repository URL, reading from src (size bytes)
// and writing the rewritten archive to dst.
// This is required when pushing to GitHub Packages, which mandates a repository
// association via the <repository url="..." /> element in the .nuspec.
func RewriteNuPkgRepository(src io.ReaderAt, size int64, dst io.Writer, repoURL string) (retErr error) {
	r, err := zip.NewReader(src, size)
	if err != nil {
		return fmt.Errorf("failed to open nupkg: %w", err)
	}

	w := zip.NewWriter(dst)
	defer func() {
		if err := w.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("failed to finalize nupkg: %w", err)
		}
	}()

	for _, f := range r.File {
		fhCopy := f.FileHeader
		fw, err := w.CreateHeader(&fhCopy)
		if err != nil {
			return fmt.Errorf("failed to create zip entry %s: %w", f.Name, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip entry %s: %w", f.Name, err)
		}

		if strings.HasSuffix(strings.ToLower(f.Name), ".nuspec") {
			content, readErr := io.ReadAll(rc)
			if closeErr := rc.Close(); closeErr != nil && readErr == nil {
				return fmt.Errorf("failed to close zip entry %s: %w", f.Name, closeErr)
			}
			if readErr != nil {
				return fmt.Errorf("failed to read zip entry %s: %w", f.Name, readErr)
			}
			content = rewriteNuspecRepository(content, repoURL)
			if _, err := fw.Write(content); err != nil {
				return fmt.Errorf("failed to write zip entry %s: %w", f.Name, err)
			}
		} else {
			if _, err := io.Copy(fw, rc); err != nil {
				if closeErr := rc.Close(); closeErr != nil {
					return fmt.Errorf("failed to copy zip entry %s: %w", f.Name, errors.Join(err, closeErr))
				}
				return fmt.Errorf("failed to copy zip entry %s: %w", f.Name, err)
			}
			if err := rc.Close(); err != nil {
				return fmt.Errorf("failed to close zip entry %s: %w", f.Name, err)
			}
		}
	}

	return nil
}

// DownloadNuGetPackage downloads a .nupkg file from the GitHub NuGet registry and
// streams it to dst.
// GitHub Packages may redirect the download to external storage (e.g. Azure Blob Storage).
// To avoid forwarding the GitHub Authorization header to a different host, redirects are
// followed manually using a plain HTTP client without auth headers.
func (g *GitHubClient) DownloadNuGetPackage(ctx context.Context, owner, packageName, version string, dst io.Writer) (retErr error) {
	body, err := g.fetchNuGetPackageBody(ctx, owner, packageName, version)
	if err != nil {
		return err
	}
	defer func() {
		if err := body.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("failed to close nupkg body: %w", err)
		}
	}()
	if _, err := io.Copy(dst, body); err != nil {
		return fmt.Errorf("failed to write nupkg: %w", err)
	}
	return nil
}

// fetchNuGetPackageBody resolves the final download URL and returns the response body.
//
// The correct download URL is discovered via the NuGet V3 service index (index.json)
// so the code works regardless of how a GHES instance configures its PackageBaseAddress.
// Falls back to the hardcoded URL pattern if discovery fails.
//
// Redirect handling is done manually hop-by-hop:
//   - GitHub host or its subdomains: follow with Basic auth (username:PAT).
//   - Any other host (e.g. Azure Blob Storage): follow without auth.
func (g *GitHubClient) fetchNuGetPackageBody(ctx context.Context, owner, packageName, version string) (io.ReadCloser, error) {
	githubHost := g.Host()
	token := g.bearerToken()

	// nugetAuthClient uses the raw transport so we can set Basic auth manually.
	// The GitHub NuGet registry (github.com and GHES) accepts any non-empty username
	// with a PAT as the password; "x-token-auth" is the conventional placeholder.
	nugetAuthClient := &http.Client{
		Transport: g.rawHTTPTransport(),
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	// plainClient is used for non-GitHub hosts (e.g. Azure Blob Storage) without auth.
	plainClient := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Discover the correct download URL via the NuGet V3 service index.
	downloadURL := g.nugetResolveDownloadURL(ctx, owner, packageName, version)
	slog.Debug("NuGet download", "url", downloadURL, "host", githubHost)

	currentURL := downloadURL
	const maxHops = 10
	for hop := range maxHops {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, currentURL, nil)
		if err != nil {
			return nil, err
		}

		isGitHub := nugetIsGitHubHost(currentURL, githubHost)
		slog.Debug("NuGet download hop", "hop", hop, "url", currentURL, "github_host", isGitHub)

		var resp *http.Response
		if isGitHub {
			if token != "" {
				req.SetBasicAuth("x-token-auth", token)
			}
			resp, err = nugetAuthClient.Do(req)
		} else {
			resp, err = plainClient.Do(req)
		}
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			slog.Debug("NuGet download complete", "hops", hop)
			return resp.Body, nil
		case http.StatusFound, http.StatusMovedPermanently, http.StatusSeeOther,
			http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
			if err := resp.Body.Close(); err != nil {
				return nil, fmt.Errorf("failed to close redirect response body: %w", err)
			}
			loc := resp.Header.Get("Location")
			if loc == "" {
				return nil, fmt.Errorf("redirect response missing Location header")
			}
			// A redirect to the login page means authentication failed.
			if nugetIsLoginRedirect(loc) {
				return nil, fmt.Errorf("authentication failed: server redirected to login page; ensure the gh auth token has 'read:packages' scope (run: gh auth refresh -h %s --scopes read:packages)", githubHost)
			}
			// Log redirect location without query parameters to avoid leaking credentials (e.g. SAS tokens).
			if u, err := url.Parse(loc); err == nil {
				u.RawQuery = ""
				u.Fragment = ""
				slog.Debug("NuGet download redirect", "location", u.String())
			} else {
				slog.Debug("NuGet download redirect", "location", "invalid redirect URL")
			}
			parsedLoc, err := url.Parse(loc)
			if err != nil {
				return nil, fmt.Errorf("invalid redirect Location header: %w", err)
			}
			resolvedURL := resp.Request.URL.ResolveReference(parsedLoc)
			currentURL = resolvedURL.String()
		default:
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			statusErr := fmt.Errorf("download failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			return nil, errors.Join(statusErr, resp.Body.Close())
		}
	}
	return nil, fmt.Errorf("too many redirects downloading NuGet package")
}

// nugetServiceIndex is the NuGet V3 service index response.
type nugetServiceIndex struct {
	Resources []struct {
		ID      string `json:"@id"`
		ResType string `json:"@type"`
	} `json:"resources"`
}

// nugetResolveDownloadURL fetches the NuGet V3 service index and returns the download URL
// for the given package. Falls back to the hardcoded URL pattern if discovery fails.
func (g *GitHubClient) nugetResolveDownloadURL(ctx context.Context, owner, packageName, version string) string {
	fallback := NuGetDownloadURL(g.Host(), owner, packageName, version)

	serviceIndexURL := NuGetServiceIndexURL(g.Host(), owner)
	slog.Debug("NuGet service index", "url", serviceIndexURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serviceIndexURL, nil)
	if err != nil {
		slog.Debug("NuGet service index: failed to create request", "error", err)
		return fallback
	}
	req.Header.Set("Accept", "*/*")
	if token := g.bearerToken(); token != "" {
		req.SetBasicAuth("x-token-auth", token)
	}

	// Do not follow redirects: a redirect to /login means auth failed; a redirect
	// to another URL should not be silently followed (it would return HTML).
	indexClient := &http.Client{
		Transport: g.rawHTTPTransport(),
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := indexClient.Do(req)
	if err != nil {
		slog.Debug("NuGet service index: request failed", "error", err)
		return fallback
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Debug("NuGet service index: failed to close response body", "error", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		// good
	case http.StatusFound, http.StatusMovedPermanently, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		loc := resp.Header.Get("Location")
		_ = resp.Body.Close()
		if nugetIsLoginRedirect(loc) {
			slog.Debug("NuGet service index: auth failed (login redirect)", "location", loc)
		} else {
			slog.Debug("NuGet service index: unexpected redirect", "status", resp.StatusCode, "location", loc)
		}
		return fallback
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		_ = resp.Body.Close()
		slog.Debug("NuGet service index: non-200 response", "status", resp.StatusCode, "body", strings.TrimSpace(string(body)))
		return fallback
	}

	var index nugetServiceIndex
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		slog.Debug("NuGet service index: failed to parse JSON", "error", err)
		return fallback
	}

	for _, r := range index.Resources {
		if r.ResType == "PackageBaseAddress/3.0.0" {
			base := strings.TrimRight(r.ID, "/")
			lower := strings.ToLower(packageName)
			discovered := fmt.Sprintf("%s/%s/%s/%s.%s.nupkg", base, lower, version, lower, version)
			slog.Debug("NuGet download URL from service index", "url", discovered, "base", r.ID)
			return discovered
		}
	}

	slog.Debug("NuGet service index: PackageBaseAddress not found, using fallback")
	return fallback
}

// nugetIsGitHubHost returns true if rawURL belongs to the GitHub instance
// identified by githubHost. Subdomains are also matched (e.g. nuget.pkg.github.com
// for github.com) to cover GitHub Packages registry subdomains.
func nugetIsGitHubHost(rawURL, githubHost string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	h := u.Hostname()
	return h == githubHost || strings.HasSuffix(h, "."+githubHost)
}

// nugetIsLoginRedirect returns true if the redirect URL is a GitHub login page,
// which indicates that authentication failed (e.g. insufficient token scopes).
func nugetIsLoginRedirect(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.HasPrefix(u.Path, "/login")
}

// nugetBasicAuthTransport wraps an existing RoundTripper and converts
// "Authorization: token X" or "Authorization: Bearer X" headers to
// Basic auth as required by the GitHub NuGet registry.
type nugetBasicAuthTransport struct {
	base http.RoundTripper
}

func (t *nugetBasicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	if auth := r2.Header.Get("Authorization"); auth != "" {
		var token string
		switch {
		case strings.HasPrefix(auth, "token "):
			token = strings.TrimPrefix(auth, "token ")
		case strings.HasPrefix(auth, "Bearer "):
			token = strings.TrimPrefix(auth, "Bearer ")
		}
		if token != "" {
			creds := base64.StdEncoding.EncodeToString([]byte("x-token:" + token))
			r2.Header.Set("Authorization", "Basic "+creds)
		}
	}
	return t.base.RoundTrip(r2)
}

// PushNuGetPackage pushes a .nupkg file to the GitHub NuGet registry, streaming
// the contents of r via a multipart request without buffering the full payload in memory.
// GitHub NuGet registry requires Basic auth (username + PAT); the go-github OAuth
// transport sends "Authorization: token <TOKEN>" which is rejected by nuget.pkg.github.com.
// nugetBasicAuthTransport converts the token auth to Basic auth transparently.
func (g *GitHubClient) PushNuGetPackage(ctx context.Context, owner string, r io.Reader) (retErr error) {
	url := NuGetPushURL(g.Host(), owner)

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, pr)
	if err != nil {
		// Close PipeWriter so that the writer goroutine cannot block
		_ = pw.CloseWithError(err)
		return err
	}

	go func() {
		part, err := writer.CreateFormFile("package", "package.nupkg")
		if err != nil {
			_ = pw.CloseWithError(fmt.Errorf("failed to create multipart form: %w", err))
			return
		}
		if _, err := io.Copy(part, r); err != nil {
			_ = pw.CloseWithError(fmt.Errorf("failed to write package data: %w", err))
			return
		}
		if err := writer.Close(); err != nil {
			_ = pw.CloseWithError(fmt.Errorf("failed to close multipart writer: %w", err))
			return
		}
		if err := pw.Close(); err != nil {
			_ = pw.CloseWithError(err)
		}
	}()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Wrap the existing transport to convert bearer/token auth to Basic auth
	// as required by the GitHub NuGet registry.
	httpClient := &http.Client{
		Transport: &nugetBasicAuthTransport{base: g.client.Client().Transport},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		// Close PipeWriter so that the writer goroutine cannot block if the HTTP request fails early
		_ = pw.CloseWithError(err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("push failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
