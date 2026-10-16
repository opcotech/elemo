package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/goccy/go-json"

	"github.com/opcotech/elemo/internal/pkg"
)

const (
	tokenSeparator = ";"
)

// GenerateToken creates a bearer token and returns the public value plus a
// hash that authenticates the complete token, including all embedded claims.
func GenerateToken(kind string, data map[string]any) (string, string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", "", err
	}

	secret := pkg.GenerateRandomString(36)
	raw := []byte(strings.Join([]string{kind, secret, string(jsonData)}, tokenSeparator))
	public := base64.RawURLEncoding.EncodeToString(raw)
	return public, HashToken(public), nil
}

// SplitToken splits the token to the encapsulated data and secret.
func SplitToken(token string) (string, string, map[string]any) {
	decoded := make([]byte, base64.RawURLEncoding.DecodedLen(len(token)))
	_, _ = base64.RawURLEncoding.Decode(decoded, []byte(token))

	parts := strings.Split(string(decoded), tokenSeparator)
	if len(parts) < 3 {
		return "", "", nil
	}

	// Reassemble the data if it contained any token separators
	var data map[string]any
	if err := json.Unmarshal([]byte(strings.Join(parts[2:], tokenSeparator)), &data); err != nil {
		return "", "", nil
	}

	return parts[0], parts[1], data
}

// HashToken returns a password-grade hash of the complete bearer token.
func HashToken(token string) string {
	return HashPassword(tokenDigest(token))
}

// IsTokenMatching validates the complete token, including its claims, against
// the hash persisted at issuance.
func IsTokenMatching(hash, token string) bool {
	return IsPasswordMatching(hash, tokenDigest(token))
}

func tokenDigest(token string) string {
	digest := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}
