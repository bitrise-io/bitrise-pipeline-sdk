package pipeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func TestBuilder_Validate_Valid(t *testing.T) {
	p := pipeline.New("other").
		AddWorkflow("setup", workflow.New().AddStep(step.GitClone())).
		AddWorkflow("primary", workflow.New().
			WithBeforeRun("setup").
			AddStep(step.Script("go test ./...")))

	result, err := p.Validate()
	require.NoError(t, err)
	assert.True(t, result.IsValid(), "errors: %v", result.Errors)
}

func TestBuilder_Validate_Invalid(t *testing.T) {
	p := pipeline.New("other").
		AddWorkflow("primary", workflow.New().WithBeforeRun("nonexistent"))

	result, err := p.Validate()
	require.NoError(t, err)
	assert.False(t, result.IsValid())
	require.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Error(), "nonexistent")
}
