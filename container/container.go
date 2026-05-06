// Package container provides builders for Docker container definitions used in Bitrise workflows.
package container

import (
	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"
)

// Builder constructs a Container model.
type Builder struct {
	model bitriseModels.Container
}

// NewExecution creates a builder for an execution container (runs steps inside).
func NewExecution(image string) *Builder {
	return &Builder{
		model: bitriseModels.Container{
			Type:  bitriseModels.ContainerTypeExecution,
			Image: image,
		},
	}
}

// NewService creates a builder for a service container (runs alongside steps).
func NewService(image string) *Builder {
	return &Builder{
		model: bitriseModels.Container{
			Type:  bitriseModels.ContainerTypeService,
			Image: image,
		},
	}
}

// WithCredentials sets Docker registry credentials for private images.
func (b *Builder) WithCredentials(username, password, server string) *Builder {
	b.model.Credentials = bitriseModels.DockerCredentials{
		Username: username,
		Password: password,
		Server:   server,
	}
	return b
}

// WithPort maps a container port (e.g. "5432:5432").
func (b *Builder) WithPort(mapping string) *Builder {
	b.model.Ports = append(b.model.Ports, mapping)
	return b
}

// WithEnv appends an environment variable to the container.
func (b *Builder) WithEnv(key, value string) *Builder {
	b.model.Envs = append(b.model.Envs, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// WithOptions sets additional Docker options string passed to `docker run`.
func (b *Builder) WithOptions(options string) *Builder {
	b.model.Options = options
	return b
}

// Build returns the Container model.
func (b *Builder) Build() bitriseModels.Container {
	return b.model
}
