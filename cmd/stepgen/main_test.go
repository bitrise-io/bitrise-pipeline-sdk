package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- helpers ---------------------------------------------------------------

func TestLatestVersion(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		want     string
	}{
		{"single", []string{"1.0.0"}, "1.0.0"},
		{"already sorted", []string{"1.0.0", "1.1.0", "2.0.0"}, "2.0.0"},
		{"reverse order", []string{"2.0.0", "1.0.0", "1.1.0"}, "2.0.0"},
		{"patch wins", []string{"1.0.0", "1.0.1", "1.0.2"}, "1.0.2"},
		{"multi-digit major", []string{"9.0.0", "10.0.0", "2.0.0"}, "10.0.0"},
		{"minor tiebreak", []string{"1.2.0", "1.10.0", "1.9.0"}, "1.10.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, latestVersion(tt.versions))
		})
	}
}

func TestCmpVer(t *testing.T) {
	assert.Equal(t, -1, cmpVer("1.0.0", "2.0.0"))
	assert.Equal(t, 1, cmpVer("2.0.0", "1.0.0"))
	assert.Equal(t, 0, cmpVer("1.2.3", "1.2.3"))
	assert.Equal(t, -1, cmpVer("1.9.0", "1.10.0"), "numeric minor comparison")
}

func TestMajorVersion(t *testing.T) {
	assert.Equal(t, "3", majorVersion("3.2.1"))
	assert.Equal(t, "10", majorVersion("10.0.0"))
	assert.Equal(t, "1", majorVersion("1"))
}

func TestNormalizeID(t *testing.T) {
	assert.Equal(t, "git_clone", normalizeID("git-clone"))
	assert.Equal(t, "android_unit_test", normalizeID("android-unit-test"))
	assert.Equal(t, "script", normalizeID("script"))
}

func TestToTypeName(t *testing.T) {
	assert.Equal(t, "GitClone", toTypeName("git-clone"))
	assert.Equal(t, "XcodeTest", toTypeName("xcode-test"))
	assert.Equal(t, "AndroidUnitTest", toTypeName("android-unit-test"))
	assert.Equal(t, "Script", toTypeName("script"))
	assert.Equal(t, "DeployToBitriseIo", toTypeName("deploy-to-bitrise-io"))
}

func TestToMethodName(t *testing.T) {
	assert.Equal(t, "SshRsaPrivateKey", toMethodName("ssh_rsa_private_key"))
	assert.Equal(t, "Content", toMethodName("content"))
	assert.Equal(t, "WorkingDir", toMethodName("working_dir"))
}

func TestIsValidIdentifier(t *testing.T) {
	assert.True(t, isValidIdentifier("content"))
	assert.True(t, isValidIdentifier("working_dir"))
	assert.True(t, isValidIdentifier("_private"))
	assert.False(t, isValidIdentifier(""))
	assert.False(t, isValidIdentifier("1bad"))
	assert.False(t, isValidIdentifier("has-hyphen"))
	assert.False(t, isValidIdentifier("has space"))
}

// ---- YAML parsers -----------------------------------------------------------

func TestParseStepInfo_Empty(t *testing.T) {
	dep, err := parseStepInfo(nil)
	require.NoError(t, err)
	assert.Nil(t, dep)

	dep, err = parseStepInfo([]byte{})
	require.NoError(t, err)
	assert.Nil(t, dep)
}

func TestParseStepInfo_Deprecated(t *testing.T) {
	yml := []byte(`
deprecate_notes: Use the new-step instead.
removal_date: "2025-01-01"
`)
	dep, err := parseStepInfo(yml)
	require.NoError(t, err)
	require.NotNil(t, dep)
	assert.Equal(t, "Use the new-step instead.", dep.Notes)
	assert.Equal(t, "2025-01-01", dep.RemovalDate)
}

func TestParseStepInfo_NotDeprecated(t *testing.T) {
	yml := []byte(`maintainer: bitrise`)
	dep, err := parseStepInfo(yml)
	require.NoError(t, err)
	assert.Nil(t, dep)
}

func TestParseStepInfo_InvalidYAML(t *testing.T) {
	_, err := parseStepInfo([]byte(":\tbad: yaml: ["))
	require.Error(t, err)
}

func TestParseStepYML_Basic(t *testing.T) {
	yml := []byte(`
title: My Step
inputs:
  - content: ""
    opts:
      title: Script content
  - working_dir: ""
    opts:
      title: Working directory
`)
	def, err := parseStepYML("my-step", "2.1.0", yml)
	require.NoError(t, err)
	assert.Equal(t, "my-step", def.StepID)
	assert.Equal(t, "2.1.0", def.Version)
	assert.Equal(t, "2", def.MajorVersion)
	assert.Equal(t, "MyStep", def.TypeName)
	require.Len(t, def.Inputs, 2)
	assert.Equal(t, "content", def.Inputs[0].Key)
	assert.Equal(t, "Content", def.Inputs[0].MethodName)
	assert.Equal(t, "working_dir", def.Inputs[1].Key)
	assert.Equal(t, "WorkingDir", def.Inputs[1].MethodName)
}

