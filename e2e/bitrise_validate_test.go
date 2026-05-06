//go:build integration

// Run with: go test -tags=integration ./e2e/...
package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bitriseValidate runs `bitrise validate -c <path>` and returns (stdout+stderr, error).
func bitriseValidate(t *testing.T, configPath string) (string, error) {
	t.Helper()
	cmd := exec.Command("bitrise", "validate", "-c", configPath)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// writeGoldenConfig writes a named golden file to a temp file and returns its path.
func writeGoldenConfig(t *testing.T, name string) string {
	t.Helper()
	golden, err := os.ReadFile(filepath.Join("testdata", name+".golden"))
	require.NoError(t, err, "golden file %s.golden must exist — run: go test ./e2e/ -run TestGolden -update", name)

	tmp, err := os.CreateTemp(t.TempDir(), "*.yml")
	require.NoError(t, err)
	_, err = tmp.Write(golden)
	require.NoError(t, err)
	require.NoError(t, tmp.Close())
	return tmp.Name()
}

func TestBitriseValidate_Basic(t *testing.T) {
	path := writeGoldenConfig(t, "basic")
	out, err := bitriseValidate(t, path)
	assert.NoError(t, err, "bitrise validate should succeed for basic config\n%s", out)
}

func TestBitriseValidate_iOS(t *testing.T) {
	path := writeGoldenConfig(t, "ios")
	out, err := bitriseValidate(t, path)
	assert.NoError(t, err, "bitrise validate should succeed for iOS config\n%s", out)
}

func TestBitriseValidate_Android(t *testing.T) {
	path := writeGoldenConfig(t, "android")
	out, err := bitriseValidate(t, path)
	assert.NoError(t, err, "bitrise validate should succeed for Android config\n%s", out)
}

func TestBitriseValidate_GraphPipeline(t *testing.T) {
	path := writeGoldenConfig(t, "graph_pipeline")
	out, err := bitriseValidate(t, path)
	assert.NoError(t, err, "bitrise validate should succeed for graph pipeline config\n%s", out)
}

func TestBitriseValidate_Monorepo(t *testing.T) {
	path := writeGoldenConfig(t, "monorepo")
	out, err := bitriseValidate(t, path)
	assert.NoError(t, err, "bitrise validate should succeed for monorepo config\n%s", out)
}
