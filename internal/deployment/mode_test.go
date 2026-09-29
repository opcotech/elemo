package deployment_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/deployment"
)

func TestModeEnumeration(t *testing.T) {
	require.Equal(t, []deployment.Mode{deployment.ModeSelfHosted, deployment.ModeAirGap}, deployment.ModeValues())
	require.Equal(t, []string{"self_hosted", "airgap"}, deployment.ModeStrings())

	for _, mode := range deployment.ModeValues() {
		parsed, err := deployment.ModeString(mode.String())
		require.NoError(t, err)
		require.Equal(t, mode, parsed)
		require.True(t, mode.IsAMode())
	}

	require.False(t, deployment.Mode(0).IsAMode())
	_, err := deployment.ModeString("unknown")
	require.Error(t, err)
}
