// Package copilotext provides reusable install/uninstall/list/status commands for
// GitHub Copilot CLI canvas extensions bundled with a host gh CLI extension.
package copilotext

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/srz-zumix/go-gh-extension/pkg/parser"
)

// Extension describes a single Copilot CLI canvas extension bundled with a host CLI tool.
type Extension struct {
	// Name is the extension's directory name, matching the "name" field in its
	// copilot-extension.json manifest.
	Name string
	// URL is a GitHub tree (directory) URL pointing at the extension's source directory,
	// e.g. "https://github.com/owner/repo/tree/v1.0.0/.github/extensions/my-extension".
	// The ref segment (the single path component immediately after "tree/") is used as
	// the default ref; it must not contain a slash. Pass an explicit ref (e.g. via the
	// command's --ref flag) to install from a branch name containing a slash.
	URL string
}

// Config configures the extension subcommands for a single host CLI tool.
type Config struct {
	// ToolName is the host CLI's name, recorded in installed extensions' metadata.
	ToolName string
	// ToolVersion is the host CLI's version, recorded in installed extensions' metadata.
	ToolVersion string
	// Extensions lists the extensions bundled with the host CLI.
	Extensions []Extension
}

// find returns the extension named name from cfg.Extensions.
func (cfg Config) find(name string) (Extension, error) {
	for _, ext := range cfg.Extensions {
		if ext.Name == name {
			return ext, nil
		}
	}
	return Extension{}, fmt.Errorf("unknown extension %q", name)
}

// selectExtensions returns the extensions named in names, or all of cfg.Extensions when
// names is empty.
func (cfg Config) selectExtensions(names []string) ([]Extension, error) {
	if len(names) == 0 {
		return cfg.Extensions, nil
	}
	exts := make([]Extension, 0, len(names))
	for _, name := range names {
		ext, err := cfg.find(name)
		if err != nil {
			return nil, err
		}
		exts = append(exts, ext)
	}
	return exts, nil
}

// source is the resolved location of an extension: a repository, ref, and subdirectory path.
type source struct {
	Repo repository.Repository
	Ref  string
	Path string
}

// resolve parses ext.URL into a source, overriding the ref with overrideRef when non-empty.
func resolve(ext Extension, overrideRef string) (*source, error) {
	treeURL, err := parser.ParseTreeURL(ext.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URL for extension %q: %w", ext.Name, err)
	}
	if treeURL == nil {
		return nil, fmt.Errorf("source URL for extension %q is not a GitHub tree URL: %s", ext.Name, ext.URL)
	}

	ref := treeURL.Ref
	if overrideRef != "" {
		ref = overrideRef
	}

	return &source{
		Repo: *treeURL.Repo,
		Ref:  ref,
		Path: treeURL.Path,
	}, nil
}
