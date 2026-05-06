package step

const idDeployStep = "deploy-to-bitrise-io"

// DeployBuilder builds a deploy-to-bitrise-io step with typed input methods.
type DeployBuilder struct{ *Builder }

// DeployToBitriseIO creates a deploy-to-bitrise-io step builder.
func DeployToBitriseIO() *DeployBuilder {
	return &DeployBuilder{Builder: From(idDeployStep, "1")}
}

// WithDeployPath sets the directory or file path to deploy.
func (b *DeployBuilder) WithDeployPath(path string) *DeployBuilder {
	b.Builder.WithInput("deploy_path", path)
	return b
}

// WithNotifyUserGroups sets which user groups receive install-link notifications.
func (b *DeployBuilder) WithNotifyUserGroups(groups string) *DeployBuilder {
	b.Builder.WithInput("notify_user_groups", groups)
	return b
}

// WithPublicInstallPage controls whether a public install page is generated.
func (b *DeployBuilder) WithPublicInstallPage(enabled bool) *DeployBuilder {
	b.Builder.WithInput("is_enable_public_page", btoa(enabled))
	return b
}

// WithPipelineIntermediateFiles registers intermediate files to be passed between pipeline stages.
func (b *DeployBuilder) WithPipelineIntermediateFiles(files string) *DeployBuilder {
	b.Builder.WithInput("pipeline_intermediate_files", files)
	return b
}
