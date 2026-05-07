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
