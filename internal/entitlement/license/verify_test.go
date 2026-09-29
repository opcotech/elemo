package license_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/entitlement/license"
)

func testPayload(installationID string, now time.Time) license.Payload {
	return license.Payload{
		ID:             uuid.NewString(),
		Customer:       "ACME",
		InstallationID: installationID,
		Seats:          5,
		IssuedAt:       now.Add(-time.Hour),
		NotBefore:      now.Add(-time.Minute),
		ExpiresAt:      now.Add(24 * time.Hour),
	}
}

func signWith(t *testing.T, keyID string, priv ed25519.PrivateKey, payload license.Payload) []byte {
	t.Helper()
	raw, err := license.Sign(keyID, priv, payload)
	require.NoError(t, err)
	return raw
}

func TestVerifyAndEvaluate(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	trust := license.Trust{"test-key": pub}
	installationID := uuid.NewString()
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)

	t.Run("valid signature", func(t *testing.T) {
		raw := signWith(t, "test-key", priv, testPayload(installationID, now))
		lic, err := license.Verify(raw, trust)
		require.NoError(t, err)
		eval := license.EvaluateLicense(lic, now, installationID)
		require.Equal(t, license.StateValid, eval.State)
	})

	t.Run("equivalent UUID spelling remains bound", func(t *testing.T) {
		raw := signWith(t, "test-key", priv, testPayload(strings.ToUpper(installationID), now))
		lic, err := license.Verify(raw, trust)
		require.NoError(t, err)
		eval := license.EvaluateLicense(lic, now, installationID)
		require.Equal(t, license.StateValid, eval.State)
	})

	t.Run("tampered signature", func(t *testing.T) {
		raw := signWith(t, "test-key", priv, testPayload(installationID, now))
		var envelope map[string]string
		require.NoError(t, json.Unmarshal(raw, &envelope))
		sig, err := base64.StdEncoding.DecodeString(envelope["signature"])
		require.NoError(t, err)
		sig[0] ^= 0xff
		envelope["signature"] = base64.StdEncoding.EncodeToString(sig)
		tampered, err := json.Marshal(envelope)
		require.NoError(t, err)
		_, err = license.Verify(tampered, trust)
		require.ErrorIs(t, err, license.ErrLicenseInvalidSignature)
	})

	t.Run("unknown key id", func(t *testing.T) {
		raw := signWith(t, "test-key", priv, testPayload(installationID, now))
		_, err := license.Verify(raw, license.Trust{"other": pub})
		require.ErrorIs(t, err, license.ErrLicenseUnknownKey)
	})

	t.Run("backdated not-before remains valid", func(t *testing.T) {
		payload := testPayload(installationID, now)
		payload.NotBefore = now.Add(-24 * time.Hour)
		payload.IssuedAt = now
		raw := signWith(t, "test-key", priv, payload)
		_, err := license.Verify(raw, trust)
		require.NoError(t, err)
	})

	t.Run("not yet valid", func(t *testing.T) {
		payload := testPayload(installationID, now)
		payload.NotBefore = now.Add(time.Hour)
		raw := signWith(t, "test-key", priv, payload)
		lic, err := license.Verify(raw, trust)
		require.NoError(t, err)
		eval := license.EvaluateLicense(lic, now, installationID)
		require.Equal(t, license.StateNotYetValid, eval.State)
	})

	t.Run("grace", func(t *testing.T) {
		payload := testPayload(installationID, now)
		payload.ExpiresAt = now.Add(-time.Hour)
		payload.NotBefore = now.Add(-48 * time.Hour)
		payload.IssuedAt = now.Add(-72 * time.Hour)
		raw := signWith(t, "test-key", priv, payload)
		lic, err := license.Verify(raw, trust)
		require.NoError(t, err)
		eval := license.EvaluateLicense(lic, now, installationID)
		require.Equal(t, license.StateGrace, eval.State)
		require.Equal(t, payload.ExpiresAt.UTC().Add(license.GracePeriod), eval.GraceEndsAt)
	})

	t.Run("expired beyond grace", func(t *testing.T) {
		payload := testPayload(installationID, now)
		payload.ExpiresAt = now.Add(-31 * 24 * time.Hour)
		payload.NotBefore = now.Add(-60 * 24 * time.Hour)
		payload.IssuedAt = now.Add(-90 * 24 * time.Hour)
		raw := signWith(t, "test-key", priv, payload)
		lic, err := license.Verify(raw, trust)
		require.NoError(t, err)
		eval := license.EvaluateLicense(lic, now, installationID)
		require.Equal(t, license.StateExpired, eval.State)
	})

	t.Run("wrong installation", func(t *testing.T) {
		raw := signWith(t, "test-key", priv, testPayload(installationID, now))
		lic, err := license.Verify(raw, trust)
		require.NoError(t, err)
		eval := license.EvaluateLicense(lic, now, uuid.NewString())
		require.Equal(t, license.StateInvalid, eval.State)
		require.Contains(t, eval.Reason, "installation")
	})

	t.Run("unknown field rejected", func(t *testing.T) {
		raw := signWith(t, "test-key", priv, testPayload(installationID, now))
		var envelope map[string]any
		require.NoError(t, json.Unmarshal(raw, &envelope))
		envelope["extra"] = "nope"
		tampered, err := json.Marshal(envelope)
		require.NoError(t, err)
		_, err = license.Verify(tampered, trust)
		require.Error(t, err)
	})

	t.Run("key rotation accepts a second trusted key", func(t *testing.T) {
		pub2, priv2, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)
		rotated := license.Trust{"test-key": pub, "test-key-2027": pub2}
		raw := signWith(t, "test-key-2027", priv2, testPayload(installationID, now))
		lic, err := license.Verify(raw, rotated)
		require.NoError(t, err)
		require.Equal(t, "test-key-2027", lic.KeyID)
	})
}

