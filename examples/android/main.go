// This example generates a Bitrise configuration for an Android project.
// It runs unit tests on every push and pull request, and publishes
// an AAB to Bitrise artifacts on version tags.
//
// Usage:
//
//	go run ./examples/android/main.go
//	go run ./examples/android/main.go | bitrise run --config -
package main

import (
	"log"

	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/serialize"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/trigger"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func main() {
	// test — runs Android unit tests.
	test := workflow.New().
		WithTitle("Android Tests").
		AddStep(step.ActivateSshKey()).
		AddStep(step.GitClone()).
		AddStep(step.AndroidUnitTest().
			WithProjectLocation("./").
			WithModule("app").
			WithVariant("Debug"))

	// build — assembles a Release AAB and uploads to Bitrise.
	build := workflow.New().
		WithTitle("Android Build").
		WithBeforeRun("test").
		AddStep(step.AndroidBuild().
			WithProjectLocation("./").
			WithModule("app").
			WithVariant("Release").
			WithBuildType("aab")).
		AddStep(step.DeployToBitriseIo())

	cfg := pipeline.New("android").
		WithTitle("Android CI/CD").
		AddWorkflow("test", test).
		AddWorkflow("build", build).
		AddTrigger(trigger.OnPush("test", "").WithBranch("*").Build()).
		AddTrigger(trigger.OnPullRequest("test", "").WithTargetBranch("main").Build()).
		AddTrigger(trigger.OnTag("build", "").WithTag("v*").Build())

	if err := serialize.ValidatedPrint(cfg.Build()); err != nil {
		log.Fatal(err)
	}
}
