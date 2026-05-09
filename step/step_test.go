package step_test

import (
	"testing"
	"time"

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
	// Generic methods return the concrete typed builder, so typed input
	// methods remain callable in any order in the chain.
	item := step.XcodeTest().
		WithTitle("Run Tests").   // generic — must return *XcodeTestV6Builder
		WithIsSkippable(true).    // generic — must return *XcodeTestV6Builder
		WithScheme("MyApp").      // typed input — only accessible via *XcodeTestV6Builder
		Build()
	assert.Contains(t, item, "xcode-test@6")
}

func TestTypedBuilder_GenericMethodsReturnConcreteType(t *testing.T) {
	// Compile-time assertions: each generic method must return the concrete
	// typed builder, not *Builder.
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithVersion("6.0.0")
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithInput("scheme", "MyApp")
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithTitle("t")
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithRunIf("true")
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithIsAlwaysRun(true)
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithIsSkippable(true)
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithTimeout(300)
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithNoOutputTimeout(60)
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithExecutionContainer("c")
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithServiceContainers("db")
}

func TestTypedBuilder_WithVersion_Chain(t *testing.T) {
	// WithVersion must preserve the typed chain so typed input methods remain callable.
	item := step.XcodeTest().
		WithVersion("6.0.1").
		WithScheme("MyApp"). // typed — only on *XcodeTestV6Builder
		Build()
	assert.Contains(t, item, "xcode-test@6.0.1")
}

func TestTypedBuilder_WithInput_Chain(t *testing.T) {
	// WithInput (escape hatch) must also preserve the typed chain.
	item := step.XcodeTest().
		WithInput("scheme", "MyApp").
		WithDestination("platform=iOS Simulator,name=iPhone 15"). // typed
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

// --- Typed enum constants ---------------------------------------------------

func TestEnumConstants_XcodeArchive_DistributionMethod(t *testing.T) {
	// Typed constants hold the correct string values.
	assert.Equal(t, step.XcodeArchiveV6DistributionMethod("development"), step.XcodeArchiveV6DistributionMethodDevelopment)
	assert.Equal(t, step.XcodeArchiveV6DistributionMethod("app-store"), step.XcodeArchiveV6DistributionMethodAppStore)
	assert.Equal(t, step.XcodeArchiveV6DistributionMethod("ad-hoc"), step.XcodeArchiveV6DistributionMethodAdHoc)
	assert.Equal(t, step.XcodeArchiveV6DistributionMethod("enterprise"), step.XcodeArchiveV6DistributionMethodEnterprise)
}

func TestEnumConstants_Alias(t *testing.T) {
	// The unversioned alias type is interchangeable with the versioned type.
	var _ step.XcodeArchiveDistributionMethod = step.XcodeArchiveV6DistributionMethodAppStore
}

func TestEnumConstants_BuilderAcceptsTypedValue(t *testing.T) {
	// Typed constant can be passed directly to the With* method.
	item := step.XcodeArchive().
		WithDistributionMethod(step.XcodeArchiveV6DistributionMethodAppStore).
		Build()
	require.Contains(t, item, "xcode-archive@6")
}

func TestEnumConstants_BuilderAcceptsStringLiteral(t *testing.T) {
	// Untyped string literals are still accepted (backward compatible).
	item := step.XcodeArchive().
		WithDistributionMethod("app-store").
		Build()
	require.Contains(t, item, "xcode-archive@6")
}

// --- RunIfExpr constants & helpers ------------------------------------------

func TestRunIfConstants_Values(t *testing.T) {
	assert.Equal(t, "", step.RunIfAlways)
	assert.Equal(t, "false", step.RunIfNever)
	assert.Equal(t, ".IsCI", step.RunIfCI)
	assert.Equal(t, "not .IsCI", step.RunIfNotCI)
	assert.Equal(t, ".IsBuildFailed", step.RunIfBuildFailed)
	assert.Equal(t, "not .IsBuildFailed", step.RunIfBuildSucceeded)
	assert.Equal(t, ".IsPR", step.RunIfPR)
	assert.Equal(t, "not .IsPR", step.RunIfNotPR)
}

func TestRunIfEnvEq(t *testing.T) {
	assert.Equal(t, `enveq "DEPLOY_ENV" "production"`, step.RunIfEnvEq("DEPLOY_ENV", "production"))
	assert.Equal(t, `enveq "MY_VAR" "1"`, step.RunIfEnvEq("MY_VAR", "1"))
}

func TestRunIfExpr_IsAssignableToString(t *testing.T) {
	// RunIfExpr is a type alias for string — existing string callers are unaffected.
	var _ string = step.RunIfCI
	var _ step.RunIfExpr = ".IsCI"
}

func TestWithRunIf_UsesConstant(t *testing.T) {
	item := step.Script().WithRunIf(step.RunIfCI).Build()
	require.Contains(t, item, "script@1")
}

func TestWithRunIf_UsesEnvEqHelper(t *testing.T) {
	item := step.Script().WithRunIf(step.RunIfEnvEq("DEPLOY_ENV", "prod")).Build()
	require.Contains(t, item, "script@1")
}

func TestWithRunIf_TypedBuilderChain(t *testing.T) {
	// WithRunIf must return the concrete typed builder so the chain continues.
	item := step.XcodeTest().
		WithRunIf(step.RunIfCI).
		WithScheme("MyApp"). // typed method — only available on *XcodeTestV6Builder
		Build()
	assert.Contains(t, item, "xcode-test@6")
}

func TestWithRunIf_TypedBuilderReturnsConcreteType(t *testing.T) {
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithRunIf(step.RunIfCI)
}

// --- Duration timeout variants -----------------------------------------------

func TestWithTimeoutDuration(t *testing.T) {
	item := step.Script().WithTimeoutDuration(10 * time.Minute).Build()
	require.Contains(t, item, "script@1")
}

func TestWithNoOutputTimeoutDuration(t *testing.T) {
	item := step.Script().WithNoOutputTimeoutDuration(30 * time.Second).Build()
	require.Contains(t, item, "script@1")
}

func TestWithTimeoutDuration_SubSecondTruncated(t *testing.T) {
	// 90.9 seconds truncates to 90 — same result as WithTimeout(90).
	a := step.Script().WithTimeoutDuration(90*time.Second + 900*time.Millisecond).Build()
	b := step.Script().WithTimeout(90).Build()
	assert.Equal(t, a, b)
}

func TestWithTimeoutDuration_TypedBuilderChain(t *testing.T) {
	// WithTimeoutDuration on a typed builder must return the concrete type.
	item := step.XcodeTest().
		WithTimeoutDuration(20 * time.Minute).
		WithScheme("MyApp"). // typed — verifies the chain wasn't broken
		Build()
	assert.Contains(t, item, "xcode-test@6")
}

func TestWithTimeoutDuration_TypedBuilderReturnsConcreteType(t *testing.T) {
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithTimeoutDuration(5 * time.Minute)
	var _ *step.XcodeTestV6Builder = step.XcodeTest().WithNoOutputTimeoutDuration(30 * time.Second)
}
