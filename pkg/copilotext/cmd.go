package copilotext

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewExtensionCmd creates the "extension" command and its install/uninstall/list/status/
// update subcommands, configured for cfg's bundled extensions.
func NewExtensionCmd(cfg Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Manage Copilot CLI canvas extensions",
		Long:  `Install, update, and manage GitHub Copilot CLI canvas extensions bundled with this tool.`,
	}
	cmd.AddCommand(newExtensionListCmd(cfg))
	cmd.AddCommand(newExtensionStatusCmd(cfg))
	cmd.AddCommand(newExtensionInstallCmd(cfg))
	cmd.AddCommand(newExtensionUpdateCmd(cfg))
	cmd.AddCommand(newExtensionUninstallCmd(cfg))
	return cmd
}

// addScopeFlags registers the --scope and --prefix flags shared by the subcommands that
// resolve an installation directory.
func addScopeFlags(cmd *cobra.Command, scope *string, prefix *string) {
	cmd.Flags().StringVar(scope, "scope", string(ScopeUser), "installation scope (\"user\" or \"repo\")")
	cmd.Flags().StringVar(prefix, "prefix", "", "install directory, overriding --scope")
}

func newExtensionListCmd(cfg Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List extensions bundled with this tool",
		Long:  `List the Copilot CLI canvas extensions bundled with this tool, along with their default source ref.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, ext := range cfg.Extensions {
				src, err := resolve(ext, "")
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", ext.Name, ext.URL, src.Ref); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newExtensionStatusCmd(cfg Config) *cobra.Command {
	var scope, prefix string
	cmd := &cobra.Command{
		Use:   "status [name...]",
		Short: "Show installation status of extensions",
		Long:  `Show whether the given extensions (or all bundled extensions, when none are given) are installed and, if so, which ref they were installed from. Status only inspects the local filesystem and does not query GitHub.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			exts, err := cfg.selectExtensions(args)
			if err != nil {
				return err
			}
			for _, ext := range exts {
				status, err := GetStatus(cmd.Context(), cfg, ext.Name, Scope(scope), prefix)
				if err != nil {
					return fmt.Errorf("failed to get status of extension %q: %w", ext.Name, err)
				}
				var writeErr error
				switch {
				case !status.Installed:
					_, writeErr = fmt.Fprintf(cmd.OutOrStdout(), "%s\tnot installed\t%s\n", status.Name, status.Dir)
				case !status.Managed:
					_, writeErr = fmt.Fprintf(cmd.OutOrStdout(), "%s\tinstalled (unmanaged)\t%s\n", status.Name, status.Dir)
				default:
					_, writeErr = fmt.Fprintf(cmd.OutOrStdout(), "%s\tinstalled\tref=%s\tcommit=%s\t%s\n", status.Name, status.Ref, status.CommitSHA, status.Dir)
				}
				if writeErr != nil {
					return writeErr
				}
			}
			return nil
		},
	}
	addScopeFlags(cmd, &scope, &prefix)
	return cmd
}

func newExtensionInstallCmd(cfg Config) *cobra.Command {
	var opts InstallOptions
	var scope string
	cmd := &cobra.Command{
		Use:   "install [name...]",
		Short: "Install extensions",
		Long:  `Download and install the given extensions (or all bundled extensions, when none are given). Fails if the destination directory already exists and is not managed by this command, unless --force is given.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Scope = Scope(scope)
			exts, err := cfg.selectExtensions(args)
			if err != nil {
				return err
			}
			for _, ext := range exts {
				result, err := Install(cmd.Context(), cfg, ext.Name, opts)
				if err != nil {
					return fmt.Errorf("failed to install extension %q: %w", ext.Name, err)
				}
				if err := printInstallResult(cmd, "install", "installed", result, opts.DryRun); err != nil {
					return err
				}
			}
			return nil
		},
	}
	addScopeFlags(cmd, &scope, &opts.Prefix)
	cmd.Flags().StringVar(&opts.Ref, "ref", "", "git ref to install, overriding the extension's default ref")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "resolve and print what would be installed without writing any files")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "overwrite an existing, unmanaged destination directory")
	return cmd
}

func newExtensionUpdateCmd(cfg Config) *cobra.Command {
	var opts InstallOptions
	var scope string
	cmd := &cobra.Command{
		Use:   "update [name...]",
		Short: "Update installed extensions",
		Long:  `Re-install the given extensions (or all bundled extensions, when none are given) when their resolved ref points at a different commit than the one installed, or when --force is given.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Scope = Scope(scope)
			exts, err := cfg.selectExtensions(args)
			if err != nil {
				return err
			}
			for _, ext := range exts {
				result, err := Update(cmd.Context(), cfg, ext.Name, opts)
				if err != nil {
					return fmt.Errorf("failed to update extension %q: %w", ext.Name, err)
				}
				if !result.Changed {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\talready up to date\t%s\n", result.Name, result.Dir); err != nil {
						return err
					}
					continue
				}
				if err := printInstallResult(cmd, "update", "updated", result, opts.DryRun); err != nil {
					return err
				}
			}
			return nil
		},
	}
	addScopeFlags(cmd, &scope, &opts.Prefix)
	cmd.Flags().StringVar(&opts.Ref, "ref", "", "git ref to update to, overriding the extension's default ref")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "resolve and print what would be updated without writing any files")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "re-install even if already up to date, or overwrite an unmanaged destination directory")
	return cmd
}

func newExtensionUninstallCmd(cfg Config) *cobra.Command {
	var scope, prefix string
	var dryRun, force bool
	cmd := &cobra.Command{
		Use:   "uninstall [name...]",
		Short: "Uninstall extensions",
		Long:  `Remove the given extensions (or all bundled extensions, when none are given). Refuses to remove a destination directory that is not managed by this command, unless --force is given.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			exts, err := cfg.selectExtensions(args)
			if err != nil {
				return err
			}
			for _, ext := range exts {
				if err := Uninstall(cmd.Context(), cfg, ext.Name, Scope(scope), prefix, dryRun, force); err != nil {
					return fmt.Errorf("failed to uninstall extension %q: %w", ext.Name, err)
				}
				if dryRun {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\twould be uninstalled\n", ext.Name); err != nil {
						return err
					}
				} else {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\tuninstalled\n", ext.Name); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	addScopeFlags(cmd, &scope, &prefix)
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print what would be uninstalled without removing any files")
	cmd.Flags().BoolVar(&force, "force", false, "remove the destination directory even if it is not managed by this command")
	return cmd
}

// printInstallResult prints the outcome of an install or update operation. verb is the base
// form (e.g. "install") used for dry-run output and pastVerb is its past tense (e.g.
// "installed") used for the completed-action output.
func printInstallResult(cmd *cobra.Command, verb, pastVerb string, result *InstallResult, dryRun bool) error {
	var err error
	if dryRun {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\twould %s\tref=%s\tcommit=%s\t%s\n", result.Name, verb, result.Ref, result.CommitSHA, result.Dir)
	} else {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\tref=%s\tcommit=%s\t%s\n", result.Name, pastVerb, result.Ref, result.CommitSHA, result.Dir)
	}
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	return nil
}
