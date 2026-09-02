package plugin

import (
	"archive/zip"
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteZipFile_RejectsUncompressedSizeMismatch(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	header := &zip.FileHeader{Name: "plugin.yaml", Method: zip.Store}
	w, err := zw.CreateHeader(header)
	require.NoError(t, err)
	_, err = w.Write([]byte("schemaVersion: 1\n"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	require.Len(t, reader.File, 1)
	reader.File[0].UncompressedSize64 = 1

	root, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = root.Close()
	})

	err = writeZipFile(root, reader.File[0], "plugin.yaml")
	require.Error(t, err)
}
