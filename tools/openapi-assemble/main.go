// Command openapi-assemble merges split OpenAPI YAML fragments into a single spec.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	srcDir := flag.String("src", "api/openapi/src", "Directory containing OpenAPI fragments")
	outPath := flag.String("out", "api/openapi/openapi.yaml", "Assembled OpenAPI output path")
	splitFrom := flag.String("split-from", "", "Optional bundled spec to extract into -src before assembling")
	flag.Parse()

	if err := run(*srcDir, *outPath, *splitFrom); err != nil {
		fmt.Fprintf(os.Stderr, "openapi-assemble: %v\n", err)
		os.Exit(1)
	}
}

func run(srcDir, outPath, splitFrom string) error {
	srcDir = resolveRepoPath(srcDir)
	outPath = resolveRepoPath(outPath)
	splitFrom = resolveRepoPath(splitFrom)
	layout := elemoLayout()

	if splitFrom != "" {
		if err := split(splitFrom, srcDir, layout); err != nil {
			return fmt.Errorf("split: %w", err)
		}
	}

	node, err := assemble(srcDir, layout)
	if err != nil {
		return err
	}

	return writeBundle(outPath, node)
}

// repoRoot is the Elemo repository root. Relative -src/-out flags are resolved
// from there so `go -C tools/openapi-assemble` does not look inside the tool dir.
func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func resolveRepoPath(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	root := repoRoot()
	if root == "" {
		return p
	}
	return filepath.Join(root, p)
}
