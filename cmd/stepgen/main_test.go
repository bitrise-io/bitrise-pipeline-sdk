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

func TestParseStepYML_Outputs(t *testing.T) {
	yml := []byte(`
title: My Step
outputs:
  - MY_OUTPUT_KEY: null
    opts:
      title: The output value
  - ANOTHER_OUTPUT: null
    opts:
      summary: A summary description
`)
	def, err := parseStepYML("my-step", "1.0.0", yml)
	require.NoError(t, err)
	require.Len(t, def.Outputs, 2)

	assert.Equal(t, "MY_OUTPUT_KEY", def.Outputs[0].Key)
	assert.Equal(t, "MyOutputKey", def.Outputs[0].FieldName)
	assert.Contains(t, def.Outputs[0].Comment, "The output value")

	assert.Equal(t, "ANOTHER_OUTPUT", def.Outputs[1].Key)
	assert.Equal(t, "AnotherOutput", def.Outputs[1].FieldName)
	assert.Contains(t, def.Outputs[1].Comment, "A summary description")
}

func TestParseStepYML_NoOutputs(t *testing.T) {
	yml := []byte(`title: Simple Step`)
	def, err := parseStepYML("simple", "1.0.0", yml)
	require.NoError(t, err)
	assert.Empty(t, def.Outputs)
}

// ---- toOutputFieldName helper -----------------------------------------------

func TestToOutputFieldName(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"GIT_CLONE_COMMIT_HASH", "GitCloneCommitHash"},
		{"BITRISE_IPA_PATH", "BitriseIpaPath"},
		{"MY_OUTPUT", "MyOutput"},
		{"SINGLE", "Single"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, toOutputFieldName(tt.key), "key=%q", tt.key)
	}
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

func newTmpls(t *testing.T) stepTemplates {
	t.Helper()
	return stepTemplates{
		builder: template.Must(template.New("builder").Parse(builderTmpl)),
		alias:   template.Must(template.New("alias").Parse(aliasBuilderTmpl)),
	}
}

func TestGenerateStep_Tombstone(t *testing.T) {
	src := fakeSource{versions: map[string][]string{"dead-step": {}}}
	out, err := generateStep("dead-step", t.TempDir(), newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)
	assert.True(t, out.removed)
	assert.False(t, out.upToDate)
	assert.Nil(t, out.dep)
}

func TestGenerateStep_UpToDate(t *testing.T) {
	outputDir := t.TempDir()

	// Single-major: versions are 2.0.0 and 2.1.0 — latest is 2.1.0.
	// Pre-write a cached file at the latest version.
	cached := "// Code generated by cmd/stepgen. DO NOT EDIT.\n// Step: my-step (2.1.0)\npackage step\n"
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "gen_my_step.go"), []byte(cached), 0644))

	src := fakeSource{
		versions: map[string][]string{"my-step": {"2.0.0", "2.1.0"}},
		stepYML:  map[string][]byte{"my-step/2.1.0": minimalStepYML("My Step")},
	}

	out, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
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

	out, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, true /* force */)
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

	out, err := generateStep("script", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
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

func TestGenerateStep_WritesOutputs(t *testing.T) {
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0"}},
		stepYML: map[string][]byte{"my-step/1.0.0": []byte(`
title: My Step
outputs:
  - MY_RESULT: null
    opts:
      title: The result value
  - MY_STATUS: null
    opts:
      summary: The status code
`)},
	}

	_, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	data, _ := os.ReadFile(filepath.Join(outputDir, "gen_my_step.go"))
	content := string(data)
	assert.Contains(t, content, "type myStepOutputs struct")
	assert.Contains(t, content, "MyResult string")
	assert.Contains(t, content, "MyStatus string")
	assert.Contains(t, content, "var MyStepOutputs = myStepOutputs{")
	assert.Contains(t, content, `MyResult: "MY_RESULT"`)
	assert.Contains(t, content, `MyStatus: "MY_STATUS"`)
}

func TestGenerateStep_NoOutputs(t *testing.T) {
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"script": {"1.0.0"}},
		stepYML:  map[string][]byte{"script/1.0.0": minimalStepYML("Script")},
	}

	_, err := generateStep("script", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	data, _ := os.ReadFile(filepath.Join(outputDir, "gen_script.go"))
	assert.NotContains(t, string(data), "Outputs", "no outputs section when step has no outputs")
}

