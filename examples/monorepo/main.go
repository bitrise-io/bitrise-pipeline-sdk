// This example generates a Bitrise configuration for a Go monorepo with
// multiple services. Each service has its own test workflow sharing a
// common setup, and a single deploy workflow publishes all services on
// release tags.
//
// Usage:
//
//	go run ./examples/monorepo/main.go
//	go run ./examples/monorepo/main.go | bitrise run --config -
package main

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
	// setup — checked out once; subsequent workflows use before_run to depend on it.
	setup := workflow.New().
		WithTitle("Shared Setup").
		AddStep(step.ActivateSshKey()).
		AddStep(step.GitClone()).
		AddStep(step.CachePull())

	// Per-service test workflows.
	serviceA := workflow.New().
		WithTitle("Service A Tests").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("cd services/a && go test ./...").WithTitle("Test Service A")).
		AddStep(step.CachePush())

	serviceB := workflow.New().
		WithTitle("Service B Tests").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("cd services/b && go test ./...").WithTitle("Test Service B")).
		AddStep(step.CachePush())

	serviceC := workflow.New().
		WithTitle("Service C Tests").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("cd services/c && go test ./...").WithTitle("Test Service C")).
		AddStep(step.CachePush())

	// deploy — runs all services after tests pass.
	deploy := workflow.New().
		WithTitle("Deploy All Services").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("make build-all").WithTitle("Build all services")).
		AddStep(step.Script().WithContent("make deploy").WithTitle("Deploy")).
		AddStep(step.DeployToBitriseIo())

	// Graph pipeline — service tests run in parallel, deploy waits for all.
	ci := graphpipeline.New().
		WithTitle("Monorepo CI").
		AddWorkflow("setup", graphpipeline.NewWorkflow()).
		AddWorkflow("service-a", graphpipeline.NewWorkflow().WithDependsOn("setup")).
		AddWorkflow("service-b", graphpipeline.NewWorkflow().WithDependsOn("setup")).
		AddWorkflow("service-c", graphpipeline.NewWorkflow().WithDependsOn("setup")).
		AddWorkflow("deploy", graphpipeline.NewWorkflow().
			WithDependsOn("service-a", "service-b", "service-c").
			WithAbortOnFail(true))

	cfg := pipeline.New("other").
		WithTitle("Monorepo CI/CD").
		AddWorkflow("setup", setup).
		AddWorkflow("service-a", serviceA).
		AddWorkflow("service-b", serviceB).
		AddWorkflow("service-c", serviceC).
		AddWorkflow("deploy", deploy).
		AddGraphPipeline("ci", ci).
		AddTrigger(trigger.OnPush("", "ci").WithBranch("*").Build()).
		AddTrigger(trigger.OnPullRequest("", "ci").WithTargetBranch("main").Build()).
		AddTrigger(trigger.OnTag("deploy", "").WithTag("release/*").Build())

	if err := serialize.ValidatedPrint(cfg.Build()); err != nil {
		log.Fatal(err)
	}
}
