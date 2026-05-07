package main

import (
	"flag"
	"fmt"
	"os"
)

const scaffoldUsage = `Usage: bitrise-gen scaffold [--output=pipeline.go]

Write a starter pipeline script to the given path. The generated file is a
complete, runnable Go program that uses the bitrise-pipeline-sdk to build and
print a basic Bitrise configuration.

Flags:
  --output string   output file path (default "pipeline.go")

Example:
  bitrise-gen scaffold
  bitrise-gen scaffold --output=ci/pipeline.go
  bitrise-gen run ./pipeline.go | bitrise run --config -
`

// scaffoldTemplate is the content written by 'bitrise-gen scaffold'.
// It is a complete, runnable pipeline script using the SDK.
const scaffoldTemplate = `package main

import (
	"log"

	"github.com/bitrise-io/bitrise-pipeline-sdk/graphpipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/serialize"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/trigger"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func main() {
	primary := workflow.New().
		WithTitle("Primary").
		AddStep(step.ActivateSshKey()).
		AddStep(step.GitClone()).
		AddStep(step.Script().WithContent("echo \"add your build steps here\""))

	ci := graphpipeline.New().
		WithTitle("CI").
		AddWorkflow("primary", graphpipeline.NewWorkflow())

	cfg := pipeline.New("other").
		WithTitle("My Pipeline").
		AddWorkflow("primary", primary).
		AddGraphPipeline("ci", ci).
		AddTrigger(trigger.OnPush("", "ci").WithBranch("main").Build())

	if err := serialize.ValidatedPrint(cfg.Build()); err != nil {
		log.Fatal(err)
	}
}
`

func scaffoldCommand(args []string) {
	fs := flag.NewFlagSet("scaffold", flag.ExitOnError)
	outputPath := fs.String("output", "pipeline.go", "output file path")
	fs.Usage = func() { fmt.Fprint(os.Stderr, scaffoldUsage) }

	if err := fs.Parse(args); err != nil {
		fatalf("%v", err)
	}

	if len(fs.Args()) > 0 && (fs.Args()[0] == "--help" || fs.Args()[0] == "-h") {
		fmt.Fprint(os.Stderr, scaffoldUsage)
		os.Exit(0)
	}

	if _, err := os.Stat(*outputPath); err == nil {
		fatalf("%s already exists — use --output to choose a different path", *outputPath)
	}

	if err := os.WriteFile(*outputPath, []byte(scaffoldTemplate), 0644); err != nil {
		fatalf("could not write file: %v", err)
	}

	fmt.Printf("Created %s\n\n", *outputPath)
	fmt.Printf("Make sure your go.mod requires:\n")
	fmt.Printf("  github.com/bitrise-io/bitrise-pipeline-sdk\n\n")
	fmt.Printf("Then run:\n")
	fmt.Printf("  bitrise-gen run %s | bitrise run --config -\n", *outputPath)
}