func TestGenerateStep_MultiMajor_WritesOutputsVarAndAlias(t *testing.T) {
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0", "2.0.0"}},
		stepYML: map[string][]byte{
			"my-step/1.0.0": []byte("title: My Step\noutputs:\n  - OLD_OUTPUT: null\n    opts:\n      title: Old\n"),
			"my-step/2.0.0": []byte("title: My Step\noutputs:\n  - NEW_OUTPUT: null\n    opts:\n      title: New\n"),
		},
	}

	_, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	// Versioned files each have their own outputs var.
	v1Data, _ := os.ReadFile(filepath.Join(outputDir, "gen_my_step_v1.go"))
	assert.Contains(t, string(v1Data), "var MyStepV1Outputs = myStepV1Outputs{")
	assert.Contains(t, string(v1Data), `OldOutput: "OLD_OUTPUT"`)

	v2Data, _ := os.ReadFile(filepath.Join(outputDir, "gen_my_step_v2.go"))
	assert.Contains(t, string(v2Data), "var MyStepV2Outputs = myStepV2Outputs{")
	assert.Contains(t, string(v2Data), `NewOutput: "NEW_OUTPUT"`)

	// Alias file re-exports the latest major's outputs.
	aliasData, _ := os.ReadFile(filepath.Join(outputDir, "gen_my_step_alias.go"))
	assert.Contains(t, string(aliasData), "var MyStepOutputs = MyStepV2Outputs")
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

	out, err := generateStep("old-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
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

	_, err := generateStep("xcode-test", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	_, errMissing := os.Stat(filepath.Join(outputDir, "gen_xcode_test.go"))
	assert.True(t, os.IsNotExist(errMissing), "gen_xcode_test.go must not exist")

	_, errPresent := os.Stat(filepath.Join(outputDir, "gen_xcode_test_builder.go"))
	assert.NoError(t, errPresent, "gen_xcode_test_builder.go must exist")
}

func TestGenerateStep_NewVersionReplacesCache(t *testing.T) {
	outputDir := t.TempDir()

	// Single-major: cached file is at 1.0.0; steplib now has 1.1.0.
	cached := "// Code generated by cmd/stepgen. DO NOT EDIT.\n// Step: script (1.0.0)\npackage step\n"
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "gen_script.go"), []byte(cached), 0644))

	src := fakeSource{
		versions: map[string][]string{"script": {"1.0.0", "1.1.0"}},
		stepYML:  map[string][]byte{"script/1.1.0": minimalStepYML("Script")},
	}

	out, err := generateStep("script", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)
	assert.False(t, out.upToDate)

	data, _ := os.ReadFile(filepath.Join(outputDir, "gen_script.go"))
	assert.Contains(t, string(data), "// Step: script (1.1.0)", "cached file should be updated")
}

// ---- majorVersionNums / latestOfMajor helpers ------------------------------

func TestMajorVersionNums(t *testing.T) {
	tests := []struct {
		versions []string
		want     []int
	}{
		{[]string{"1.0.0", "1.1.0", "1.2.0"}, []int{1}},
		{[]string{"1.0.0", "2.0.0", "3.0.0"}, []int{1, 2, 3}},
		{[]string{"0.9.0", "1.0.0"}, []int{0, 1}},
		{[]string{"5.0.0", "5.1.0", "6.0.0", "6.2.0"}, []int{5, 6}},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, majorVersionNums(tt.versions), "versions=%v", tt.versions)
	}
}

func TestLatestOfMajor(t *testing.T) {
	versions := []string{"1.0.0", "1.2.0", "2.0.0", "2.3.1", "3.0.0"}
	assert.Equal(t, "1.2.0", latestOfMajor(versions, 1))
	assert.Equal(t, "2.3.1", latestOfMajor(versions, 2))
	assert.Equal(t, "3.0.0", latestOfMajor(versions, 3))
	assert.Equal(t, "", latestOfMajor(versions, 4), "missing major → empty string")
}

// ---- generateStep multi-major -----------------------------------------------

