package safepath

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrEmptyPath     = errors.New("path is empty")
	ErrAbsolutePath  = errors.New("absolute paths are not allowed")
	ErrInvalidPath   = errors.New("invalid path")
	ErrPathTraversal = errors.New("path escapes root")
)

// Normalize validates a user-supplied relative filesystem path and returns its
// normalized absolute path underneath root.
//
// Examples, given root "/srv/uploads":
//
//	"foo/bar.txt"     -> "/srv/uploads/foo/bar.txt"
//	"foo/../bar.txt"  -> "/srv/uploads/bar.txt"
//	"../etc/passwd"   -> ErrPathTraversal
//	"/etc/passwd"     -> ErrAbsolutePath
func Normalize(root, input string) (string, error) {
	if input == "" {
		return "", ErrEmptyPath
	}

	// NUL cannot appear in valid OS paths and can cause problems when paths
	// cross API boundaries.
	if strings.IndexByte(input, 0) >= 0 {
		return "", ErrInvalidPath
	}

	if filepath.IsAbs(input) {
		return "", ErrAbsolutePath
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	root = filepath.Clean(root)

	// Clean removes ".", duplicate separators, and resolves ".."
	// lexically.
	clean := filepath.Clean(input)

	candidate := filepath.Join(root, clean)

	// Rel is safer than checking strings.HasPrefix(candidate, root).
	// For example, "/srv/uploads2" must not count as being inside
	// "/srv/uploads".
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return "", ErrInvalidPath
	}

	if rel == ".." ||
		strings.HasPrefix(rel, ".."+string(filepath.Separator)) ||
		filepath.IsAbs(rel) {
		return "", ErrPathTraversal
	}

	return candidate, nil
}
