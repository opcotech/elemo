package plugin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrManifestNotFound = errors.New("plugin.yaml not found")
	ErrManifestParse    = errors.New("failed to parse plugin manifest")
)

// ParseManifest decodes and validates a plugin.yaml document.
func ParseManifest(data []byte) (model.PluginManifest, error) {
	var manifest model.PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return model.PluginManifest{}, errors.Join(ErrManifestParse, err)
	}
	if err := manifest.Validate(); err != nil {
		return model.PluginManifest{}, err
	}
	return manifest, nil
}

// ParseManifestFile reads plugin.yaml from path.
func ParseManifestFile(path string) (model.PluginManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return model.PluginManifest{}, ErrManifestNotFound
		}
		return model.PluginManifest{}, errors.Join(ErrManifestParse, err)
	}
	return ParseManifest(data)
}

// ValidatePackageLayout checks declared backend/frontend files exist under root.
func ValidatePackageLayout(root string, manifest model.PluginManifest) error {
	if manifest.Backend != nil {
		entry := manifest.Backend.Entry
		if entry == "" {
			entry = model.PluginBackendWASMPath
		}
		if err := requireRelativeFile(root, entry); err != nil {
			return fmt.Errorf("backend entry: %w", err)
		}
	}
	if manifest.Frontend != nil {
		entry := manifest.Frontend.Entry
		if entry == "" {
			entry = model.PluginFrontendEntryDefault
		}
		if err := requireRelativeFile(root, entry); err != nil {
			return fmt.Errorf("frontend entry: %w", err)
		}
	}
	return nil
}

func requireRelativeFile(root, rel string) error {
	if rel == "" || strings.Contains(rel, "..") || filepath.IsAbs(rel) {
		return errors.New("invalid package path")
	}
	full := filepath.Join(root, filepath.FromSlash(rel))
	info, err := os.Stat(full)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("package path is a directory")
	}
	return nil
}
