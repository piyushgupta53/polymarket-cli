package main

import (
	"os"

	"github.com/piyushgupta/polymarket-cli/cmd"
	"github.com/piyushgupta/polymarket-cli/internal/output"
)

// Set at build time via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		jsonMode := cmd.GetResolvedOutputFormat() == "json"
		output.PrintError(err, jsonMode)
		os.Exit(output.ExitCodeFromError(err))
	}
}
