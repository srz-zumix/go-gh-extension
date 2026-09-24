package copilotext

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

// VSCodeSession describes one VS Code Copilot Chat debug-logs session (one main.jsonl
// file under a workspaceStorage entry).
type VSCodeSession struct {
	ID            string
	Dir           string
	LogPath       string
	WorkspaceHash string
	// Folder is the absolute path decoded from the owning workspace's workspace.json
	// "folder" field. It is empty when workspace.json is missing, unreadable, or uses
	// the multi-root "workspace" key instead of "folder".
	Folder     string
	ModifiedAt time.Time
}

// vscodeStorageRoot returns the directory containing one subdirectory per VS Code
// workspace, each potentially holding a GitHub Copilot Chat debug-logs tree.
//
// Only the stable macOS Code (non-Insiders) location is supported.
func vscodeStorageRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(home, "Library", "Application Support", "Code", "User", "workspaceStorage"), nil
}

// workspaceJSON mirrors the fields read from a workspaceStorage entry's workspace.json.
type workspaceJSON struct {
	Folder    string `json:"folder"`
	Workspace string `json:"workspace"`
}

// readWorkspaceFolder returns the absolute folder path recorded in hashDir's
// workspace.json, or "" when the file is absent, unreadable, or does not record a single
// folder (e.g. a multi-root ".code-workspace" file, recorded under the "workspace" key).
func readWorkspaceFolder(hashDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(hashDir, "workspace.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read workspace.json: %w", err)
	}
	var w workspaceJSON
	if err := json.Unmarshal(data, &w); err != nil {
		return "", fmt.Errorf("failed to parse workspace.json: %w", err)
	}
	if w.Folder == "" {
		return "", nil
	}
	return folderURIToPath(w.Folder)
}

// folderURIToPath converts a "file://" URI, as recorded in workspace.json, to an absolute
// filesystem path. Remote workspace URIs (e.g. "vscode-remote://...", used for SSH/dev
// container workspaces) are not resolvable to a local path; folderURIToPath returns "",
// nil for these rather than an error, since they are an expected, unsupported case, not a
// data integrity problem. Some remote URIs also fail net/url's strict percent-encoding
// validation in the host component, so the scheme is checked before parsing.
func folderURIToPath(uri string) (string, error) {
	if !strings.HasPrefix(uri, "file://") {
		return "", nil
	}
	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("failed to parse folder URI %q: %w", uri, err)
	}
	return u.Path, nil
}

// ListVSCodeSessions returns every VS Code Copilot Chat debug-logs session found under
// root (a workspaceStorage directory). Workspaces or sessions that cannot be read are
// skipped with a warning rather than failing the whole scan.
func ListVSCodeSessions(root string) ([]VSCodeSession, error) {
	hashEntries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list workspace storage directory %q: %w", root, err)
	}

	var sessions []VSCodeSession
	for _, h := range hashEntries {
		if !h.IsDir() {
			continue
		}
		hashDir := filepath.Join(root, h.Name())

		folder, err := readWorkspaceFolder(hashDir)
		if err != nil {
			logger.Warn("skipping workspace with unreadable workspace.json", "dir", hashDir, "error", err)
		}

		debugLogsDir := filepath.Join(hashDir, "GitHub.copilot-chat", "debug-logs")
		sessionEntries, err := os.ReadDir(debugLogsDir)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Warn("skipping unreadable debug-logs directory", "dir", debugLogsDir, "error", err)
			}
			continue
		}

		for _, se := range sessionEntries {
			if !se.IsDir() {
				continue
			}
			sessionDir := filepath.Join(debugLogsDir, se.Name())
			logPath := filepath.Join(sessionDir, "main.jsonl")
			info, err := os.Stat(logPath)
			if err != nil {
				if !os.IsNotExist(err) {
					logger.Warn("skipping unreadable session log", "path", logPath, "error", err)
				}
				continue
			}
			sessions = append(sessions, VSCodeSession{
				ID:            se.Name(),
				Dir:           sessionDir,
				LogPath:       logPath,
				WorkspaceHash: h.Name(),
				Folder:        folder,
				ModifiedAt:    info.ModTime(),
			})
		}
	}
	return sessions, nil
}
