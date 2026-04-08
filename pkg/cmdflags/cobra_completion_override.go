package cmdflags

// This file provides direct write access to cobra's unexported flagCompletionFunctions map
// via go:linkname. cobra.RegisterFlagCompletionFunc returns an error when called for a flag
// that already has a completion registered, with no public override API. Libraries such as
// cmdutil.AddFormatFlags (via StringEnumFlag) register completions before we can, so we
// must update the map directly.
//
// The go:linkname directive is fragile with respect to cobra internal renames; if it stops
// compiling after a cobra upgrade, check whether flagCompletionFunctions was renamed.

import (
	"fmt"
	"sync"
	_ "unsafe" // required for go:linkname

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//go:linkname cobraFlagCompletionFunctions github.com/spf13/cobra.flagCompletionFunctions
var cobraFlagCompletionFunctions map[*pflag.Flag]cobra.CompletionFunc

//go:linkname cobraFlagCompletionMutex github.com/spf13/cobra.flagCompletionMutex
var cobraFlagCompletionMutex *sync.RWMutex

// isAlreadyRegisteredError reports whether err is the specific error returned by
// cobra.RegisterFlagCompletionFunc when a completion for that flag was already set.
func isAlreadyRegisteredError(err error, flagName string) bool {
	return err != nil && err.Error() == fmt.Sprintf("RegisterFlagCompletionFunc: flag '%s' already registered", flagName)
}

// overrideFlagCompletion forcibly sets the completion function for flag, replacing any
// previously registered function. Use this only when RegisterFlagCompletionFunc cannot
// be used because the completion was already registered by an imported library.
// Returns an error if the go:linkname targets are nil, which indicates that cobra
// internals have changed and the linkname variables no longer bind correctly.
func overrideFlagCompletion(flag *pflag.Flag, f cobra.CompletionFunc) error {
	if cobraFlagCompletionMutex == nil {
		return fmt.Errorf("overrideFlagCompletion: cobraFlagCompletionMutex is nil; cobra internals may have changed")
	}
	if cobraFlagCompletionFunctions == nil {
		return fmt.Errorf("overrideFlagCompletion: cobraFlagCompletionFunctions is nil; cobra internals may have changed")
	}
	cobraFlagCompletionMutex.Lock()
	defer cobraFlagCompletionMutex.Unlock()
	cobraFlagCompletionFunctions[flag] = f
	return nil
}
