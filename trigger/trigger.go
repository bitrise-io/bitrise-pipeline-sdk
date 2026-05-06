// Package trigger provides helpers for building Bitrise trigger map entries.
package trigger

import bitriseModels "github.com/bitrise-io/bitrise/v2/models"

// Item is an alias for the underlying trigger map item model.
type Item = bitriseModels.TriggerMapItemModel

// Map is a slice of trigger items, matching the TriggerMapModel type.
type Map = bitriseModels.TriggerMapModel

// PushBuilder constructs a code-push trigger item.
type PushBuilder struct {
	item bitriseModels.TriggerMapItemModel
}

// OnPush returns a builder for a code-push trigger targeting the given workflow or pipeline.
// Provide either workflowID or pipelineID (leave the other empty).
func OnPush(workflowID, pipelineID string) *PushBuilder {
	return &PushBuilder{item: bitriseModels.TriggerMapItemModel{
		Type:       bitriseModels.CodePushType,
		WorkflowID: workflowID,
		PipelineID: pipelineID,
	}}
}

// WithBranch restricts the trigger to the given branch pattern (glob supported).
func (b *PushBuilder) WithBranch(pattern string) *PushBuilder {
	b.item.PushBranch = pattern
	return b
}

// Enabled controls whether the trigger is active.
func (b *PushBuilder) Enabled(v bool) *PushBuilder {
	b.item.Enabled = &v
	return b
}

// Build returns the TriggerMapItemModel.
func (b *PushBuilder) Build() bitriseModels.TriggerMapItemModel { return b.item }

// PRBuilder constructs a pull-request trigger item.
type PRBuilder struct {
	item bitriseModels.TriggerMapItemModel
}

// OnPullRequest returns a builder for a pull-request trigger targeting the given workflow or pipeline.
func OnPullRequest(workflowID, pipelineID string) *PRBuilder {
	return &PRBuilder{item: bitriseModels.TriggerMapItemModel{
		Type:       bitriseModels.PullRequestType,
		WorkflowID: workflowID,
		PipelineID: pipelineID,
	}}
}

// WithSourceBranch restricts the trigger to PRs from the given branch pattern.
func (b *PRBuilder) WithSourceBranch(pattern string) *PRBuilder {
	b.item.PullRequestSourceBranch = pattern
	return b
}

// WithTargetBranch restricts the trigger to PRs targeting the given branch pattern.
func (b *PRBuilder) WithTargetBranch(pattern string) *PRBuilder {
	b.item.PullRequestTargetBranch = pattern
	return b
}

// WithDraftPREnabled controls whether draft PRs fire this trigger.
func (b *PRBuilder) WithDraftPREnabled(v bool) *PRBuilder {
	b.item.DraftPullRequestEnabled = &v
	return b
}

// Enabled controls whether the trigger is active.
func (b *PRBuilder) Enabled(v bool) *PRBuilder {
	b.item.Enabled = &v
	return b
}

// Build returns the TriggerMapItemModel.
func (b *PRBuilder) Build() bitriseModels.TriggerMapItemModel { return b.item }

// TagBuilder constructs a tag-push trigger item.
type TagBuilder struct {
	item bitriseModels.TriggerMapItemModel
}

// OnTag returns a builder for a tag-push trigger targeting the given workflow or pipeline.
func OnTag(workflowID, pipelineID string) *TagBuilder {
	return &TagBuilder{item: bitriseModels.TriggerMapItemModel{
		Type:       bitriseModels.TagPushType,
		WorkflowID: workflowID,
		PipelineID: pipelineID,
	}}
}

// WithTag restricts the trigger to tags matching the given pattern.
func (b *TagBuilder) WithTag(pattern string) *TagBuilder {
	b.item.Tag = pattern
	return b
}

// Enabled controls whether the trigger is active.
func (b *TagBuilder) Enabled(v bool) *TagBuilder {
	b.item.Enabled = &v
	return b
}

// Build returns the TriggerMapItemModel.
func (b *TagBuilder) Build() bitriseModels.TriggerMapItemModel { return b.item }
