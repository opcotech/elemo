package auth

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestIsTokenMatching(t *testing.T) {
	token, hash, err := GenerateToken(model.UserTokenContextConfirm.String(), map[string]any{
		"role_id": "member",
	})
	require.NoError(t, err)
	require.True(t, IsTokenMatching(hash, token))
	require.False(t, IsTokenMatching(hash, "not-matching"))

	decoded, err := base64.RawURLEncoding.DecodeString(token)
	require.NoError(t, err)
	tampered := strings.Replace(string(decoded), `"member"`, `"administrator"`, 1)
	require.NotEqual(t, string(decoded), tampered)
	tamperedToken := base64.RawURLEncoding.EncodeToString([]byte(tampered))
	require.False(t, IsTokenMatching(hash, tamperedToken))
}
