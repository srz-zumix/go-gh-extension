package cmdflags

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/srz-zumix/go-gh-extension/pkg/gh"
	"github.com/srz-zumix/go-gh-extension/pkg/gh/guardrails"
	"github.com/srz-zumix/go-gh-extension/pkg/logger"
)

var (
	logLevel    string
	readOnly    bool
	httpTimeout time.Duration
)

// AddPersistentFlags registers the options shared by every command and installs
// the PersistentPreRun hook that applies them. It also enables cobra.EnableTraverseRunHooks so persistent hooks run for the full command chain.
func AddPersistentFlags(cmd *cobra.Command) {
	// Run every persistent hook in the command chain, not just the closest one,
	// so descendant commands that define their own PersistentPreRun(E) do not
	// silently bypass the shared setup registered here.
	cobra.EnableTraverseRunHooks = true

	f := cmd.PersistentFlags()
	logger.AddCmdFlag(cmd, f, &logLevel, "log-level", "L")
	f.BoolVar(&readOnly, "read-only", false, "Run in read-only mode (prevent write operations)")
	f.DurationVar(&httpTimeout, "http-timeout", gh.DefaultHTTPTimeout, "Timeout for each GitHub API request")

	// Chain onto whatever hook the caller already installed instead of replacing it.
	// Cobra runs PersistentPreRunE in preference to PersistentPreRun, so mirror that here.
	prev, prevE := cmd.PersistentPreRun, cmd.PersistentPreRunE
	cmd.PersistentPreRun = nil
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := PersistentPreRun(cmd, args); err != nil {
			return err
		}
		if prevE != nil {
			return prevE(cmd, args)
		}
		if prev != nil {
			prev(cmd, args)
		}
		return nil
	}
}

// PersistentPreRun applies the options registered by AddPersistentFlags.
func PersistentPreRun(cmd *cobra.Command, args []string) error {
	logger.SetLogLevel(logLevel)
	guardrails.NewGuardrail(guardrails.ReadOnlyOption(readOnly))
	if httpTimeout <= 0 {
		return fmt.Errorf("invalid --http-timeout %s: expected a positive duration", httpTimeout)
	}
	gh.SetHTTPTimeout(httpTimeout)
	return nil
}
