package client

// RubyGems registry URL helpers and HTTP functions.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

const rubygemsDefaultHost = DefaultHost

// RubyGemsRegistryBase returns the RubyGems registry base URL for the given GitHub host and owner.
// For github.com, it returns "https://rubygems.pkg.github.com/<owner>".
// For GitHub Enterprise Server with subdomain isolation enabled, it returns "https://rubygems.<host>/<owner>".
func RubyGemsRegistryBase(host, owner string) string {
	if host == "" || host == rubygemsDefaultHost {
		return fmt.Sprintf("https://rubygems.pkg.github.com/%s", owner)
	}
	return fmt.Sprintf("https://rubygems.%s/%s", host, owner)
}

// RubyGemsDownloadURL returns the URL to download a .gem file from the GitHub RubyGems registry.
func RubyGemsDownloadURL(host, owner, packageName, version string) string {
	base := RubyGemsRegistryBase(host, owner)
	return fmt.Sprintf("%s/gems/%s-%s.gem", base, packageName, version)
}

// RubyGemsPushURL returns the URL to push a .gem file to the GitHub RubyGems registry.
func RubyGemsPushURL(host, owner string) string {
	return RubyGemsRegistryBase(host, owner) + "/api/v1/gems"
}

// DownloadRubyGemsPackage downloads a .gem file from the GitHub RubyGems registry.
// Returns the gem content as bytes.
// GitHub Packages redirects the download to Azure Blob Storage; the auth token must
// not be forwarded to that third-party URL, so redirects are handled manually.
func (g *GitHubClient) DownloadRubyGemsPackage(ctx context.Context, owner, packageName, version string) ([]byte, error) {
	url := RubyGemsDownloadURL(g.Host(), owner, packageName, version)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	// Do not follow redirects so the GitHub auth token is not sent to Azure Blob Storage.
	noRedirect := g.basicAuthHTTPClient()
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := noRedirect.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download gem: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		defer resp.Body.Close() // nolint
		return io.ReadAll(resp.Body)
	case http.StatusFound, http.StatusMovedPermanently, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		if err := resp.Body.Close(); err != nil {
			return nil, fmt.Errorf("failed to close redirect response body: %w", err)
		}
		loc := resp.Header.Get("Location")
		if loc == "" {
			return nil, fmt.Errorf("redirect response missing Location header")
		}
		// Resolve Location relative to the original request URL to handle relative redirects.
		redirectURL, err := req.URL.Parse(loc)
		if err != nil {
			return nil, err
		}
		// Follow the redirect without auth headers, using a clone of the base client
		// (inherits Timeout, proxy settings, etc.) with the raw transport so credentials
		// are not forwarded to third-party storage (e.g. Azure Blob Storage).
		plainReq, err := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL.String(), nil)
		if err != nil {
			return nil, err
		}
		cdnBase := g.client.Client()
		cdnClone := *cdnBase
		cdnClone.Transport = g.rawHTTPTransport()
		cdnClone.CheckRedirect = nil
		plainResp, err := cdnClone.Do(plainReq)
		if err != nil {
			return nil, err
		}
		defer plainResp.Body.Close() // nolint
		if plainResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(plainResp.Body, 512))
			return nil, fmt.Errorf("download failed with status %d: %s", plainResp.StatusCode, strings.TrimSpace(string(body)))
		}
		return io.ReadAll(plainResp.Body)
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// PushRubyGemsPackage pushes a .gem file to the GitHub RubyGems registry.
// The gem data is sent as raw body with application/octet-stream content type.
func (g *GitHubClient) PushRubyGemsPackage(ctx context.Context, owner string, r io.Reader) (retErr error) {
	pushURL := RubyGemsPushURL(g.Host(), owner)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pushURL, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	httpClient := g.basicAuthHTTPClient()

	resp, err := httpClient.Do(req)
	if err != nil {
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

type gemEntry struct {
	name   string
	header *tar.Header
	data   []byte
}

// RewriteGemGitHubRepo rewrites the github_repo metadata field in a .gem archive
// to reference the destination host/owner/repo, and updates checksums.yaml.gz accordingly.
// This is required when migrating gems between GitHub instances because GitHub validates
// that the github_repo URL points to a repository on the target instance.
// If the gem has no github_repo field, it is returned unchanged.
func RewriteGemGitHubRepo(gemData []byte, host, owner, repo string) ([]byte, error) {
	entries, err := readGemEntries(gemData)
	if err != nil {
		return nil, err
	}

	// Rewrite metadata.gz
	rewritten := false
	for i, e := range entries {
		if e.name == "metadata.gz" {
			newData, err := rewriteMetadataGzGitHubRepo(e.data, host, owner, repo)
			if err != nil {
				return nil, err
			}
			entries[i].data = newData
			rewritten = true
			break
		}
	}

	if !rewritten {
		return gemData, nil
	}

	// Rebuild checksums.yaml.gz with updated hashes
	for i, e := range entries {
		if e.name == "checksums.yaml.gz" {
			newData, err := rebuildGemChecksums(entries)
			if err != nil {
				return nil, err
			}
			entries[i].data = newData
			break
		}
	}

	return writeGemEntries(entries)
}

func readGemEntries(gemData []byte) ([]gemEntry, error) {
	r := tar.NewReader(bytes.NewReader(gemData))
	var entries []gemEntry
	for {
		hdr, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read gem archive: %w", err)
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s from gem: %w", hdr.Name, err)
		}
		entries = append(entries, gemEntry{name: hdr.Name, header: hdr, data: data})
	}
	return entries, nil
}

func writeGemEntries(entries []gemEntry) ([]byte, error) {
	var out bytes.Buffer
	tw := tar.NewWriter(&out)
	for _, e := range entries {
		hdr := *e.header
		hdr.Size = int64(len(e.data))
		if err := tw.WriteHeader(&hdr); err != nil {
			return nil, fmt.Errorf("failed to write header for %s: %w", e.name, err)
		}
		if _, err := tw.Write(e.data); err != nil {
			return nil, fmt.Errorf("failed to write data for %s: %w", e.name, err)
		}
	}
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize gem archive: %w", err)
	}
	return out.Bytes(), nil
}

