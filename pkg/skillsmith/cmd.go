// Package skillsmith provides a cobra.Command wrapper for the Songmu/skillsmith library.
// It enables CLI tools to easily expose a "skills" subcommand for managing agent skills.
package skillsmith

import (
	"io/fs"

	"github.com/Songmu/skillsmith"
	"github.com/spf13/cobra"
)

// NewSkillsCmd creates a cobra.Command that wraps skillsmith's Run method.
// name is the hosting CLI tool name, version is the tool's version string,
// and skillsFS is the embedded filesystem containing the skills directory.
func NewSkillsCmd(name, version string, skillsFS fs.FS) *cobra.Command {
	return &cobra.Command{
		Use:                "skills",
		Short:              "Manage agent skills",
		Long:               `Install, update, and manage agent skills bundled with this tool.`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := skillsmith.New(name, version, skillsFS)
			if err != nil {
				return err
			}
			return s.Run(cmd.Context(), args)
		},
	}
}
