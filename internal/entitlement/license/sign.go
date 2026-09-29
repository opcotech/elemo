package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type signedPayload struct {
	ID             string `json:"id"`
	Customer       string `json:"customer"`
	InstallationID string `json:"installation_id"`
	Seats          uint32 `json:"seats"`
	IssuedAt       string `json:"issued_at"`
	NotBefore      string `json:"not_before"`
	ExpiresAt      string `json:"expires_at"`
}

func marshalPayload(p Payload) ([]byte, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	return json.Marshal(signedPayload{
		ID:             p.ID,
		Customer:       p.Customer,
		InstallationID: p.InstallationID,
		Seats:          p.Seats,
		IssuedAt:       p.IssuedAt.UTC().Format(time.RFC3339),
		NotBefore:      p.NotBefore.UTC().Format(time.RFC3339),
		ExpiresAt:      p.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

// Sign produces a verified-format envelope for the payload using the given
// private key. The private key is supplied by the caller and is never read
// from the repository or configuration.
func Sign(keyID string, privateKey ed25519.PrivateKey, payload Payload) ([]byte, error) {
	if keyID == "" {
		return nil, fmt.Errorf("%w: missing key id", ErrLicenseMalformed)
	}
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("%w: invalid private key", ErrLicenseInvalidSignature)
	}

	payloadBytes, err := marshalPayload(payload)
	if err != nil {
		return nil, err
	}

	sig := ed25519.Sign(privateKey, signedMessage(keyID, payloadBytes))
	envelope := Envelope{
		Format:    FormatV1,
		KeyID:     keyID,
		Payload:   base64.StdEncoding.EncodeToString(payloadBytes),
		Signature: base64.StdEncoding.EncodeToString(sig),
	}
	return json.Marshal(envelope)
}

// ReadFile reads a license without allocating beyond the accepted size.
func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	raw, err := io.ReadAll(io.LimitReader(file, MaxLicenseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > MaxLicenseBytes {
		return nil, ErrLicenseTooLarge
	}
	return raw, nil
}

const licenseFilePerm os.FileMode = 0o600

// WriteFile atomically writes signed license bytes with restrictive permissions.
func WriteFile(path string, raw []byte) (err error) {
	dir := filepath.Dir(path)
	file, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	closed := false
	defer func() {
		if !closed {
			if closeErr := file.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
		}
		if removeErr := os.Remove(tempPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) && err == nil {
			err = removeErr
		}
	}()

	if err = file.Chmod(licenseFilePerm); err != nil {
		return err
	}
	content := append(append([]byte(nil), raw...), '\n')
	if _, err = file.Write(content); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	closed = true
	return os.Rename(tempPath, path)
}
