package safepath

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	absRoot, err := filepath.Abs(root)
	require.NoError(t, err)
	absRoot = filepath.Clean(absRoot)

	tests := []struct {
		name    string
		root    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "simple relative file",
			root:  root,
			input: "foo/bar.txt",
			want:  filepath.Join(absRoot, "foo", "bar.txt"),
		},
		{
			name:  "nested directories",
			root:  root,
			input: "a/b/c/d.txt",
			want:  filepath.Join(absRoot, "a", "b", "c", "d.txt"),
		},
		{
			name:  "lexical parent stays inside root",
			root:  root,
			input: "foo/../bar.txt",
			want:  filepath.Join(absRoot, "bar.txt"),
		},
		{
			name:  "dot segment is cleaned",
			root:  root,
			input: "foo/./bar.txt",
			want:  filepath.Join(absRoot, "foo", "bar.txt"),
		},
		{
			name:  "duplicate separators are cleaned",
			root:  root,
			input: "foo//bar.txt",
			want:  filepath.Join(absRoot, "foo", "bar.txt"),
		},
		{
			name:  "leading dot slash",
			root:  root,
			input: "./foo.txt",
			want:  filepath.Join(absRoot, "foo.txt"),
		},
		{
			name:  "current directory resolves to root",
			root:  root,
			input: ".",
			want:  absRoot,
		},
		{
			name:  "parent after child resolves to root",
			root:  root,
			input: "foo/..",
			want:  absRoot,
		},
		{
			name:  "hidden file",
			root:  root,
			input: ".gitignore",
			want:  filepath.Join(absRoot, ".gitignore"),
		},
		{
			name:  "double dots in filename are not traversal",
			root:  root,
			input: "foo..bar.txt",
			want:  filepath.Join(absRoot, "foo..bar.txt"),
		},
		{
			name:  "spaces in path",
			root:  root,
			input: "my files/my file.txt",
			want:  filepath.Join(absRoot, "my files", "my file.txt"),
		},
		{
			name:  "unicode path",
			root:  root,
			input: "földer/файл.txt",
			want:  filepath.Join(absRoot, "földer", "файл.txt"),
		},
		{
			name:    "empty path",
			root:    root,
			input:   "",
			wantErr: ErrEmptyPath,
		},
		{
			name:    "nul byte",
			root:    root,
			input:   "foo\x00bar.txt",
			wantErr: ErrInvalidPath,
		},
		{
			name:    "absolute path",
			root:    root,
			input:   "/etc/passwd",
			wantErr: ErrAbsolutePath,
		},
		{
			name:    "path traversal to parent",
			root:    root,
			input:   "..",
			wantErr: ErrPathTraversal,
		},
		{
			name:    "path traversal to sibling of root",
			root:    root,
			input:   "../etc/passwd",
			wantErr: ErrPathTraversal,
		},
		{
			name:    "path traversal nested",
			root:    root,
			input:   "foo/../../etc/passwd",
			wantErr: ErrPathTraversal,
		},
		{
			name:    "path traversal with extra dots",
			root:    root,
			input:   "foo/bar/../../../secret",
			wantErr: ErrPathTraversal,
		},
		{
			name:  "godoc example simple path",
			root:  "/srv/uploads",
			input: "foo/bar.txt",
			want:  "/srv/uploads/foo/bar.txt",
		},
		{
			name:  "godoc example cleaned parent",
			root:  "/srv/uploads",
			input: "foo/../bar.txt",
			want:  "/srv/uploads/bar.txt",
		},
		{
			name:    "godoc example path traversal",
			root:    "/srv/uploads",
			input:   "../etc/passwd",
			wantErr: ErrPathTraversal,
		},
		{
			name:    "godoc example absolute path",
			root:    "/srv/uploads",
			input:   "/etc/passwd",
			wantErr: ErrAbsolutePath,
		},
		{
			name:  "root with trailing separator",
			root:  "/srv/uploads/",
			input: "foo.txt",
			want:  "/srv/uploads/foo.txt",
		},
		{
			name:  "filesystem root cannot be escaped",
			root:  "/",
			input: "../etc/passwd",
			want:  "/etc/passwd",
		},
		{
			name:  "filesystem root current directory",
			root:  "/",
			input: ".",
			want:  "/",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Normalize(tt.root, tt.input)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalize_PrefixDoesNotMatchNeighbor(t *testing.T) {
	t.Parallel()

	root := "/srv/uploads"
	got, err := Normalize(root, "foo.txt")
	require.NoError(t, err)
	assert.Equal(t, "/srv/uploads/foo.txt", got)
	assert.False(t, strings.HasPrefix(got, "/srv/uploads2"))
}

func FuzzNormalize(f *testing.F) {
	root := f.TempDir()

	f.Add("foo/bar.txt")
	f.Add("")
	f.Add(".")
	f.Add("..")
	f.Add("../etc/passwd")
	f.Add("/etc/passwd")
	f.Add("foo/../bar")
	f.Add("foo/../../bar")
	f.Add("./foo")
	f.Add("foo//bar")
	f.Add("foo\x00bar")
	f.Add("foo..bar")
	f.Add(".hidden")

	f.Fuzz(func(t *testing.T, input string) {
		got, err := Normalize(root, input)
		if err != nil {
			assert.Empty(t, got)
			return
		}

		absRoot, absErr := filepath.Abs(root)
		require.NoError(t, absErr)
		absRoot = filepath.Clean(absRoot)

		require.True(t, filepath.IsAbs(got), "successful result must be absolute")

		rel, relErr := filepath.Rel(absRoot, got)
		require.NoError(t, relErr)
		assert.NotEqual(t, "..", rel)
		assert.False(t, strings.HasPrefix(rel, ".."+string(filepath.Separator)))
		assert.False(t, filepath.IsAbs(rel), "relative path from root must not be absolute")
	})
}
