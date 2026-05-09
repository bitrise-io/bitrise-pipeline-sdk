// Package withgroup provides a builder for Bitrise "with" groups — a set of steps
// that run inside a specific execution container with optional service containers.
//
// Usage:
//
//	wg := withgroup.New("my-docker-image").
//	    WithServices("postgres", "redis").
//	    AddStep(
//	        step.Script().WithContent("go test ./..."),
//	        step.Script().WithContent("go vet ./..."),
//	    )
//
//	workflow.AddWithGroup(wg)
package withgroup

import (
	bitriseModels "github.com/bitrise-io/bitrise/v2/models"

	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
)

// Builder constructs a StepListWithItemModel — a "with" group inside a workflow.
type Builder struct {
	model bitriseModels.WithModel
}

// New creates a with group that runs its steps inside the named execution container.
// The container must be defined in the pipeline config.
func New(containerID string) *Builder {
	return &Builder{
		model: bitriseModels.WithModel{ContainerID: containerID},
	}
}

// WithServices attaches one or more service container IDs to all steps in the group.
func (b *Builder) WithServices(containerIDs ...string) *Builder {
	b.model.ServiceIDs = append(b.model.ServiceIDs, containerIDs...)
	return b
}

// AddStep appends one or more steps to the with group.
// Accepts *step.Builder or any typed step builder.
//
//	wg.AddStep(
//	    step.Script().WithContent("go build ./..."),
//	    step.Script().WithContent("go test ./..."),
//	)
func (b *Builder) AddStep(steps ...step.BundleBuildable) *Builder {
	for _, s := range steps {
		b.model.Steps = append(b.model.Steps, s.BuildForWithGroup())
	}
	return b
}

// Build returns the StepListItemModel that inserts this with group into a workflow step list.
func (b *Builder) Build() bitriseModels.StepListItemModel {
	return bitriseModels.StepListItemModel{
		bitriseModels.StepListItemWithKey: b.model,
	}
}
