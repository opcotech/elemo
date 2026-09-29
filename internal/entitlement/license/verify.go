package license

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

// Verify parses and verifies an License envelope against trust.
// Signature verification happens before the payload is interpreted.
func Verify(raw []byte, trust Trust) (License, error) {
	if len(raw) == 0 {
		return License{}, ErrLicenseMissing
	}
	if len(raw) > MaxLicenseBytes {
		return License{}, ErrLicenseTooLarge
	}
	if trust == nil {
		return License{}, ErrNoTrust
	}

	trimmed := bytes.TrimSpace(raw)
	if err := requireObjectKeyOrder(trimmed, envelopeFieldOrder); err != nil {
		return License{}, err
	}

	var envelope Envelope
	if err := strictDecode(trimmed, &envelope); err != nil {
		return License{}, err
	}
	if envelope.Format != FormatV1 {
		return License{}, fmt.Errorf("%w: unsupported format %q", ErrLicenseMalformed, envelope.Format)
	}
	if envelope.KeyID == "" {
		return License{}, fmt.Errorf("%w: missing key id", ErrLicenseMalformed)
	}

	pub, ok := trust.PublicKey(envelope.KeyID)
	if !ok {
		return License{}, fmt.Errorf("%w: %s", ErrLicenseUnknownKey, envelope.KeyID)
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		return License{}, fmt.Errorf("%w: payload is not standard base64", ErrLicenseMalformed)
	}
	signature, err := base64.StdEncoding.DecodeString(envelope.Signature)
	if err != nil {
		return License{}, fmt.Errorf("%w: signature is not standard base64", ErrLicenseMalformed)
	}
	if len(signature) != ed25519.SignatureSize {
		return License{}, fmt.Errorf("%w: signature has invalid length", ErrLicenseInvalidSignature)
	}

	if !ed25519.Verify(pub, signedMessage(envelope.KeyID, payloadBytes), signature) {
		return License{}, ErrLicenseInvalidSignature
	}

	if err := requireObjectKeyOrder(payloadBytes, payloadFieldOrder); err != nil {
		return License{}, err
	}

	var payload Payload
	if err := strictDecode(payloadBytes, &payload); err != nil {
		return License{}, errors.Join(ErrLicenseInvalidPayload, err)
	}
	if err := payload.validate(); err != nil {
		return License{}, errors.Join(ErrLicenseInvalidPayload, err)
	}

	return License{KeyID: envelope.KeyID, Payload: payload}, nil
}

func (p Payload) validate() error {
	if _, err := uuid.Parse(p.ID); err != nil {
		return fmt.Errorf("id must be a UUID: %w", err)
	}
	customer := strings.TrimSpace(p.Customer)
	if customer == "" || customer != p.Customer {
		return errors.New("customer is required")
	}
	if len(p.Customer) > 256 {
		return errors.New("customer exceeds 256 characters")
	}
	if _, err := uuid.Parse(p.InstallationID); err != nil {
		return fmt.Errorf("installation_id must be a UUID: %w", err)
	}
	if p.Seats < 1 {
		return errors.New("seats must be at least 1")
	}
	if p.IssuedAt.IsZero() || p.NotBefore.IsZero() || p.ExpiresAt.IsZero() {
		return errors.New("issued_at, not_before, and expires_at are required")
	}
	issued := p.IssuedAt.UTC()
	notBefore := p.NotBefore.UTC()
	expires := p.ExpiresAt.UTC()
	if !notBefore.Before(expires) {
		return errors.New("not_before must be before expires_at")
	}
	if !issued.Before(expires) {
		return errors.New("issued_at must be before expires_at")
	}
	return nil
}

func strictDecode(raw []byte, dest any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		if isUnknownFieldError(err) {
			return errors.Join(ErrLicenseUnknownField, err)
		}
		return errors.Join(ErrLicenseMalformed, err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: trailing data", ErrLicenseMalformed)
	}
	return nil
}

func isUnknownFieldError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unknown field")
}

func requireObjectKeyOrder(raw []byte, expected []string) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return errors.Join(ErrLicenseMalformed, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return fmt.Errorf("%w: expected JSON object", ErrLicenseMalformed)
	}

	i := 0
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return errors.Join(ErrLicenseMalformed, err)
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("%w: expected object key", ErrLicenseMalformed)
		}
		if i >= len(expected) || key != expected[i] {
			return fmt.Errorf("%w: expected %q at position %d, got %q", ErrLicenseFieldOrder, expectedAt(expected, i), i, key)
		}
		if err := skipValue(dec); err != nil {
			return err
		}
		i++
	}
	if i != len(expected) {
		return fmt.Errorf("%w: missing fields after %d keys", ErrLicenseFieldOrder, i)
	}

	end, err := dec.Token()
	if err != nil {
		return errors.Join(ErrLicenseMalformed, err)
	}
	if delim, ok := end.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("%w: expected end of object", ErrLicenseMalformed)
	}
	return nil
}

func expectedAt(expected []string, i int) string {
	if i < len(expected) {
		return expected[i]
	}
	return "<end>"
}

func skipValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return errors.Join(ErrLicenseMalformed, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	for dec.More() {
		if delim == '{' {
			if _, err := dec.Token(); err != nil {
				return errors.Join(ErrLicenseMalformed, err)
			}
		}
		if err := skipValue(dec); err != nil {
			return err
		}
	}
	_, err = dec.Token()
	if err != nil {
		return errors.Join(ErrLicenseMalformed, err)
	}
	return nil
}
