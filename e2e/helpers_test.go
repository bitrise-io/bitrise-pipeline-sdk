package e2e_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files instead of comparing")

// checkGolden compares got against the golden file testdata/<name>.golden.
// When -update is passed the golden file is (re)written instead.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")

	if *update {
		require.NoError(t, os.WriteFile(path, []byte(got), 0644),
			"failed to write golden file %s", path)
		t.Logf("updated %s", path)
		return
	}

	want, err := os.ReadFile(path)
	require.NoError(t, err, "golden file missing — run with -update to create it")
	assert.Equal(t, string(want), got)
}
