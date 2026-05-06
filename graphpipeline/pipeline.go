package graphpipeline

import bitriseModels "github.com/bitrise-io/bitrise/v2/models"

// Builder constructs a graph (DAG-based) PipelineModel.
type Builder struct {
	title       string
	summary     string
	description string
	priority    *int
	workflows   bitriseModels.GraphPipelineWorkflowListItemModel
}

// New returns an empty graph pipeline builder.
func New() *Builder {
	return &Builder{
		workflows: bitriseModels.GraphPipelineWorkflowListItemModel{},
	}
}

// WithTitle sets the pipeline title.
func (b *Builder) WithTitle(title string) *Builder {
	b.title = title
	return b
}

// WithSummary sets the pipeline summary.
func (b *Builder) WithSummary(summary string) *Builder {
	b.summary = summary
	return b
}

// WithDescription sets the pipeline description.
func (b *Builder) WithDescription(desc string) *Builder {
	b.description = desc
	return b
}

// WithPriority sets the pipeline build priority (-20 to 20).
func (b *Builder) WithPriority(priority int) *Builder {
	b.priority = &priority
	return b
}

// AddWorkflow adds a workflow node to the graph pipeline.
func (b *Builder) AddWorkflow(id string, wf *WorkflowBuilder) *Builder {
	b.workflows[id] = wf.Build()
	return b
}

// Build returns the PipelineModel.
func (b *Builder) Build() bitriseModels.PipelineModel {
	return bitriseModels.PipelineModel{
		Title:       b.title,
		Summary:     b.summary,
		Description: b.description,
		Priority:    b.priority,
		Workflows:   b.workflows,
	}
}
