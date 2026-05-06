package env

import envmanModels "github.com/bitrise-io/envman/v2/models"

// New creates an EnvironmentItemModel from a key-value pair.
func New(key, value string) envmanModels.EnvironmentItemModel {
	return envmanModels.EnvironmentItemModel{key: value}
}

// List builds a slice of EnvironmentItemModel from alternating key-value pairs.
// Panics if an odd number of arguments is provided.
func List(pairs ...string) []envmanModels.EnvironmentItemModel {
	if len(pairs)%2 != 0 {
		panic("env.List requires an even number of arguments (key, value pairs)")
	}
	envs := make([]envmanModels.EnvironmentItemModel, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		envs = append(envs, New(pairs[i], pairs[i+1]))
	}
	return envs
}
