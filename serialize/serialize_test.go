package serialize_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/serialize"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func validConfig() func() interface{ Build() interface{} } {
	return nil
}

func buildValid() interface{} {
	return pipeline.New("other").
		AddWorkflow("primary", workflow.New().AddStep(step.Script("echo hi"))).
		Build()
}

func TestToYAML(t *testing.T) {
	data := pipeline.New("ios").
		AddWorkflow("primary", workflow.New().AddStep(step.GitClone())).
		Build()

	out, err := serialize.ToYAML(data)
	require.NoError(t, err)
	assert.Contains(t, string(out), "format_version")
	assert.Contains(t, string(out), "git-clone@1")
}

func TestToJSON(t *testing.T) {
	data := pipeline.New("android").Build()

	out, err := serialize.ToJSON(data)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"project_type"`)
	assert.Contains(t, string(out), `"android"`)
}

func TestNormalize_ReturnsCopy(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().AddStep(step.Script("echo hi"))).
		Build()

	normalized, err := serialize.Normalize(data)
	require.NoError(t, err)
	// Normalized copy should still be a valid model.
	assert.Equal(t, data.ProjectType, normalized.ProjectType)
	assert.Equal(t, data.FormatVersion, normalized.FormatVersion)
}

func TestNormalize_DoesNotMutateOriginal(t *testing.T) {
	data := pipeline.New("other").Build()
	original := data.ProjectType

	_, err := serialize.Normalize(data)
	require.NoError(t, err)
	assert.Equal(t, original, data.ProjectType)
}

func TestFillMissingDefaults(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().AddStep(step.Script("echo hi"))).
		Build()

	filled, err := serialize.FillMissingDefaults(data)
	require.NoError(t, err)
	assert.NotEmpty(t, filled.FormatVersion)
}

func TestValidatedToYAML_ValidConfig(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().AddStep(step.Script("echo hi"))).
		Build()

	out, err := serialize.ValidatedToYAML(data)
	require.NoError(t, err)
	assert.Contains(t, string(out), "format_version")
	assert.Contains(t, string(out), "script@1")
}

func TestValidatedToYAML_InvalidConfig(t *testing.T) {
	// Workflow references a before_run that doesn't exist.
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().WithBeforeRun("ghost")).
		Build()

	_, err := serialize.ValidatedToYAML(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost")
}

func TestValidatedToJSON_ValidConfig(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().AddStep(step.Script("echo hi"))).
		Build()

	out, err := serialize.ValidatedToJSON(data)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"project_type"`)
}

func TestValidatedToJSON_InvalidConfig(t *testing.T) {
	data := pipeline.New("other").
		AddWorkflow("primary", workflow.New().WithBeforeRun("ghost")).
		Build()

	_, err := serialize.ValidatedToJSON(data)
	require.Error(t, err)
}
