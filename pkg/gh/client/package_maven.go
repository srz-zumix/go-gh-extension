package client

// Maven registry URL helpers and HTTP functions.
// Maven packages on GitHub Packages require a repository context in the URL.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const mavenDefaultHost = DefaultHost

// MavenRegistryBase returns the Maven registry base URL for the given GitHub host, owner, and repository.
// For github.com, it returns "https://maven.pkg.github.com/<owner>/<repo>".
// For GitHub Enterprise Server with subdomain isolation enabled, it returns "https://maven.<host>/<owner>/<repo>".
func MavenRegistryBase(host, owner, repo string) string {
	if host == "" || host == mavenDefaultHost {
		return fmt.Sprintf("https://maven.pkg.github.com/%s/%s", owner, repo)
	}
	return fmt.Sprintf("https://maven.%s/%s/%s", host, owner, repo)
}

// ParseMavenPackageName splits a Maven package name into groupID and artifactID.
// Accepts colon-separated format ("com.example:my-artifact") as well as the
// GitHub Packages dot-separated format ("com.example.my-artifact"), where the
// last dot-delimited segment is treated as the artifactId.
// Returns an error if the package name is not in a recognized format.
func ParseMavenPackageName(packageName string) (groupID, artifactID string, err error) {
	if idx := strings.LastIndex(packageName, ":"); idx >= 0 {
		groupID = packageName[:idx]
		artifactID = packageName[idx+1:]
	} else if idx := strings.LastIndex(packageName, "."); idx >= 0 {
		// GitHub Packages stores Maven package names as groupId.artifactId (dots throughout).
		// Split at the last dot: everything before is groupId, last segment is artifactId.
		groupID = packageName[:idx]
		artifactID = packageName[idx+1:]
	} else {
		return "", "", fmt.Errorf("invalid Maven package name %q: expected format <groupId>:<artifactId> or <groupId>.<artifactId>", packageName)
	}
	if groupID == "" || artifactID == "" {
		return "", "", fmt.Errorf("invalid Maven package name %q: groupId and artifactId must not be empty", packageName)
	}
	return groupID, artifactID, nil
}

// MavenArtifactURL returns the URL for a Maven artifact file.
// classifier may be empty for the primary artifact.
func MavenArtifactURL(host, owner, repo, groupID, artifactID, version, classifier, ext string) string {
	base := MavenRegistryBase(host, owner, repo)
	groupPath := strings.ReplaceAll(groupID, ".", "/")
	var filename string
	if classifier != "" {
		filename = fmt.Sprintf("%s-%s-%s.%s", artifactID, version, classifier, ext)
	} else {
		filename = fmt.Sprintf("%s-%s.%s", artifactID, version, ext)
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s", base, groupPath, artifactID, version, filename)
}

// MavenDownloadError is returned when a Maven artifact download fails with an HTTP error status.
type MavenDownloadError struct {
	StatusCode int
	Message    string
}

func (e *MavenDownloadError) Error() string {
	return fmt.Sprintf("download failed with status %d: %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the download failed because the artifact does not exist (HTTP 404).
func (e *MavenDownloadError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// fetchMavenArtifactBody fetches an artifact file body, handling redirects without
// forwarding auth headers to third-party storage (e.g. Azure Blob Storage).
func (g *GitHubClient) fetchMavenArtifactBody(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	noRedirect := &http.Client{
		Transport: &basicAuthTransport{base: g.client.Client().Transport},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := noRedirect.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
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
		redirectURL, err := resp.Request.URL.Parse(loc)
		if err != nil {
			return nil, err
		}
		plainReq, err := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL.String(), nil)
		if err != nil {
			return nil, err
		}
		plainResp, err := http.DefaultClient.Do(plainReq)
		if err != nil {
			return nil, err
		}
		if plainResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(plainResp.Body, 512))
			statusErr := &MavenDownloadError{StatusCode: plainResp.StatusCode, Message: strings.TrimSpace(string(body))}
			return nil, errors.Join(statusErr, plainResp.Body.Close())
		}
		return plainResp.Body, nil
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		statusErr := &MavenDownloadError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(body))}
		return nil, errors.Join(statusErr, resp.Body.Close())
	}
}

// MavenArtifact holds the content of a single Maven artifact file.
type MavenArtifact struct {
	Classifier string
	Ext        string
	Data       []byte
}

// DownloadMavenArtifacts downloads the primary artifacts (.pom and .jar) for a given Maven
// package version. The .jar download is treated as optional; if it returns 404 the error is
// silently ignored (POM-only packages are valid in Maven).
func (g *GitHubClient) DownloadMavenArtifacts(ctx context.Context, owner, repo, packageName, version string) ([]MavenArtifact, error) {
	groupID, artifactID, err := ParseMavenPackageName(packageName)
	if err != nil {
		return nil, err
	}

	var artifacts []MavenArtifact

	for _, item := range []struct{ classifier, ext string }{
		{"", "jar"},
		{"", "pom"},
	} {
		url := MavenArtifactURL(g.Host(), owner, repo, groupID, artifactID, version, item.classifier, item.ext)
		body, err := g.fetchMavenArtifactBody(ctx, url)
		if err != nil {
			// .jar is optional — some Maven packages are POM-only.
			// Only silently skip on 404; surface all other errors (network, auth, etc.).
			if item.ext == "jar" {
				var dlErr *MavenDownloadError
				if errors.As(err, &dlErr) && dlErr.IsNotFound() {
					continue
				}
			}
			return nil, fmt.Errorf("failed to download %s %s: %w", packageName, item.ext, err)
		}
		data, readErr := io.ReadAll(body)
		if closeErr := body.Close(); closeErr != nil && readErr == nil {
			readErr = closeErr
		}
		if readErr != nil {
			return nil, fmt.Errorf("failed to read %s %s: %w", packageName, item.ext, readErr)
		}
		artifacts = append(artifacts, MavenArtifact{Classifier: item.classifier, Ext: item.ext, Data: data})
	}

	return artifacts, nil
}

// MavenPushError is returned when a Maven artifact push fails with an HTTP error status.
type MavenPushError struct {
	StatusCode int
	Message    string
}

func (e *MavenPushError) Error() string {
	return fmt.Sprintf("push failed with status %d: %s", e.StatusCode, e.Message)
}

// IsConflict returns true if the push failed because the artifact already exists (HTTP 409 Conflict).
func (e *MavenPushError) IsConflict() bool {
	return e.StatusCode == http.StatusConflict
}

// PushMavenArtifact pushes a single Maven artifact file to the registry.
// GitHub Packages Maven registry follows the standard Maven repository protocol:
// a plain HTTP PUT with the raw artifact bytes as the request body.
func (g *GitHubClient) PushMavenArtifact(ctx context.Context, owner, repo, packageName, version, classifier, ext string, data []byte) (retErr error) {
	groupID, artifactID, err := ParseMavenPackageName(packageName)
	if err != nil {
		return err
	}

	url := MavenArtifactURL(g.Host(), owner, repo, groupID, artifactID, version, classifier, ext)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
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
		return &MavenPushError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(body))}
	}
	return nil
}
