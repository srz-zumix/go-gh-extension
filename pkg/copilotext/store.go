package copilotext

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/gitutil"
	"github.com/srz-zumix/go-gh-extension/pkg/ioutil"
)

// Scope selects where extensions are installed.
type Scope string

const (
	// ScopeUser installs into the Copilot CLI's per-user extensions directory.
	ScopeUser Scope = "user"
	// ScopeRepo installs into the current repository's .github/extensions directory.
	ScopeRepo Scope = "repo"
)

// metadataFileName is the hidden file recording how an extension directory was
// installed, so later commands can tell it apart from a directory the user manages by hand.
const metadataFileName = ".copilot-extension-install.json"

// metadata records how an installed extension directory was produced.
type metadata struct {
	Tool        string    `json:"tool"`
	ToolVersion string    `json:"tool_version"`
	Host        string    `json:"host"`
	Owner       string    `json:"owner"`
	Repo        string    `json:"repo"`
	Path        string    `json:"path"`
	Ref         string    `json:"ref"`
	CommitSHA   string    `json:"commit_sha"`
	InstalledAt time.Time `json:"installed_at"`
}

// extensionsRoot returns the directory under which extensions are installed for the given
// scope. prefix, when non-empty, overrides the scope and is returned as-is.
func extensionsRoot(ctx context.Context, scope Scope, prefix string) (string, error) {
	if prefix != "" {
		return prefix, nil
	}
	switch scope {
	case ScopeRepo:
		dir, err := gitutil.NewClient().ToplevelDir(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to determine repository root: %w", err)
		}
		return filepath.Join(dir, ".github", "extensions"), nil
	case ScopeUser, "":
		if home := os.Getenv("COPILOT_HOME"); home != "" {
			return filepath.Join(home, "extensions"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to determine home directory: %w", err)
		}
		return filepath.Join(home, ".copilot", "extensions"), nil
	default:
		return "", fmt.Errorf("unknown scope %q", scope)
	}
}

// destDir returns the installation directory for extension name under root.
func destDir(root, name string) string {
	return filepath.Join(root, name)
}

// metadataPath returns the path to dir's install metadata file.
func metadataPath(dir string) string {
	return filepath.Join(dir, metadataFileName)
}

// readMetadata reads dir's install metadata. It returns nil, nil (not an error) when dir
// has no metadata file or the file cannot be parsed, since both cases mean the directory
// is not managed by this command.
func readMetadata(dir string) (*metadata, error) {
	data, err := os.ReadFile(metadataPath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read extension metadata in %q: %w", dir, err)
	}
	var m metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, nil
	}
	return &m, nil
}

// writeMetadata writes m into dir's install metadata file.
func writeMetadata(dir string, m metadata) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode extension metadata: %w", err)
	}
	if err := ioutil.WriteFileAtomic(metadataPath(dir), data, 0644); err != nil {
		return fmt.Errorf("failed to write extension metadata in %q: %w", dir, err)
	}
	return nil
}
