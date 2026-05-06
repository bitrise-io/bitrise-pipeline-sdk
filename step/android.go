package step

const (
	idAndroidBuildStep = "android-build"
	idAndroidTestStep  = "android-unit-test"
)

// AndroidBuildBuilder builds an android-build step with typed input methods.
type AndroidBuildBuilder struct{ *Builder }

// AndroidBuild creates an android-build step builder.
func AndroidBuild() *AndroidBuildBuilder {
	return &AndroidBuildBuilder{Builder: From(idAndroidBuildStep, "1")}
}

// WithProjectLocation sets the path to the root of the Android project.
func (b *AndroidBuildBuilder) WithProjectLocation(path string) *AndroidBuildBuilder {
	b.Builder.WithInput("project_location", path)
	return b
}

// WithModule sets the Gradle module to build (e.g. "app").
func (b *AndroidBuildBuilder) WithModule(module string) *AndroidBuildBuilder {
	b.Builder.WithInput("module", module)
	return b
}

// WithVariant sets the build variant (e.g. "Release", "Debug").
func (b *AndroidBuildBuilder) WithVariant(variant string) *AndroidBuildBuilder {
	b.Builder.WithInput("variant", variant)
	return b
}

// WithBuildType sets what to build: "apk" or "aab".
func (b *AndroidBuildBuilder) WithBuildType(buildType string) *AndroidBuildBuilder {
	b.Builder.WithInput("build_type", buildType)
	return b
}

// WithArguments appends extra Gradle arguments.
func (b *AndroidBuildBuilder) WithArguments(args string) *AndroidBuildBuilder {
	b.Builder.WithInput("arguments", args)
	return b
}

// AndroidTestBuilder builds an android-unit-test step with typed input methods.
type AndroidTestBuilder struct{ *Builder }

// AndroidTest creates an android-unit-test step builder.
func AndroidTest() *AndroidTestBuilder {
	return &AndroidTestBuilder{Builder: From(idAndroidTestStep, "1")}
}

// WithProjectLocation sets the path to the root of the Android project.
func (b *AndroidTestBuilder) WithProjectLocation(path string) *AndroidTestBuilder {
	b.Builder.WithInput("project_location", path)
	return b
}

// WithModule sets the Gradle module whose tests to run.
func (b *AndroidTestBuilder) WithModule(module string) *AndroidTestBuilder {
	b.Builder.WithInput("module", module)
	return b
}

// WithVariant sets the build variant to test.
func (b *AndroidTestBuilder) WithVariant(variant string) *AndroidTestBuilder {
	b.Builder.WithInput("variant", variant)
	return b
}

// WithArguments appends extra Gradle arguments.
func (b *AndroidTestBuilder) WithArguments(args string) *AndroidTestBuilder {
	b.Builder.WithInput("arguments", args)
	return b
}
