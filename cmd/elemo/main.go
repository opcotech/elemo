package main

import (
	"runtime"
	"time"

	"github.com/opcotech/elemo/cmd/elemo/cli"
)

var (
	version   = "dev"
	commit    = "dirty"
	date      = time.Now().UTC().Format(time.RFC3339)
	goVersion = runtime.Version()
)

func main() {
	cli.Execute(version, commit, date, goVersion)
}
