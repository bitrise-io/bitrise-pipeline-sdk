// Command bitrise-gen is a thin CLI wrapper for running and validating
// Bitrise pipeline scripts written with the bitrise-pipeline-sdk.
//
// Install:
//
//	go install github.com/bitrise-io/bitrise-pipeline-sdk/cmd/bitrise-gen@latest
//
// Usage:
//
//	bitrise-gen run      ./pipeline.go        # run script, print YAML to stdout
//	bitrise-gen validate ./pipeline.go        # run script, validate output, report result
//	bitrise-gen scaffold [--output=file.go]   # write a starter pipeline script
package main

import (
	"fmt"
	"os"
)

const usage = `bitrise-gen — Bitrise pipeline script runner

Usage:
  bitrise-gen <command> [flags]

Commands:
  run       <script.go>              Compile and run a pipeline script, stream YAML to stdout
  validate  <script.go>              Run a pipeline script and validate its YAML output
  scaffold  [--output=pipeline.go]   Write a starter pipeline script

Examples:
  bitrise-gen scaffold
  bitrise-gen run      ./pipeline.go | bitrise run --config -
  bitrise-gen validate ./pipeline.go
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCommand(os.Args[2:])
	case "validate":
		validateCommand(os.Args[2:])
	case "scaffold":
		scaffoldCommand(os.Args[2:])
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}
