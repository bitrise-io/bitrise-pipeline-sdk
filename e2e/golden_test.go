package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/graphpipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/serialize"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/trigger"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func buildYAML(t *testing.T, p *pipeline.Builder) string {
	t.Helper()
	out, err := serialize.ToYAML(p.Build())
	require.NoError(t, err)
	return string(out)
}

// TestGolden_Basic covers a minimal single-workflow config.
func TestGolden_Basic(t *testing.T) {
	wf := workflow.New().
		AddStep(step.ActivateSshKey()).
		AddStep(step.GitClone()).
		AddStep(step.Script().WithContent("echo hello"))

	cfg := pipeline.New("other").
		AddWorkflow("primary", wf)

	checkGolden(t, "basic", buildYAML(t, cfg))
}

// TestGolden_iOS covers an iOS project with Xcode test + archive steps.
func TestGolden_iOS(t *testing.T) {
	test := workflow.New().
		WithTitle("iOS Tests").
		AddStep(step.ActivateSshKey()).
		AddStep(step.GitClone()).
		AddStep(step.XcodeTest().
			WithScheme("MyApp").
			WithProjectPath("MyApp.xcworkspace").
			WithDestination("platform=iOS Simulator,name=iPhone 15,OS=latest"))

	archive := workflow.New().
		WithTitle("iOS Archive").
		WithBeforeRun("test").
		AddStep(step.XcodeArchive().
			WithScheme("MyApp").
			WithProjectPath("MyApp.xcworkspace").
			WithDistributionMethod("app-store").
			WithAutomaticCodeSigning("api-key")).
		AddStep(step.DeployToBitriseIo())

	cfg := pipeline.New("ios").
		WithTitle("iOS CI/CD").
		AddWorkflow("test", test).
		AddWorkflow("archive", archive).
		AddTrigger(trigger.OnPush("test", "").WithBranch("*").Build()).
		AddTrigger(trigger.OnTag("archive", "").WithTag("v*").Build())

	checkGolden(t, "ios", buildYAML(t, cfg))
}

// TestGolden_Android covers an Android project with test + build steps.
func TestGolden_Android(t *testing.T) {
	test := workflow.New().
		WithTitle("Android Tests").
		AddStep(step.ActivateSshKey()).
		AddStep(step.GitClone()).
		AddStep(step.AndroidUnitTest().
			WithProjectLocation("./").
			WithModule("app").
			WithVariant("Debug"))

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

	checkGolden(t, "android", buildYAML(t, cfg))
}

// TestGolden_GraphPipeline covers a graph pipeline with parallel workflows.
func TestGolden_GraphPipeline(t *testing.T) {
	setup := workflow.New().
		WithTitle("Setup").
		AddStep(step.ActivateSshKey()).
		AddStep(step.GitClone()).
		AddStep(step.CachePull())

	unitTests := workflow.New().
		WithTitle("Unit Tests").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("go test ./...").WithTitle("Run tests")).
		AddStep(step.CachePush())

	lint := workflow.New().
		WithTitle("Lint").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("golangci-lint run").WithTitle("Lint"))

	deploy := workflow.New().
		WithTitle("Deploy").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("go build -o bin/app ./cmd/app").WithTitle("Build")).
		AddStep(step.DeployToBitriseIo())

	ci := graphpipeline.New().
		WithTitle("CI").
		AddWorkflow("setup", graphpipeline.NewWorkflow()).
		AddWorkflow("unit-tests", graphpipeline.NewWorkflow().WithDependsOn("setup")).
		AddWorkflow("lint", graphpipeline.NewWorkflow().WithDependsOn("setup")).
		AddWorkflow("deploy", graphpipeline.NewWorkflow().WithDependsOn("unit-tests", "lint").WithAbortOnFail(true))

	cfg := pipeline.New("other").
		WithTitle("Go Service CI").
		AddWorkflow("setup", setup).
		AddWorkflow("unit-tests", unitTests).
		AddWorkflow("lint", lint).
		AddWorkflow("deploy", deploy).
		AddGraphPipeline("ci", ci).
		AddTrigger(trigger.OnPush("", "ci").WithBranch("main").Build()).
		AddTrigger(trigger.OnPullRequest("", "ci").WithTargetBranch("main").Build())

	checkGolden(t, "graph_pipeline", buildYAML(t, cfg))
}

// TestGolden_Monorepo covers a monorepo setup with per-service workflows.
func TestGolden_Monorepo(t *testing.T) {
	sharedSetup := workflow.New().
		WithTitle("Shared Setup").
		AddStep(step.ActivateSshKey()).
		AddStep(step.GitClone())

	serviceA := workflow.New().
		WithTitle("Service A Tests").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("cd services/a && go test ./...").WithTitle("Test Service A"))

	serviceB := workflow.New().
		WithTitle("Service B Tests").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("cd services/b && go test ./...").WithTitle("Test Service B"))

	deployAll := workflow.New().
		WithTitle("Deploy All").
		WithBeforeRun("setup").
		AddStep(step.Script().WithContent("make deploy-all").WithTitle("Deploy")).
		AddStep(step.DeployToBitriseIo())

	cfg := pipeline.New("other").
		WithTitle("Monorepo CI").
		AddWorkflow("setup", sharedSetup).
		AddWorkflow("service-a", serviceA).
		AddWorkflow("service-b", serviceB).
		AddWorkflow("deploy", deployAll).
		AddTrigger(trigger.OnPush("service-a", "").WithBranch("*").Build()).
		AddTrigger(trigger.OnPullRequest("service-b", "").WithTargetBranch("main").Build()).
		AddTrigger(trigger.OnTag("deploy", "").WithTag("release/*").Build())

	checkGolden(t, "monorepo", buildYAML(t, cfg))
}
