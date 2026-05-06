package stepbundle_test

import (
	"testing"

	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/stepbundle"
)

func TestNew_Metadata(t *testing.T) {
	sb := stepbundle.New().
		WithTitle("Lint").
		WithSummary("Run linters").
		WithDescription("Runs golangci-lint and go vet").
		Build()

	assert.Equal(t, "Lint", sb.Title)
	assert.Equal(t, "Run linters", sb.Summary)
}

func TestNew_AddStep(t *testing.T) {
	sb := stepbundle.New().
		AddStep(step.Script("golangci-lint run ./...")).
		AddStep(step.Script("go vet ./...")).
		Build()

	assert.Len(t, sb.Steps, 2)
}

func TestNew_Inputs(t *testing.T) {
	sb := stepbundle.New().
		WithInput("flags", "--fix").
		WithEnv("LINT_TIMEOUT", "5m").
		Build()

	require.Len(t, sb.Inputs, 1)
	assert.Equal(t, "--fix", sb.Inputs[0]["flags"])
	require.Len(t, sb.Environments, 1)
}

func TestNew_RunIf(t *testing.T) {
	sb := stepbundle.New().WithRunIf(".IsCI").Build()
	assert.Equal(t, ".IsCI", sb.RunIf)
}

func TestRef_BuildStepListItem(t *testing.T) {
	item := stepbundle.Ref().
		WithInput("flags", "--fix").
		BuildStepListItem("lint")

	key := bitriseModels.StepListItemStepBundleKeyPrefix + "lint"
	require.Contains(t, item, key)
}
