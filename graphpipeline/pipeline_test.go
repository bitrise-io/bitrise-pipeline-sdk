package graphpipeline_test

import (
	"testing"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/graphpipeline"
)

func TestPipeline_Metadata(t *testing.T) {
	p := graphpipeline.New().
		WithTitle("My Pipeline").
		WithSummary("Runs CI").
		WithDescription("Full CI/CD pipeline").
		Build()

	assert.Equal(t, "My Pipeline", p.Title)
	assert.Equal(t, "Runs CI", p.Summary)
	assert.Equal(t, "Full CI/CD pipeline", p.Description)
}

func TestPipeline_Workflows(t *testing.T) {
	p := graphpipeline.New().
		AddWorkflow("test", graphpipeline.NewWorkflow()).
		AddWorkflow("deploy", graphpipeline.NewWorkflow().WithDependsOn("test")).
		Build()

	require.Contains(t, p.Workflows, "test")
	require.Contains(t, p.Workflows, "deploy")
	assert.Equal(t, []string{"test"}, p.Workflows["deploy"].DependsOn)
}

func TestPipeline_Priority(t *testing.T) {
	priority := -5
	p := graphpipeline.New().WithPriority(priority).Build()
	require.NotNil(t, p.Priority)
	assert.Equal(t, priority, *p.Priority)
}

func TestWorkflow_AbortOnFail(t *testing.T) {
	wf := graphpipeline.NewWorkflow().WithAbortOnFail(true).Build()
	assert.True(t, wf.AbortOnFail)
}

func TestWorkflow_ShouldAlwaysRun(t *testing.T) {
	wf := graphpipeline.NewWorkflow().
		WithShouldAlwaysRun(bitriseModels.GraphPipelineAlwaysRunModeWorkflow).
		Build()
	assert.Equal(t, bitriseModels.GraphPipelineAlwaysRunModeWorkflow, wf.ShouldAlwaysRun)
}

func TestWorkflow_RunIf(t *testing.T) {
	wf := graphpipeline.NewWorkflow().WithRunIf(".IsBuildFailed").Build()
	assert.Equal(t, ".IsBuildFailed", wf.RunIf.Expression)
}

func TestWorkflow_Inputs(t *testing.T) {
	wf := graphpipeline.NewWorkflow().
		WithInput("scheme", "MyApp").
		WithInput("configuration", "Release").
		Build()

	require.Len(t, wf.Inputs, 2)
}
