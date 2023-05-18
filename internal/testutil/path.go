package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// NewTempFile creates a new temporary file with the given name and content.
func NewTempFile(t *testing.T, name string, content string) string {
	file, err := os.CreateTemp("", name)
	require.NoError(t, err)

	_, err = file.WriteString(content)
	require.NoError(t, err)

	return file.Name()
}