func TestGenerateStep_MultiMajor_WritesVersionedAndAliasFiles(t *testing.T) {
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0", "1.2.0", "2.0.0", "2.3.0"}},
		stepYML: map[string][]byte{
			"my-step/1.2.0": []byte("title: My Step\ninputs:\n  - old_input: \"\"\n"),
			"my-step/2.3.0": []byte("title: My Step\ninputs:\n  - new_input: \"\"\n"),
		},
	}

	out, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)
	assert.False(t, out.removed)
	assert.False(t, out.upToDate)
	assert.Nil(t, out.dep)

	// Versioned files must exist.
	v1Path := filepath.Join(outputDir, "gen_my_step_v1.go")
	v2Path := filepath.Join(outputDir, "gen_my_step_v2.go")
	aliasPath := filepath.Join(outputDir, "gen_my_step_alias.go")
	assert.FileExists(t, v1Path)
	assert.FileExists(t, v2Path)
	assert.FileExists(t, aliasPath)

	// The old single-major file must NOT exist.
	assert.NoFileExists(t, filepath.Join(outputDir, "gen_my_step.go"))

	// v1 builder uses major-versioned type name and v1-specific inputs.
	v1Data, _ := os.ReadFile(v1Path)
	v1Content := string(v1Data)
	assert.Contains(t, v1Content, "type MyStepV1Builder struct")
	assert.Contains(t, v1Content, "func MyStepV1(version ...string) *MyStepV1Builder {")
	assert.Contains(t, v1Content, `v := "1"`)
	assert.Contains(t, v1Content, "func (b *MyStepV1Builder) WithOldInput(")

	// v2 builder uses major-versioned type name and v2-specific inputs.
	v2Data, _ := os.ReadFile(v2Path)
	v2Content := string(v2Data)
	assert.Contains(t, v2Content, "type MyStepV2Builder struct")
	assert.Contains(t, v2Content, "func MyStepV2(version ...string) *MyStepV2Builder {")
	assert.Contains(t, v2Content, `v := "2"`)
	assert.Contains(t, v2Content, "func (b *MyStepV2Builder) WithNewInput(")

	// Alias file re-exports v2 (the latest) under the plain name.
	aliasData, _ := os.ReadFile(aliasPath)
	aliasContent := string(aliasData)
	assert.Contains(t, aliasContent, "// Step: my-step (2.3.0)")
	assert.Contains(t, aliasContent, "type MyStepBuilder = MyStepV2Builder")
	assert.Contains(t, aliasContent, "func MyStep(version ...string) *MyStepV2Builder {")
	assert.Contains(t, aliasContent, "return MyStepV2(version...)")
}

func TestGenerateStep_MultiMajor_UpToDate(t *testing.T) {
	outputDir := t.TempDir()

	// Pre-write the alias file at the latest version (2.3.0).
	aliasCached := "// Code generated by cmd/stepgen. DO NOT EDIT.\n// Step: my-step (2.3.0)\npackage step\n"
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "gen_my_step_alias.go"), []byte(aliasCached), 0644))

	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0", "2.3.0"}},
		// No step.yml entries needed — should not be fetched.
	}

	out, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)
	assert.True(t, out.upToDate)
	assert.False(t, out.removed)
}

func TestGenerateStep_MultiMajor_DeletesOldSingleFile(t *testing.T) {
	outputDir := t.TempDir()

	// Simulate a pre-existing single-major file from a previous generation run.
	oldFile := filepath.Join(outputDir, "gen_my_step.go")
	require.NoError(t, os.WriteFile(oldFile, []byte("package step\n"), 0644))

	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0", "2.0.0"}},
		stepYML: map[string][]byte{
			"my-step/1.0.0": minimalStepYML("My Step"),
			"my-step/2.0.0": minimalStepYML("My Step"),
		},
	}

	_, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	// Old single-major file must be gone.
	assert.NoFileExists(t, oldFile, "old single-major file should be deleted")

	// New multi-major files must exist.
	assert.FileExists(t, filepath.Join(outputDir, "gen_my_step_v1.go"))
	assert.FileExists(t, filepath.Join(outputDir, "gen_my_step_v2.go"))
	assert.FileExists(t, filepath.Join(outputDir, "gen_my_step_alias.go"))
}

func TestGenerateStep_MultiMajor_Force(t *testing.T) {
	outputDir := t.TempDir()

	// Pre-write an up-to-date alias file — force should bypass it.
	aliasCached := "// Code generated by cmd/stepgen. DO NOT EDIT.\n// Step: my-step (2.0.0)\npackage step\n"
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "gen_my_step_alias.go"), []byte(aliasCached), 0644))

	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0", "2.0.0"}},
		stepYML: map[string][]byte{
			"my-step/1.0.0": minimalStepYML("My Step"),
			"my-step/2.0.0": minimalStepYML("My Step"),
		},
	}

	out, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, true /* force */)
	require.NoError(t, err)
	assert.False(t, out.upToDate, "--force should bypass fingerprinting")
}

