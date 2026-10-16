//go:build !airgap

package deployment_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/deployment"
)

func TestCurrentSelfHosted(t *testing.T) {
	require.Equal(t, deployment.ModeSelfHosted, deployment.Current())
	require.False(t, deployment.IsAirGap())
}
