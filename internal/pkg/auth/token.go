package auth

import (
	"encoding/base64"
	"strings"

	"github.com/goccy/go-json"

	"github.com/opcotech/elemo/internal/pkg"
)

const (
	tokenSeparator = ";"
)

// GenerateToken creates a new token and returns both the unencrypted and
// encrypted pair.
func GenerateToken(kind string, data map[string]any) (string, string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", "", err
	}

	secret := pkg.GenerateRandomString(36)
	public := []byte(strings.Join([]string{kind, secret, string(jsonData)}, tokenSeparator))
	return base64.RawURLEncoding.EncodeToString(public), HashPassword(secret), nil
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

// IsTokenMatching validates if the provided token matches the original
// token hash.
func IsTokenMatching(hash, token string) bool {
	_, secret, _ := SplitToken(token)
	return IsPasswordMatching(hash, secret)
}
