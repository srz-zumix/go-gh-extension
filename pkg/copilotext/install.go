package copilotext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/v90/github"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/ioutil"
)

// InstallOptions configures Install and Update.
type InstallOptions struct {
	Scope  Scope
	Prefix string
	Ref    string
	DryRun bool
	Force  bool
}

// InstallResult describes the outcome of Install or Update.
type InstallResult struct {
	Name      string
	Dir       string
	Ref       string
	CommitSHA string
	// Changed is false when Update found the extension already up to date.
	Changed bool
}

// Install downloads and installs the named extension. If the destination already exists
// without install metadata, Install fails unless opts.Force is set.
func Install(ctx context.Context, cfg Config, name string, opts InstallOptions) (*InstallResult, error) {
	ext, err := cfg.find(name)
	if err != nil {
		return nil, err
	}
	return installOrUpdate(ctx, cfg, ext, opts, false)
}

// Update re-installs the named extension when the resolved ref points at a different
// commit than the one installed, or when opts.Force is set. It reports Changed=false when
// already up to date.
func Update(ctx context.Context, cfg Config, name string, opts InstallOptions) (*InstallResult, error) {
	ext, err := cfg.find(name)
	if err != nil {
		return nil, err
	}
	return installOrUpdate(ctx, cfg, ext, opts, true)
}

func installOrUpdate(ctx context.Context, cfg Config, ext Extension, opts InstallOptions, isUpdate bool) (*InstallResult, error) {
	src, err := resolve(ext, opts.Ref)
	if err != nil {
		return nil, err
	}

	root, err := extensionsRoot(ctx, opts.Scope, opts.Prefix)
	if err != nil {
		return nil, err
	}
	dir := destDir(root, ext.Name)

	// Determine the local installation state before making any network calls so that a
	// misuse (e.g. updating an extension that is not installed) fails fast without hitting
	// GitHub.
	existing, err := readMetadata(dir)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		if isUpdate {
			// Update only operates on an already-installed extension. A missing
			// destination means it was never installed, which --force does not
			// bootstrap; the user must run install first.
			if _, statErr := os.Stat(dir); statErr != nil {
				if os.IsNotExist(statErr) {
					return nil, fmt.Errorf("extension %q is not installed; run install first", ext.Name)
				}
				return nil, fmt.Errorf("failed to inspect %q: %w", dir, statErr)
			}
			// The destination exists but is unmanaged: require --force to overwrite it,
			// preserving the actionable error from requireOverwritable.
		}
		if err := requireOverwritable(dir, opts.Force); err != nil {
			return nil, err
		}
	}

	client, err := gh.NewGitHubClientWithRepo(src.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client for %s/%s: %w", src.Repo.Owner, src.Repo.Name, err)
	}
	sha, err := gh.GetCommitSHA1(ctx, client, src.Repo, src.Ref)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ref %q for %s/%s: %w", src.Ref, src.Repo.Owner, src.Repo.Name, err)
	}

	if isUpdate && !opts.Force && existing != nil && existing.CommitSHA == sha && existing.Ref == src.Ref {
		return &InstallResult{Name: ext.Name, Dir: dir, Ref: src.Ref, CommitSHA: sha, Changed: false}, nil
	}

	result := &InstallResult{Name: ext.Name, Dir: dir, Ref: src.Ref, CommitSHA: sha, Changed: true}
	if opts.DryRun {
		return result, nil
	}

	if err := installArchive(ctx, client, cfg, src, ext.Name, dir, sha); err != nil {
		return nil, err
	}
	return result, nil
}

// requireOverwritable returns an error if dir exists, has no install metadata (i.e. is
// not managed by this command), and force is false.
func requireOverwritable(dir string, force bool) error {
	if force {
		return nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to inspect %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q already exists and is not a directory; use --force to overwrite", dir)
	}
	m, err := readMetadata(dir)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%q already exists and is not managed by this command; use --force to overwrite", dir)
	}
	return nil
}

// installArchive downloads src's tarball, extracts src.Path into a temporary directory,
// writes install metadata, then atomically replaces dir with the new contents.
func installArchive(ctx context.Context, client *gh.GitHubClient, cfg Config, src *source, name, dir, sha string) error {
	body, err := gh.DownloadRepositoryArchive(ctx, client, src.Repo, src.Ref, github.Tarball)
	if err != nil {
		return err
	}
	defer body.Close() //nolint:errcheck

	root := filepath.Dir(dir)
	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("failed to create extensions directory %q: %w", root, err)
	}

	tmpDir, err := os.MkdirTemp(root, ".copilot-extension-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}
	defer os.RemoveAll(tmpDir) //nolint:errcheck

	if err := ioutil.ExtractTarGzSubdir(body, src.Path, tmpDir); err != nil {
		return fmt.Errorf("failed to extract extension %q archive: %w", name, err)
	}

	m := metadata{
		Tool:        cfg.ToolName,
		ToolVersion: cfg.ToolVersion,
		Host:        src.Repo.Host,
		Owner:       src.Repo.Owner,
		Repo:        src.Repo.Name,
		Path:        src.Path,
		Ref:         src.Ref,
		CommitSHA:   sha,
		InstalledAt: time.Now().UTC(),
	}
	if err := writeMetadata(tmpDir, m); err != nil {
		return err
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove existing extension directory %q: %w", dir, err)
	}
	if err := os.Rename(tmpDir, dir); err != nil {
		return fmt.Errorf("failed to install extension %q to %q: %w", name, dir, err)
	}
	return nil
}

// Uninstall removes the named extension's installed directory. It refuses to remove a
// directory that has no install metadata unless force is true.
func Uninstall(ctx context.Context, cfg Config, name string, scope Scope, prefix string, dryRun, force bool) error {
	if _, err := cfg.find(name); err != nil {
		return err
	}
	root, err := extensionsRoot(ctx, scope, prefix)
	if err != nil {
		return err
	}
	dir := destDir(root, name)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("extension %q is not installed at %q", name, dir)
		}
		return fmt.Errorf("failed to inspect %q: %w", dir, err)
	}

	if !force {
		m, err := readMetadata(dir)
		if err != nil {
			return err
		}
		if m == nil {
			return fmt.Errorf("%q is not managed by this command; use --force to remove it anyway", dir)
		}
	}

	if dryRun {
		return nil
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove %q: %w", dir, err)
	}
	return nil
}

// Status describes an extension's installation state.
type Status struct {
	Name        string
	Dir         string
	Installed   bool
	Managed     bool
	Ref         string
	CommitSHA   string
	InstalledAt time.Time
}

// GetStatus reports the installation state of the named extension. It only inspects the
// local filesystem and does not make any network calls.
func GetStatus(ctx context.Context, cfg Config, name string, scope Scope, prefix string) (*Status, error) {
	if _, err := cfg.find(name); err != nil {
		return nil, err
	}
	root, err := extensionsRoot(ctx, scope, prefix)
	if err != nil {
		return nil, err
	}
	dir := destDir(root, name)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return &Status{Name: name, Dir: dir}, nil
		}
		return nil, fmt.Errorf("failed to inspect %q: %w", dir, err)
	}

	m, err := readMetadata(dir)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return &Status{Name: name, Dir: dir, Installed: true}, nil
	}
	return &Status{
		Name:        name,
		Dir:         dir,
		Installed:   true,
		Managed:     true,
		Ref:         m.Ref,
		CommitSHA:   m.CommitSHA,
		InstalledAt: m.InstalledAt,
	}, nil
}
