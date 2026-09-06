// Package ghexec runs the gh CLI for the features that have no API wrapper.
package ghexec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	ghcli "github.com/cli/go-gh/v2"
)

// Run executes the gh CLI and returns its standard output. GH_HOST and GH_REPO
// are removed from the environment to prevent an inherited host or repository
// override from retargeting the command; the features that need this helper are
// expected to run against github.com.
func Run(ctx context.Context, args ...string) (string, error) {
	path, err := ghcli.Path()
	if err != nil {
		return "", fmt.Errorf("failed to locate the gh CLI: %w", err)
	}
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = env()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return stdout.String(), fmt.Errorf("%w: %s", err, msg)
		}
		return stdout.String(), err
	}
	return stdout.String(), nil
}

// env returns the current environment without the gh host (GH_HOST) and
// repository (GH_REPO) overrides.
func env() []string {
	current := os.Environ()
	filtered := make([]string, 0, len(current))
	for _, kv := range current {
		key, _, _ := strings.Cut(kv, "=")
		if key == "GH_HOST" || key == "GH_REPO" {
			continue
		}
		filtered = append(filtered, kv)
	}
	return filtered
}
