package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteKeyPair(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	dir := t.TempDir()
	privatePath := filepath.Join(dir, "vendor.key")
	publicPath := filepath.Join(dir, "vendor.pub")
	require.NoError(t, writeKeyPair(privatePath, publicPath, priv, pub))

	for _, path := range []string{privatePath, publicPath} {
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}

	require.Error(t, writeKeyPair(privatePath, publicPath, priv, pub))
}

func TestWriteKeyPairRemovesPartialPrivateKey(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	dir := t.TempDir()
	privatePath := filepath.Join(dir, "vendor.key")
	publicPath := filepath.Join(dir, "vendor.pub")
	require.NoError(t, os.WriteFile(publicPath, []byte("existing"), 0o600))

	require.Error(t, writeKeyPair(privatePath, publicPath, priv, pub))
	_, err = os.Stat(privatePath)
	require.ErrorIs(t, err, os.ErrNotExist)
}
