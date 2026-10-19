package copilotext

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/logger"
	"gopkg.in/yaml.v3"
)

// Log source kinds accepted by LogOptions.Sources.
const (
	// LogSourceCLISessions is the Copilot CLI session-state directory.
	LogSourceCLISessions = "cli-sessions"
	// LogSourceCLILogs is the Copilot CLI logs directory.
	LogSourceCLILogs = "cli-logs"
	// LogSourceVSCodeChat is VS Code's chatSessions directories.
	LogSourceVSCodeChat = "vscode-chat"
	// LogSourceVSCodeDebug is the Copilot Chat debug-logs, transcripts and
	// chat-session-resources directories of VS Code.
	LogSourceVSCodeDebug = "vscode-debug"
	// LogSourcePath marks files read from an explicit path.
	LogSourcePath = "path"
)

// LogSources lists the selectable default log locations.
var LogSources = []string{
	LogSourceCLISessions,
	LogSourceCLILogs,
	LogSourceVSCodeChat,
	LogSourceVSCodeDebug,
}

var sessionUUIDPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// LogOptions selects which Copilot log files WalkLogFiles visits.
type LogOptions struct {
	// Sources selects default locations; empty means all of LogSources.
	Sources []string
	// Paths replaces the default locations when non-empty.
	Paths []string
	// Since skips files last modified before it; zero disables the filter.
	Since time.Time
	// SessionPrefixes keeps only files whose session ID starts with one of
	// these (case-insensitive). Files without a session ID are skipped.
	SessionPrefixes []string
	// CopilotHome and VSCodeUserDir override the detected locations.
	CopilotHome   string
	VSCodeUserDir string
}

// LogFile is a Copilot log file found by WalkLogFiles.
type LogFile struct {
	// Source is one of the LogSource* kinds.
	Source string
	// Path is the absolute file path.
	Path string
	// SessionID is the lower-case session UUID taken from the path, or "".
	SessionID   string
	ModifiedAt  time.Time
	Workspace   string
	SessionName string
}

// CopilotHome returns the Copilot CLI's per-user home directory, honoring the
// COPILOT_HOME override.
func CopilotHome() (string, error) {
	return copilotHome()
}

// VSCodeUserDir returns the VS Code (stable) user data directory.
func VSCodeUserDir() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user config directory: %w", err)
	}
	return filepath.Join(cfg, "Code", "User"), nil
}

type logRoot struct {
	source string
	path   string
}

// WalkLogFiles calls fn for every regular file under the selected Copilot log
// locations. Missing default locations and unreadable entries are skipped with
// a log message; a missing explicit path is an error. Symlinks and other
// non-regular files are never visited. fn's error stops the walk and is
// returned unchanged.
func WalkLogFiles(opts LogOptions, fn func(LogFile) error) error {
	roots, err := logRoots(opts)
	if err != nil {
		return err
	}
	prefixes := make([]string, 0, len(opts.SessionPrefixes))
	for _, p := range opts.SessionPrefixes {
		prefixes = append(prefixes, strings.ToLower(p))
	}

	for _, root := range roots {
		names := make(map[string]string)
		logger.Debug("walking copilot logs", "source", root.source, "path", root.path)
		walkErr := filepath.WalkDir(root.path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				logger.Warn("skipping unreadable path", "path", path, "error", err)
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.IsDir() {
				if d.Name() == ".session-operation-locks" {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				logger.Warn("skipping file", "path", path, "error", err)
				return nil
			}
			if !opts.Since.IsZero() && info.ModTime().Before(opts.Since) {
				return nil
			}

			rel, err := filepath.Rel(root.path, path)
			if err != nil || rel == "." {
				rel = filepath.Base(path)
			}
			id := SessionIDFromPath(rel)
			if id == "" && root.source == LogSourcePath {
				id = SessionIDFromPath(path)
			}
			if len(prefixes) > 0 && !hasSessionPrefix(id, prefixes) {
				return nil
			}
			key := logSessionMetadataPath(path, id)
			name, ok := names[key]
			if !ok {
				name = readLogSessionName(key)
				names[key] = name
			}
			return fn(LogFile{Source: root.source, Path: path, SessionID: id, ModifiedAt: info.ModTime(), Workspace: logWorkspace(path), SessionName: name})
		})
		if walkErr != nil {
			return walkErr
		}
	}
	return nil
}

func logSessionMetadataPath(path, id string) string {
	if id == "" {
		return ""
	}
	for _, ext := range []string{".jsonl", ".json"} {
		if strings.EqualFold(filepath.Base(path), id+ext) {
			return path
		}
	}
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		if strings.EqualFold(filepath.Base(dir), id) {
			metadata := filepath.Join(dir, "workspace.yaml")
			if _, err := os.Stat(metadata); err == nil {
				return metadata
			}
		}
		var chatDir string
		switch filepath.Base(filepath.Dir(dir)) {
		case "workspaceStorage":
			chatDir = filepath.Join(dir, "chatSessions")
		case "User":
			if filepath.Base(dir) == "globalStorage" {
				chatDir = filepath.Join(dir, "emptyWindowChatSessions")
			}
		}
		if chatDir != "" {
			for _, ext := range []string{".jsonl", ".json"} {
				metadata := filepath.Join(chatDir, id+ext)
				if _, err := os.Stat(metadata); err == nil {
					return metadata
				}
			}
			return ""
		}
		if filepath.Dir(dir) == dir {
			return ""
		}
	}
}

