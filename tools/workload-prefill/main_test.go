package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultQueriesDirectoryContainsBootstrapScripts(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))

	for _, name := range []string{"bootstrap.cypher", "bootstrap.sql"} {
		info, err := os.Stat(filepath.Join(root, defaultQueriesDir, name))
		require.NoError(t, err)
		require.False(t, info.IsDir())
	}
}
