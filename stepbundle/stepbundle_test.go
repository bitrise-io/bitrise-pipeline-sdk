package stepbundle_test

import (
	"testing"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/stepbundle"
)

// ---- Builder ----------------------------------------------------------------

func TestNew_Metadata(t *testing.T) {
	sb := stepbundle.New().
		WithTitle("Lint").
		WithSummary("Run linters").
		WithDescription("Runs golangci-lint and go vet").
		Build()

	assert.Equal(t, "Lint", sb.Title)
	assert.Equal(t, "Run linters", sb.Summary)
	assert.Equal(t, "Runs golangci-lint and go vet", sb.Description)
}

func TestNew_AddStep_Single(t *testing.T) {
	sb := stepbundle.New().
		AddStep(step.Script().WithContent("golangci-lint run ./...")).
		AddStep(step.Script().WithContent("go vet ./...")).
		Build()

	assert.Len(t, sb.Steps, 2)
}

func TestNew_AddStep_Variadic(t *testing.T) {
	// Multiple steps can be passed in a single AddStep call.
	sb := stepbundle.New().
		AddStep(
			step.Script().WithContent("golangci-lint run ./..."),
			step.Script().WithContent("go vet ./..."),
			step.Script().WithContent("go test ./..."),
		).
		Build()

	assert.Len(t, sb.Steps, 3)
}

func TestNew_AddStep_Mixed(t *testing.T) {
	// Variadic and chained AddStep calls can be combined.
	sb := stepbundle.New().
		AddStep(
			step.Script().WithContent("step1"),
			step.Script().WithContent("step2"),
		).
		AddStep(step.Script().WithContent("step3")).
		Build()

	assert.Len(t, sb.Steps, 3)
}

func TestNew_Inputs(t *testing.T) {
	sb := stepbundle.New().
		WithInput("flags", "--fix").
		WithEnv("LINT_TIMEOUT", "5m").
		Build()

	require.Len(t, sb.Inputs, 1)
	assert.Equal(t, "--fix", sb.Inputs[0]["flags"])
	require.Len(t, sb.Environments, 1)
	assert.Equal(t, "5m", sb.Environments[0]["LINT_TIMEOUT"])
}

func TestNew_RunIf_PlainString(t *testing.T) {
	sb := stepbundle.New().WithRunIf(".IsCI").Build()
	assert.Equal(t, ".IsCI", sb.RunIf)
}

func TestNew_RunIf_Constant(t *testing.T) {
	sb := stepbundle.New().WithRunIf(step.RunIfCI).Build()
	assert.Equal(t, step.RunIfCI, sb.RunIf)
}

func TestNew_RunIf_EnvEqHelper(t *testing.T) {
	sb := stepbundle.New().WithRunIf(step.RunIfEnvEq("DEPLOY_ENV", "prod")).Build()
	assert.Equal(t, `enveq "DEPLOY_ENV" "prod"`, sb.RunIf)
}

func TestNew_ExecutionContainer(t *testing.T) {
	sb := stepbundle.New().WithExecutionContainer("my-container").Build()
	assert.Equal(t, "my-container", sb.ExecutionContainer)
}

func TestNew_ServiceContainers(t *testing.T) {
	sb := stepbundle.New().WithServiceContainers("postgres", "redis").Build()
	require.Len(t, sb.ServiceContainers, 2)
	assert.Equal(t, "postgres", sb.ServiceContainers[0])
	assert.Equal(t, "redis", sb.ServiceContainers[1])
}

func TestNew_AddStepBundleRef(t *testing.T) {
	// Bundles can nest other bundle references in their step list.
	inner := stepbundle.Ref().WithInput("flags", "--fix")
	sb := stepbundle.New().
		AddStep(step.Script().WithContent("pre-check")).
		AddStepBundleRef("inner-lint", inner).
		Build()

	require.Len(t, sb.Steps, 2)
	key := bitriseModels.StepListItemStepBundleKeyPrefix + "inner-lint"
	assert.Contains(t, sb.Steps[1], key)
}

// ---- RefBuilder -------------------------------------------------------------

func TestRef_BuildStepListItem(t *testing.T) {
	item := stepbundle.Ref().
		WithInput("flags", "--fix").
		BuildStepListItem("lint")

	key := bitriseModels.StepListItemStepBundleKeyPrefix + "lint"
	require.Contains(t, item, key)
}

func TestRef_Metadata(t *testing.T) {
	ref := stepbundle.Ref().
		WithTitle("Override Title").
		WithSummary("Override Summary").
		WithDescription("Override Description")
	item := ref.BuildStepListItem("lint")

	key := bitriseModels.StepListItemStepBundleKeyPrefix + "lint"
	val := item[key].(bitriseModels.StepBundleListItemModel)
	assert.Equal(t, "Override Title", val.Title)
	assert.Equal(t, "Override Summary", val.Summary)
	assert.Equal(t, "Override Description", val.Description)
}

func TestRef_RunIf_PlainString(t *testing.T) {
	item := stepbundle.Ref().WithRunIf(".IsCI").BuildStepListItem("lint")
	key := bitriseModels.StepListItemStepBundleKeyPrefix + "lint"
	val := item[key].(bitriseModels.StepBundleListItemModel)
	require.NotNil(t, val.RunIf)
	assert.Equal(t, ".IsCI", *val.RunIf)
}

func TestRef_RunIf_Constant(t *testing.T) {
	item := stepbundle.Ref().WithRunIf(step.RunIfBuildFailed).BuildStepListItem("cleanup")
	key := bitriseModels.StepListItemStepBundleKeyPrefix + "cleanup"
	val := item[key].(bitriseModels.StepBundleListItemModel)
	require.NotNil(t, val.RunIf)
	assert.Equal(t, step.RunIfBuildFailed, *val.RunIf)
}

func TestRef_InputsAndEnvs(t *testing.T) {
	item := stepbundle.Ref().
		WithInput("flags", "--fix").
		WithEnv("CI_MODE", "true").
		BuildStepListItem("lint")

	key := bitriseModels.StepListItemStepBundleKeyPrefix + "lint"
	val := item[key].(bitriseModels.StepBundleListItemModel)
	require.Len(t, val.Inputs, 1)
	assert.Equal(t, "--fix", val.Inputs[0]["flags"])
	require.Len(t, val.Environments, 1)
	assert.Equal(t, "true", val.Environments[0]["CI_MODE"])
}

func TestRef_ExecutionContainer(t *testing.T) {
	item := stepbundle.Ref().WithExecutionContainer("my-container").BuildStepListItem("lint")
	key := bitriseModels.StepListItemStepBundleKeyPrefix + "lint"
	val := item[key].(bitriseModels.StepBundleListItemModel)
	assert.Equal(t, "my-container", val.ExecutionContainer)
}

func TestRef_ServiceContainers(t *testing.T) {
	item := stepbundle.Ref().WithServiceContainers("postgres", "redis").BuildStepListItem("lint")
	key := bitriseModels.StepListItemStepBundleKeyPrefix + "lint"
	val := item[key].(bitriseModels.StepBundleListItemModel)
	require.Len(t, val.ServiceContainers, 2)
	assert.Equal(t, "postgres", val.ServiceContainers[0])
}
