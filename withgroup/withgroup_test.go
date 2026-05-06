package withgroup_test

import (
	"testing"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/withgroup"
)

func TestNew_ContainerID(t *testing.T) {
	item := withgroup.New("my-container").Build()
	raw, ok := item[bitriseModels.StepListItemWithKey]
	require.True(t, ok)

	wm, ok := raw.(bitriseModels.WithModel)
	require.True(t, ok)
	assert.Equal(t, "my-container", wm.ContainerID)
}

func TestWithServices(t *testing.T) {
	item := withgroup.New("runner").WithServices("postgres", "redis").Build()
	wm := item[bitriseModels.StepListItemWithKey].(bitriseModels.WithModel)
	assert.Equal(t, []string{"postgres", "redis"}, wm.ServiceIDs)
}

func TestAddStep(t *testing.T) {
	item := withgroup.New("runner").
		AddStep(step.Script("go test ./...")).
		AddStep(step.Script("go vet ./...")).
		Build()

	wm := item[bitriseModels.StepListItemWithKey].(bitriseModels.WithModel)
	assert.Len(t, wm.Steps, 2)
}