func TestParseStepYML_NoInputs(t *testing.T) {
	yml := []byte(`title: Minimal Step`)
	def, err := parseStepYML("minimal", "1.0.0", yml)
	require.NoError(t, err)
	assert.Empty(t, def.Inputs)
}

func TestParseStepYML_SkipsInvalidIdentifiers(t *testing.T) {
	yml := []byte(`
title: Step
inputs:
  - valid_key: ""
  - "has-hyphen": ""
  - "1starts_with_digit": ""
  - opts: {}
`)
	def, err := parseStepYML("step", "1.0.0", yml)
	require.NoError(t, err)
	require.Len(t, def.Inputs, 1)
	assert.Equal(t, "valid_key", def.Inputs[0].Key)
}

func TestParseStepYML_DeduplicatesMethodNames(t *testing.T) {
	// Some step.yml files list the same key more than once with different defaults.
	yml := []byte(`
title: Step
inputs:
  - content: "default1"
  - content: "default2"
`)
	def, err := parseStepYML("step", "1.0.0", yml)
	require.NoError(t, err)
	require.Len(t, def.Inputs, 1, "duplicate key should be deduplicated")
}

func TestParseStepYML_InvalidYAML(t *testing.T) {
	_, err := parseStepYML("step", "1.0.0", []byte(":\tbad"))
	require.Error(t, err)
}

// ---- file I/O ---------------------------------------------------------------

func TestReadCachedVersion_Missing(t *testing.T) {
	assert.Equal(t, "", readCachedVersion("/nonexistent/path.go"))
}

func TestReadCachedVersion_Valid(t *testing.T) {
	f := writeTempFile(t, "// Code generated by cmd/stepgen. DO NOT EDIT.\n// Step: git-clone (8.2.1)\npackage step\n")
	assert.Equal(t, "8.2.1", readCachedVersion(f))
}

func TestReadCachedVersion_MalformedHeader(t *testing.T) {
	// Line 2 exists but has no parenthesised version.
	f := writeTempFile(t, "// Code generated by cmd/stepgen. DO NOT EDIT.\n// Step: git-clone\npackage step\n")
	assert.Equal(t, "", readCachedVersion(f))
}

func TestReadCachedVersion_SingleLine(t *testing.T) {
	f := writeTempFile(t, "// only one line")
	assert.Equal(t, "", readCachedVersion(f))
}

func TestLoadConfig_Missing(t *testing.T) {
	cfg, err := loadConfig("/nonexistent/stepgen.json")
	require.NoError(t, err)
	assert.Empty(t, cfg.Steps)
	assert.Empty(t, cfg.Skip)
}

func TestLoadConfig_Valid(t *testing.T) {
	f := writeTempFile(t, `{"steps":["git-clone","script"],"skip":["old-step"]}`)
	cfg, err := loadConfig(f)
	require.NoError(t, err)
	assert.Equal(t, []string{"git-clone", "script"}, cfg.Steps)
	assert.Equal(t, []string{"old-step"}, cfg.Skip)
}

func TestLoadConfig_NoSkip(t *testing.T) {
	f := writeTempFile(t, `{"steps":["git-clone"]}`)
	cfg, err := loadConfig(f)
	require.NoError(t, err)
	assert.Empty(t, cfg.Skip)
}

// ---- pruneConfig ------------------------------------------------------------

func TestPruneConfig_NothingToPrune(t *testing.T) {
	f := writeTempFile(t, `{"steps":["git-clone","script"]}`)
	err := pruneConfig(f, nil, nil)
	require.NoError(t, err)
	// File should be unchanged (steps still present).
	cfg, _ := loadConfig(f)
	assert.Equal(t, []string{"git-clone", "script"}, cfg.Steps)
}

func TestPruneConfig_MovesRemovedToSkip(t *testing.T) {
	f := writeTempFile(t, `{"steps":["git-clone","old-step","script"]}`)
	err := pruneConfig(f, []string{"old-step"}, nil)
	require.NoError(t, err)

	cfg, _ := loadConfig(f)
	assert.Equal(t, []string{"git-clone", "script"}, cfg.Steps)
	assert.Equal(t, []string{"old-step"}, cfg.Skip)
}

func TestPruneConfig_DropsDeprecatedEntirely(t *testing.T) {
	f := writeTempFile(t, `{"steps":["git-clone","deprecated-step","script"]}`)
	err := pruneConfig(f, nil, []string{"deprecated-step"})
	require.NoError(t, err)

	cfg, _ := loadConfig(f)
	assert.Equal(t, []string{"git-clone", "script"}, cfg.Steps)
	assert.Empty(t, cfg.Skip, "deprecated steps should not be added to skip list")
}

