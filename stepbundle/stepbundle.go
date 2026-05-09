// Package stepbundle provides builders for defining and referencing Bitrise step bundles.
//
// Step bundles are reusable groups of steps defined at the top level of a config and
// referenced by name inside workflows.
//
// Usage:
//
//	// 1. Define the bundle in the pipeline:
//	p.AddStepBundle("lint", stepbundle.New().
//	    WithInput("flags", "--fix").
//	    AddStep(step.Script().WithContent("golangci-lint run ./...")),
//	    AddStep(step.Script().WithContent("go vet ./...")))
//
//	// 2. Reference it inside a workflow, optionally overriding inputs:
//	wf.AddStepBundleRef("lint", stepbundle.Ref().
//	    WithRunIf(step.RunIfCI).
//	    WithInput("flags", "--fix --timeout 5m"))
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

// WithRunIf sets a conditional expression evaluated before the bundle runs.
// Use the step.RunIf* constants for the most common conditions:
//
//	bundle.WithRunIf(step.RunIfCI)
//	bundle.WithRunIf(step.RunIfBuildFailed)
//	bundle.WithRunIf(step.RunIfEnvEq("DEPLOY_ENV", "production"))
func (b *Builder) WithRunIf(expr step.RunIfExpr) *Builder {
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

// WithExecutionContainer pins all steps in this bundle to run inside the named container.
// The container must be defined in the pipeline config.
func (b *Builder) WithExecutionContainer(containerID string) *Builder {
	b.model.ExecutionContainer = containerID
	return b
}

// WithServiceContainers attaches one or more named service containers to all steps
// in this bundle. The containers must be defined in the pipeline config.
func (b *Builder) WithServiceContainers(containerIDs ...string) *Builder {
	for _, id := range containerIDs {
		b.model.ServiceContainers = append(b.model.ServiceContainers, id)
	}
	return b
}

// AddStep appends one or more steps to the bundle.
// Accepts *step.Builder or any typed step builder.
//
//	bundle.AddStep(
//	    step.Script().WithContent("golangci-lint run ./..."),
//	    step.Script().WithContent("go vet ./..."),
//	)
func (b *Builder) AddStep(steps ...step.BundleBuildable) *Builder {
	for _, s := range steps {
		item := bitriseModels.StepListItemStepOrBundleModel{s.Ref(): s.Build()[s.Ref()]}
		b.model.Steps = append(b.model.Steps, item)
	}
	return b
}

// AddStepBundleRef appends a reference to another named step bundle inside this bundle's
// step list, with optional call-site overrides. This enables bundle composition.
//
//	outer.AddStepBundleRef("inner-lint", stepbundle.Ref().WithInput("flags", "--fix"))
func (b *Builder) AddStepBundleRef(bundleID string, ref *RefBuilder) *Builder {
	key := bitriseModels.StepListItemStepBundleKeyPrefix + bundleID
	item := bitriseModels.StepListItemStepOrBundleModel{key: ref.model}
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

// WithSummary overrides the bundle summary at this call site.
func (b *RefBuilder) WithSummary(summary string) *RefBuilder {
	b.model.Summary = summary
	return b
}

// WithDescription overrides the bundle description at this call site.
func (b *RefBuilder) WithDescription(desc string) *RefBuilder {
	b.model.Description = desc
	return b
}

// WithRunIf overrides the run_if expression at this call site.
// Use the step.RunIf* constants for the most common conditions:
//
//	ref.WithRunIf(step.RunIfCI)
//	ref.WithRunIf(step.RunIfEnvEq("DEPLOY_ENV", "production"))
func (b *RefBuilder) WithRunIf(expr step.RunIfExpr) *RefBuilder {
	b.model.RunIf = &expr
	return b
}

// WithInput overrides or provides a value for a declared bundle input at this call site.
func (b *RefBuilder) WithInput(key, value string) *RefBuilder {
	b.model.Inputs = append(b.model.Inputs, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// WithEnv adds an environment variable at this call site.
func (b *RefBuilder) WithEnv(key, value string) *RefBuilder {
	b.model.Environments = append(b.model.Environments, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// WithExecutionContainer overrides the execution container at this call site.
func (b *RefBuilder) WithExecutionContainer(containerID string) *RefBuilder {
	b.model.ExecutionContainer = containerID
	return b
}

// WithServiceContainers overrides the service containers at this call site.
func (b *RefBuilder) WithServiceContainers(containerIDs ...string) *RefBuilder {
	for _, id := range containerIDs {
		b.model.ServiceContainers = append(b.model.ServiceContainers, id)
	}
	return b
}

// BuildStepListItem returns the StepListItemModel that places this bundle reference in a workflow.
func (b *RefBuilder) BuildStepListItem(bundleID string) bitriseModels.StepListItemModel {
	key := bitriseModels.StepListItemStepBundleKeyPrefix + bundleID
	return bitriseModels.StepListItemModel{key: b.model}
}
