package workflow_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func TestWorkflow_Metadata(t *testing.T) {
	wf := workflow.New().
		WithTitle("Build").
		WithSummary("Builds the app").
		WithDescription("Full build workflow").
		Build()

	assert.Equal(t, "Build", wf.Title)
	assert.Equal(t, "Builds the app", wf.Summary)
	assert.Equal(t, "Full build workflow", wf.Description)
}

func TestWorkflow_BeforeAfterRun(t *testing.T) {
	wf := workflow.New().
		WithBeforeRun("setup").
		WithAfterRun("notify").
		Build()

	assert.Equal(t, []string{"setup"}, wf.BeforeRun)
	assert.Equal(t, []string{"notify"}, wf.AfterRun)
}

func TestWorkflow_Envs(t *testing.T) {
	wf := workflow.New().
		WithEnv("FOO", "bar").
		WithEnv("BAZ", "qux").
		Build()

	require.Len(t, wf.Environments, 2)
	assert.Equal(t, "bar", wf.Environments[0]["FOO"])
	assert.Equal(t, "qux", wf.Environments[1]["BAZ"])
}

func TestWorkflow_Steps(t *testing.T) {
	wf := workflow.New().
		AddStep(step.GitClone()).
		AddStep(step.Script("echo done")).
		Build()

	require.Len(t, wf.Steps, 2)
	assert.Contains(t, wf.Steps[0], "git-clone@1")
	assert.Contains(t, wf.Steps[1], "script@1")
}

func TestWorkflow_Priority(t *testing.T) {
	priority := 5
	wf := workflow.New().WithPriority(priority).Build()
	require.NotNil(t, wf.Priority)
	assert.Equal(t, priority, *wf.Priority)
}
