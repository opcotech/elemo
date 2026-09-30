//go:build airgap

package cli

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/deployment"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/entitlement/license"
)

func TestAirGapLicensePolicyStartup(t *testing.T) {
	now := time.Date(2026, time.September, 7, 12, 0, 0, 0, time.UTC)
	installationID := uuid.NewString()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	raw, err := license.Sign("test", privateKey, license.Payload{
		ID:             uuid.NewString(),
		Customer:       "Test Customer",
		InstallationID: installationID,
		Seats:          17,
		IssuedAt:       now.Add(-time.Hour),
		NotBefore:      now.Add(-time.Minute),
		ExpiresAt:      now.Add(time.Hour),
	})
	require.NoError(t, err)

	evaluation, err := evaluateAirGapLicense(
		raw,
		license.Trust{"test": publicKey},
		now,
		installationID,
	)
	require.NoError(t, err)
	require.Equal(t, deployment.ModeAirGap, deployment.Current())
	require.Equal(t, license.StateValid, evaluation.State)

	policy := entitlement.NewAirGapPolicy(evaluation, func() time.Time { return now }, nil)
	limit, err := policy.ActivationLimit(t.Context())
	require.NoError(t, err)
	require.False(t, limit.Unlimited)
	require.Equal(t, uint32(17), limit.Max)
	require.NoError(t, policy.AllowsMutation(t.Context()))
}