func rewriteMetadataGzGitHubRepo(data []byte, host, owner, repo string) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decompress metadata.gz: %w", err)
	}
	metaYAML, readErr := io.ReadAll(gr)
	if closeErr := gr.Close(); closeErr != nil && readErr == nil {
		readErr = closeErr
	}
	if readErr != nil {
		return nil, fmt.Errorf("failed to read metadata.gz: %w", readErr)
	}

	// Replace the github_repo SSH URL with the destination repository URL.
	newURL := fmt.Sprintf("ssh://%s/%s/%s", host, owner, repo)
	re := regexp.MustCompile(`(github_repo:\s+)ssh://\S+`)
	newYAML := re.ReplaceAll(metaYAML, []byte("${1}"+newURL))

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(newYAML); err != nil {
		return nil, fmt.Errorf("failed to compress rewritten metadata: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize metadata compression: %w", err)
	}
	return buf.Bytes(), nil
}

// rebuildGemChecksums regenerates checksums.yaml.gz for all gem entries except itself.
// SHA256 values are unquoted; SHA512 values are single-quoted to match Ruby Psych YAML output.
func rebuildGemChecksums(entries []gemEntry) ([]byte, error) {
	type fileSums struct {
		sha256 string
		sha512 string
	}
	sums := make(map[string]fileSums)
	for _, e := range entries {
		if e.name == "checksums.yaml.gz" {
			continue
		}
		s256 := sha256.Sum256(e.data)
		s512 := sha512.Sum512(e.data)
		sums[e.name] = fileSums{
			sha256: hex.EncodeToString(s256[:]),
			sha512: hex.EncodeToString(s512[:]),
		}
	}

	names := make([]string, 0, len(sums))
	for n := range sums {
		names = append(names, n)
	}
	sort.Strings(names)

	var buf bytes.Buffer
	buf.WriteString("---\nSHA256:\n")
	for _, n := range names {
		fmt.Fprintf(&buf, "  %s: %s\n", n, sums[n].sha256)
	}
	buf.WriteString("SHA512:\n")
	for _, n := range names {
		fmt.Fprintf(&buf, "  %s: '%s'\n", n, sums[n].sha512)
	}

	var out bytes.Buffer
	gw := gzip.NewWriter(&out)
	if _, err := gw.Write(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to compress checksums: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize checksums compression: %w", err)
	}
	return out.Bytes(), nil
}
