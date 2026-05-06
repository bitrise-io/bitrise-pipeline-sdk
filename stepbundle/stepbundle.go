// Package stepbundle provides builders for defining and referencing Bitrise step bundles.
//
// Step bundles are reusable groups of steps defined at the top level of a config and
// referenced by name inside workflows.
//
// Usage:
//
//	// 1. Define the bundle in the pipeline:
//	p.AddStepBundle("lint", stepbundle.New().
//	    AddStep(step.Script("golangci-lint run ./...")).
//	    AddStep(step.Script("go vet ./...")))
//
//	// 2. Reference it inside a workflow:
//	wf.AddStepBundleRef("lint", stepbundle.Ref())
package stepbundle

import (
	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"

	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
)

// Builder constructs a StepBundleModel — the definition stored at the top level.
type Builder struct {
	model bitriseModels.StepBundleModel
}

// New returns an empty step bundle definition builder.
func New() *Builder {
	return &Builder{}
}

// WithTitle sets the bundle title.
func (b *Builder) WithTitle(title string) *Builder {
	b.model.Title = title
	return b
}

// WithSummary sets the bundle summary.
func (b *Builder) WithSummary(summary string) *Builder {
	b.model.Summary = summary
	return b
}

// WithDescription sets the bundle description.
func (b *Builder) WithDescription(desc string) *Builder {
	b.model.Description = desc
	return b
}

// WithRunIf sets a conditional expression; the bundle is skipped when it evaluates to false.
func (b *Builder) WithRunIf(expr string) *Builder {
	b.model.RunIf = expr
	return b
}

// WithInput declares an input that callers can override when referencing the bundle.
func (b *Builder) WithInput(key, value string) *Builder {
	b.model.Inputs = append(b.model.Inputs, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// WithEnv appends an environment variable scoped to all steps in this bundle.
func (b *Builder) WithEnv(key, value string) *Builder {
	b.model.Environments = append(b.model.Environments, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// AddStep appends a step to the bundle. Accepts *step.Builder or any typed step builder.
func (b *Builder) AddStep(s step.BundleBuildable) *Builder {
	item := bitriseModels.StepListItemStepOrBundleModel{s.Ref(): s.Build()[s.Ref()]}
	b.model.Steps = append(b.model.Steps, item)
	return b
}

// Build returns the StepBundleModel.
func (b *Builder) Build() bitriseModels.StepBundleModel {
	return b.model
}

// RefBuilder constructs the call-site reference to a step bundle inside a workflow step list.
// It carries optional override values (title, run_if, inputs, envs) applied at the call site.
type RefBuilder struct {
	model bitriseModels.StepBundleListItemModel
}

// Ref returns a RefBuilder for referencing a step bundle by name.
// Use workflow.AddStepBundleRef(id, stepbundle.Ref()) to add it to a workflow.
func Ref() *RefBuilder {
	return &RefBuilder{}
}

// WithTitle overrides the bundle title at this call site.
func (b *RefBuilder) WithTitle(title string) *RefBuilder {
	b.model.Title = title
	return b
}

// WithRunIf overrides the run_if expression at this call site.
func (b *RefBuilder) WithRunIf(expr string) *RefBuilder {
	b.model.RunIf = &expr
	return b
}

// WithInput overrides or provides a value for a declared bundle input.
func (b *RefBuilder) WithInput(key, value string) *RefBuilder {
	b.model.Inputs = append(b.model.Inputs, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// WithEnv adds an environment variable at this call site.
func (b *RefBuilder) WithEnv(key, value string) *RefBuilder {
	b.model.Environments = append(b.model.Environments, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// BuildStepListItem returns the StepListItemModel that places this bundle reference in a workflow.
func (b *RefBuilder) BuildStepListItem(bundleID string) bitriseModels.StepListItemModel {
	key := bitriseModels.StepListItemStepBundleKeyPrefix + bundleID
	return bitriseModels.StepListItemModel{key: b.model}
}
