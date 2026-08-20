package cmdflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddPersistentFlags_ChainsExistingPersistentPreRunE(t *testing.T) {
	called := false
	cmd := &cobra.Command{Use: "root"}
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		called = true
		return nil
	}

	AddPersistentFlags(cmd)

	require.NoError(t, cmd.PersistentPreRunE(cmd, nil))
	assert.True(t, called, "existing PersistentPreRunE should still be invoked")
}

func TestAddPersistentFlags_ChainsExistingPersistentPreRun(t *testing.T) {
	called := false
	cmd := &cobra.Command{Use: "root"}
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		called = true
	}

	AddPersistentFlags(cmd)

	require.Nil(t, cmd.PersistentPreRun, "PersistentPreRun should be folded into PersistentPreRunE")
	require.NoError(t, cmd.PersistentPreRunE(cmd, nil))
	assert.True(t, called, "existing PersistentPreRun should still be invoked")
}

func TestAddPersistentFlags_RejectsNonPositiveHTTPTimeout(t *testing.T) {
	cmd := &cobra.Command{Use: "root"}
	AddPersistentFlags(cmd)

	require.NoError(t, cmd.PersistentFlags().Set("http-timeout", "0s"))

	err := cmd.PersistentPreRunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http-timeout")
}
