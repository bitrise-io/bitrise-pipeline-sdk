// This example generates a Bitrise configuration with a graph pipeline and
// multiple workflows, validates it, then prints it as YAML to stdout.
//
// Usage:
//
//	go run ./examples/basic/main.go
//	go run ./examples/basic/main.go | bitrise run --config -
package main

import (
	"log"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"

	"github.com/bitrise-io/bitrise-pipeline-sdk/container"
	"github.com/bitrise-io/bitrise-pipeline-sdk/graphpipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/serialize"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/trigger"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func main() {
	// Shared setup workflow — runs before the main jobs.
	setup := workflow.New().
		WithTitle("Setup").
		AddStep(step.ActivateSSHKey()).
		AddStep(step.GitClone()).
		AddStep(step.CachePull())

	// Test workflow — runs unit tests.
	test := workflow.New().
		WithTitle("Test").
		WithBeforeRun("setup").
		AddStep(step.Script("go test ./...").WithTitle("Run unit tests")).
		AddStep(step.CachePush())

	// Deploy workflow — builds and uploads artifacts.
	deploy := workflow.New().
		WithTitle("Deploy").
		WithBeforeRun("setup").
		AddStep(step.Script("go build -o bin/app ./cmd/app").WithTitle("Build binary")).
		AddStep(step.DeployToBitriseIO())

	// Graph pipeline — test and deploy run in parallel after setup.
	ci := graphpipeline.New().
		WithTitle("CI Pipeline").
		AddWorkflow("setup", graphpipeline.NewWorkflow()).
		AddWorkflow("test", graphpipeline.NewWorkflow().WithDependsOn("setup")).
		AddWorkflow("deploy", graphpipeline.NewWorkflow().
			WithDependsOn("test").
			WithAbortOnFail(true))

	// Postgres service container used by integration tests.
	postgres := container.NewService("postgres:15").
		WithPort("5432:5432").
		WithEnv("POSTGRES_PASSWORD", "secret")

	cfg := pipeline.New("other").
		WithTitle("My App CI/CD").
		WithAppEnv("APP_ENV", "ci").
		AddWorkflow("setup", setup).
		AddWorkflow("test", test).
		AddWorkflow("deploy", deploy).
		AddGraphPipeline("ci", ci).
		AddContainer("postgres", postgres).
		WithTool(bitriseModels.ToolID("golang"), "1.22").
		AddTrigger(trigger.OnPush("", "ci").WithBranch("main").Build()).
		AddTrigger(trigger.OnPullRequest("", "ci").WithTargetBranch("main").Build()).
		AddTrigger(trigger.OnTag("deploy", "").WithTag("v*").Build())

	// Validate before serializing — warnings go to stderr, errors abort.
	if err := serialize.ValidatedPrint(cfg.Build()); err != nil {
		log.Fatal(err)
	}
}