func TestPruneConfig_BothRemovedAndDeprecated(t *testing.T) {
	f := writeTempFile(t, `{"steps":["a","removed","b","deprecated","c"]}`)
	err := pruneConfig(f, []string{"removed"}, []string{"deprecated"})
	require.NoError(t, err)

	cfg, _ := loadConfig(f)
	assert.Equal(t, []string{"a", "b", "c"}, cfg.Steps)
	assert.Equal(t, []string{"removed"}, cfg.Skip)
}

func TestPruneConfig_DeduplicatesSkipList(t *testing.T) {
	// "old-step" is already in the skip list; pruning it again should not duplicate it.
	f := writeTempFile(t, `{"steps":["git-clone","old-step"],"skip":["old-step"]}`)
	err := pruneConfig(f, []string{"old-step"}, nil)
	require.NoError(t, err)

	cfg, _ := loadConfig(f)
	assert.Equal(t, []string{"old-step"}, cfg.Skip, "should not be duplicated")
}

func TestPruneConfig_SortsSkipList(t *testing.T) {
	f := writeTempFile(t, `{"steps":["z-step","a-step","m-step"]}`)
	err := pruneConfig(f, []string{"z-step", "a-step", "m-step"}, nil)
	require.NoError(t, err)

	cfg, _ := loadConfig(f)
	assert.Equal(t, []string{"a-step", "m-step", "z-step"}, cfg.Skip)
}

// ---- generateStep (fake stepSource) ----------------------------------------

// fakeSource implements stepSource entirely in memory.
type fakeSource struct {
	versions map[string][]string // stepID → version list
	stepYML  map[string][]byte   // "stepID/version" → step.yml bytes
	stepInfo map[string][]byte   // stepID → step-info.yml bytes
}

func (f fakeSource) listVersions(stepID string) ([]string, error) {
	v, ok := f.versions[stepID]
	if !ok {
		return nil, fmt.Errorf("step %q not found", stepID)
	}
	return v, nil
}

func (f fakeSource) readStepYML(stepID, version string) ([]byte, error) {
	key := stepID + "/" + version
	data, ok := f.stepYML[key]
	if !ok {
		return nil, fmt.Errorf("step.yml not found for %s@%s", stepID, version)
	}
	return data, nil
}

func (f fakeSource) readStepInfo(stepID string) ([]byte, error) {
	return f.stepInfo[stepID], nil // nil is fine — means not deprecated
}

// minimalStepYML returns the bare minimum step.yml for a given step ID.
func minimalStepYML(title string) []byte {
	return []byte("title: " + title + "\n")
}

func newTmpl(t *testing.T) *template.Template {
	t.Helper()
	return template.Must(template.New("builder").Parse(builderTmpl))
}

func TestGenerateStep_Tombstone(t *testing.T) {
	src := fakeSource{versions: map[string][]string{"dead-step": {}}}
	out, err := generateStep("dead-step", t.TempDir(), newTmpl(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)
	assert.True(t, out.removed)
	assert.False(t, out.upToDate)
	assert.Nil(t, out.dep)
}

func TestGenerateStep_UpToDate(t *testing.T) {
	outputDir := t.TempDir()

	// Pre-write a cached file at version 2.0.0.
	cached := "// Code generated by cmd/stepgen. DO NOT EDIT.\n// Step: my-step (2.0.0)\npackage step\n"
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "gen_my_step.go"), []byte(cached), 0644))

	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0", "2.0.0"}},
		stepYML:  map[string][]byte{"my-step/2.0.0": minimalStepYML("My Step")},
	}

	out, err := generateStep("my-step", outputDir, newTmpl(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)
	assert.True(t, out.upToDate)
	assert.False(t, out.removed)
}

func TestGenerateStep_ForceBypassesCache(t *testing.T) {
	outputDir := t.TempDir()

	// Cached file is already at the latest version.
	cached := "// Code generated by cmd/stepgen. DO NOT EDIT.\n// Step: my-step (2.0.0)\npackage step\n"
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "gen_my_step.go"), []byte(cached), 0644))

	src := fakeSource{
		versions: map[string][]string{"my-step": {"2.0.0"}},
		stepYML:  map[string][]byte{"my-step/2.0.0": minimalStepYML("My Step")},
	}

	out, err := generateStep("my-step", outputDir, newTmpl(t), src, &sync.Mutex{}, true /* force */)
	require.NoError(t, err)
	assert.False(t, out.upToDate, "force should bypass fingerprinting")
	assert.False(t, out.removed)
}

