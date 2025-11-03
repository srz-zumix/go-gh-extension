package cmdflags

import "github.com/spf13/cobra"

type MutuallyExclusiveBoolFlags struct {
	Enabled  bool
	Disabled bool
}

func (m *MutuallyExclusiveBoolFlags) AddNoPrefixFlag(cmd *cobra.Command, name string, enableCaseUsage string, disableCaseUsage string) {
	m.AddFlag(cmd, name, "no-"+name, enableCaseUsage, disableCaseUsage)
}

func (m *MutuallyExclusiveBoolFlags) AddFlag(cmd *cobra.Command, enableCaseName string, disableCaseName string, enableCaseUsage string, disableCaseUsage string) {
	cmd.Flags().BoolVar(&m.Enabled, enableCaseName, false, enableCaseUsage)
	cmd.Flags().BoolVar(&m.Disabled, disableCaseName, false, disableCaseUsage)
	cmd.MarkFlagsMutuallyExclusive(enableCaseName, disableCaseName)
}

func (m *MutuallyExclusiveBoolFlags) IsEnabled() bool {
	return m.Enabled
}

func (m *MutuallyExclusiveBoolFlags) IsTrue() bool {
	return m.Enabled
}

func (m *MutuallyExclusiveBoolFlags) IsDisabled() bool {
	return m.Disabled
}

func (m *MutuallyExclusiveBoolFlags) IsFalse() bool {
	return m.Disabled
}

func (m *MutuallyExclusiveBoolFlags) IsSet() bool {
	return m.Enabled || m.Disabled
}

func (m *MutuallyExclusiveBoolFlags) GetValue() *bool {
	if m.Enabled {
		val := true
		return &val
	} else if m.Disabled {
		val := false
		return &val
	}
	return nil
}
