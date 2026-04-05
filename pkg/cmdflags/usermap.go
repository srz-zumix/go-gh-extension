package cmdflags

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/srz-zumix/go-gh-extension/pkg/settings"
)

// AddUsermapFlag adds the --usermap flag to cmd and sets up PreRunE to load the mapping file.
// The loaded *settings.CompiledMappings is written to *mappings when the flag is provided.
// usage is the flag description shown in help text.
func AddUsermapFlag(cmd *cobra.Command, mappings **settings.CompiledMappings, usage string) {
	var mapFile string
	cmd.Flags().StringVar(&mapFile, "usermap", "", usage)

	oldPreRunE := cmd.PreRunE
	cmd.PreRunE = func(c *cobra.Command, args []string) error {
		if oldPreRunE != nil {
			if err := oldPreRunE(c, args); err != nil {
				return err
			}
		}
		if mapFile == "" {
			return nil
		}
		compiled, err := settings.NewCompiledMappingsFromFile(mapFile)
		if err != nil {
			return fmt.Errorf("error loading mapping file '%s': %w", mapFile, err)
		}
		*mappings = compiled
		return nil
	}
}
