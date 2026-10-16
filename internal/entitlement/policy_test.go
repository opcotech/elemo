package entitlement_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/deployment"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/entitlement/license"
)

func TestUnrestricted(t *testing.T) {
	policy := entitlement.Unrestricted()
	limit, err := policy.ActivationLimit(context.Background())
	require.NoError(t, err)
	require.True(t, limit.Unlimited)
	require.NoError(t, policy.AllowsMutation(context.Background()))

	status, err := policy.Status(context.Background())
	require.NoError(t, err)
	require.Equal(t, deployment.ModeSelfHosted, status.DeploymentMode)
	require.Nil(t, status.AirGap)
}

func TestAirGapPolicyActivation(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	installationID := uuid.NewString()
	lic := license.License{
		KeyID: license.ProductionKeyID,
		Payload: license.Payload{
			ID:             uuid.NewString(),
			Customer:       "ACME",
			InstallationID: installationID,
			Seats:          3,
			IssuedAt:       now.Add(-time.Hour),
			NotBefore:      now.Add(-time.Minute),
			ExpiresAt:      now.Add(time.Hour),
		},
	}

	t.Run("valid seats", func(t *testing.T) {
		policy := entitlement.NewAirGapPolicy(license.EvaluateLicense(lic, now, installationID), func() time.Time { return now }, nil)
		limit, err := policy.ActivationLimit(context.Background())
		require.NoError(t, err)
		require.False(t, limit.Unlimited)
		require.Equal(t, uint32(3), limit.Max)
	})

	t.Run("grace allows activation", func(t *testing.T) {
		grace := lic
		grace.Payload.ExpiresAt = now.Add(-time.Hour)
		grace.Payload.NotBefore = now.Add(-48 * time.Hour)
		grace.Payload.IssuedAt = now.Add(-72 * time.Hour)
		policy := entitlement.NewAirGapPolicy(license.EvaluateLicense(grace, now, installationID), func() time.Time { return now }, nil)
		limit, err := policy.ActivationLimit(context.Background())
		require.NoError(t, err)
		require.Equal(t, uint32(3), limit.Max)
	})

	t.Run("expired denies activation", func(t *testing.T) {
		expired := lic
		expired.Payload.ExpiresAt = now.Add(-40 * 24 * time.Hour)
		expired.Payload.NotBefore = now.Add(-80 * 24 * time.Hour)
		expired.Payload.IssuedAt = now.Add(-90 * 24 * time.Hour)
		policy := entitlement.NewAirGapPolicy(license.EvaluateLicense(expired, now, installationID), func() time.Time { return now }, nil)
		_, err := policy.ActivationLimit(context.Background())
		require.ErrorIs(t, err, entitlement.ErrActivationDenied)
	})

	t.Run("missing denies activation", func(t *testing.T) {
		policy := entitlement.NewAirGapPolicy(license.MissingEvaluation(now, installationID, "missing"), func() time.Time { return now }, nil)
		_, err := policy.ActivationLimit(context.Background())
		require.ErrorIs(t, err, entitlement.ErrActivationDenied)
	})
}

