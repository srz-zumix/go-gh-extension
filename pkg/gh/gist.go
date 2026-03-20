package gh

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/cli/v2/git"
	"github.com/google/go-github/v79/github"
)

// ListGists lists all gists for the authenticated user on the given client.
func ListGists(ctx context.Context, g *GitHubClient) ([]*github.Gist, error) {
	return g.ListGists(ctx, "")
}

// GetGist retrieves a gist by its ID.
func GetGist(ctx context.Context, g *GitHubClient, gistID string) (*github.Gist, error) {
	return g.GetGist(ctx, gistID)
}

// CreateGist creates a new gist with the specified parameters.
func CreateGist(ctx context.Context, g *GitHubClient, gist *github.Gist) (*github.Gist, error) {
	return g.CreateGist(ctx, gist)
}

// CopyGist copies the current file content of a gist from src to dst without preserving git history.
func CopyGist(ctx context.Context, src, dst *GitHubClient, gistID string) (*github.Gist, error) {
	gist, err := src.GetGist(ctx, gistID)
	if err != nil {
		return nil, err
	}

	newGist := &github.Gist{
		Description: gist.Description,
		Public:      gist.Public,
		Files:       make(map[github.GistFilename]github.GistFile),
	}
	for name, file := range gist.Files {
		newGist.Files[name] = github.GistFile{
			Filename: file.Filename,
			Content:  file.Content,
		}
	}

	return dst.CreateGist(ctx, newGist)
}

// MigrateGist migrates a gist from src to dst, preserving the full git history
// via git clone --mirror + git push --mirror.
// Tokens are obtained from the clients' bearer token.
func MigrateGist(ctx context.Context, src, dst *GitHubClient, gistID string) (*github.Gist, error) {
	// Fetch source gist metadata to get git URL and to set up destination.
	srcGist, err := src.GetGist(ctx, gistID)
	if err != nil {
		return nil, fmt.Errorf("get source gist: %w", err)
	}

	srcURL := src.GitURLWithToken(srcGist.GetGitPullURL())

	tmpDir, err := os.MkdirTemp("", "gh-my-kit-migrate-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	mirrorDir := filepath.Join(tmpDir, "repo.git")
	cloneClient := &git.Client{Stderr: os.Stderr, Stdout: os.Stderr}
	cloneCmd, err := cloneClient.Command(ctx, "clone", "--mirror", srcURL, mirrorDir)
	if err != nil {
		return nil, fmt.Errorf("prepare git clone --mirror: %w", err)
	}
	cloneCmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if err := cloneCmd.Run(); err != nil {
		return nil, fmt.Errorf("git clone --mirror: %w", err)
	}

	// Create destination gist after successful clone to avoid orphan gists on failure.
	// GitHub requires at least one file when creating a gist, so we use a
	// single dummy file. The content is irrelevant because git push --mirror
	// (which implies --force) will overwrite all refs with the source history.
	// The placeholder's initial commit becomes a dangling (unreachable) object
	// after the push and is invisible in normal git operations.
	placeholder := &github.Gist{
		Description: srcGist.Description,
		Public:      srcGist.Public,
		Files: map[github.GistFilename]github.GistFile{
			".gitkeep": {Content: github.Ptr("placeholder")},
		},
	}
	dstGist, err := dst.CreateGist(ctx, placeholder)
	if err != nil {
		return nil, fmt.Errorf("create destination gist: %w", err)
	}
	// If anything fails after this point, clean up the placeholder gist.
	var migrateErr error
	defer func() {
		if migrateErr != nil {
			_ = dst.DeleteGist(context.Background(), dstGist.GetID())
		}
	}()

	dstURL := dst.GitURLWithToken(dstGist.GetGitPushURL())

	pushClient := &git.Client{RepoDir: mirrorDir, Stderr: os.Stderr, Stdout: os.Stderr}
	pushCmd, err := pushClient.Command(ctx, "push", "--mirror", dstURL)
	if err != nil {
		migrateErr = fmt.Errorf("prepare git push --mirror: %w", err)
		return nil, migrateErr
	}
	pushCmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if err := pushCmd.Run(); err != nil {
		migrateErr = fmt.Errorf("git push --mirror: %w", err)
		return nil, migrateErr
	}

	// git push --mirror prunes refs on the destination that don't exist in the
	// source (e.g. the placeholder's 'main' branch). However, the remote HEAD
	// symref still points to 'main' which no longer exists, causing GitHub to
	// show "We couldn't find any files to show."
	// Fix: read the source's HEAD branch from the mirror and force-push it to
	// the destination's default branch name ('main') so the web UI can resolve it.
	symrefCmd, err := pushClient.Command(ctx, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		migrateErr = fmt.Errorf("prepare git symbolic-ref: %w", err)
		return nil, migrateErr
	}
	symrefOut, err := symrefCmd.Output()
	if err != nil {
		migrateErr = fmt.Errorf("git symbolic-ref: %w", err)
		return nil, migrateErr
	}
	headBranch := strings.TrimSpace(string(symrefOut))
	if headBranch != "main" {
		fixCmd, err := pushClient.Command(ctx, "push", "--force", dstURL, headBranch+":main")
		if err != nil {
			migrateErr = fmt.Errorf("prepare git push HEAD fix: %w", err)
			return nil, migrateErr
		}
		fixCmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		if err := fixCmd.Run(); err != nil {
			migrateErr = fmt.Errorf("git push HEAD fix: %w", err)
			return nil, migrateErr
		}
	}

	return dstGist, nil
}
