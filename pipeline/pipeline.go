// Package pipeline is the primary entry point for building Bitrise configuration programmatically.
//
// Basic usage:
//
//	p := pipeline.New("ios").
//	    AddWorkflow("primary", workflow.New().
//	        AddStep(step.GitClone()).
//	        AddStep(step.XcodeTest().WithInput("scheme", "MyApp"))).
//	    AddGraphPipeline("ci", graphpipeline.New().
//	        AddWorkflow("test", graphpipeline.NewWorkflow()).
//	        AddWorkflow("deploy", graphpipeline.NewWorkflow().WithDependsOn("test")))
//
//	if err := serialize.Print(p.Build()); err != nil {
//	    log.Fatal(err)
//	}
package pipeline

import (
	bitriseModels "github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"

	"github.com/bitrise-io/bitrise-pipeline-sdk/container"
	"github.com/bitrise-io/bitrise-pipeline-sdk/graphpipeline"
	"github.com/bitrise-io/bitrise-pipeline-sdk/stage"
	"github.com/bitrise-io/bitrise-pipeline-sdk/stepbundle"
	"github.com/bitrise-io/bitrise-pipeline-sdk/trigger"
	"github.com/bitrise-io/bitrise-pipeline-sdk/validate"
	"github.com/bitrise-io/bitrise-pipeline-sdk/workflow"
)

const (
	defaultFormatVersion    = bitriseModels.FormatVersion
	defaultStepLibSource    = "https://github.com/bitrise-io/bitrise-steplib.git"
)

// Builder assembles a complete BitriseDataModel.
type Builder struct {
	projectType string
	title       string
	summary     string
	description string
	appEnvs     []envmanModels.EnvironmentItemModel
	workflows   map[string]bitriseModels.WorkflowModel
	pipelines   map[string]bitriseModels.PipelineModel
	stages      map[string]bitriseModels.StageModel
	stepBundles map[string]bitriseModels.StepBundleModel
	containers  map[string]bitriseModels.Container
	tools       bitriseModels.ToolsModel
	triggerMap  bitriseModels.TriggerMapModel
	meta        map[string]interface{}
}

// New returns a pipeline builder for the given project type (e.g. "ios", "android", "other").
func New(projectType string) *Builder {
	return &Builder{
		projectType: projectType,
		workflows:   map[string]bitriseModels.WorkflowModel{},
		pipelines:   map[string]bitriseModels.PipelineModel{},
		stages:      map[string]bitriseModels.StageModel{},
		stepBundles: map[string]bitriseModels.StepBundleModel{},
		containers:  map[string]bitriseModels.Container{},
	}
}

// WithTitle sets the top-level config title.
func (b *Builder) WithTitle(title string) *Builder {
	b.title = title
	return b
}

// WithSummary sets the top-level config summary.
func (b *Builder) WithSummary(summary string) *Builder {
	b.summary = summary
	return b
}

// WithDescription sets the top-level config description.
func (b *Builder) WithDescription(desc string) *Builder {
	b.description = desc
	return b
}

// WithAppEnv appends a single app-level environment variable.
func (b *Builder) WithAppEnv(key, value string) *Builder {
	b.appEnvs = append(b.appEnvs, envmanModels.EnvironmentItemModel{key: value})
	return b
}

// WithAppEnvs appends multiple app-level environment variables.
func (b *Builder) WithAppEnvs(envs ...envmanModels.EnvironmentItemModel) *Builder {
	b.appEnvs = append(b.appEnvs, envs...)
	return b
}

// AddWorkflow adds a workflow to the configuration.
func (b *Builder) AddWorkflow(id string, wf *workflow.Builder) *Builder {
	b.workflows[id] = wf.Build()
	return b
}

// AddGraphPipeline adds a graph (DAG-based) pipeline to the configuration.
func (b *Builder) AddGraphPipeline(id string, p *graphpipeline.Builder) *Builder {
	b.pipelines[id] = p.Build()
	return b
}

// AddStage adds a stage to the configuration (used by stage-based pipelines).
func (b *Builder) AddStage(id string, s *stage.Builder) *Builder {
	b.stages[id] = s.Build()
	return b
}

// AddStepBundle defines a reusable step bundle at the top level of the config.
// Reference it inside a workflow with workflow.AddStepBundleRef.
func (b *Builder) AddStepBundle(id string, sb *stepbundle.Builder) *Builder {
	b.stepBundles[id] = sb.Build()
	return b
}

// AddContainer adds a container definition (execution or service) to the configuration.
func (b *Builder) AddContainer(id string, c *container.Builder) *Builder {
	b.containers[id] = c.Build()
	return b
}

// AddTrigger appends a trigger map entry. Use the trigger package helpers to construct entries:
//
//	p.AddTrigger(trigger.OnPush("primary", "").WithBranch("main").Build())
func (b *Builder) AddTrigger(item trigger.Item) *Builder {
	b.triggerMap = append(b.triggerMap, item)
	return b
}

// WithTool adds a tool version requirement at the app level.
func (b *Builder) WithTool(id bitriseModels.ToolID, version string) *Builder {
	if b.tools == nil {
		b.tools = bitriseModels.ToolsModel{}
	}
	b.tools[id] = version
	return b
}

// WithMeta sets arbitrary metadata on the configuration.
func (b *Builder) WithMeta(meta map[string]interface{}) *Builder {
	b.meta = meta
	return b
}

// Validate builds the config and runs the full validation pipeline (SDK structural checks
// + upstream bitrise/v2 Normalize + Validate). Use this to catch problems before serializing.
//
//	result, err := p.Validate()
//	if err != nil { log.Fatal(err) }
//	if !result.IsValid() { log.Fatal(result.Errors) }
//	for _, w := range result.Warnings { log.Println("WARNING:", w) }
func (b *Builder) Validate() (validate.Result, error) {
	return validate.Full(b.Build())
}

// Build assembles and returns the BitriseDataModel.
func (b *Builder) Build() bitriseModels.BitriseDataModel {
	return bitriseModels.BitriseDataModel{
		FormatVersion:        defaultFormatVersion,
		DefaultStepLibSource: defaultStepLibSource,
		ProjectType:          b.projectType,
		Title:                b.title,
		Summary:              b.summary,
		Description:          b.description,
		App:                  bitriseModels.AppModel{Environments: b.appEnvs},
		Workflows:            b.workflows,
		Pipelines:            b.pipelines,
		Stages:               b.stages,
		StepBundles:          b.stepBundles,
		Containers:           b.containers,
		Tools:                b.tools,
		TriggerMap:           b.triggerMap,
		Meta:                 b.meta,
	}
}
