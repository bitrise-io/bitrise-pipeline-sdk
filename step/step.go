// Package step provides helpers for building Bitrise step list items.
package step

import (
	"fmt"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

// Buildable is implemented by Builder and all typed step builders.
// Pass any of them to workflow.Builder.AddStep.
type Buildable interface {
	Build() bitriseModels.StepListItemModel
}

// BundleBuildable extends Buildable with Ref and BuildForWithGroup, required when adding
// steps to a StepBundle or WithGroup definition.
type BundleBuildable interface {
	Buildable
	Ref() string
	BuildForWithGroup() bitriseModels.StepListStepItemModel
}

// Builder constructs a single StepListItemModel.
type Builder struct {
	id      string
	version string
	model   stepmanModels.StepModel
}

// From creates a step builder referencing a step by its steplib ID and version.
// Version may be a full semver ("1.2.3"), a major version ("1"), or "latest".
func From(id, version string) *Builder {
	return &Builder{id: id, version: version}
}

// WithVersion overrides the step version set at construction time.
// Version may be a full semver ("1.2.3"), a major version ("1"), or "latest".
func (b *Builder) WithVersion(version string) *Builder {
	b.version = version
	return b
}

// WithInput adds or overwrites a step input.
func (b *Builder) WithInput(key, value string) *Builder {
	b.model.Inputs = append(b.model.Inputs, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// WithRunIf sets the run_if expression for the step.
func (b *Builder) WithRunIf(expr string) *Builder {
	b.model.RunIf = &expr
	return b
}

// WithIsAlwaysRun configures whether the step runs even when a previous step failed.
func (b *Builder) WithIsAlwaysRun(v bool) *Builder {
	b.model.IsAlwaysRun = &v
	return b
}

// WithIsSkippable marks the step as skippable so a failure does not fail the build.
func (b *Builder) WithIsSkippable(v bool) *Builder {
	b.model.IsSkippable = &v
	return b
}

// WithTitle overrides the step title shown in the build log.
func (b *Builder) WithTitle(title string) *Builder {
	b.model.Title = &title
	return b
}

// WithTimeout sets the maximum execution time in seconds. 0 disables the timeout.
func (b *Builder) WithTimeout(seconds int) *Builder {
	b.model.Timeout = &seconds
	return b
}

// WithNoOutputTimeout sets the maximum time in seconds the step may run without
// producing any stdout/stderr output before it is aborted. 0 disables the timeout.
func (b *Builder) WithNoOutputTimeout(seconds int) *Builder {
	b.model.NoOutputTimeout = &seconds
	return b
}

// WithExecutionContainer pins this step to run inside the named container.
// The container must be defined in the pipeline config.
func (b *Builder) WithExecutionContainer(containerID string) *Builder {
	b.model.ExecutionContainer = containerID
	return b
}

// WithServiceContainers attaches one or more named service containers to this step.
func (b *Builder) WithServiceContainers(containerIDs ...string) *Builder {
	for _, id := range containerIDs {
		b.model.ServiceContainers = append(b.model.ServiceContainers, id)
	}
	return b
}

// Build returns the StepListItemModel ready to add to a workflow.
func (b *Builder) Build() bitriseModels.StepListItemModel {
	ref := fmt.Sprintf("%s@%s", b.id, b.version)
	return bitriseModels.StepListItemModel{ref: b.model}
}

// BuildForWithGroup returns a StepListStepItemModel for use inside a WithGroup.
// Unlike Build, the value type is stepmanModels.StepModel (required by WithModel.Steps).
func (b *Builder) BuildForWithGroup() bitriseModels.StepListStepItemModel {
	ref := fmt.Sprintf("%s@%s", b.id, b.version)
	return bitriseModels.StepListStepItemModel{ref: b.model}
}

// Ref returns the step reference string (e.g. "script@1").
func (b *Builder) Ref() string {
	return fmt.Sprintf("%s@%s", b.id, b.version)
}
