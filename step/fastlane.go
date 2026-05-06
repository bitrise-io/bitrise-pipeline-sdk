package step

const idFastlaneStep = "fastlane"

// FastlaneBuilder builds a fastlane step with typed input methods.
type FastlaneBuilder struct{ *Builder }

// Fastlane creates a fastlane step builder.
func Fastlane() *FastlaneBuilder {
	return &FastlaneBuilder{Builder: From(idFastlaneStep, "1")}
}

// WithLane sets the lane to run (e.g. "ios beta", "android deploy").
func (b *FastlaneBuilder) WithLane(lane string) *FastlaneBuilder {
	b.Builder.WithInput("lane", lane)
	return b
}

// WithWorkDir sets the directory where the Fastfile is located.
func (b *FastlaneBuilder) WithWorkDir(dir string) *FastlaneBuilder {
	b.Builder.WithInput("work_dir", dir)
	return b
}

// WithGemfilePath sets the path to the Gemfile used for Fastlane.
func (b *FastlaneBuilder) WithGemfilePath(path string) *FastlaneBuilder {
	b.Builder.WithInput("gemfile_path", path)
	return b
}
