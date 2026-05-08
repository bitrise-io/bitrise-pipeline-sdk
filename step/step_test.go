package step_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise-pipeline-sdk/step"
)

func TestFrom_Ref(t *testing.T) {
	s := step.From("my-step", "2.1.0")
	assert.Equal(t, "my-step@2.1.0", s.Ref())
}

func TestBuild_StepKey(t *testing.T) {
	item := step.From("script", "1").Build()
	require.Contains(t, item, "script@1")
}

func TestWithInput(t *testing.T) {
	item := step.From("script", "1").WithInput("content", "echo hi").Build()
	model := item["script@1"]
	require.NotNil(t, model)
}

func TestWithTimeout(t *testing.T) {
	s := step.From("script", "1").WithTimeout(300)
	item := s.Build()
	model := item["script@1"]
	require.NotNil(t, model)
}

func TestWithNoOutputTimeout(t *testing.T) {
	s := step.From("script", "1").WithNoOutputTimeout(60)
	item := s.Build()
	require.NotNil(t, item["script@1"])
}

func TestWithExecutionContainer(t *testing.T) {
	s := step.From("script", "1").WithExecutionContainer("my-container")
	item := s.Build()
	require.NotNil(t, item["script@1"])
}

func TestWithServiceContainers(t *testing.T) {
	s := step.From("script", "1").WithServiceContainers("postgres", "redis")
	item := s.Build()
	require.NotNil(t, item["script@1"])
}

func TestBuildForWithGroup(t *testing.T) {
	item := step.From("script", "1").BuildForWithGroup()
	require.Contains(t, item, "script@1")
}

// --- Typed builders ---

func TestGitClone(t *testing.T) {
	item := step.GitClone().WithBranch("main").WithCloneDepth("1").Build()
	require.Contains(t, item, "git-clone@8")
}

func TestScript(t *testing.T) {
	item := step.Script().WithContent("echo hello").WithWorkingDir("/tmp").Build()
	require.Contains(t, item, "script@1")
}

func TestXcodeTest(t *testing.T) {
	item := step.XcodeTest().
		WithScheme("MyApp").
		WithProjectPath("MyApp.xcworkspace").
		WithDestination("platform=iOS Simulator,name=iPhone 15").
		Build()
	require.Contains(t, item, "xcode-test@6")
}

func TestXcodeArchive(t *testing.T) {
	item := step.XcodeArchive().
		WithScheme("MyApp").
		WithDistributionMethod("app-store").
		Build()
	require.Contains(t, item, "xcode-archive@6")
}

func TestAndroidBuild(t *testing.T) {
	item := step.AndroidBuild().
		WithModule("app").
		WithVariant("Release").
		WithBuildType("aab").
		Build()
	require.Contains(t, item, "android-build@1")
}

func TestAndroidUnitTest(t *testing.T) {
	item := step.AndroidUnitTest().WithModule("app").WithVariant("Debug").Build()
	require.Contains(t, item, "android-unit-test@1")
}

func TestDeployToBitriseIo(t *testing.T) {
	item := step.DeployToBitriseIo().WithIsEnablePublicPage("false").Build()
	require.Contains(t, item, "deploy-to-bitrise-io@2")
}

func TestFastlane(t *testing.T) {
	item := step.Fastlane().WithLane("ios beta").Build()
	require.Contains(t, item, "fastlane@3")
}

func TestTypedBuilder_GenericMethodChain(t *testing.T) {
	// Verify that generic *Builder methods are accessible from typed builders
	// via embedding promotion, and the step still builds correctly.
	item := step.XcodeTest().
		WithScheme("MyApp").
		WithTitle("Run Tests").  // promoted *Builder method
		WithIsSkippable(true).   // promoted *Builder method
		Build()
	assert.Contains(t, item, "xcode-test@6")
}

// --- Version override -------------------------------------------------------

func TestTypedBuilder_DefaultVersion(t *testing.T) {
	// Calling the constructor with no argument uses the baked-in default.
	assert.Contains(t, step.GitClone().Build(), "git-clone@8")
	assert.Contains(t, step.Script().Build(), "script@1")
	assert.Contains(t, step.XcodeTest().Build(), "xcode-test@6")
}

func TestTypedBuilder_ExplicitVersion(t *testing.T) {
	// Passing a version string selects a different major version.
	assert.Contains(t, step.GitClone("7").Build(), "git-clone@7")
	assert.Contains(t, step.Script("2").Build(), "script@2")
	assert.Contains(t, step.XcodeTest("5").Build(), "xcode-test@5")
}

func TestTypedBuilder_VersionOverride_PreservesTypedMethods(t *testing.T) {
	// The typed input methods must still be callable after a version override —
	// the constructor returns *TypedBuilder, not *Builder.
	item := step.XcodeTest("5").
		WithScheme("MyApp").
		WithProjectPath("App.xcworkspace").
		Build()
	assert.Contains(t, item, "xcode-test@5")
}

func TestTypedBuilder_EmptyStringUsesDefault(t *testing.T) {
	// An explicit empty string should fall back to the default version.
	assert.Contains(t, step.GitClone("").Build(), "git-clone@8")
}

// --- Typed outputs -----------------------------------------------------------

func TestOutputs_GitClone(t *testing.T) {
	// The unversioned alias points to the latest major (v8).
	assert.Equal(t, "GIT_CLONE_COMMIT_HASH", step.GitCloneOutputs.GitCloneCommitHash)
	assert.Equal(t, "GIT_CLONE_COMMIT_MESSAGE_SUBJECT", step.GitCloneOutputs.GitCloneCommitMessageSubject)
	// The versioned var is also accessible directly.
	assert.Equal(t, "GIT_CLONE_COMMIT_HASH", step.GitCloneV8Outputs.GitCloneCommitHash)
}

func TestOutputs_XcodeArchive(t *testing.T) {
	assert.Equal(t, "BITRISE_IPA_PATH", step.XcodeArchiveOutputs.BitriseIpaPath)
	assert.Equal(t, "BITRISE_DSYM_PATH", step.XcodeArchiveOutputs.BitriseDsymPath)
	assert.Equal(t, "BITRISE_XCARCHIVE_PATH", step.XcodeArchiveOutputs.BitriseXcarchivePath)
}

func TestOutputs_DeployToBitriseIo(t *testing.T) {
	assert.Equal(t, "BITRISE_PUBLIC_INSTALL_PAGE_URL", step.DeployToBitriseIoOutputs.BitrisePublicInstallPageUrl)
}

func TestOutputs_VersionedVsAlias(t *testing.T) {
	// The unversioned alias and the latest versioned var hold the same values.
	assert.Equal(t, step.XcodeArchiveV6Outputs.BitriseIpaPath, step.XcodeArchiveOutputs.BitriseIpaPath)
}