func TestGenerateStep_WritesFile(t *testing.T) {
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"script": {"1.2.3"}},
		stepYML: map[string][]byte{"script/1.2.3": []byte(`
title: Script
inputs:
  - content: ""
    opts:
      title: Script content
`)},
	}

	out, err := generateStep("script", outputDir, newTmpl(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)
	assert.False(t, out.removed)
	assert.False(t, out.upToDate)
	assert.Nil(t, out.dep)

	data, err := os.ReadFile(filepath.Join(outputDir, "gen_script.go"))
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "// Step: script (1.2.3)")
	assert.Contains(t, content, "type ScriptBuilder struct")
	assert.Contains(t, content, "func Script(version ...string) *ScriptBuilder {")
	assert.Contains(t, content, `v := "1"`)
	assert.Contains(t, content, "func (b *ScriptBuilder) WithContent(")
}

func TestGenerateStep_DeprecatedStep(t *testing.T) {
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"old-step": {"3.0.0"}},
		stepYML:  map[string][]byte{"old-step/3.0.0": minimalStepYML("Old Step")},
		stepInfo: map[string][]byte{
			"old-step": []byte("deprecate_notes: Use new-step instead.\nremoval_date: \"2025-06-01\"\n"),
		},
	}

	out, err := generateStep("old-step", outputDir, newTmpl(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)
	require.NotNil(t, out.dep)
	assert.Equal(t, "Use new-step instead.", out.dep.Notes)

	data, err := os.ReadFile(filepath.Join(outputDir, "gen_old_step.go"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "// Deprecated: Use new-step instead.")
}

func TestGenerateStep_TestSuffixGetsBuilderFilename(t *testing.T) {
	// Step IDs ending in "-test" would produce a "_test.go" filename which the
	// Go build system treats as a test file. The generator appends "_builder".
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"xcode-test": {"6.0.0"}},
		stepYML:  map[string][]byte{"xcode-test/6.0.0": minimalStepYML("Xcode Test")},
	}

	_, err := generateStep("xcode-test", outputDir, newTmpl(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	_, errMissing := os.Stat(filepath.Join(outputDir, "gen_xcode_test.go"))
	assert.True(t, os.IsNotExist(errMissing), "gen_xcode_test.go must not exist")

	_, errPresent := os.Stat(filepath.Join(outputDir, "gen_xcode_test_builder.go"))
	assert.NoError(t, errPresent, "gen_xcode_test_builder.go must exist")
}

func TestGenerateStep_NewVersionReplacesCache(t *testing.T) {
	outputDir := t.TempDir()

	// Cached file is at 1.0.0; steplib now has 2.0.0.
	cached := "// Code generated by cmd/stepgen. DO NOT EDIT.\n// Step: script (1.0.0)\npackage step\n"
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "gen_script.go"), []byte(cached), 0644))

	src := fakeSource{
		versions: map[string][]string{"script": {"1.0.0", "2.0.0"}},
		stepYML:  map[string][]byte{"script/2.0.0": minimalStepYML("Script")},
	}

	out, err := generateStep("script", outputDir, newTmpl(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)
	assert.False(t, out.upToDate)

	data, _ := os.ReadFile(filepath.Join(outputDir, "gen_script.go"))
	assert.Contains(t, string(data), "// Step: script (2.0.0)", "cached file should be updated")
}

// ---- deleteDeprecatedFiles --------------------------------------------------

func TestDeleteDeprecatedFiles(t *testing.T) {
	outputDir := t.TempDir()

	// Create two files that should be deleted.
	for _, name := range []string{"gen_old_step.go", "gen_xcode_test_builder.go"} {
		require.NoError(t, os.WriteFile(filepath.Join(outputDir, name), []byte("package step\n"), 0644))
	}
	// A file that should NOT be deleted.
	keep := filepath.Join(outputDir, "gen_script.go")
	require.NoError(t, os.WriteFile(keep, []byte("package step\n"), 0644))

	deps := []deprecationInfo{
		{StepID: "old-step"},
		{StepID: "xcode-test"}, // ends in -test → _builder suffix
	}
	require.NoError(t, deleteDeprecatedFiles(outputDir, deps))

	assert.NoFileExists(t, filepath.Join(outputDir, "gen_old_step.go"))
	assert.NoFileExists(t, filepath.Join(outputDir, "gen_xcode_test_builder.go"))
	assert.FileExists(t, keep, "unrelated file must not be deleted")
}

func TestDeleteDeprecatedFiles_MissingFileIsNotAnError(t *testing.T) {
	err := deleteDeprecatedFiles(t.TempDir(), []deprecationInfo{{StepID: "ghost-step"}})
	assert.NoError(t, err)
}

// ---- test helpers ----------------------------------------------------------

// writeTempFile writes content to a new temp file and returns its path.
func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stepgen-test-*.json")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

// writeTempFile is reused for non-JSON content (Go source, YAML) throughout
// this file. The .json extension is cosmetic and does not affect test behaviour.