func TestGenerateStep_SingleMajor_CleansUpStaleMajorFiles(t *testing.T) {
	// Regression test: when a step transitions from multi-major (e.g. due to a
	// stale "assets" directory in the steplib being counted as a version) to
	// truly single-major, the generator must remove any leftover alias and
	// versioned files so the package still compiles.
	outputDir := t.TempDir()

	// Simulate stale files left by a previous multi-major run.
	staleAlias := filepath.Join(outputDir, "gen_script_alias.go")
	staleV1 := filepath.Join(outputDir, "gen_script_v1.go")
	require.NoError(t, os.WriteFile(staleAlias, []byte("package step\n"), 0644))
	require.NoError(t, os.WriteFile(staleV1, []byte("package step\n"), 0644))

	src := fakeSource{
		versions: map[string][]string{"script": {"1.2.3"}}, // truly single-major
		stepYML:  map[string][]byte{"script/1.2.3": minimalStepYML("Script")},
	}

	_, err := generateStep("script", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	// Single-major file must exist.
	assert.FileExists(t, filepath.Join(outputDir, "gen_script.go"))
	// Stale multi-major files must be gone.
	assert.NoFileExists(t, staleAlias, "stale alias file should be cleaned up")
	assert.NoFileExists(t, staleV1, "stale versioned file should be cleaned up")
}

// ---- toEnumConstSuffix / buildValueOptions ----------------------------------

func TestToEnumConstSuffix(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"development", "Development"},
		{"app-store", "AppStore"},
		{"ad-hoc", "AdHoc"},
		{"iOS", "IOS"},
		{"watchOS", "WatchOS"},
		{"up_until_maximum_repetitions", "UpUntilMaximumRepetitions"},
		{"yes", "Yes"},
		{"no", "No"},
		{"off", "Off"},
		{"api-key", "ApiKey"},
		{"xcbeautify", "Xcbeautify"},
		{"1", "1"},
		{"0.05", "005"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, toEnumConstSuffix(tt.value), "value=%q", tt.value)
	}
}

func TestToEnumConstSuffix_Empty(t *testing.T) {
	assert.Equal(t, "", toEnumConstSuffix(""))
	assert.Equal(t, "", toEnumConstSuffix("-"))
	assert.Equal(t, "", toEnumConstSuffix("'-'"))
}

func TestBuildValueOptions_Valid(t *testing.T) {
	raw := []interface{}{"development", "app-store", "ad-hoc", "enterprise"}
	opts := buildValueOptions("XcodeArchiveV6DistributionMethod", raw)
	require.NotNil(t, opts)
	require.Len(t, opts, 4)
	assert.Equal(t, "development", opts[0].Value)
	assert.Equal(t, "XcodeArchiveV6DistributionMethodDevelopment", opts[0].ConstName)
	assert.Equal(t, "app-store", opts[1].Value)
	assert.Equal(t, "XcodeArchiveV6DistributionMethodAppStore", opts[1].ConstName)
}

func TestBuildValueOptions_EmptySuffix(t *testing.T) {
	// "'-'" produces an empty suffix → the whole input should be skipped.
	raw := []interface{}{"yes", "'-'"}
	opts := buildValueOptions("SomeType", raw)
	assert.Nil(t, opts, "a value that produces empty suffix should return nil")
}

func TestBuildValueOptions_Collision(t *testing.T) {
	// "app-store" and "app_store" both produce "AppStore" → collision.
	raw := []interface{}{"app-store", "app_store"}
	opts := buildValueOptions("SomeType", raw)
	assert.Nil(t, opts, "duplicate const names should return nil")
}

func TestBuildValueOptions_SingleValue(t *testing.T) {
	// A single value_option is not useful as an enum — caller requires >= 2.
	// buildValueOptions itself doesn't enforce this but caller in parseStepYML does.
	raw := []interface{}{"only"}
	opts := buildValueOptions("SomeType", raw)
	require.NotNil(t, opts)
	assert.Len(t, opts, 1)
}

// ---- parseStepYML value_options ---------------------------------------------

func TestParseStepYML_ValueOptions(t *testing.T) {
	yml := []byte(`
title: My Step
inputs:
  - distribution_method: development
    opts:
      title: Distribution method
      value_options:
      - development
      - app-store
      - ad-hoc
      - enterprise
  - platform: detect
    opts:
      title: Platform
      value_options:
      - detect
      - iOS
      - watchOS
`)
	def, err := parseStepYML("my-step", "1.0.0", yml)
	require.NoError(t, err)
	require.Len(t, def.Inputs, 2)

	dist := def.Inputs[0]
	assert.Equal(t, "distribution_method", dist.Key)
	assert.Equal(t, "MyStepDistributionMethod", dist.EnumTypeName)
	require.Len(t, dist.Options, 4)
	assert.Equal(t, "development", dist.Options[0].Value)
	assert.Equal(t, "MyStepDistributionMethodDevelopment", dist.Options[0].ConstName)
	assert.Equal(t, "app-store", dist.Options[1].Value)
	assert.Equal(t, "MyStepDistributionMethodAppStore", dist.Options[1].ConstName)

	platform := def.Inputs[1]
	assert.Equal(t, "platform", platform.Key)
	assert.Equal(t, "MyStepPlatform", platform.EnumTypeName)
	require.Len(t, platform.Options, 3)
	assert.Equal(t, "iOS", platform.Options[1].Value)
	assert.Equal(t, "MyStepPlatformIOS", platform.Options[1].ConstName)
}

