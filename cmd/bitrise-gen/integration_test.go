//go:build integration

// Run with: go test -tags=integration ./cmd/bitrise-gen/...
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// binaryPath is set by TestMain to the path of the compiled bitrise-gen binary.
var binaryPath string

// TestMain builds the binary once before running all integration tests.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "bitrise-gen-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	binaryPath = filepath.Join(dir, "bitrise-gen")
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build bitrise-gen: %v\n%s\n", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// validScript is a minimal pipeline script that outputs valid YAML.
const validScript = `package main

import (
	"log"
	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/serialize"
	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func main() {
	cfg := pipeline.New("other").
		AddWorkflow("primary", workflow.New().AddStep(step.Script().WithContent("echo hi"))).
		Build()
	if err := serialize.Print(cfg); err != nil {
		log.Fatal(err)
	}
}
`

// invalidScript outputs a config with a broken before_run reference.
const invalidScript = `package main

import (
	"log"
	"github.com/bitrise-io/bitrise-pipeline-sdk/pipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/serialize"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

func main() {
	cfg := pipeline.New("other").
		AddWorkflow("primary", workflow.New().WithBeforeRun("ghost")).
		Build()
	if err := serialize.Print(cfg); err != nil {
		log.Fatal(err)
	}
}
`

func writeScript(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "pipeline.go")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestRun_ValidScript(t *testing.T) {
	script := writeScript(t, validScript)
	out, err := exec.Command(binaryPath, "run", script).Output()
	require.NoError(t, err, "run should succeed for a valid script")
	assert.Contains(t, string(out), "format_version")
	assert.Contains(t, string(out), "script@1")
}

func TestRun_MissingScript(t *testing.T) {
	err := exec.Command(binaryPath, "run", "/nonexistent/pipeline.go").Run()
	assert.Error(t, err, "run should fail for a missing script")
}

func TestValidate_ValidScript(t *testing.T) {
	script := writeScript(t, validScript)
	out, err := exec.Command(binaryPath, "validate", script).Output()
	require.NoError(t, err, "validate should succeed for a valid script")
	assert.Contains(t, strings.ToLower(string(out)), "valid")
}

func TestValidate_InvalidScript(t *testing.T) {
	script := writeScript(t, invalidScript)
	// Use Output() so that ExitError.Stderr is populated with the command's stderr.
	_, err := exec.Command(binaryPath, "validate", script).Output()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
	assert.Contains(t, string(exitErr.Stderr), "ghost")
}

func TestScaffold_Integration(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "pipeline.go")
	out, err := exec.Command(binaryPath, "scaffold", "--output="+outPath).Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), outPath)

	content, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "package main")
	assert.Contains(t, string(content), "serialize.ValidatedPrint")
}

func TestScaffold_ExistsError(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "pipeline.go")
	require.NoError(t, os.WriteFile(outPath, []byte("exists"), 0644))

	err := exec.Command(binaryPath, "scaffold", "--output="+outPath).Run()
	assert.Error(t, err, "scaffold should fail if output file already exists")
}
