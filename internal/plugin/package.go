package plugin

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/opcotech/elemo/internal/model"
)

const (
	maxPackageBytes = 32 << 20
	maxFileBytes    = 24 << 20
	maxZipEntries   = 4096
)

var (
	ErrPackageInvalid  = errors.New("invalid plugin package")
	ErrPackageTooLarge = errors.New("plugin package is too large")
)

// ExtractedPackage is a validated plugin on disk.
type ExtractedPackage struct {
	Root     string
	Manifest model.PluginManifest
}

// InstallDirectory returns {directory}/{pluginID}/{version}.
func InstallDirectory(base, pluginID, version string) string {
	return filepath.Join(base, pluginID, version)
}

// ExtractZip unpacks a plugin zip into dest, validating paths and the manifest.
func ExtractZip(zipBytes []byte, dest string) (ExtractedPackage, error) {
	if len(zipBytes) == 0 {
		return ExtractedPackage{}, ErrPackageInvalid
	}
	if len(zipBytes) > maxPackageBytes {
		return ExtractedPackage{}, ErrPackageTooLarge
	}

	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return ExtractedPackage{}, errors.Join(ErrPackageInvalid, err)
	}
	if len(reader.File) == 0 || len(reader.File) > maxZipEntries {
		return ExtractedPackage{}, ErrPackageInvalid
	}

	if err := os.MkdirAll(dest, 0o750); err != nil {
		return ExtractedPackage{}, err
	}
	root, err := os.OpenRoot(dest)
	if err != nil {
		return ExtractedPackage{}, err
	}
	defer root.Close()

	var foundManifest bool
	for _, file := range reader.File {
		name := strings.ReplaceAll(file.Name, "\\", "/")
		name = strings.TrimPrefix(name, "./")
		if name == "" || strings.HasSuffix(name, "/") {
			continue
		}
		if strings.HasPrefix(filepath.Base(name), ".") {
			continue
		}
		if !filepath.IsLocal(name) {
			return ExtractedPackage{}, ErrPackageInvalid
		}
		if !file.Mode().IsRegular() {
			return ExtractedPackage{}, ErrPackageInvalid
		}
		if file.UncompressedSize64 > maxFileBytes {
			return ExtractedPackage{}, ErrPackageTooLarge
		}
		if err := writeZipFile(root, file, filepath.FromSlash(name)); err != nil {
			return ExtractedPackage{}, errors.Join(ErrPackageInvalid, err)
		}
		if filepath.Base(name) == model.PluginManifestFileName {
			foundManifest = true
		}
	}
	if !foundManifest {
		return ExtractedPackage{}, ErrManifestNotFound
	}

	manifestPath := filepath.Join(dest, model.PluginManifestFileName)
	if _, err := os.Stat(manifestPath); err != nil {
		matches, walkErr := filepath.Glob(filepath.Join(dest, "*", model.PluginManifestFileName))
		if walkErr != nil || len(matches) != 1 {
			return ExtractedPackage{}, ErrManifestNotFound
		}
		manifestPath = matches[0]
		dest = filepath.Dir(manifestPath)
	}

	manifest, err := ParseManifestFile(manifestPath)
	if err != nil {
		return ExtractedPackage{}, err
	}
	if err := ValidatePackageLayout(dest, manifest); err != nil {
		return ExtractedPackage{}, errors.Join(ErrPackageInvalid, err)
	}

	return ExtractedPackage{Root: dest, Manifest: manifest}, nil
}

func writeZipFile(root *os.Root, file *zip.File, name string) (err error) {
	dir := filepath.Dir(name)
	if dir != "." {
		if err := root.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o640)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dst.Close(); err == nil {
			err = cerr
		}
	}()

	if file.UncompressedSize64 > uint64(math.MaxInt64)-1 {
		return ErrPackageTooLarge
	}
	n, copyErr := io.CopyN(dst, src, int64(file.UncompressedSize64)+1) //nolint:gosec // UncompressedSize64 is capped below MaxInt64
	if copyErr != nil && !errors.Is(copyErr, io.EOF) {
		return copyErr
	}
	if n > int64(file.UncompressedSize64) {
		return ErrPackageTooLarge
	}
	return nil
}

// ReadWASM loads the backend wasm from an extracted package.
func ReadWASM(pkg ExtractedPackage) ([]byte, error) {
	if pkg.Manifest.Backend == nil {
		return nil, nil
	}
	entry := pkg.Manifest.Backend.Entry
	if entry == "" {
		entry = model.PluginBackendWASMPath
	}
	path := filepath.Join(pkg.Root, filepath.FromSlash(entry))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read wasm: %w", err)
	}
	return data, nil
}

// RemoveVersion deletes a plugin version directory.
func RemoveVersion(base, pluginID, version string) error {
	dir := InstallDirectory(base, pluginID, version)
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	parent := filepath.Join(base, pluginID)
	entries, err := os.ReadDir(parent)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(entries) == 0 {
		return os.Remove(parent)
	}
	return nil
}