func TestLicenseFileIO(t *testing.T) {
	t.Run("atomic overwrite restricts permissions", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "license.json")
		require.NoError(t, os.WriteFile(path, []byte("old"), 0o666)) //nolint:gosec // The test verifies overwrite hardening.
		require.NoError(t, license.WriteFile(path, []byte("new")))

		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		raw, err := license.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, "new\n", string(raw))
	})

	t.Run("oversized file is rejected", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "license.json")
		require.NoError(t, os.WriteFile(path, make([]byte, license.MaxLicenseBytes+1), 0o600))
		_, err := license.ReadFile(path)
		require.ErrorIs(t, err, license.ErrLicenseTooLarge)
	})
}

func TestStateEnumeration(t *testing.T) {
	cases := []struct {
		state license.State
		text  string
	}{
		{license.StateValid, "valid"},
		{license.StateGrace, "grace"},
		{license.StateExpired, "expired"},
		{license.StateNotYetValid, "not_yet_valid"},
		{license.StateInvalid, "invalid"},
		{license.StateMissing, "missing"},
	}

	require.Equal(t, len(cases), len(license.StateValues()))
	require.Equal(t, []string{"valid", "grace", "expired", "not_yet_valid", "invalid", "missing"}, license.StateStrings())
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			require.Equal(t, tc.text, tc.state.String())
			require.True(t, tc.state.IsAState())
			parsed, err := license.StateString(tc.text)
			require.NoError(t, err)
			require.Equal(t, tc.state, parsed)
			encoded, err := tc.state.MarshalText()
			require.NoError(t, err)
			require.Equal(t, tc.text, string(encoded))
			var decoded license.State
			require.NoError(t, decoded.UnmarshalText(encoded))
			require.Equal(t, tc.state, decoded)
		})
	}

	invalid := license.State(0)
	require.False(t, invalid.IsAState())
	_, err := license.StateString("unknown")
	require.Error(t, err)
}

func TestVendorTrustKeyIDAndFingerprint(t *testing.T) {
	vendor, err := license.VendorTrust()
	require.NoError(t, err)
	require.Contains(t, strings.Join(vendor.KeyIDs(), ","), license.ProductionKeyID)
	publicKey, ok := vendor.PublicKey(license.ProductionKeyID)
	require.True(t, ok)
	require.Len(t, publicKey, ed25519.PublicKeySize)

	fingerprint := sha256.Sum256(publicKey)
	require.Equal(t, "f9a27c913cc8163bc43dba8b160b7fa301d56da276dedb84b01e7803e5be71f7", hex.EncodeToString(fingerprint[:]))

	_, err = license.Verify([]byte("{}"), nil)
	require.ErrorIs(t, err, license.ErrNoTrust)
}
