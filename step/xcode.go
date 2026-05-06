package step

const (
	idXcodeTestStep    = "xcode-test"
	idXcodeArchiveStep = "xcode-archive"
)

// XcodeTestBuilder builds an xcode-test step with typed input methods.
type XcodeTestBuilder struct{ *Builder }

// XcodeTest creates an xcode-test step builder.
func XcodeTest() *XcodeTestBuilder {
	return &XcodeTestBuilder{Builder: From(idXcodeTestStep, "1")}
}

// WithScheme sets the Xcode scheme to test.
func (b *XcodeTestBuilder) WithScheme(scheme string) *XcodeTestBuilder {
	b.Builder.WithInput("scheme", scheme)
	return b
}

// WithProjectPath sets the path to the .xcodeproj or .xcworkspace file.
func (b *XcodeTestBuilder) WithProjectPath(path string) *XcodeTestBuilder {
	b.Builder.WithInput("project_path", path)
	return b
}

// WithDestination sets the -destination argument passed to xcodebuild.
func (b *XcodeTestBuilder) WithDestination(destination string) *XcodeTestBuilder {
	b.Builder.WithInput("destination", destination)
	return b
}

// WithTestPlan selects an Xcode Test Plan (.xctestplan) by name.
func (b *XcodeTestBuilder) WithTestPlan(plan string) *XcodeTestBuilder {
	b.Builder.WithInput("test_plan", plan)
	return b
}

// WithConfiguration sets the build configuration (e.g. "Debug").
func (b *XcodeTestBuilder) WithConfiguration(config string) *XcodeTestBuilder {
	b.Builder.WithInput("configuration", config)
	return b
}

// WithXcconfigContent inlines an xcconfig file content to override build settings.
func (b *XcodeTestBuilder) WithXcconfigContent(content string) *XcodeTestBuilder {
	b.Builder.WithInput("xcconfig_content", content)
	return b
}

// WithXcodebuildOptions appends extra flags to the xcodebuild test invocation.
func (b *XcodeTestBuilder) WithXcodebuildOptions(opts string) *XcodeTestBuilder {
	b.Builder.WithInput("xcodebuild_test_options", opts)
	return b
}

// XcodeArchiveBuilder builds an xcode-archive step with typed input methods.
type XcodeArchiveBuilder struct{ *Builder }

// XcodeArchive creates an xcode-archive step builder.
func XcodeArchive() *XcodeArchiveBuilder {
	return &XcodeArchiveBuilder{Builder: From(idXcodeArchiveStep, "1")}
}

// WithScheme sets the Xcode scheme to archive.
func (b *XcodeArchiveBuilder) WithScheme(scheme string) *XcodeArchiveBuilder {
	b.Builder.WithInput("scheme", scheme)
	return b
}

// WithProjectPath sets the path to the .xcodeproj or .xcworkspace file.
func (b *XcodeArchiveBuilder) WithProjectPath(path string) *XcodeArchiveBuilder {
	b.Builder.WithInput("project_path", path)
	return b
}

// WithDistributionMethod sets the export method (e.g. "app-store", "ad-hoc", "development").
func (b *XcodeArchiveBuilder) WithDistributionMethod(method string) *XcodeArchiveBuilder {
	b.Builder.WithInput("distribution_method", method)
	return b
}

// WithConfiguration sets the build configuration (e.g. "Release").
func (b *XcodeArchiveBuilder) WithConfiguration(config string) *XcodeArchiveBuilder {
	b.Builder.WithInput("configuration", config)
	return b
}

// WithXcconfigContent inlines an xcconfig file content to override build settings.
func (b *XcodeArchiveBuilder) WithXcconfigContent(content string) *XcodeArchiveBuilder {
	b.Builder.WithInput("xcconfig_content", content)
	return b
}

// WithAutomaticCodeSigning sets the code signing mode ("api-key" or "apple-id").
func (b *XcodeArchiveBuilder) WithAutomaticCodeSigning(mode string) *XcodeArchiveBuilder {
	b.Builder.WithInput("automatic_code_signing", mode)
	return b
}
