package pipeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/graphpipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/serialize"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/trigger"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func TestBuild_ProjectType(t *testing.T) {
	data := pipeline.New("ios").Build()
	assert.Equal(t, "ios", data.ProjectType)
}

func TestBuild_FormatVersionSet(t *testing.T) {
	data := pipeline.New("other").Build()
	assert.NotEmpty(t, data.FormatVersion)
}

func TestBuild_DefaultStepLibSource(t *testing.T) {
	data := pipeline.New("other").Build()
	assert.Equal(t, "https://github.com/bitrise-io/bitrise-steplib.git", data.DefaultStepLibSource)
}

func TestBuild_AppEnvs(t *testing.T) {
	data := pipeline.New("other").
		WithAppEnv("MY_KEY", "my_value").
		Build()

	require.Len(t, data.App.Environments, 1)
	assert.Equal(t, "my_value", data.App.Environments[0]["MY_KEY"])
}

func TestBuild_Workflow(t *testing.T) {
	wf := workflow.New().
		WithTitle("Primary").
		AddStep(step.Script("echo hello"))

	data := pipeline.New("other").
		AddWorkflow("primary", wf).
		Build()

	require.Contains(t, data.Workflows, "primary")
	assert.Equal(t, "Primary", data.Workflows["primary"].Title)
	require.Len(t, data.Workflows["primary"].Steps, 1)
}

func TestBuild_GraphPipeline(t *testing.T) {
	p := graphpipeline.New().
		WithTitle("CI Pipeline").
		AddWorkflow("test", graphpipeline.NewWorkflow()).
		AddWorkflow("deploy", graphpipeline.NewWorkflow().WithDependsOn("test"))

	data := pipeline.New("other").
		AddGraphPipeline("ci", p).
		Build()

	require.Contains(t, data.Pipelines, "ci")
	assert.Equal(t, "CI Pipeline", data.Pipelines["ci"].Title)
	require.Contains(t, data.Pipelines["ci"].Workflows, "deploy")
	assert.Equal(t, []string{"test"}, data.Pipelines["ci"].Workflows["deploy"].DependsOn)
}

func TestBuild_Triggers(t *testing.T) {
	data := pipeline.New("other").
		AddTrigger(trigger.OnPush("primary", "").WithBranch("main").Build()).
		AddTrigger(trigger.OnPullRequest("primary", "").WithTargetBranch("main").Build()).
		AddTrigger(trigger.OnTag("deploy", "").WithTag("v*").Build()).
		Build()

	require.Len(t, data.TriggerMap, 3)
}

func TestSerialize_ToYAML(t *testing.T) {
	data := pipeline.New("ios").
		AddWorkflow("primary", workflow.New().AddStep(step.GitClone())).
		Build()

	out, err := serialize.ToYAML(data)
	require.NoError(t, err)
	assert.Contains(t, string(out), "format_version")
	assert.Contains(t, string(out), "primary")
	assert.Contains(t, string(out), "git-clone")
}

func TestSerialize_ToJSON(t *testing.T) {
	data := pipeline.New("android").Build()

	out, err := serialize.ToJSON(data)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"project_type"`)
	assert.Contains(t, string(out), `"android"`)
}
