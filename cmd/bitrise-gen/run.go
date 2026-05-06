package main

import (
	"fmt"
	"os"
	"os/exec"
)

const runUsage = `Usage: bitrise-gen run <script.go> [-- extra args...]

Compile and run a pipeline script using 'go run'. The script's stdout is
streamed directly to stdout, making it easy to pipe into the Bitrise CLI:

  bitrise-gen run ./pipeline.go | bitrise run --config -

The script controls its own output. Use serialize.ValidatedPrint inside the
script to validate before printing, or use 'bitrise-gen validate' separately.
`

func runCommand(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(os.Stderr, runUsage)
		os.Exit(0)
	}

	scriptPath := args[0]
	extraArgs := args[1:]

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fatalf("script not found: %s", scriptPath)
	}

	goArgs := append([]string{"run", scriptPath}, extraArgs...)
	cmd := exec.Command("go", goArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fatalf("failed to run script: %v", err)
	}
}