func TestAirGapPolicyMutation(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	installationID := uuid.NewString()
	lic := license.License{
		KeyID: license.ProductionKeyID,
		Payload: license.Payload{
			ID:             uuid.NewString(),
			Customer:       "ACME",
			InstallationID: installationID,
			Seats:          3,
			IssuedAt:       now.Add(-time.Hour),
			NotBefore:      now.Add(-time.Minute),
			ExpiresAt:      now.Add(time.Hour),
		},
	}

	t.Run("valid allows mutation", func(t *testing.T) {
		policy := entitlement.NewAirGapPolicy(license.EvaluateLicense(lic, now, installationID), func() time.Time { return now }, nil)
		require.NoError(t, policy.AllowsMutation(context.Background()))
	})

	t.Run("grace allows mutation", func(t *testing.T) {
		grace := lic
		grace.Payload.ExpiresAt = now.Add(-time.Hour)
		grace.Payload.NotBefore = now.Add(-48 * time.Hour)
		grace.Payload.IssuedAt = now.Add(-72 * time.Hour)
		policy := entitlement.NewAirGapPolicy(license.EvaluateLicense(grace, now, installationID), func() time.Time { return now }, nil)
		require.NoError(t, policy.AllowsMutation(context.Background()))
	})

	t.Run("grace ends at the grace boundary", func(t *testing.T) {
		expiredAt := now.Add(-license.GracePeriod)
		boundary := lic
		boundary.Payload.ExpiresAt = expiredAt
		boundary.Payload.NotBefore = now.Add(-48 * time.Hour)
		boundary.Payload.IssuedAt = now.Add(-72 * time.Hour)
		atBoundary := entitlement.NewAirGapPolicy(license.EvaluateLicense(boundary, now, installationID), func() time.Time { return now }, nil)
		require.NoError(t, atBoundary.AllowsMutation(context.Background()))

		past := now.Add(time.Nanosecond)
		pastPolicy := entitlement.NewAirGapPolicy(license.EvaluateLicense(boundary, past, installationID), func() time.Time { return past }, nil)
		require.ErrorIs(t, pastPolicy.AllowsMutation(context.Background()), entitlement.ErrMutationDenied)
	})

	t.Run("expired denies mutation", func(t *testing.T) {
		expired := lic
		expired.Payload.ExpiresAt = now.Add(-40 * 24 * time.Hour)
		expired.Payload.NotBefore = now.Add(-80 * 24 * time.Hour)
		expired.Payload.IssuedAt = now.Add(-90 * 24 * time.Hour)
		policy := entitlement.NewAirGapPolicy(license.EvaluateLicense(expired, now, installationID), func() time.Time { return now }, nil)
		require.ErrorIs(t, policy.AllowsMutation(context.Background()), entitlement.ErrMutationDenied)
	})

	t.Run("missing denies mutation", func(t *testing.T) {
		policy := entitlement.NewAirGapPolicy(license.MissingEvaluation(now, installationID, "missing"), func() time.Time { return now }, nil)
		require.ErrorIs(t, policy.AllowsMutation(context.Background()), entitlement.ErrMutationDenied)
	})

	t.Run("invalid denies mutation", func(t *testing.T) {
		policy := entitlement.NewAirGapPolicy(license.InvalidEvaluation(now, installationID, "invalid"), func() time.Time { return now }, nil)
		require.ErrorIs(t, policy.AllowsMutation(context.Background()), entitlement.ErrMutationDenied)
	})

	t.Run("not yet valid denies mutation", func(t *testing.T) {
		pending := lic
		pending.Payload.NotBefore = now.Add(time.Hour)
		pending.Payload.ExpiresAt = now.Add(24 * time.Hour)
		policy := entitlement.NewAirGapPolicy(license.EvaluateLicense(pending, now, installationID), func() time.Time { return now }, nil)
		require.ErrorIs(t, policy.AllowsMutation(context.Background()), entitlement.ErrMutationDenied)
	})
}

type stubActiveHumans int

func (s stubActiveHumans) ActiveHumanCount(context.Context) (int, error) {
	return int(s), nil
}

func TestAirGapPolicyStatus(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	installationID := uuid.NewString()
	lic := license.License{
		KeyID: license.ProductionKeyID,
		Payload: license.Payload{
			ID:             uuid.NewString(),
			Customer:       "ACME",
			InstallationID: installationID,
			Seats:          3,
			IssuedAt:       now.Add(-time.Hour),
			NotBefore:      now.Add(-time.Minute),
			ExpiresAt:      now.Add(time.Hour),
		},
	}

	t.Run("requires an active human counter", func(t *testing.T) {
		policy := entitlement.NewAirGapPolicy(license.EvaluateLicense(lic, now, installationID), func() time.Time { return now }, nil)
		_, err := policy.Status(context.Background())
		require.ErrorIs(t, err, entitlement.ErrNoActiveHumanCounter)
	})

	t.Run("reports licensed and active seats", func(t *testing.T) {
		policy := entitlement.NewAirGapPolicy(license.EvaluateLicense(lic, now, installationID), func() time.Time { return now }, stubActiveHumans(2))
		status, err := policy.Status(context.Background())
		require.NoError(t, err)
		require.Equal(t, deployment.ModeAirGap, status.DeploymentMode)
		require.NotNil(t, status.AirGap)
		require.Equal(t, 3, status.AirGap.SeatsLicensed)
		require.Equal(t, 2, status.AirGap.SeatsActive)
	})
}
