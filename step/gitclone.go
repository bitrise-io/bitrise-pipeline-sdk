package step

const idGitCloneStep = "git-clone"

// GitCloneBuilder builds a git-clone step with typed input methods.
// Embed promotes all *Builder methods; step-specific methods return *GitCloneBuilder
// for fluent chaining of typed inputs.
type GitCloneBuilder struct{ *Builder }

// GitClone creates a git-clone step builder at the given version.
func GitClone() *GitCloneBuilder { return &GitCloneBuilder{Builder: From(idGitCloneStep, "1")} }

// WithRepositoryURL sets the repository URL to clone.
func (b *GitCloneBuilder) WithRepositoryURL(url string) *GitCloneBuilder {
	b.Builder.WithInput("repository_url", url)
	return b
}

// WithBranch checks out the given branch.
func (b *GitCloneBuilder) WithBranch(branch string) *GitCloneBuilder {
	b.Builder.WithInput("branch", branch)
	return b
}

// WithTag checks out the given tag instead of a branch.
func (b *GitCloneBuilder) WithTag(tag string) *GitCloneBuilder {
	b.Builder.WithInput("tag", tag)
	return b
}

// WithCommit checks out a specific commit SHA.
func (b *GitCloneBuilder) WithCommit(sha string) *GitCloneBuilder {
	b.Builder.WithInput("commit", sha)
	return b
}

// WithCloneDepth sets the clone depth for shallow clones.
func (b *GitCloneBuilder) WithCloneDepth(depth int) *GitCloneBuilder {
	b.Builder.WithInput("clone_depth", itoa(depth))
	return b
}

// WithMergeBase enables or disables merging the PR source branch into the target (default true for PRs).
func (b *GitCloneBuilder) WithMergeBase(enabled bool) *GitCloneBuilder {
	b.Builder.WithInput("merge_base", btoa(enabled))
	return b
}
