package main

import (
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/entitlement/license"
)

func TestValidateOptionsSeatBounds(t *testing.T) {
	originalKeyID := keyID
	originalPrivateKeyFile := privateKeyFile
	originalOutputLicense := outputLicense
	originalCustomer := customer
	originalInstallationID := installationID
	originalSeats := seats
	originalValidityDays := validityDays
	t.Cleanup(func() {
		keyID = originalKeyID
		privateKeyFile = originalPrivateKeyFile
		outputLicense = originalOutputLicense
		customer = originalCustomer
		installationID = originalInstallationID
		seats = originalSeats
		validityDays = originalValidityDays
	})

	keyID = license.ProductionKeyID
	privateKeyFile = "vendor.key"
	outputLicense = "license.json"
	customer = "ACME"
	canonicalInstallationID := uuid.NewString()
	installationID = strings.ToUpper(canonicalInstallationID)
	validityDays = 365

	seats = math.MaxUint32
	require.NoError(t, validateOptions())
	require.Equal(t, canonicalInstallationID, installationID)

	if uint64(^uint(0)) > math.MaxUint32 {
		seats = uint(math.MaxUint32) + 1
		require.EqualError(t, validateOptions(), "seats must not exceed 4294967295")
	}
}