func readLogSessionName(path string) string {
	if path == "" {
		return ""
	}
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	if filepath.Ext(path) == ".yaml" {
		var metadata struct {
			Name    string `yaml:"name"`
			Summary string `yaml:"summary"`
		}
		if yaml.NewDecoder(file).Decode(&metadata) != nil {
			return ""
		}
		if metadata.Name != "" {
			return metadata.Name
		}
		return metadata.Summary
	}
	decoder := json.NewDecoder(file)
	name := ""
	for {
		var entry struct {
			Kind        int             `json:"kind"`
			Key         []any           `json:"k"`
			Value       json.RawMessage `json:"v"`
			CustomTitle string          `json:"customTitle"`
		}
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return name
			}
			return ""
		}
		if filepath.Ext(path) == ".json" {
			return entry.CustomTitle
		}
		if entry.Kind == 0 || entry.Kind == 1 && len(entry.Key) == 0 {
			var snapshot struct {
				CustomTitle string `json:"customTitle"`
			}
			if json.Unmarshal(entry.Value, &snapshot) == nil {
				name = snapshot.CustomTitle
			}
		} else if entry.Kind == 1 && len(entry.Key) == 1 && entry.Key[0] == any("customTitle") {
			var title string
			if json.Unmarshal(entry.Value, &title) != nil {
				return ""
			}
			name = title
		}
	}
}

func logWorkspace(path string) string {
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		if filepath.Base(filepath.Dir(dir)) == "workspaceStorage" {
			folder, _ := readWorkspaceFolder(dir)
			return folder
		}
		if sessionUUIDPattern.MatchString(filepath.Base(dir)) {
			data, err := os.ReadFile(filepath.Join(dir, "workspace.yaml"))
			if err == nil {
				var metadata workspaceMetadata
				if yaml.Unmarshal(data, &metadata) == nil && metadata.CWD != "" {
					return metadata.CWD
				}
			}
		}
		if filepath.Dir(dir) == dir {
			return ""
		}
	}
}

// SessionIDFromPath returns the lower-case UUID found in the last component of
// the relative path that contains one, or "" when the path has no session ID.
func SessionIDFromPath(rel string) string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if id := sessionUUIDPattern.FindString(parts[i]); id != "" {
			return strings.ToLower(id)
		}
	}
	return ""
}

func hasSessionPrefix(id string, prefixes []string) bool {
	if id == "" {
		return false
	}
	for _, p := range prefixes {
		if strings.HasPrefix(id, p) {
			return true
		}
	}
	return false
}

func logRoots(opts LogOptions) ([]logRoot, error) {
	if len(opts.Paths) > 0 {
		roots := make([]logRoot, 0, len(opts.Paths))
		for _, p := range opts.Paths {
			abs, err := filepath.Abs(p)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve path %q: %w", p, err)
			}
			if _, err := os.Stat(abs); err != nil {
				return nil, fmt.Errorf("failed to access path %q: %w", p, err)
			}
			roots = append(roots, logRoot{source: LogSourcePath, path: abs})
		}
		return roots, nil
	}

	sources := opts.Sources
	if len(sources) == 0 {
		sources = LogSources
	}
	home := opts.CopilotHome
	if home == "" {
		var err error
		if home, err = copilotHome(); err != nil {
			return nil, err
		}
	}
	userDir := opts.VSCodeUserDir
	if userDir == "" {
		var err error
		if userDir, err = VSCodeUserDir(); err != nil {
			return nil, err
		}
	}

	var candidates []logRoot
	for _, source := range sources {
		switch source {
		case LogSourceCLISessions:
			candidates = append(candidates, logRoot{source, filepath.Join(home, "session-state")})
		case LogSourceCLILogs:
			candidates = append(candidates, logRoot{source, filepath.Join(home, "logs")})
		case LogSourceVSCodeChat:
			for _, ws := range workspaceStorageDirs(userDir) {
				candidates = append(candidates, logRoot{source, filepath.Join(ws, "chatSessions")})
			}
			candidates = append(candidates, logRoot{source, filepath.Join(userDir, "globalStorage", "emptyWindowChatSessions")})
		case LogSourceVSCodeDebug:
			chatDirs := []string{filepath.Join(userDir, "globalStorage", "github.copilot-chat")}
			for _, ws := range workspaceStorageDirs(userDir) {
				chatDirs = append(chatDirs, filepath.Join(ws, "GitHub.copilot-chat"))
			}
			for _, dir := range chatDirs {
				for _, sub := range []string{"debug-logs", "transcripts", "chat-session-resources"} {
					candidates = append(candidates, logRoot{source, filepath.Join(dir, sub)})
				}
			}
		default:
			return nil, fmt.Errorf("unknown copilot log source %q", source)
		}
	}

	var roots []logRoot
	for _, c := range candidates {
		if _, err := os.Stat(c.path); err != nil {
			logger.Debug("copilot log location not found", "source", c.source, "path", c.path)
			continue
		}
		roots = append(roots, c)
	}
	return roots, nil
}

// workspaceStorageDirs lists the per-workspace directories of a VS Code user dir.
func workspaceStorageDirs(userDir string) []string {
	base := filepath.Join(userDir, "workspaceStorage")
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(base, e.Name()))
		}
	}
	return dirs
}
