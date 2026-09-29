//go:build airgap

package deployment_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/deployment"
)

func TestCurrentAirGap(t *testing.T) {
	require.Equal(t, deployment.ModeAirGap, deployment.Current())
	require.True(t, deployment.IsAirGap())
}
