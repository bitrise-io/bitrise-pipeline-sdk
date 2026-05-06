// Package stage provides builders for legacy stage-based Bitrise pipelines.
// For new pipelines, prefer the graphpipeline package which uses a DAG model.
package stage

import bitriseModels "github.com/bitrise-io/bitrise/v2/models"

// WorkflowBuilder constructs a StageWorkflowModel entry for a stage.
type WorkflowBuilder struct {
	model bitriseModels.StageWorkflowModel
}

// NewWorkflow returns a stage workflow builder.
func NewWorkflow() *WorkflowBuilder {
	return &WorkflowBuilder{}
}

// WithRunIf sets a conditional expression for the workflow within the stage.
func (b *WorkflowBuilder) WithRunIf(expr string) *WorkflowBuilder {
	b.model.RunIf = expr
	return b
}

// Build returns the StageWorkflowModel.
func (b *WorkflowBuilder) Build() bitriseModels.StageWorkflowModel {
	return b.model
}

// Builder constructs a StageModel.
type Builder struct {
	model bitriseModels.StageModel
}

// New returns an empty stage builder.
func New() *Builder {
	return &Builder{}
}

// WithTitle sets the stage title.
func (b *Builder) WithTitle(title string) *Builder {
	b.model.Title = title
	return b
}

// WithSummary sets the stage summary.
func (b *Builder) WithSummary(summary string) *Builder {
	b.model.Summary = summary
	return b
}

// WithDescription sets the stage description.
func (b *Builder) WithDescription(desc string) *Builder {
	b.model.Description = desc
	return b
}

// WithAbortOnFail sets whether the pipeline aborts when this stage fails.
func (b *Builder) WithAbortOnFail(abort bool) *Builder {
	b.model.AbortOnFail = abort
	return b
}

// WithShouldAlwaysRun sets whether this stage runs even if a previous stage failed.
func (b *Builder) WithShouldAlwaysRun(always bool) *Builder {
	b.model.ShouldAlwaysRun = always
	return b
}

// WithRunIf sets a conditional expression for the stage.
func (b *Builder) WithRunIf(expr string) *Builder {
	b.model.RunIf = expr
	return b
}

// AddWorkflow adds a workflow entry to this stage.
func (b *Builder) AddWorkflow(id string, wf *WorkflowBuilder) *Builder {
	b.model.Workflows = append(b.model.Workflows, bitriseModels.StageWorkflowListItemModel{id: wf.Build()})
	return b
}

// Build returns the StageModel.
func (b *Builder) Build() bitriseModels.StageModel {
	return b.model
}
