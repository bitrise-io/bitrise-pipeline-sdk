package main

import (
	"fmt"
	"os"
	"os/exec"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/bitrise-pipeline-sdk/validate"
)

const validateUsage = `Usage: bitrise-gen validate <script.go>

Run a pipeline script, capture its YAML output, and validate it using the
full bitrise-pipeline-sdk validation pipeline (structural checks + upstream
bitrise/v2 normalization and validation).

Warnings are printed to stderr. Errors are printed to stderr and the command
exits with code 1.

Examples:
  bitrise-gen validate ./pipeline.go
  bitrise-gen validate ./pipeline.go && echo "ready to deploy"
`

func validateCommand(args []string) {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(os.Stderr, validateUsage)
		os.Exit(0)
	}

	scriptPath := args[0]

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fatalf("script not found: %s", scriptPath)
	}

	// Run the script: capture stdout (YAML), pass stderr through.
	cmd := exec.Command("go", "run", scriptPath)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "error: script exited with code %d\n", exitErr.ExitCode())
			os.Exit(exitErr.ExitCode())
		}
		fatalf("failed to run script: %v", err)
	}

	if len(output) == 0 {
		fatalf("script produced no output")
	}

	// Parse the captured YAML.
	var data bitriseModels.BitriseDataModel
	if err := yaml.Unmarshal(output, &data); err != nil {
		fatalf("failed to parse YAML output: %v", err)
	}

	// Run full validation.
	result, err := validate.Full(data)
	if err != nil {
		fatalf("validation error: %v", err)
	}

	// Print warnings to stderr.
	if len(result.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "%d warning(s):\n", len(result.Warnings))
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", w)
		}
	}

	if !result.IsValid() {
		fmt.Fprintf(os.Stderr, "\nConfig has %d error(s):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e.Error())
		}
		os.Exit(1)
	}

	msg := "Config is valid"
	if len(result.Warnings) > 0 {
		msg += fmt.Sprintf(" (%d warning(s))", len(result.Warnings))
	}
	fmt.Println(msg)
}
