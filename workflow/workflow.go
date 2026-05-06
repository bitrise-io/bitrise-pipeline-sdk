// Package workflow provides a builder for Bitrise workflow definitions.
package workflow

import (
	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"

	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/stepbundle"
	"github.com/bitrise-io/bitrise-pipeline-sdk/withgroup"
)

// Builder constructs a WorkflowModel.
type Builder struct {
	model bitriseModels.WorkflowModel
}

// New returns an empty workflow builder.
func New() *Builder {
	return &Builder{}
}

// WithTitle sets the workflow title.
func (b *Builder) WithTitle(title string) *Builder {
	b.model.Title = title
	return b
}

// WithSummary sets the workflow summary.
func (b *Builder) WithSummary(summary string) *Builder {
	b.model.Summary = summary
	return b
}

// WithDescription sets the workflow description.
func (b *Builder) WithDescription(desc string) *Builder {
	b.model.Description = desc
	return b
}

// WithBeforeRun sets the workflow IDs to run before this workflow.
func (b *Builder) WithBeforeRun(workflows ...string) *Builder {
	b.model.BeforeRun = append(b.model.BeforeRun, workflows...)
	return b
}

// WithAfterRun sets the workflow IDs to run after this workflow.
func (b *Builder) WithAfterRun(workflows ...string) *Builder {
	b.model.AfterRun = append(b.model.AfterRun, workflows...)
	return b
}

// WithEnv appends a single environment variable to the workflow.
func (b *Builder) WithEnv(key, value string) *Builder {
	b.model.Environments = append(b.model.Environments, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// WithEnvs appends multiple environment variables to the workflow.
func (b *Builder) WithEnvs(envs ...envmanModels.EnvironmentItemModel) *Builder {
	b.model.Environments = append(b.model.Environments, envs...)
	return b
}

// AddStep appends a step to the workflow. Accepts *step.Builder or any typed step builder
// (e.g. *step.GitCloneBuilder, *step.XcodeTestBuilder).
func (b *Builder) AddStep(s step.Buildable) *Builder {
	b.model.Steps = append(b.model.Steps, s.Build())
	return b
}

// AddStepListItem appends a raw StepListItemModel to the workflow.
// Use this when you need full control over step configuration.
func (b *Builder) AddStepListItem(item bitriseModels.StepListItemModel) *Builder {
	b.model.Steps = append(b.model.Steps, item)
	return b
}

// AddStepBundleRef appends a reference to a named step bundle. The ref builder
// allows overriding inputs, envs, run_if, and title at the call site.
//
//	wf.AddStepBundleRef("lint", stepbundle.Ref().WithInput("flags", "--fix"))
func (b *Builder) AddStepBundleRef(bundleID string, ref *stepbundle.RefBuilder) *Builder {
	b.model.Steps = append(b.model.Steps, ref.BuildStepListItem(bundleID))
	return b
}

// AddWithGroup appends a "with" group — a set of steps that run inside a specific container.
//
//	wf.AddWithGroup(withgroup.New("postgres").AddStep(step.Script("./run-migrations.sh")))
func (b *Builder) AddWithGroup(wg *withgroup.Builder) *Builder {
	b.model.Steps = append(b.model.Steps, wg.Build())
	return b
}

// WithPriority sets the workflow build priority (-20 to 20).
func (b *Builder) WithPriority(priority int) *Builder {
	b.model.Priority = &priority
	return b
}

// WithTool adds a tool version requirement to the workflow.
func (b *Builder) WithTool(id bitriseModels.ToolID, version string) *Builder {
	if b.model.Tools == nil {
		b.model.Tools = bitriseModels.ToolsModel{}
	}
	b.model.Tools[id] = version
	return b
}

// Build returns the WorkflowModel.
func (b *Builder) Build() bitriseModels.WorkflowModel {
	return b.model
}
