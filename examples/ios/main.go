// This example generates a Bitrise configuration for an iOS project.
// It runs Xcode tests on every push, archives on version tags, and
// sends a Slack notification on failure.
//
// Usage:
//
//	go run ./examples/ios/main.go
//	go run ./examples/ios/main.go | bitrise run --config -
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
	// test — runs Xcode tests on the simulator.
	test := workflow.New().
		WithTitle("iOS Tests").
		AddStep(step.ActivateSSHKey()).
		AddStep(step.GitClone()).
		AddStep(step.XcodeTest().
			WithScheme("MyApp").
			WithProjectPath("MyApp.xcworkspace").
			WithDestination("platform=iOS Simulator,name=iPhone 15,OS=latest"))

	// archive — builds and uploads an App Store IPA on version tags.
	archive := workflow.New().
		WithTitle("iOS Archive").
		WithBeforeRun("test").
		AddStep(step.XcodeArchive().
			WithScheme("MyApp").
			WithProjectPath("MyApp.xcworkspace").
			WithDistributionMethod("app-store").
			WithConfiguration("Release").
			WithAutomaticCodeSigning("api-key")).
		AddStep(step.DeployToBitriseIO())

	cfg := pipeline.New("ios").
		WithTitle("iOS CI/CD").
		AddWorkflow("test", test).
		AddWorkflow("archive", archive).
		AddTrigger(trigger.OnPush("test", "").WithBranch("*").Build()).
		AddTrigger(trigger.OnPullRequest("test", "").WithTargetBranch("main").Build()).
		AddTrigger(trigger.OnTag("archive", "").WithTag("v*").Build())

	if err := serialize.ValidatedPrint(cfg.Build()); err != nil {
		log.Fatal(err)
	}
}
