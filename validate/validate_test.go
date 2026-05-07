package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/container"
	"github.com/bitrise-io/bitrise-pipeline-sdk/graphpipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/stepbundle"
	"github.com/bitrise-io/bitrise-pipeline-sdk/validate"
	"github.com/bitrise-io/bitrise-pipeline-sdk/withgroup"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func TestValidate_CleanConfig(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("setup", workflow.New()).
		AddWorkflow("primary", workflow.New().WithBeforeRun("setup")).
		Build()

	errs := validate.Config(data)
	assert.Empty(t, errs)
}

func TestValidate_UnknownBeforeRun(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().WithBeforeRun("nonexistent")).
		Build()

	errs := validate.Config(data)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "nonexistent")
	assert.Contains(t, errs[0].Location, "before_run")
}

func TestValidate_UnknownAfterRun(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().WithAfterRun("ghost")).
		Build()

	errs := validate.Config(data)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Location, "after_run")
}

func TestValidate_GraphPipeline_ValidDeps(t *testing.T) {
	data := pipeline.New("other").
		AddGraphPipeline("ci", graphpipeline.New().
			AddWorkflow("test", graphpipeline.NewWorkflow()).
			AddWorkflow("deploy", graphpipeline.NewWorkflow().WithDependsOn("test"))).
		Build()

	assert.Empty(t, validate.Config(data))
}

func TestValidate_GraphPipeline_BrokenDep(t *testing.T) {
	data := pipeline.New("other").
		AddGraphPipeline("ci", graphpipeline.New().
			AddWorkflow("deploy", graphpipeline.NewWorkflow().WithDependsOn("test"))).
		Build()

	errs := validate.Config(data)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "test")
	assert.Contains(t, errs[0].Location, "pipeline.ci")
}

func TestValidate_WithGroup_ValidContainer(t *testing.T) {
	data := pipeline.New("other").
		AddContainer("runner", container.NewExecution("golang:1.22")).
		AddWorkflow("primary", workflow.New().
			AddWithGroup(withgroup.New("runner").AddStep(step.Script().WithContent("go test ./...")))).
		Build()

	assert.Empty(t, validate.Config(data))
}

func TestValidate_WithGroup_UnknownContainer(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().
			AddWithGroup(withgroup.New("ghost-container").AddStep(step.Script().WithContent("echo hi")))).
		Build()

	errs := validate.Config(data)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "ghost-container")
}

func TestValidate_StepBundle_ValidRef(t *testing.T) {
	data := pipeline.New("other").
		AddStepBundle("lint", stepbundle.New().AddStep(step.Script().WithContent("golangci-lint run"))).
		AddWorkflow("primary", workflow.New().
			AddStepBundleRef("lint", stepbundle.Ref())).
		Build()

	assert.Empty(t, validate.Config(data))
}

func TestValidate_StepBundle_UndefinedRef(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().
			AddStepBundleRef("missing-bundle", stepbundle.Ref())).
		Build()

	errs := validate.Config(data)
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "missing-bundle")
}
