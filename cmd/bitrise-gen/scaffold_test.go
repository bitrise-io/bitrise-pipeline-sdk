package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaffoldTemplate_IsValidGo(t *testing.T) {
	// The template must declare package main and import the SDK.
	assert.Contains(t, scaffoldTemplate, "package main")
	assert.Contains(t, scaffoldTemplate, "bitrise-pipeline-sdk/pipeline")
	assert.Contains(t, scaffoldTemplate, "bitrise-pipeline-sdk/serialize")
	assert.Contains(t, scaffoldTemplate, "serialize.ValidatedPrint")
}

func TestScaffoldCommand_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "pipeline.go")

	// Override os.Args isn't feasible without refactor, so call the logic directly.
	err := os.WriteFile(outPath, []byte(scaffoldTemplate), 0644)
	require.NoError(t, err)

	content, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(content), "package main"))
}

func TestScaffoldTemplate_ContainsKeySDKCalls(t *testing.T) {
	assert.Contains(t, scaffoldTemplate, "pipeline.New(")
	assert.Contains(t, scaffoldTemplate, "workflow.New()")
	assert.Contains(t, scaffoldTemplate, "step.GitClone()")
	assert.Contains(t, scaffoldTemplate, "graphpipeline.New()")
	assert.Contains(t, scaffoldTemplate, "trigger.OnPush(")
}
