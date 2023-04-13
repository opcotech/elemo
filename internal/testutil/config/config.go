package config

import (
	"path/filepath"
	"runtime"
)

var (
	_, f, _, _ = runtime.Caller(0)
	RootDir    = filepath.Join(filepath.Dir(f), "..", "..", "..")
)
