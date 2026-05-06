// Package graphpipeline provides builders for Bitrise graph (DAG-based) pipelines.
package graphpipeline

import bitriseModels "github.com/bitrise-io/bitrise/v2/models"

// WorkflowBuilder constructs a GraphPipelineWorkflowModel — a workflow node within a graph pipeline.
type WorkflowBuilder struct {
	model bitriseModels.GraphPipelineWorkflowModel
}

// NewWorkflow returns an empty graph pipeline workflow builder.
func NewWorkflow() *WorkflowBuilder {
	return &WorkflowBuilder{}
}

// WithDependsOn declares which other workflow IDs must complete before this one starts.
func (b *WorkflowBuilder) WithDependsOn(workflows ...string) *WorkflowBuilder {
	b.model.DependsOn = append(b.model.DependsOn, workflows...)
	return b
}

// WithAbortOnFail sets whether the pipeline aborts when this workflow fails.
func (b *WorkflowBuilder) WithAbortOnFail(abort bool) *WorkflowBuilder {
	b.model.AbortOnFail = abort
	return b
}

// WithShouldAlwaysRun sets whether this workflow runs even if dependencies failed.
// Use bitriseModels.GraphPipelineAlwaysRunModeWorkflow to always run.
func (b *WorkflowBuilder) WithShouldAlwaysRun(mode bitriseModels.GraphPipelineAlwaysRunMode) *WorkflowBuilder {
	b.model.ShouldAlwaysRun = mode
	return b
}

// WithRunIf sets a conditional expression that controls whether the workflow runs.
func (b *WorkflowBuilder) WithRunIf(expression string) *WorkflowBuilder {
	b.model.RunIf = bitriseModels.GraphPipelineRunIfModel{Expression: expression}
	return b
}

// WithUses references a reusable pipeline or workflow by its identifier.
func (b *WorkflowBuilder) WithUses(uses string) *WorkflowBuilder {
	b.model.Uses = uses
	return b
}

// WithParallel sets the parallelism level for this workflow node.
func (b *WorkflowBuilder) WithParallel(parallel string) *WorkflowBuilder {
	b.model.Parallel = parallel
	return b
}

// WithInput passes an input value to the workflow (used with `uses`).
func (b *WorkflowBuilder) WithInput(key string, value interface{}) *WorkflowBuilder {
	b.model.Inputs = append(b.model.Inputs, bitriseModels.GraphPipelineWorkflowModelInput{key: value})
	return b
}

// Build returns the GraphPipelineWorkflowModel.
func (b *WorkflowBuilder) Build() bitriseModels.GraphPipelineWorkflowModel {
	return b.model
}
