package step

const idScriptStep = "script"

// ScriptBuilder builds a script step with typed input methods.
type ScriptBuilder struct{ *Builder }

// Script creates a script step with the provided shell content.
func Script(content string) *ScriptBuilder {
	return &ScriptBuilder{Builder: From(idScriptStep, "1").WithInput("content", content)}
}

// WithWorkingDir sets the directory in which the script runs.
func (b *ScriptBuilder) WithWorkingDir(dir string) *ScriptBuilder {
	b.Builder.WithInput("working_dir", dir)
	return b
}

// WithRunner sets the interpreter binary (default: /bin/bash).
func (b *ScriptBuilder) WithRunner(bin string) *ScriptBuilder {
	b.Builder.WithInput("runner_bin", bin)
	return b
}
