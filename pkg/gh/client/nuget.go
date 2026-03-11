package client

// NuGet registry URL helpers and HTTP functions.
// URL construction uses the host derived from the GitHubClient's BaseURL.

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
)

const nugetDefaultHost = "github.com"

// repositoryElemRe matches the full <repository .../> or <repository ...>...</repository> element
// in a .nuspec file, including any child content between the tags.
// (?is) enables case-insensitive matching and makes '.' match newlines for multi-line elements.
var repositoryElemRe = regexp.MustCompile(`(?is)<repository\b[^>]*(?:/>|>[\s\S]*?</repository\s*>)`)

// NuGetRegistryBase returns the NuGet registry base URL for the given GitHub host and owner.
// For github.com, it returns "https://nuget.pkg.github.com/<owner>".
// For GitHub Enterprise Server, it returns "https://<host>/_registry/nuget/<owner>".
func NuGetRegistryBase(host, owner string) string {
	if host == "" || host == nugetDefaultHost {
		return fmt.Sprintf("https://nuget.pkg.github.com/%s", owner)
	}
	return fmt.Sprintf("https://%s/_registry/nuget/%s", host, owner)
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
func RewriteNuPkgRepository(src io.ReaderAt, size int64, dst io.Writer, repoURL string) error {
	r, err := zip.NewReader(src, size)
	if err != nil {
		return fmt.Errorf("failed to open nupkg: %w", err)
	}

	w := zip.NewWriter(dst)

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
				_ = rc.Close()
				return fmt.Errorf("failed to copy zip entry %s: %w", f.Name, err)
			}
			if err := rc.Close(); err != nil {
				return fmt.Errorf("failed to close zip entry %s: %w", f.Name, err)
			}
		}
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to finalize nupkg: %w", err)
	}
	return nil
}

// DownloadNuGetPackage downloads a .nupkg file from the GitHub NuGet registry and
// streams it to dst.
func (g *GitHubClient) DownloadNuGetPackage(ctx context.Context, owner, packageName, version string, dst io.Writer) (retErr error) {
	url := NuGetDownloadURL(g.Host(), owner, packageName, version)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := g.client.Client().Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if _, err := io.Copy(dst, resp.Body); err != nil {
		return fmt.Errorf("failed to write nupkg: %w", err)
	}
	return nil
}

// PushNuGetPackage pushes a .nupkg file to the GitHub NuGet registry, streaming
// the contents of r via a multipart request without buffering the full payload in memory.
func (g *GitHubClient) PushNuGetPackage(ctx context.Context, owner string, r io.Reader) (retErr error) {
	url := NuGetPushURL(g.Host(), owner)

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	go func() {
		part, err := writer.CreateFormFile("package", "package.nupkg")
		if err != nil {
			pw.CloseWithError(fmt.Errorf("failed to create multipart form: %w", err))
			return
		}
		if _, err := io.Copy(part, r); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to write package data: %w", err))
			return
		}
		if err := writer.Close(); err != nil {
			pw.CloseWithError(fmt.Errorf("failed to close multipart writer: %w", err))
			return
		}
		if err := pw.Close(); err != nil {
			pw.CloseWithError(err)
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, pr)
	if err != nil {
		pr.CloseWithError(err)
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := g.client.Client().Do(req)
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
