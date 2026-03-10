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

// repositoryElemRe matches the <repository .../> element in a .nuspec file.
var repositoryElemRe = regexp.MustCompile(`(?i)<repository\b[^>]*/?>(?:\s*</repository>)?`)

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
// a .nupkg (ZIP archive) to use the given repository URL.
// This is required when pushing to GitHub Packages, which mandates a repository
// association via the <repository url="..." /> element in the .nuspec.
func RewriteNuPkgRepository(data []byte, repoURL string) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open nupkg: %w", err)
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	for _, f := range r.File {
		fhCopy := f.FileHeader
		fw, err := w.CreateHeader(&fhCopy)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry %s: %w", f.Name, err)
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open zip entry %s: %w", f.Name, err)
		}
		content, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			return nil, fmt.Errorf("failed to read zip entry %s: %w", f.Name, readErr)
		}

		if strings.HasSuffix(strings.ToLower(f.Name), ".nuspec") {
			content = rewriteNuspecRepository(content, repoURL)
		}

		if _, err := fw.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write zip entry %s: %w", f.Name, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize nupkg: %w", err)
	}
	return buf.Bytes(), nil
}

// DownloadNuGetPackage downloads a .nupkg file from the GitHub NuGet registry using the authenticated HTTP client.
func (g *GitHubClient) DownloadNuGetPackage(ctx context.Context, owner, packageName, version string) ([]byte, error) {
	url := NuGetDownloadURL(g.Host(), owner, packageName, version)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.Client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// PushNuGetPackage pushes a .nupkg file to the GitHub NuGet registry using the authenticated HTTP client.
func (g *GitHubClient) PushNuGetPackage(ctx context.Context, owner string, data []byte) error {
	url := NuGetPushURL(g.Host(), owner)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("package", "package.nupkg")
	if err != nil {
		return fmt.Errorf("failed to create multipart form: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("failed to write package data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := g.client.Client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
