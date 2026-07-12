package main

import (
	"os"

	"github.com/emulith/emulith/internal/cli"
)

var (
	version = "dev"
	commit  = "unknown"
	built   = "unknown"
)

func main() {
	if err := cli.Execute(os.Stdout, os.Stderr, version, commit, built); err != nil {
		os.Exit(1)
	}
}
