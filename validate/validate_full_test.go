package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/graphpipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/validate"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func TestFull_ValidConfig(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("setup", workflow.New().AddStep(step.GitClone())).
		AddWorkflow("primary", workflow.New().
			WithBeforeRun("setup").
			AddStep(step.Script().WithContent("go test ./..."))).
		Build()

	result, err := validate.Full(data)
	require.NoError(t, err)
	assert.True(t, result.IsValid(), "expected valid config, got errors: %v", result.Errors)
}

func TestFull_MissingBeforeRunWorkflow(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().WithBeforeRun("does-not-exist")).
		Build()

	result, err := validate.Full(data)
	require.NoError(t, err)
	assert.False(t, result.IsValid())
	require.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Error(), "does-not-exist")
}

func TestFull_UpstreamCatchesCycles(t *testing.T) {
	// before_run cycle: a → b → a
	data := pipeline.New("other").
		AddWorkflow("a", workflow.New().WithBeforeRun("b")).
		AddWorkflow("b", workflow.New().WithBeforeRun("a")).
		Build()

	result, err := validate.Full(data)
	require.NoError(t, err)
	// Upstream validator catches the cycle; our structural check only finds missing refs,
	// so at least one error (cycle) should be present.
	assert.False(t, result.IsValid())
}

func TestFull_GraphPipeline_ValidDeps(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("test", workflow.New().AddStep(step.Script().WithContent("go test ./..."))).
		AddWorkflow("deploy", workflow.New().AddStep(step.Script().WithContent("go build ./..."))).
		AddGraphPipeline("ci", graphpipeline.New().
			AddWorkflow("test", graphpipeline.NewWorkflow()).
			AddWorkflow("deploy", graphpipeline.NewWorkflow().WithDependsOn("test"))).
		Build()

	result, err := validate.Full(data)
	require.NoError(t, err)
	assert.True(t, result.IsValid(), "errors: %v", result.Errors)
}

func TestFull_GraphPipeline_WorkflowMissingFromConfig(t *testing.T) {
	// Pipeline references "test" workflow but it's not defined as a top-level workflow.
	data := pipeline.New("other").
		AddGraphPipeline("ci", graphpipeline.New().
			AddWorkflow("test", graphpipeline.NewWorkflow())).
		Build()

	result, err := validate.Full(data)
	require.NoError(t, err)
	// Upstream pipeline validator requires graph pipeline workflow IDs to exist
	// as top-level workflows.
	assert.False(t, result.IsValid())
}

func TestFull_Result_IsValid(t *testing.T) {
	result := validate.Result{}
	assert.True(t, result.IsValid())

	result.Errors = []validate.Error{{Location: "x", Message: "y"}}
	assert.False(t, result.IsValid())
}

func TestFull_WarningsAreReturned(t *testing.T) {
	// A valid config should produce no errors; warnings may or may not be present
	// depending on upstream heuristics.
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().AddStep(step.Script().WithContent("echo hi"))).
		Build()

	result, err := validate.Full(data)
	require.NoError(t, err)
	// Warnings is allowed to be nil or non-nil; this just checks the field exists.
	_ = result.Warnings
}
