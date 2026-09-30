package license

import "time"

const (
	// FormatV1 is the only supported License envelope format.
	FormatV1 = "elemo.airgap.license.v1"

	// ProductionKeyID is the key identifier of the first vendor signing key.
	ProductionKeyID = "elemo-airgap-2026"

	// GracePeriod is the fixed window after expires_at during which
	// seat-limited activations remain allowed.
	GracePeriod = 30 * 24 * time.Hour

	// NotBeforeSkew is the allowed clock skew when evaluating not_before.
	NotBeforeSkew = 5 * time.Minute

	// MaxLicenseBytes is the maximum accepted license file size.
	MaxLicenseBytes = 64 * 1024

	signedMessagePrefix = FormatV1
)

var (
	envelopeFieldOrder = []string{"format", "key_id", "payload", "signature"}
	payloadFieldOrder  = []string{
		"id",
		"customer",
		"installation_id",
		"seats",
		"issued_at",
		"not_before",
		"expires_at",
	}
)

// Envelope is the signed license wrapper. Payload and Signature are standard
// base64 encodings of the exact payload bytes and the Ed25519 signature.
type Envelope struct {
	Format    string `json:"format"`
	KeyID     string `json:"key_id"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

// Payload is the unsigned AirGap entitlement. Times are RFC3339 UTC.
type Payload struct {
	ID             string    `json:"id"`
	Customer       string    `json:"customer"`
	InstallationID string    `json:"installation_id"`
	Seats          uint32    `json:"seats"`
	IssuedAt       time.Time `json:"issued_at"`
	NotBefore      time.Time `json:"not_before"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// License is a verified AirGap entitlement.
type License struct {
	KeyID   string
	Payload Payload
}

func signedMessage(keyID string, payload []byte) []byte {
	msg := make([]byte, 0, len(signedMessagePrefix)+1+len(keyID)+1+len(payload))
	msg = append(msg, signedMessagePrefix...)
	msg = append(msg, '|')
	msg = append(msg, keyID...)
	msg = append(msg, '|')
	msg = append(msg, payload...)
	return msg
}
