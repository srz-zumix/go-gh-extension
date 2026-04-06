package gh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/client"
)

// NuGetRegistryBase returns the NuGet registry base URL for the given GitHub host and owner.
// For github.com, it returns "https://nuget.pkg.github.com/<owner>".
// For GitHub Enterprise Server, it returns "https://nuget.<host>/<owner>".
func NuGetRegistryBase(host, owner string) string {
	return client.NuGetRegistryBase(host, owner)
}

// NuGetDownloadURL returns the URL to download a .nupkg file from the GitHub NuGet registry.
// Package name is lowercased to comply with NuGet V3 API conventions.
func NuGetDownloadURL(host, owner, packageName, version string) string {
	return client.NuGetDownloadURL(host, owner, packageName, version)
}

// NuGetPushURL returns the URL to push a .nupkg file to the GitHub NuGet registry.
func NuGetPushURL(host, owner string) string {
	return client.NuGetPushURL(host, owner)
}

// RewriteNuPkgRepository rewrites the <repository> element in the .nuspec inside the src
// .nupkg file to use the given repository URL.
// If destPath is empty, a new temporary file is created and returned.
// If destPath equals src.Name(), the file is rewritten in place (truncate + overwrite).
// Otherwise the result is written to a new file at destPath.
// The caller is responsible for closing the returned file. When destPath is empty, the caller is also responsible for removing the temporary file.
func RewriteNuPkgRepository(src *os.File, repoURL, destPath string) (_ *os.File, retErr error) {
	info, err := src.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat nupkg: %w", err)
	}

	// Always write to a temp file first to handle the same-file case safely.
	tmp, err := os.CreateTemp("", "nupkg-rewritten-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for rewrite: %w", err)
	}

	if err := client.RewriteNuPkgRepository(src, info.Size(), tmp, repoURL); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return nil, err
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return nil, fmt.Errorf("failed to seek rewritten nupkg: %w", err)
	}

	if destPath == "" {
		// Return the temp file directly; caller is responsible for cleanup.
		return tmp, nil
	}

	// Copy temp content to destPath, then close and remove temp regardless.
	defer func() {
		if err := tmp.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("failed to close temp nupkg: %w", err)
		}
		_ = os.Remove(tmp.Name())
	}()

	var dst *os.File
	if destPath == src.Name() {
		// In-place rewrite: truncate src and overwrite its content.
		if err := src.Truncate(0); err != nil {
			return nil, fmt.Errorf("failed to truncate %s: %w", destPath, err)
		}
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek %s: %w", destPath, err)
		}
		dst = src
	} else {
		var createErr error
		dst, createErr = os.Create(destPath)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create file %s for rewrite: %w", destPath, createErr)
		}
		defer func() {
			if retErr != nil {
				_ = dst.Close()
				_ = os.Remove(dst.Name())
			}
		}()
	}

	if _, err := io.Copy(dst, tmp); err != nil {
		return nil, fmt.Errorf("failed to write rewritten nupkg to %s: %w", destPath, err)
	}
	if _, err := dst.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek %s: %w", destPath, err)
	}
	return dst, nil
}

// DownloadNuGetPackage downloads a .nupkg file from the GitHub NuGet registry.
// If destPath is empty, a temporary file is created; otherwise the file at destPath is
// created (or truncated). The caller is responsible for closing the returned file and for
// removing it when destPath is empty (temporary file). For a non-empty destPath, the file
// is created at the specified location and is not treated as a temporary file by default.
func DownloadNuGetPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, packageName, version, destPath string) (_ *os.File, retErr error) {
	var (
		tmp *os.File
		err error
	)
	if destPath == "" {
		tmp, err = os.CreateTemp("", "nupkg-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file for download: %w", err)
		}
	} else {
		tmp, err = os.Create(destPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s for download: %w", destPath, err)
		}
	}
	defer func() {
		if retErr != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()
	if err := g.DownloadNuGetPackage(ctx, repo.Owner, packageName, version, tmp); err != nil {
		return nil, fmt.Errorf("failed to download NuGet package '%s' version '%s': %w", packageName, version, err)
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek downloaded nupkg: %w", err)
	}
	return tmp, nil
}

// PushNuGetPackage pushes a .nupkg from the given reader to the GitHub NuGet registry,
// streaming the payload without buffering the full content in memory.
func PushNuGetPackage(ctx context.Context, g *GitHubClient, repo repository.Repository, r io.Reader) error {
	if err := g.PushNuGetPackage(ctx, repo.Owner, r); err != nil {
		return fmt.Errorf("failed to push NuGet package: %w", err)
	}
	return nil
}

// IsNuGetConflictError returns true if err indicates that the NuGet package version already exists (HTTP 409 Conflict).
func IsNuGetConflictError(err error) bool {
	var pushErr *client.NuGetPushError
	return errors.As(err, &pushErr) && pushErr.IsConflict()
}