func TestParseStepYML_ValueOptions_Messy_Skipped(t *testing.T) {
	// A value_options list containing an entry that produces an empty suffix
	// should not generate an enum type for that input.
	yml := []byte(`
title: My Step
inputs:
  - mode: auto
    opts:
      title: Mode
      value_options:
      - auto
      - "'-'"
`)
	def, err := parseStepYML("my-step", "1.0.0", yml)
	require.NoError(t, err)
	require.Len(t, def.Inputs, 1)
	assert.Empty(t, def.Inputs[0].EnumTypeName, "messy value_options should not generate enum")
	assert.Empty(t, def.Inputs[0].Options)
}

func TestParseStepYML_ValueOptions_SingleValue_Skipped(t *testing.T) {
	yml := []byte(`
title: My Step
inputs:
  - mode: auto
    opts:
      title: Mode
      value_options:
      - auto
`)
	def, err := parseStepYML("my-step", "1.0.0", yml)
	require.NoError(t, err)
	require.Len(t, def.Inputs, 1)
	assert.Empty(t, def.Inputs[0].EnumTypeName, "single value_option should not generate enum")
}

// ---- generateStep enum code generation -------------------------------------

func TestGenerateStep_WritesEnumType(t *testing.T) {
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0"}},
		stepYML: map[string][]byte{"my-step/1.0.0": []byte(`
title: My Step
inputs:
  - distribution_method: development
    opts:
      title: Distribution method
      value_options:
      - development
      - app-store
      - ad-hoc
`)},
	}

	_, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	data, _ := os.ReadFile(filepath.Join(outputDir, "gen_my_step.go"))
	content := string(data)
	assert.Contains(t, content, "type MyStepDistributionMethod string")
	// gofmt aligns const values with tabs, so check name and value separately.
	assert.Contains(t, content, "MyStepDistributionMethodDevelopment")
	assert.Contains(t, content, `MyStepDistributionMethod = "development"`)
	assert.Contains(t, content, "MyStepDistributionMethodAppStore")
	assert.Contains(t, content, `MyStepDistributionMethod = "app-store"`)
	assert.Contains(t, content, "MyStepDistributionMethodAdHoc")
	assert.Contains(t, content, `MyStepDistributionMethod = "ad-hoc"`)
	// Method should accept the enum type, not plain string.
	assert.Contains(t, content, "WithDistributionMethod(value MyStepDistributionMethod)")
	assert.Contains(t, content, `b.Builder.WithInput("distribution_method", string(value))`)
}

func TestGenerateStep_NoEnumForMessyOptions(t *testing.T) {
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0"}},
		stepYML: map[string][]byte{"my-step/1.0.0": []byte(`
title: My Step
inputs:
  - mode: auto
    opts:
      title: Mode
      value_options:
      - auto
      - "'-'"
`)},
	}

	_, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	data, _ := os.ReadFile(filepath.Join(outputDir, "gen_my_step.go"))
	content := string(data)
	assert.NotContains(t, content, "type MyStepMode string", "messy options must not produce enum type")
	assert.Contains(t, content, "WithMode(value string)", "plain string method must remain")
}

func TestGenerateStep_MultiMajor_WritesEnumAliases(t *testing.T) {
	outputDir := t.TempDir()
	src := fakeSource{
		versions: map[string][]string{"my-step": {"1.0.0", "2.0.0"}},
		stepYML: map[string][]byte{
			"my-step/1.0.0": []byte("title: My Step\n"),
			"my-step/2.0.0": []byte(`
title: My Step
inputs:
  - method: development
    opts:
      title: Method
      value_options:
      - development
      - app-store
`),
		},
	}

	_, err := generateStep("my-step", outputDir, newTmpls(t), src, &sync.Mutex{}, false)
	require.NoError(t, err)

	// v2 file has the enum type.
	v2Data, _ := os.ReadFile(filepath.Join(outputDir, "gen_my_step_v2.go"))
	assert.Contains(t, string(v2Data), "type MyStepV2Method string")

	// Alias file exposes the unversioned type alias.
	aliasData, _ := os.ReadFile(filepath.Join(outputDir, "gen_my_step_alias.go"))
	assert.Contains(t, string(aliasData), "type MyStepMethod = MyStepV2Method")
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
