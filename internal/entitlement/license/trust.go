package license

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io/fs"
	"strings"
	"sync"
)

// Trust maps key IDs to Ed25519 public keys. Production binaries embed the
// vendor keyring. Tests inject an isolated Trust and never reuse production
// private keys.
type Trust map[string]ed25519.PublicKey

func (t Trust) PublicKey(keyID string) (ed25519.PublicKey, bool) {
	if t == nil {
		return nil, false
	}
	key, ok := t[keyID]
	return key, ok
}

func (t Trust) KeyIDs() []string {
	ids := make([]string, 0, len(t))
	for id := range t {
		ids = append(ids, id)
	}
	return ids
}

var (
	vendorTrustOnce sync.Once
	vendorTrust     Trust
	errVendorTrust  error
)

// VendorTrust returns the embedded production keyring. Test keys are not
// compiled into this set.
func VendorTrust() (Trust, error) {
	vendorTrustOnce.Do(func() {
		vendorTrust, errVendorTrust = loadTrust(vendorKeyFS, "vendorkeys")
	})
	if errVendorTrust != nil {
		return nil, errVendorTrust
	}
	return cloneTrust(vendorTrust), nil
}

func loadTrust(files fs.FS, dir string) (Trust, error) {
	entries, err := fs.ReadDir(files, dir)
	if err != nil {
		return nil, err
	}

	trust := make(Trust)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pub") {
			continue
		}
		keyID := strings.TrimSuffix(entry.Name(), ".pub")
		raw, err := fs.ReadFile(files, dir+"/"+entry.Name())
		if err != nil {
			return nil, err
		}
		hexKey := strings.TrimSpace(string(raw))
		decoded, err := hex.DecodeString(hexKey)
		if err != nil {
			return nil, fmt.Errorf("decode vendor public key %s: %w", keyID, err)
		}
		if len(decoded) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("vendor public key %s has invalid length %d", keyID, len(decoded))
		}
		if _, exists := trust[keyID]; exists {
			return nil, fmt.Errorf("duplicate vendor key id %s", keyID)
		}
		trust[keyID] = ed25519.PublicKey(decoded)
	}
	if len(trust) == 0 {
		return nil, fmt.Errorf("vendor trust store is empty")
	}
	return trust, nil
}

func cloneTrust(src Trust) Trust {
	out := make(Trust, len(src))
	for id, key := range src {
		out[id] = append(ed25519.PublicKey(nil), key...)
	}
	return out
}
